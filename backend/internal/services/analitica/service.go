// Package analitica implementa el servicio de la HU-08 (spec-008): agregados
// del volumen de viajes finalizados por hora del día (00–23) y día de la
// semana (lunes=0…domingo=6) para un rango inclusivo en Europe/Madrid.
// Solo LEE viajes/estaciones (const. III: cero migraciones); un único
// construirFiltro alimenta agregados y CSV (FR-006, research D3).
package analitica

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/apierr"
	"si16/backend/internal/models"
)

// zonaHoraria del filtro y de la agrupación (FR-001/FR-002, clarificaciones).
const zonaHoraria = "Europe/Madrid"

// Parametros es la entrada ya validada: rango inclusivo obligatorio.
type Parametros struct {
	Desde, Hasta time.Time
}

// Service agrupa las dependencias de lectura de la analítica.
type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func errInterno() *apierr.Error {
	return apierr.New(500, "ERROR_INTERNO", "Error interno al consultar la analítica")
}

// ValidarEntrada exige desde/hasta obligatorios con formato estricto
// YYYY-MM-DD (copiar el criterio de services/viajes.ParsearFechaParam sin
// importar ni modificar ese paquete — independencia por spec, const. V) y
// desde ≤ hasta (clarificación 1: rango invertido → 422).
func ValidarEntrada(desde, hasta string) (Parametros, *apierr.Error) {
	p := Parametros{}
	f1, err := time.Parse("2006-01-02", desde)
	if err != nil || f1.Format("2006-01-02") != desde {
		return p, errEntrada()
	}
	f2, err := time.Parse("2006-01-02", hasta)
	if err != nil || f2.Format("2006-01-02") != hasta {
		return p, errEntrada()
	}
	if f1.After(f2) {
		return p, errEntrada()
	}
	return Parametros{Desde: f1, Hasta: f2}, nil
}

func errEntrada() *apierr.Error {
	return apierr.New(422, "ENTRADA_INVALIDA",
		"desde y hasta son obligatorios, en formato YYYY-MM-DD, y desde no puede ser posterior a hasta")
}

// nuevosBuckets inicializa la respuesta con arrays de 24/7 posiciones a 0
// (research D2: zero-fill en Go; schema global minItems/maxItems).
func nuevosBuckets() models.AnaliticaViajes {
	return models.AnaliticaViajes{
		PorHora:      make([]int, 24),
		PorDiaSemana: make([]int, 7),
	}
}

// indiceDiaSemana mapea la fecha local al bucket lunes=0 … domingo=6
// (PostgreSQL dow: 0=domingo; la fórmula desplaza a lunes=0 — T8c).
func indiceDiaSemana(fecha time.Time) int {
	return (int(fecha.Weekday()) + 6) % 7
}

// construirFiltro devuelve el único predicado del rango (FR-006/V7) y sus
// args, compartidos por TODAS las consultas de agregados y del CSV.
func construirFiltro(p Parametros) (string, []any) {
	const tz = ` AT TIME ZONE '` + zonaHoraria + `')::date`
	filtro := ` AND (v.fecha_inicio` + tz + ` BETWEEN $1 AND $2`
	return filtro, []any{p.Desde, p.Hasta}
}

// Agregados ejecuta Q1 (por_hora), Q2 (por_dia_semana) y Q3 (totales) con el
// mismo WHERE/args; total_viajes = suma de por_hora (por construcción).
func (s *Service) Agregados(ctx context.Context, p Parametros) (models.AnaliticaViajes, *apierr.Error) {
	r := nuevosBuckets()
	r.Desde = p.Desde.Format("2006-01-02")
	r.Hasta = p.Hasta.Format("2006-01-02")
	filtro, args := construirFiltro(p)

	// Q1 — por_hora: hora local española del inicio (0…23).
	rows, err := s.pool.Query(ctx,
		`SELECT EXTRACT(hour FROM (v.fecha_inicio`+` AT TIME ZONE '`+zonaHoraria+`'))::int AS hora,
		        count(*)::int
		   FROM viajes v
		  WHERE v.estado = 'finalizado'`+filtro+`
		  GROUP BY 1`, args...)
	if err != nil {
		log.Printf("analitica: Q1 por_hora: %v", err)
		return r, errInterno()
	}
	for rows.Next() {
		var hora, n int
		if err := rows.Scan(&hora, &n); err != nil {
			rows.Close()
			log.Printf("analitica: scan Q1: %v", err)
			return r, errInterno()
		}
		if hora >= 0 && hora < 24 {
			r.PorHora[hora] = n
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		log.Printf("analitica: rows Q1: %v", err)
		return r, errInterno()
	}

	// Q2 — por_dia_semana: lunes=0 … domingo=6.
	rows, err = s.pool.Query(ctx,
		`SELECT ((EXTRACT(dow FROM (v.fecha_inicio`+` AT TIME ZONE '`+zonaHoraria+`'))::int + 6) % 7) AS dia,
		        count(*)::int
		   FROM viajes v
		  WHERE v.estado = 'finalizado'`+filtro+`
		  GROUP BY 1`, args...)
	if err != nil {
		log.Printf("analitica: Q2 por_dia_semana: %v", err)
		return r, errInterno()
	}
	for rows.Next() {
		var dia, n int
		if err := rows.Scan(&dia, &n); err != nil {
			rows.Close()
			log.Printf("analitica: scan Q2: %v", err)
			return r, errInterno()
		}
		if dia >= 0 && dia < 7 {
			r.PorDiaSemana[dia] = n
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		log.Printf("analitica: rows Q2: %v", err)
		return r, errInterno()
	}

	// Q3 — totales del rango (mismo predicado).
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*)::int, COALESCE(sum(v.coste_total), 0)::float8
		   FROM viajes v
		  WHERE v.estado = 'finalizado'`+filtro, args...,
	).Scan(&r.TotalViajes, &r.TotalCoste); err != nil {
		log.Printf("analitica: Q3 totales: %v", err)
		return r, errInterno()
	}

	// total_viajes se deriva de las propias barras (imposible desalinearse).
	suma := 0
	for _, n := range r.PorHora {
		suma += n
	}
	r.TotalViajes = suma
	return r, nil
}
