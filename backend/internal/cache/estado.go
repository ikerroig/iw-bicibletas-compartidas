package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ActualizarEstadoEstacion recalcula (fórmula R9 — nunca incrementos ciegos)
// y reescribe el HASH de estado de la estación. Llamar post-commit.
func ActualizarEstadoEstacion(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, estacionID int64) error {
	var (
		capacidad      int
		estadoEstacion string
	)
	err := pool.QueryRow(ctx,
		`SELECT capacidad_total, estado FROM estaciones WHERE id = $1`, estacionID,
	).Scan(&capacidad, &estadoEstacion)
	if err != nil {
		return fmt.Errorf("estación %d: %w", estacionID, err)
	}

	var anclajesLibres int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM anclajes WHERE estacion_id = $1 AND estado = 'libre'`, estacionID,
	).Scan(&anclajesLibres); err != nil {
		return fmt.Errorf("anclajes libres: %w", err)
	}

	var bicisDisponibles int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*)
		   FROM anclajes a
		   JOIN bicicletas b ON b.id = a.bicicleta_id
		  WHERE a.estacion_id = $1 AND b.estado = 'disponible'`, estacionID,
	).Scan(&bicisDisponibles); err != nil {
		return fmt.Errorf("bicis disponibles: %w", err)
	}

	ocupacion := 0.0
	if estadoEstacion == "operativa" && capacidad > 0 {
		ocupacion = 100.0 * float64(bicisDisponibles) / float64(capacidad)
	}

	// Escritura del HASH (HSet es atómico) + TTL de seguridad.
 clave := ClaveEstado(estacionID)
	if err := rdb.HSet(ctx, clave, map[string]interface{}{
		"bicis_disponibles": strconv.Itoa(bicisDisponibles),
		"anclajes_libres":   strconv.Itoa(anclajesLibres),
		"ocupacion_pct":     strconv.FormatFloat(ocupacion, 'f', 2, 64),
		"estado":            estadoEstacion,
		"actualizado_en":    time.Now().UTC().Format(time.RFC3339),
	}).Err(); err != nil {
		return fmt.Errorf("HSet %s: %w", clave, err)
	}
	if err := rdb.Expire(ctx, clave, TTLEstado).Err(); err != nil {
		return fmt.Errorf("Expire %s: %w", clave, err)
	}
	return nil
}
