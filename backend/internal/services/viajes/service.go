// Package viajes implementa el servicio de consulta del historial de viajes
// de la HU-04 (spec-004). Solo LEE: dos sentencias indexadas por petición
// (Q1 página con JOIN a estaciones, Q2 agregado del periodo sin paginar),
// filtro temporal en Europe/Madrid vía AT TIME ZONE — data-model.md y
// research D1/D2/D3. Nunca recalcula costes ni duraciones (FR-007).
package viajes

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/apierr"
	"si16/backend/internal/models"
)

// PageSizeDefecto y límites de paginación (contrato: contracts/viajes.md).
const (
	PageSizeDefecto = 50
	PageSizeMax     = 200
)

// Zona horaria del filtro temporal (FR-002, clarificación 2).
const zonaHoraria = "Europe/Madrid"

// Etiqueta del viaje gratuito (FR-004): la deriva el backend, jamás el cliente.
const etiquetaGratuito = "Tramo gratuito (0,00 €)"

// Parametros es la entrada ya validada de ListarHistorial.
// Desde/Hasta son nil cuando el parámetro no vino (rango abierto por ese lado).
type Parametros struct {
	Desde, Hasta *time.Time
	Page         int
	PageSize     int
}

// ItemHistorial = schema Viaje + nombres de estación (aditivos, patrón
// AlertaBalanceo del OpenAPI global).
type ItemHistorial struct {
	models.Viaje
	EstacionOrigenNombre  string  `json:"estacion_origen_nombre"`
	EstacionDestinoNombre *string `json:"estacion_destino_nombre"`
}

// Respuesta es el envelope del 200: { total, suma_coste, items }.
type Respuesta struct {
	Total     int             `json:"total"`
	SumaCoste float64         `json:"suma_coste"`
	Items     []ItemHistorial `json:"items"`
}

// Service agrupa las dependencias de lectura del historial.
type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func errInterno() *apierr.Error {
	return apierr.New(500, "ERROR_INTERNO", "Error interno al consultar el historial")
}

// EtiquetaParaCoste deriva la etiqueta de tramo gratuito (FR-004): solo
// cuando el coste registrado es exactamente 0.
func EtiquetaParaCoste(coste float64) *string {
	if coste == 0 {
		et := etiquetaGratuito
		return &et
	}
	return nil
}

// Offset es el desplazamiento SQL de la página solicitada (research D6).
func Offset(page, pageSize int) int { return (page - 1) * pageSize }

// ParsearFechaParam exige formato estricto YYYY-MM-DD (sin normalizar:
// un día inexistente o un dígito suelto → 422).
func ParsearFechaParam(s string) (time.Time, *apierr.Error) {
	f, err := time.Parse("2006-01-02", s)
	if err != nil || f.Format("2006-01-02") != s {
		return time.Time{}, errEntrada("desde/hasta deben tener formato YYYY-MM-DD")
	}
	return f, nil
}

func errEntrada(mensaje string) *apierr.Error {
	return apierr.New(422, "ENTRADA_INVALIDA", mensaje)
}

// ValidarEntrada parsea y valida la query del contrato (research D6):
// page ≥ 1 (default 1), page_size ∈ [1,200] (default 50), fechas
// YYYY-MM-DD estrictas y desde ≤ hasta. Cada fecha es opcional e
// independiente (clarificación del contrato).
func ValidarEntrada(desde, hasta, page, pageSize string) (Parametros, *apierr.Error) {
	p := Parametros{Page: 1, PageSize: PageSizeDefecto}

	if page != "" {
		n, err := strconv.Atoi(page)
		if err != nil || n < 1 {
			return p, errEntrada("page debe ser un entero ≥ 1")
		}
		p.Page = n
	}
	if pageSize != "" {
		n, err := strconv.Atoi(pageSize)
		if err != nil || n < 1 || n > PageSizeMax {
			return p, errEntrada("page_size debe estar entre 1 y 200")
		}
		p.PageSize = n
	}
	if desde != "" {
		f, aerr := ParsearFechaParam(desde)
		if aerr != nil {
			return p, aerr
		}
		p.Desde = &f
	}
	if hasta != "" {
		f, aerr := ParsearFechaParam(hasta)
		if aerr != nil {
			return p, aerr
		}
		p.Hasta = &f
	}
	if p.Desde != nil && p.Hasta != nil && p.Desde.After(*p.Hasta) {
		return p, errEntrada("desde no puede ser posterior a hasta")
	}
	return p, nil
}

// ListarHistorial ejecuta Q2 (agregado del periodo completo, sin paginar)
// y Q1 (página con JOIN a estaciones): el agregado no depende del
// LIMIT/OFFSET (FR-003/FR-008, V5).
func (s *Service) ListarHistorial(ctx context.Context, usuarioID int64, p Parametros) (Respuesta, *apierr.Error) {
	// Args compartidos por Q1/Q2: $1 = usuario; fechas (si vienen) en $2….
	var base []any
	base = append(base, usuarioID)
	var filtro string
	const tz = ` AT TIME ZONE '` + zonaHoraria + `')::date`
	switch {
	case p.Desde != nil && p.Hasta != nil:
		filtro = ` AND (v.fecha_inicio` + tz + ` BETWEEN $2 AND $3`
		base = append(base, *p.Desde, *p.Hasta)
	case p.Desde != nil:
		filtro = ` AND (v.fecha_inicio` + tz + ` >= $2`
		base = append(base, *p.Desde)
	case p.Hasta != nil:
		filtro = ` AND (v.fecha_inicio` + tz + ` <= $2`
		base = append(base, *p.Hasta)
	}

	// Q2 — agregado del periodo (count + sum, sin LIMIT/OFFSET).
	var r Respuesta
	err := s.pool.QueryRow(ctx,
		`SELECT count(*)::int, COALESCE(sum(v.coste_total), 0)::float8
		   FROM viajes v
		  WHERE v.usuario_id = $1 AND v.estado = 'finalizado'`+filtro,
		base...,
	).Scan(&r.Total, &r.SumaCoste)
	if err != nil {
		log.Printf("viajes: agregado usuario %d: %v", usuarioID, err)
		return r, errInterno()
	}

	// Q1 — página del historial (orden DESC, nombres por JOIN, sin N+1).
	n := len(base) // $1…$n ocupan base; LIMIT $n+1 OFFSET $n+2
	qargs := append(append([]any{}, base...), p.PageSize, Offset(p.Page, p.PageSize))
	rows, err := s.pool.Query(ctx,
		`SELECT v.id, v.usuario_id, v.bicicleta_id, v.estacion_origen_id,
		        v.estacion_destino_id, v.fecha_inicio, v.fecha_fin,
		        v.duracion_minutos, v.coste_total::float8, v.estado,
		        eo.nombre, ed.nombre
		   FROM viajes v
		   JOIN estaciones eo ON eo.id = v.estacion_origen_id
		   JOIN estaciones ed ON ed.id = v.estacion_destino_id
		  WHERE v.usuario_id = $1 AND v.estado = 'finalizado'`+filtro+`
		  ORDER BY v.fecha_inicio DESC
		  LIMIT $`+strconv.Itoa(n+1)+` OFFSET $`+strconv.Itoa(n+2),
		qargs...)
	if err != nil {
		log.Printf("viajes: página usuario %d: %v", usuarioID, err)
		return r, errInterno()
	}
	defer rows.Close()

	items := []ItemHistorial{} // nunca null en JSON
	for rows.Next() {
		var it ItemHistorial
		if err := rows.Scan(&it.ID, &it.UsuarioID, &it.BicicletaID,
			&it.EstacionOrigenID, &it.EstacionDestinoID, &it.FechaInicio,
			&it.FechaFin, &it.DuracionMinutos, &it.CosteTotal, &it.Estado,
			&it.EstacionOrigenNombre, &it.EstacionDestinoNombre); err != nil {
			log.Printf("viajes: scan usuario %d: %v", usuarioID, err)
			return r, errInterno()
		}
		it.Etiqueta = EtiquetaParaCoste(it.CosteTotal) // FR-004, server-side
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		log.Printf("viajes: rows usuario %d: %v", usuarioID, err)
		return r, errInterno()
	}
	r.Items = items
	return r, nil
}
