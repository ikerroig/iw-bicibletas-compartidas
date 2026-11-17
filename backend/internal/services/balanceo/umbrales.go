// Package balanceo implementa la HU-07: cómputo de ocupación (R9), creación y
// resolución de alertas de balanceo (FR-001…FR-004) y traslados de lote
// (FR-006/FR-007). Decisiones: research.md D1–D4; SQL: data-model.md §1–§4.
package balanceo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/services/estacion"
)

// Tipos de alerta (schema global AlertaBalanceo.tipo_alerta).
const (
	TipoVacia = "vacia_menos_10"
	TipoLlena = "llena_mas_90"
)

// Servicio agrupa las dependencias del balanceo. New deja estacion en nil (solo
// umbrales, suficiente para el worker y los tests unitarios); NewConEstacion
// añade Redis para el flujo de traslado (post-commit FR-007).
type Servicio struct {
	pool     *pgxpool.Pool
	estacion *estacion.Servicio
}

func New(pool *pgxpool.Pool) *Servicio {
	return &Servicio{pool: pool}
}

func NewConEstacion(pool *pgxpool.Pool, est *estacion.Servicio) *Servicio {
	return &Servicio{pool: pool, estacion: est}
}

// Decision es la acción de una evaluación. Función pura decidir(pct, estado)
// → acción (resolución de L4: testeable sin BD ni mocks de pgx).
type Decision struct {
	CrearTipo         string // tipo de alerta a crear ("" = ninguna)
	ResolverContrario string // umbral opuesto abierto que debe cerrarse ("" = ninguno)
	ResolverTodas     bool   // vuelve al rango [10, 90] → cerrar todas las abiertas
}

// decidir aplica los umbrales estrictos de FR-002/V2: 10,00 % y 90,00 %
// exactos NO alertan; las estaciones fuera_de_servicio se excluyen (D2).
func decidir(ocupacionPct float64, estadoEstacion string) Decision {
	if estadoEstacion != "operativa" {
		return Decision{}
	}
	switch {
	case ocupacionPct < 10:
		return Decision{CrearTipo: TipoVacia, ResolverContrario: TipoLlena}
	case ocupacionPct > 90:
		return Decision{CrearTipo: TipoLlena, ResolverContrario: TipoVacia}
	default:
		return Decision{ResolverTodas: true}
	}
}

// biciCuenta refleja FR-008: solo cuenta la bici disponible anclada (averiadas
// y en_taller excluidas; disponibles sin anclaje no pertenecen a estación).
func biciCuenta(estadoBici string, anclada bool) bool {
	return estadoBici == "disponible" && anclada
}

// ocupacionPct aplica R9: 100 × bicis disponibles / capacidad; la estación
// fuera_de_servicio queda excluida (sin capacidad computable).
func ocupacionPct(bicisDisponibles, capacidad int, estadoEstacion string) float64 {
	if estadoEstacion != "operativa" || capacidad <= 0 {
		return 0
	}
	return 100.0 * float64(bicisDisponibles) / float64(capacidad)
}

// scoreUrgencia (D3): extremos más alejados del rango [10, 90] primero
// (0 % antes que 9 %; 100 % antes que 91 %). Mismo CASE que la consulta SQL
// del panel: una única definición entre Go y SQL.
func scoreUrgencia(tipoAlerta string, porcentajeOcupacion float64) float64 {
	if tipoAlerta == TipoVacia {
		return 10 - porcentajeOcupacion
	}
	return porcentajeOcupacion - 90
}

// EvaluarEstacion calcula la ocupación de la estación (R9) y aplica
// creación/resolución (FR-002…FR-004) + refresco del snapshot de las alertas
// abiertas (D3). Sin acción para estaciones no operativas.
func (s *Servicio) EvaluarEstacion(ctx context.Context, estacionID int64) error {
	var (
		capacidad int
		estado    string
		bicis     int
	)
	err := s.pool.QueryRow(ctx,
		`SELECT e.capacidad_total, e.estado,
		        (SELECT COUNT(*)
		           FROM anclajes a
		           JOIN bicicletas b ON b.id = a.bicicleta_id
		          WHERE a.estacion_id = e.id AND b.estado = 'disponible') AS bicis_disponibles
		   FROM estaciones e
		  WHERE e.id = $1`, estacionID,
	).Scan(&capacidad, &estado, &bicis)
	if err != nil {
		return fmt.Errorf("estación %d: %w", estacionID, err)
	}
	if estado != "operativa" {
		return nil // excluida del cálculo (D2): sin creación ni resolución
	}

	pct := ocupacionPct(bicis, capacidad, estado)
	d := decidir(pct, estado)

	switch {
	case d.ResolverTodas:
		if _, err := s.pool.Exec(ctx,
			`UPDATE alertas_balanceo SET resuelta = TRUE
			  WHERE estacion_id = $1 AND NOT resuelta`, estacionID); err != nil {
			return fmt.Errorf("resolver alertas estación %d: %w", estacionID, err)
		}
	case d.ResolverContrario != "":
		if _, err := s.pool.Exec(ctx,
			`UPDATE alertas_balanceo SET resuelta = TRUE
			  WHERE estacion_id = $1 AND tipo_alerta = $2 AND NOT resuelta`,
			estacionID, d.ResolverContrario); err != nil {
			return fmt.Errorf("resolver %s estación %d: %w", d.ResolverContrario, estacionID, err)
		}
	}

	if d.CrearTipo != "" {
		if _, err := s.pool.Exec(ctx,
			`INSERT INTO alertas_balanceo (estacion_id, tipo_alerta, porcentaje_ocupacion)
			 VALUES ($1, $2, $3::numeric)
			 ON CONFLICT (estacion_id, tipo_alerta) WHERE NOT resuelta DO NOTHING`,
			estacionID, d.CrearTipo, pct); err != nil {
			return fmt.Errorf("crear %s estación %d: %w", d.CrearTipo, estacionID, err)
		}
	}

	// Refresco del snapshot de urgencia (D3): DML sobre columnas existentes.
	if _, err := s.pool.Exec(ctx,
		`UPDATE alertas_balanceo SET porcentaje_ocupacion = $2::numeric
		  WHERE estacion_id = $1 AND NOT resuelta`, estacionID, pct); err != nil {
		return fmt.Errorf("refrescar snapshot estación %d: %w", estacionID, err)
	}
	return nil
}
