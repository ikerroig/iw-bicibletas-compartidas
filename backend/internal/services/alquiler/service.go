// Package alquiler (spec-002, Iker Roig) inicia viajes: es la única autoridad
// de las reglas de abono vigente y de un solo viaje activo (constitución I).
package alquiler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/db"
	"si16/backend/internal/models"
	"si16/backend/internal/services/estacion"
)

// Causas de rechazo. Ninguna deja efectos: la transacción se revierte.
var (
	ErrAbonoVencido          = errors.New("alquiler: abono no vigente")
	ErrViajeActivo           = errors.New("alquiler: el usuario ya tiene un viaje activo")
	ErrAnclajeNoExiste       = errors.New("alquiler: anclaje inexistente")
	ErrEstacionFueraServicio = errors.New("alquiler: estación fuera de servicio")
	ErrAnclajeFueraServicio  = errors.New("alquiler: anclaje fuera de servicio")
	ErrAnclajeVacio          = errors.New("alquiler: anclaje sin bicicleta")
	ErrBicicletaNoDisponible = errors.New("alquiler: bicicleta no disponible")
)

const pgUniqueViolation = "23505"

// recalculoTimeout acota el recálculo post-commit, que ya no depende de la
// petición: si el cliente se desconecta, el viaje existe y el mapa debe verlo.
const recalculoTimeout = 5 * time.Second

const columnasViaje = `id, usuario_id, bicicleta_id, estacion_origen_id, estacion_destino_id,
	fecha_inicio, fecha_fin, duracion_minutos, coste_total::float8, estado`

// Servicio agrupa las dependencias.
type Servicio struct {
	pool     *pgxpool.Pool
	estacion *estacion.Servicio
	log      *slog.Logger
}

// Nuevo crea el servicio.
func Nuevo(pool *pgxpool.Pool, est *estacion.Servicio, log *slog.Logger) *Servicio {
	if log == nil {
		log = slog.Default()
	}
	return &Servicio{pool: pool, estacion: est, log: log}
}

// Desbloquear valida y crea el viaje en una única transacción (research R7).
//
// Bloquear el anclaje antes de validar serializa los desbloqueos del mismo
// anclaje, y cada SELECT posterior ve lo confirmado mientras se esperaba. Dos
// desbloqueos del mismo usuario en anclajes distintos no se bloquean entre sí:
// a esos los para el índice ux_viajes_usuario_activo.
func (s *Servicio) Desbloquear(ctx context.Context, usuarioID int64, codigoAnclaje string) (models.Viaje, error) {
	var viaje models.Viaje
	var estacionID int64
	err := db.EnTransaccion(ctx, s.pool, func(tx pgx.Tx) error {
		var anclajeID int64
		var estadoAnclaje string
		var bicicletaID *int64
		errAnclaje := tx.QueryRow(ctx, `
			SELECT id, estacion_id, estado, bicicleta_id
			FROM anclajes WHERE codigo_anclaje = $1
			FOR UPDATE`, codigoAnclaje).Scan(&anclajeID, &estacionID, &estadoAnclaje, &bicicletaID)
		if errAnclaje != nil && !errors.Is(errAnclaje, pgx.ErrNoRows) {
			return fmt.Errorf("alquiler: bloquear anclaje: %w", errAnclaje)
		}

		// Las reglas del usuario van antes que las del anclaje: quien no puede
		// alquilar recibe esa causa aunque también haya tecleado mal el código.
		if err := validarUsuario(ctx, tx, usuarioID); err != nil {
			return err
		}
		if errAnclaje != nil {
			return ErrAnclajeNoExiste
		}

		var estadoEstacion string
		if err := tx.QueryRow(ctx, `SELECT estado FROM estaciones WHERE id = $1`, estacionID).Scan(&estadoEstacion); err != nil {
			return fmt.Errorf("alquiler: leer estación: %w", err)
		}
		switch {
		case estadoEstacion != models.EstacionOperativa:
			return ErrEstacionFueraServicio
		case estadoAnclaje == models.AnclajeFueraDeServicio:
			return ErrAnclajeFueraServicio
		case bicicletaID == nil:
			return ErrAnclajeVacio
		}

		// El bloqueo evita desbloquear una bici que HU-05 está marcando averiada.
		var estadoBici string
		if err := tx.QueryRow(ctx, `SELECT estado FROM bicicletas WHERE id = $1 FOR UPDATE`, *bicicletaID).Scan(&estadoBici); err != nil {
			return fmt.Errorf("alquiler: leer bicicleta: %w", err)
		}
		if estadoBici != models.BicicletaDisponible {
			return ErrBicicletaNoDisponible
		}

		if _, err := tx.Exec(ctx, `UPDATE anclajes SET estado = $2, bicicleta_id = NULL WHERE id = $1`,
			anclajeID, models.AnclajeLibre); err != nil {
			return fmt.Errorf("alquiler: desacoplar: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE bicicletas SET estado = $2 WHERE id = $1`,
			*bicicletaID, models.BicicletaEnUso); err != nil {
			return fmt.Errorf("alquiler: bicicleta en uso: %w", err)
		}
		v, err := escanearViaje(tx.QueryRow(ctx, `
			INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id)
			VALUES ($1, $2, $3) RETURNING `+columnasViaje, usuarioID, *bicicletaID, estacionID))
		if err != nil {
			return traducirInsercion(err)
		}
		viaje = v
		return nil
	})
	if err != nil {
		return models.Viaje{}, err
	}

	// El commit ya está hecho: un fallo aquí no revierte nada y lo repara el barrido.
	rc, cancelar := context.WithTimeout(context.WithoutCancel(ctx), recalculoTimeout)
	defer cancelar()
	if err := s.estacion.RecalcularEstacion(rc, []int64{estacionID}); err != nil {
		s.log.Warn("alquiler: recálculo post-commit fallido", "estacion", estacionID, "error", err)
	}
	return viaje, nil
}

// Activo devuelve el viaje en curso del usuario, o nil si no tiene.
func (s *Servicio) Activo(ctx context.Context, usuarioID int64) (*models.Viaje, error) {
	v, err := escanearViaje(s.pool.QueryRow(ctx, `SELECT `+columnasViaje+`
		FROM viajes WHERE usuario_id = $1 AND estado = $2`, usuarioID, models.ViajeEnCurso))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("alquiler: viaje activo: %w", err)
	}
	return &v, nil
}

func validarUsuario(ctx context.Context, tx pgx.Tx, usuarioID int64) error {
	var u models.Usuario
	err := tx.QueryRow(ctx, `SELECT abono_activo, fecha_vencimiento_abono FROM usuarios WHERE id = $1`, usuarioID).
		Scan(&u.AbonoActivo, &u.FechaVencimientoAbono)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAbonoVencido // JWT de un usuario ya borrado: no hay abono que valga
	}
	if err != nil {
		return fmt.Errorf("alquiler: leer usuario: %w", err)
	}
	if !models.AbonoVigente(u, time.Now().UTC()) {
		return ErrAbonoVencido
	}

	var activo bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM viajes WHERE usuario_id = $1 AND estado = $2)`,
		usuarioID, models.ViajeEnCurso).Scan(&activo); err != nil {
		return fmt.Errorf("alquiler: comprobar viaje activo: %w", err)
	}
	if activo {
		return ErrViajeActivo
	}
	return nil
}

func traducirInsercion(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "ux_viajes_usuario_activo":
			return ErrViajeActivo
		case "ux_viajes_bicicleta_activa":
			return ErrBicicletaNoDisponible
		}
	}
	return fmt.Errorf("alquiler: crear viaje: %w", err)
}

func escanearViaje(row pgx.Row) (models.Viaje, error) {
	var v models.Viaje
	err := row.Scan(&v.ID, &v.UsuarioID, &v.BicicletaID, &v.EstacionOrigenID, &v.EstacionDestinoID,
		&v.FechaInicio, &v.FechaFin, &v.DuracionMinutos, &v.CosteTotal, &v.Estado)
	v.FechaInicio = v.FechaInicio.UTC()
	return v, err
}
