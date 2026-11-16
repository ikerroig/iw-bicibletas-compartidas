// Package incidencias contiene las reglas de HU-05 para reportar averias.
// El reporte y el bloqueo de la bicicleta se guardan en una unica transaccion.
package incidencias

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/db"
	"si16/backend/internal/models"
	"si16/backend/internal/services/estacion"
)

var (
	ErrBicicletaNoEncontrada = errors.New("incidencias: bicicleta no encontrada")
	ErrEstacionNoEncontrada  = errors.New("incidencias: estacion no encontrada")
	ErrBicicletaEnTaller     = errors.New("incidencias: bicicleta en taller")
	ErrBicicletaOtraEstacion = errors.New("incidencias: bicicleta en otra estacion")
)

// Reporte contiene los datos que llegan del formulario de usuario.
type Reporte struct {
	BicicletaID     int64
	BicicletaCodigo string
	EstacionID      int64
	TipoFallo       string
	Descripcion     string
}

// Resultado permite informar al usuario de que el reintento ya estaba
// registrado, sin crear otra incidencia ni repetir la transicion de estado.
type Resultado struct {
	Incidencia models.Incidencia
	Repetido   bool
}

// Servicio agrupa la base de datos y el refresco del estado de estaciones.
type Servicio struct {
	pool     *pgxpool.Pool
	estacion *estacion.Servicio
}

func Nuevo(pool *pgxpool.Pool, est *estacion.Servicio) *Servicio {
	return &Servicio{pool: pool, estacion: est}
}

const columnasIncidencia = `i.id, i.bicicleta_id, b.codigo, i.estacion_id,
	e.nombre, i.usuario_reporta_id, i.operario_id, i.tipo_fallo,
	i.descripcion_reporte, i.piezas_sustituidas, i.notas_taller,
	i.fecha_reporte, i.fecha_reparacion, i.estado`

// Reportar registra una averia y bloquea la bicicleta de forma atomica.
func (s *Servicio) Reportar(ctx context.Context, usuarioID int64, r Reporte) (Resultado, error) {
	var resultado Resultado
	var estacionID int64

	err := db.EnTransaccion(ctx, s.pool, func(tx pgx.Tx) error {
		var bicicletaID int64
		var codigo, estado string
		err := tx.QueryRow(ctx, `
			SELECT id, codigo, estado
			FROM bicicletas
			WHERE ($1 > 0 AND id = $1)
			   OR ($1 = 0 AND lower(codigo) = lower($2))
			FOR UPDATE`, r.BicicletaID, r.BicicletaCodigo).Scan(&bicicletaID, &codigo, &estado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBicicletaNoEncontrada
		}
		if err != nil {
			return fmt.Errorf("buscar bicicleta: %w", err)
		}
		if estado == models.BicicletaEnTaller {
			return ErrBicicletaEnTaller
		}

		if err := tx.QueryRow(ctx, `SELECT id FROM estaciones WHERE id = $1`, r.EstacionID).Scan(&estacionID); errors.Is(err, pgx.ErrNoRows) {
			return ErrEstacionNoEncontrada
		} else if err != nil {
			return fmt.Errorf("buscar estacion: %w", err)
		}

		// Una bicicleta anclada solo puede reportarse desde la estación donde
		// está colocada. Las bicicletas en un viaje no tienen anclaje y se
		// permiten porque la HU-05 contempla ese caso.
		var estacionBicicleta int64
		err = tx.QueryRow(ctx, `
			SELECT estacion_id
			FROM anclajes
			WHERE bicicleta_id = $1`, bicicletaID).Scan(&estacionBicicleta)
		if err == nil && estado != models.BicicletaEnUso && estacionBicicleta != estacionID {
			return ErrBicicletaOtraEstacion
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscar estacion de la bicicleta: %w", err)
		}

		// Bloquear la bicicleta antes de consultar la incidencia abierta hace
		// que dos reportes simultaneos se serialicen de forma determinista.
		var existenteID int64
		err = tx.QueryRow(ctx, `
			SELECT id
			FROM incidencias_mantenimientos
			WHERE bicicleta_id = $1 AND estado <> 'resuelta'
			ORDER BY fecha_reporte DESC
			LIMIT 1
			FOR UPDATE`, bicicletaID).Scan(&existenteID)
		if err == nil {
			resultado.Repetido = true
			if estado != models.BicicletaAveriada {
				if _, err := tx.Exec(ctx, `UPDATE bicicletas SET estado = 'averiada' WHERE id = $1`, bicicletaID); err != nil {
					return fmt.Errorf("reafirmar bloqueo de bicicleta: %w", err)
				}
			}
			return cargarIncidencia(ctx, tx, existenteID, &resultado.Incidencia)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("buscar incidencia abierta: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO incidencias_mantenimientos
				(bicicleta_id, estacion_id, usuario_reporta_id, tipo_fallo, descripcion_reporte)
			VALUES ($1, $2, $3, $4, $5)`, bicicletaID, estacionID, usuarioID, r.TipoFallo, r.Descripcion); err != nil {
			return fmt.Errorf("crear incidencia: %w", err)
		}

		// La bicicleta puede estar disponible o en un viaje activo. En ambos
		// casos queda fuera del alquiler inmediatamente.
		if _, err := tx.Exec(ctx, `UPDATE bicicletas SET estado = 'averiada' WHERE id = $1`, bicicletaID); err != nil {
			return fmt.Errorf("bloquear bicicleta: %w", err)
		}

		return cargarUltimaIncidencia(ctx, tx, bicicletaID, &resultado.Incidencia)
	})
	if err != nil {
		return Resultado{}, err
	}

	// La transaccion ya esta confirmada. Un fallo de Redis no debe deshacer el
	// reporte; el refresco periodico de estaciones lo corregira.
	if s.estacion != nil {
		if err := s.estacion.RecalcularEstacion(ctx, []int64{estacionID}); err != nil {
			// La bicicleta ya esta bloqueada y la incidencia ya existe. El GET de
			// estaciones tambien dispone de un fallback directo a PostgreSQL.
			_ = err
		}
	}
	return resultado, nil
}

// Pendientes devuelve las incidencias que debe atender el equipo de taller.
func (s *Servicio) Pendientes(ctx context.Context) ([]models.Incidencia, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT `+columnasIncidencia+`
		FROM incidencias_mantenimientos i
		JOIN bicicletas b ON b.id = i.bicicleta_id
		JOIN estaciones e ON e.id = i.estacion_id
		WHERE i.estado <> 'resuelta'
		ORDER BY i.fecha_reporte ASC, i.id ASC`)
	if err != nil {
		return nil, fmt.Errorf("listar incidencias: %w", err)
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Incidencia, error) {
		var item models.Incidencia
		return item, row.Scan(&item.ID, &item.BicicletaID, &item.BicicletaCodigo,
			&item.EstacionID, &item.EstacionNombre, &item.UsuarioReportaID,
			&item.OperarioID, &item.TipoFallo, &item.DescripcionReporte,
			&item.PiezasSustituidas, &item.NotasTaller, &item.FechaReporte,
			&item.FechaReparacion, &item.Estado)
	})
	if err != nil {
		return nil, fmt.Errorf("leer incidencias: %w", err)
	}
	if items == nil {
		items = []models.Incidencia{}
	}
	return items, nil
}

func cargarUltimaIncidencia(ctx context.Context, tx pgx.Tx, bicicletaID int64, destino *models.Incidencia) error {
	return cargarIncidenciaConQuery(ctx, tx, `
		SELECT `+columnasIncidencia+`
		FROM incidencias_mantenimientos i
		JOIN bicicletas b ON b.id = i.bicicleta_id
		JOIN estaciones e ON e.id = i.estacion_id
		WHERE i.bicicleta_id = $1
		ORDER BY i.id DESC LIMIT 1`, destino, bicicletaID)
}

func cargarIncidencia(ctx context.Context, tx pgx.Tx, id int64, destino *models.Incidencia) error {
	return cargarIncidenciaConQuery(ctx, tx, `
		SELECT `+columnasIncidencia+`
		FROM incidencias_mantenimientos i
		JOIN bicicletas b ON b.id = i.bicicleta_id
		JOIN estaciones e ON e.id = i.estacion_id
		WHERE i.id = $1`, destino, id)
}

func cargarIncidenciaConQuery(ctx context.Context, tx pgx.Tx, query string, destino *models.Incidencia, arg int64) error {
	err := tx.QueryRow(ctx, query, arg).Scan(
		&destino.ID, &destino.BicicletaID, &destino.BicicletaCodigo,
		&destino.EstacionID, &destino.EstacionNombre, &destino.UsuarioReportaID,
		&destino.OperarioID, &destino.TipoFallo, &destino.DescripcionReporte,
		&destino.PiezasSustituidas, &destino.NotasTaller, &destino.FechaReporte,
		&destino.FechaReparacion, &destino.Estado)
	destino.FechaReporte = destino.FechaReporte.UTC()
	if destino.FechaReparacion != nil {
		t := destino.FechaReparacion.UTC()
		destino.FechaReparacion = &t
	}
	return err
}
