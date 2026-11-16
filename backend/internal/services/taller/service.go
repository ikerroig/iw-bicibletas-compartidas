// Package taller contiene las reglas de HU-06 para reparar y ubicar bicicletas.
package taller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"si16/backend/internal/cache"
	"si16/backend/internal/models"
)

var (
	ErrBicicletaNoEncontrada = errors.New("taller: bicicleta no encontrada")
	ErrBicicletaEnUso        = errors.New("taller: bicicleta en uso")
	ErrAnclajeNoEncontrado   = errors.New("taller: anclaje no encontrado")
	ErrAnclajeOcupado        = errors.New("taller: anclaje ocupado")
	ErrAnclajeFueraServicio  = errors.New("taller: anclaje fuera de servicio")
	ErrEstacionFueraServicio = errors.New("taller: estacion fuera de servicio")
	ErrBicicletaNoDisponible = errors.New("taller: bicicleta no disponible")
	ErrBicicletaYaUbicada    = errors.New("taller: bicicleta ya ubicada")
)

const columnasIncidencia = `i.id, i.bicicleta_id, b.codigo,
	COALESCE(i.estacion_id, 0), COALESCE(e.nombre, ''),
	i.usuario_reporta_id, i.operario_id, i.tipo_fallo, i.descripcion_reporte,
	i.piezas_sustituidas, i.notas_taller, i.fecha_reporte, i.fecha_reparacion, i.estado`

type Servicio struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
	log  *slog.Logger
}

type BicicletaFicha struct {
	ID     int64  `json:"id"`
	Codigo string `json:"codigo"`
	Estado string `json:"estado"`
}

type Ficha struct {
	Bicicleta   BicicletaFicha      `json:"bicicleta"`
	Incidencias []models.Incidencia `json:"incidencias"`
}

type Intervencion struct {
	PiezasSustituidas *string
	NotasTaller       *string
	MarcarOperativa   bool
}

func Nuevo(pool *pgxpool.Pool, rdb *redis.Client, log *slog.Logger) *Servicio {
	if log == nil {
		log = slog.Default()
	}
	return &Servicio{pool: pool, rdb: rdb, log: log}
}

func (s *Servicio) Ficha(ctx context.Context, bicicletaID int64) (Ficha, error) {
	var ficha Ficha
	err := s.pool.QueryRow(ctx, `
		SELECT id, codigo, estado
		FROM bicicletas
		WHERE id = $1`, bicicletaID).Scan(
		&ficha.Bicicleta.ID, &ficha.Bicicleta.Codigo, &ficha.Bicicleta.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return Ficha{}, ErrBicicletaNoEncontrada
	}
	if err != nil {
		return Ficha{}, fmt.Errorf("buscar bicicleta: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
		SELECT `+columnasIncidencia+`
		FROM incidencias_mantenimientos i
		JOIN bicicletas b ON b.id = i.bicicleta_id
		LEFT JOIN estaciones e ON e.id = i.estacion_id
		WHERE i.bicicleta_id = $1
		ORDER BY i.fecha_reporte DESC, i.id DESC`, bicicletaID)
	if err != nil {
		return Ficha{}, fmt.Errorf("buscar historial: %w", err)
	}
	ficha.Incidencias, err = pgx.CollectRows(rows, scanIncidencia)
	if err != nil {
		return Ficha{}, fmt.Errorf("leer historial: %w", err)
	}
	if ficha.Incidencias == nil {
		ficha.Incidencias = []models.Incidencia{}
	}
	return ficha, nil
}

func (s *Servicio) RegistrarIntervencion(ctx context.Context, operarioID, bicicletaID int64, req Intervencion) (models.Incidencia, error) {
	var incidencia models.Incidencia
	var estacionesAfectadas []int64

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Incidencia{}, fmt.Errorf("iniciar reparacion: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op despues de COMMIT

	var estadoBicicleta string
	if err := tx.QueryRow(ctx, `
		SELECT estado
		FROM bicicletas
		WHERE id = $1
		FOR UPDATE`, bicicletaID).Scan(&estadoBicicleta); errors.Is(err, pgx.ErrNoRows) {
		return models.Incidencia{}, ErrBicicletaNoEncontrada
	} else if err != nil {
		return models.Incidencia{}, fmt.Errorf("bloquear bicicleta: %w", err)
	}
	if estadoBicicleta == models.BicicletaEnUso {
		return models.Incidencia{}, ErrBicicletaEnUso
	}

	var abiertas int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM incidencias_mantenimientos
		WHERE bicicleta_id = $1 AND estado <> 'resuelta'`, bicicletaID).Scan(&abiertas); err != nil {
		return models.Incidencia{}, fmt.Errorf("buscar incidencias abiertas: %w", err)
	}

	estadoIncidencia := models.IncidenciaEnReparacion
	if req.MarcarOperativa {
		estadoIncidencia = models.IncidenciaResuelta
	}

	if abiertas == 0 {
		var incidenciaID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO incidencias_mantenimientos
				(bicicleta_id, operario_id, tipo_fallo, descripcion_reporte,
				 piezas_sustituidas, notas_taller, fecha_reparacion, estado)
			VALUES ($1, $2, 'otro', 'Mantenimiento preventivo', $3, $4,
			        CASE WHEN $5 THEN now() ELSE NULL END, $6)
			RETURNING id`,
			bicicletaID, operarioID, req.PiezasSustituidas, req.NotasTaller,
			req.MarcarOperativa, estadoIncidencia,
		).Scan(&incidenciaID)
		if err != nil {
			return models.Incidencia{}, fmt.Errorf("guardar mantenimiento preventivo: %w", err)
		}
		if err := cargarIncidencia(ctx, tx, incidenciaID, &incidencia); err != nil {
			return models.Incidencia{}, fmt.Errorf("leer mantenimiento guardado: %w", err)
		}
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE incidencias_mantenimientos
			SET operario_id = $2,
			    piezas_sustituidas = $3,
			    notas_taller = $4,
			    fecha_reparacion = CASE WHEN $5 THEN now() ELSE fecha_reparacion END,
			    estado = $6
			WHERE bicicleta_id = $1 AND estado <> 'resuelta'`,
			bicicletaID, operarioID, req.PiezasSustituidas, req.NotasTaller,
			req.MarcarOperativa, estadoIncidencia)
		if err != nil {
			return models.Incidencia{}, fmt.Errorf("guardar parte de reparacion: %w", err)
		}
		if err := cargarUltimaIncidencia(ctx, tx, bicicletaID, &incidencia); err != nil {
			return models.Incidencia{}, fmt.Errorf("leer parte guardado: %w", err)
		}
	}

	if req.MarcarOperativa {
		if _, err := tx.Exec(ctx,
			`UPDATE bicicletas SET estado = 'disponible' WHERE id = $1`, bicicletaID); err != nil {
			return models.Incidencia{}, fmt.Errorf("reactivar bicicleta: %w", err)
		}
		estacionesAfectadas, err = liberarAnclaje(ctx, tx, bicicletaID)
		if err != nil {
			return models.Incidencia{}, err
		}
	} else if abiertas > 0 && estadoBicicleta == models.BicicletaAveriada {
		if _, err := tx.Exec(ctx,
			`UPDATE bicicletas SET estado = 'en_taller' WHERE id = $1`, bicicletaID); err != nil {
			return models.Incidencia{}, fmt.Errorf("pasar bicicleta al taller: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Incidencia{}, fmt.Errorf("confirmar reparacion: %w", err)
	}
	s.actualizarEstaciones(ctx, estacionesAfectadas)
	return incidencia, nil
}

func (s *Servicio) ConfirmarUbicacion(ctx context.Context, bicicletaID int64, codigoAnclaje string) (models.EstacionEstado, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.EstacionEstado{}, fmt.Errorf("iniciar ubicacion: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var estadoBicicleta string
	if err := tx.QueryRow(ctx, `
		SELECT estado FROM bicicletas WHERE id = $1 FOR UPDATE`, bicicletaID).Scan(&estadoBicicleta); errors.Is(err, pgx.ErrNoRows) {
		return models.EstacionEstado{}, ErrBicicletaNoEncontrada
	} else if err != nil {
		return models.EstacionEstado{}, fmt.Errorf("bloquear bicicleta: %w", err)
	}
	if estadoBicicleta != models.BicicletaDisponible {
		return models.EstacionEstado{}, ErrBicicletaNoDisponible
	}

	var estacionActual int64
	err = tx.QueryRow(ctx, `
		SELECT estacion_id
		FROM anclajes
		WHERE bicicleta_id = $1
		FOR UPDATE`, bicicletaID).Scan(&estacionActual)
	if err == nil {
		return models.EstacionEstado{}, ErrBicicletaYaUbicada
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.EstacionEstado{}, fmt.Errorf("buscar ubicacion actual: %w", err)
	}

	var anclajeID, estacionID int64
	var estadoAnclaje string
	err = tx.QueryRow(ctx, `
		SELECT a.id, a.estacion_id, a.estado
		FROM anclajes a
		WHERE a.codigo_anclaje = $1
		FOR UPDATE`, strings.TrimSpace(codigoAnclaje)).Scan(
		&anclajeID, &estacionID, &estadoAnclaje)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.EstacionEstado{}, ErrAnclajeNoEncontrado
	}
	if err != nil {
		return models.EstacionEstado{}, fmt.Errorf("buscar anclaje: %w", err)
	}
	if estadoAnclaje == models.AnclajeFueraDeServicio {
		return models.EstacionEstado{}, ErrAnclajeFueraServicio
	}
	if estadoAnclaje != models.AnclajeLibre {
		return models.EstacionEstado{}, ErrAnclajeOcupado
	}

	var estadoEstacion string
	if err := tx.QueryRow(ctx,
		`SELECT estado FROM estaciones WHERE id = $1 FOR SHARE`, estacionID).Scan(&estadoEstacion); err != nil {
		return models.EstacionEstado{}, fmt.Errorf("buscar estacion: %w", err)
	}
	if estadoEstacion != models.EstacionOperativa {
		return models.EstacionEstado{}, ErrEstacionFueraServicio
	}

	if _, err := tx.Exec(ctx, `
		UPDATE anclajes SET estado = 'ocupado', bicicleta_id = $1
		WHERE id = $2`, bicicletaID, anclajeID); err != nil {
		return models.EstacionEstado{}, fmt.Errorf("ocupar anclaje: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return models.EstacionEstado{}, fmt.Errorf("confirmar ubicacion: %w", err)
	}
	s.actualizarEstaciones(ctx, []int64{estacionID})

	return s.estadoEstacion(ctx, estacionID)
}

func liberarAnclaje(ctx context.Context, tx pgx.Tx, bicicletaID int64) ([]int64, error) {
	rows, err := tx.Query(ctx, `
		UPDATE anclajes
		SET estado = 'libre', bicicleta_id = NULL
		WHERE bicicleta_id = $1
		RETURNING estacion_id`, bicicletaID)
	if err != nil {
		return nil, fmt.Errorf("liberar anclaje: %w", err)
	}
	ids, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (int64, error) {
		var id int64
		return id, row.Scan(&id)
	})
	if err != nil {
		return nil, fmt.Errorf("leer estaciones afectadas: %w", err)
	}
	return ids, nil
}

func (s *Servicio) actualizarEstaciones(ctx context.Context, ids []int64) {
	if s.rdb == nil {
		return
	}
	for _, id := range ids {
		if err := cache.ActualizarEstadoEstacion(ctx, s.pool, s.rdb, id); err != nil {
			s.log.Warn("taller: Redis no actualizado tras cambio confirmado", "estacion_id", id, "error", err)
		}
	}
}

func (s *Servicio) estadoEstacion(ctx context.Context, id int64) (models.EstacionEstado, error) {
	var estado models.EstacionEstado
	err := s.pool.QueryRow(ctx, `
		SELECT e.id, e.nombre, e.latitud::float8, e.longitud::float8,
		       e.capacidad_total, e.estado,
		       (SELECT COUNT(*) FROM anclajes a
		         JOIN bicicletas b ON b.id = a.bicicleta_id
		        WHERE a.estacion_id = e.id AND b.estado = 'disponible'),
		       (SELECT COUNT(*) FROM anclajes
		        WHERE estacion_id = e.id AND estado = 'libre')
		FROM estaciones e
		WHERE e.id = $1`, id).Scan(
		&estado.ID, &estado.Nombre, &estado.Latitud, &estado.Longitud,
		&estado.CapacidadTotal, &estado.Estado,
		&estado.BicisDisponibles, &estado.AnclajesLibres)
	if err != nil {
		return models.EstacionEstado{}, fmt.Errorf("leer estado de estacion: %w", err)
	}
	estado.ActualizadoEn = time.Now().UTC().Format(time.RFC3339)
	if estado.CapacidadTotal > 0 && estado.Estado == models.EstacionOperativa {
		estado.OcupacionPct = float64(estado.BicisDisponibles) * 100 / float64(estado.CapacidadTotal)
	}
	return estado, nil
}

func scanIncidencia(row pgx.CollectableRow) (models.Incidencia, error) {
	var incidencia models.Incidencia
	return incidencia, row.Scan(incidenciaDestinos(&incidencia)...)
}

func incidenciaDestinos(i *models.Incidencia) []any {
	return []any{
		&i.ID, &i.BicicletaID, &i.BicicletaCodigo, &i.EstacionID, &i.EstacionNombre,
		&i.UsuarioReportaID, &i.OperarioID, &i.TipoFallo, &i.DescripcionReporte,
		&i.PiezasSustituidas,
		&i.NotasTaller, &i.FechaReporte, &i.FechaReparacion, &i.Estado,
	}
}

func cargarIncidencia(ctx context.Context, tx pgx.Tx, id int64, destino *models.Incidencia) error {
	return tx.QueryRow(ctx, `
		SELECT `+columnasIncidencia+`
		FROM incidencias_mantenimientos i
		JOIN bicicletas b ON b.id = i.bicicleta_id
		LEFT JOIN estaciones e ON e.id = i.estacion_id
		WHERE i.id = $1`, id).Scan(incidenciaDestinos(destino)...)
}

func cargarUltimaIncidencia(ctx context.Context, tx pgx.Tx, bicicletaID int64, destino *models.Incidencia) error {
	return tx.QueryRow(ctx, `
		SELECT `+columnasIncidencia+`
		FROM incidencias_mantenimientos i
		JOIN bicicletas b ON b.id = i.bicicleta_id
		LEFT JOIN estaciones e ON e.id = i.estacion_id
		WHERE i.bicicleta_id = $1
		ORDER BY i.id DESC
		LIMIT 1`, bicicletaID).Scan(incidenciaDestinos(destino)...)
}
