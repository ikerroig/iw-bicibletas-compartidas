// Package devolucion implementa el servicio de cierre y tarificación de la
// HU-03 (spec-003). Orden de transacción: research D2 — anclaje FOR UPDATE,
// viaje en_curso FOR UPDATE, reloj validado, tarifa, UPDATE condicionado,
// COMMIT y solo después la invalidación en Redis (best-effort, D4).
package devolucion

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/apierr"
	"si16/backend/internal/models"
	"si16/backend/internal/services/estacion"
)

// Service agrupa las dependencias del cierre de viaje.
type Service struct {
	pool         *pgxpool.Pool
	estacion     *estacion.Servicio
	tarifaMinuto float64
}

// New crea el servicio. est es el recálculo común del estado en vivo (T020).
func New(pool *pgxpool.Pool, est *estacion.Servicio, tarifaMinuto float64) *Service {
	return &Service{pool: pool, estacion: est, tarifaMinuto: tarifaMinuto}
}

func errInterno() *apierr.Error {
	return apierr.New(500, "ERROR_INTERNO", "Error interno al cerrar el viaje")
}

// CerrarViaje ejecuta el protocolo completo de la devolución (FR-001..FR-007):
// valida el anclaje destino y el viaje activo, calcula la tarifa con el reloj
// validado, finaliza el viaje, ancla la bicicleta (preservando su avería) y
// refresca el estado de la estación en Redis tras el COMMIT.
func (s *Service) CerrarViaje(ctx context.Context, usuarioID int64, codigoAnclaje string) (models.Viaje, *apierr.Error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Printf("devolucion: begin: %v", err)
		return models.Viaje{}, errInterno()
	}
	// No-op tras COMMIT; en fallo revierte la transacción.
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Anclaje destino — FOR UPDATE (bloquea la fila contra carreras).
	var (
		anclajeID       int64
		estacionDestino int64
		estadoAnclaje   string
		biciAnclaje     *int64
	)
	err = tx.QueryRow(ctx,
		`SELECT id, estacion_id, estado, bicicleta_id
		   FROM anclajes WHERE codigo_anclaje = $1 FOR UPDATE`,
		codigoAnclaje,
	).Scan(&anclajeID, &estacionDestino, &estadoAnclaje, &biciAnclaje)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Viaje{}, apierr.New(409, "ANCLAJE_NO_ENCONTRADO",
			"No existe ningún anclaje con ese código")
	}
	if err != nil {
		log.Printf("devolucion: anclaje %s: %v", codigoAnclaje, err)
		return models.Viaje{}, errInterno()
	}
	switch estadoAnclaje {
	case "fuera_de_servicio":
		return models.Viaje{}, apierr.New(409, "ANCLAJE_FUERA_DE_SERVICIO",
			"El anclaje está retirado del servicio")
	case "ocupado":
		return models.Viaje{}, apierr.New(409, "ANCLAJE_OCUPADO",
			"El anclaje está ocupado, elige otro")
	}

	// 2. Viaje activo del usuario — FOR UPDATE (serializa confirmaciones).
	var viaje models.Viaje
	err = tx.QueryRow(ctx,
		`SELECT id, usuario_id, bicicleta_id, estacion_origen_id, estacion_destino_id,
		        fecha_inicio, fecha_fin, duracion_minutos, coste_total::float8, estado
		   FROM viajes WHERE usuario_id = $1 AND estado = 'en_curso' FOR UPDATE`,
		usuarioID,
	).Scan(&viaje.ID, &viaje.UsuarioID, &viaje.BicicletaID, &viaje.EstacionOrigenID,
		&viaje.EstacionDestinoID, &viaje.FechaInicio, &viaje.FechaFin,
		&viaje.DuracionMinutos, &viaje.CosteTotal, &viaje.Estado)
	if errors.Is(err, pgx.ErrNoRows) {
		// ¿Segunda confirmación de un viaje recién cerrado? → VIAJE_YA_CERRADO.
		var reciente bool
		if e2 := tx.QueryRow(ctx,
			`SELECT EXISTS(
			     SELECT 1 FROM viajes
			      WHERE usuario_id = $1 AND estado = 'finalizado'
			        AND fecha_fin > now() - $2::interval)`,
			usuarioID, "5 seconds",
		).Scan(&reciente); e2 == nil && reciente {
			return models.Viaje{}, apierr.New(409, "VIAJE_YA_CERRADO",
				"Este viaje ya fue cerrado")
		}
		return models.Viaje{}, apierr.New(409, "SIN_VIAJE_ACTIVO",
			"No tienes ningún viaje en curso")
	}
	if err != nil {
		log.Printf("devolucion: viaje usuario %d: %v", usuarioID, err)
		return models.Viaje{}, errInterno()
	}

	// 3. Reloj validado + cálculo tarifario (FR-003/FR-004, D1).
	ahora := time.Now().UTC()
	segundos := int64(ahora.Sub(viaje.FechaInicio).Seconds())
	if segundos < 0 {
		segundos = 0 // reloj inconsistente: nunca negativo (T3i)
	}
	// duracion_minutos = ceil(segundos/60) (data-model R8): con los mismos
	// segundos que la tarifa, un viaje cobrado nunca figura con ≤ 30 min (SC-001).
	duracion := DuracionMinutos(segundos)
	coste := CalcularCoste(segundos, s.tarifaMinuto)

	// 4. UPDATE condicionado del viaje (rows=0 → perdió la carrera, V3).
	res, err := tx.Exec(ctx,
		`UPDATE viajes
		    SET estado = 'finalizado', fecha_fin = $1, duracion_minutos = $2,
		        coste_total = $3::numeric, estacion_destino_id = $4
		  WHERE id = $5 AND estado = 'en_curso'`,
		ahora, duracion, coste, estacionDestino, viaje.ID)
	if err != nil {
		log.Printf("devolucion: update viaje %d: %v", viaje.ID, err)
		return models.Viaje{}, errInterno()
	}
	if res.RowsAffected() == 0 {
		return models.Viaje{}, apierr.New(409, "VIAJE_YA_CERRADO",
			"Este viaje ya fue cerrado")
	}

	// 5. Anclaje → ocupado con la bicicleta del viaje (ya validado libre).
	if res, err = tx.Exec(ctx,
		`UPDATE anclajes SET estado = 'ocupado', bicicleta_id = $1
		  WHERE id = $2 AND estado = 'libre'`,
		viaje.BicicletaID, anclajeID); err != nil {
		log.Printf("devolucion: update anclaje %d: %v", anclajeID, err)
		return models.Viaje{}, errInterno()
	} else if res.RowsAffected() == 0 {
		return models.Viaje{}, apierr.New(409, "ANCLAJE_OCUPADO",
			"El anclaje está ocupado, elige otro")
	}

	// 6. Bicicleta → disponible, salvo que esté AVERIADA (la preserva, HU-05).
	if _, err = tx.Exec(ctx,
		`UPDATE bicicletas SET estado = 'disponible'
		  WHERE id = $1 AND estado <> 'averiada'`,
		viaje.BicicletaID); err != nil {
		log.Printf("devolucion: update bici %d: %v", viaje.BicicletaID, err)
		return models.Viaje{}, errInterno()
	}

	// 7. COMMIT (punto único de verdad).
	if err = tx.Commit(ctx); err != nil {
		log.Printf("devolucion: commit: %v", err)
		return models.Viaje{}, errInterno()
	}

	// 8. Redis post-commit best-effort (D4): si falla, el viaje ya está
	// cerrado; el TTL de 120 s autolimita la desviación.
	if s.estacion != nil {
		if err := s.estacion.RecalcularEstacion(ctx, []int64{estacionDestino}); err != nil {
			log.Printf("devolucion: cache estación %d (viaje cerrado igualmente): %v",
				estacionDestino, err)
		}
	}

	// Respuesta (D5): estado final + etiqueta de tramo gratuito si coste 0.
	fechaFin := ahora
	viaje.EstacionDestinoID = &estacionDestino
	viaje.FechaFin = &fechaFin
	viaje.DuracionMinutos = &duracion
	viaje.CosteTotal = coste
	viaje.Estado = "finalizado"
	if coste == 0 {
		et := "Tramo gratuito (0,00 €)"
		viaje.Etiqueta = &et
	}
	return viaje, nil
}
