// Package estacion (spec-001, Iker Roig) calcula y sirve el estado en vivo de
// las estaciones. Es además el helper compartido que las specs que mutan
// estaciones/bicicletas/anclajes (US2, US3, US5, US6, US7) llaman tras su
// commit: Servicio.RecalcularEstacion (tarea T020).
package estacion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/cache"
	"si16/backend/internal/hooks"
	"si16/backend/internal/models"
)

// timeoutHook acota a los suscriptores, con presupuesto propio: el cambio ya está
// confirmado y no deben heredar un contexto cancelado o agotado por Redis.
const timeoutHook = 3 * time.Second

// ErrNoEncontrada se devuelve cuando la estación no existe.
var ErrNoEncontrada = errors.New("estacion: no encontrada")

// ErrCache indica que el commit ya está hecho pero Redis no se pudo
// actualizar. El llamador solo debe registrarlo: el barrido del worker
// (≤ 10 s) repara el hash (research R16). Nunca debe revertir su transacción.
var ErrCache = errors.New("estacion: estado en vivo no actualizado en Redis")

// EstacionEstado es el esquema EstacionEstado de contracts/openapi.yaml.
type EstacionEstado struct {
	ID               int64     `json:"id"`
	Nombre           string    `json:"nombre"`
	Latitud          float64   `json:"latitud"`
	Longitud         float64   `json:"longitud"`
	CapacidadTotal   int       `json:"capacidad_total"`
	Estado           string    `json:"estado"`
	BicisDisponibles int       `json:"bicis_disponibles"`
	AnclajesLibres   int       `json:"anclajes_libres"`
	OcupacionPct     float64   `json:"ocupacion_pct"`
	ActualizadoEn    time.Time `json:"actualizado_en"`
}

// AnclajeDetalle es un elemento de EstacionDetalle.anclajes.
type AnclajeDetalle struct {
	CodigoAnclaje   string  `json:"codigo_anclaje"`
	Estado          string  `json:"estado"`
	BicicletaCodigo *string `json:"bicicleta_codigo"`
	BicicletaEstado *string `json:"bicicleta_estado"`
}

// EstacionDetalle es el esquema EstacionDetalle de contracts/openapi.yaml.
type EstacionDetalle struct {
	EstacionEstado
	Anclajes []AnclajeDetalle `json:"anclajes"`
}

// Servicio agrupa las dependencias.
type Servicio struct {
	pool  *pgxpool.Pool
	cache *cache.Cliente
	log   *slog.Logger
}

// Nuevo crea el servicio.
func Nuevo(pool *pgxpool.Pool, c *cache.Cliente, log *slog.Logger) *Servicio {
	if log == nil {
		log = slog.Default()
	}
	return &Servicio{pool: pool, cache: c, log: log}
}

// sqlCalcular es la fuente única de los contadores (research R9/R19):
//   - bicis_disponibles = anclajes en estado 'ocupado' cuya bicicleta está
//     'disponible' (averiadas, en_taller y anclajes fuera de servicio no suman);
//   - anclajes_libres   = anclajes en estado 'libre' (un anclaje con bici
//     averiada está 'ocupado' y uno fuera de servicio no es libre).
//
// Por eso bicis_disponibles + anclajes_libres ≤ capacidad_total (no se valida
// la igualdad). now() fija el instante del snapshot leído (actualizado_en).
const sqlCalcular = `
SELECT e.id, e.nombre, e.latitud::float8, e.longitud::float8, e.capacidad_total, e.estado,
       (COUNT(a.id) FILTER (WHERE a.estado = 'ocupado' AND b.estado = 'disponible'))::int,
       (COUNT(a.id) FILTER (WHERE a.estado = 'libre'))::int,
       now()
FROM estaciones e
LEFT JOIN anclajes   a ON a.estacion_id = e.id
LEFT JOIN bicicletas b ON b.id = a.bicicleta_id
WHERE $1::bigint[] IS NULL OR e.id = ANY($1::bigint[])
GROUP BY e.id
ORDER BY e.id`

// calcular obtiene el estado desde PostgreSQL (ids nil = todas).
func (s *Servicio) calcular(ctx context.Context, ids []int64) ([]EstacionEstado, error) {
	rows, err := s.pool.Query(ctx, sqlCalcular, ids)
	if err != nil {
		return nil, fmt.Errorf("estacion: calcular: %w", err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (EstacionEstado, error) {
		var e EstacionEstado
		err := r.Scan(&e.ID, &e.Nombre, &e.Latitud, &e.Longitud, &e.CapacidadTotal, &e.Estado,
			&e.BicisDisponibles, &e.AnclajesLibres, &e.ActualizadoEn)
		e.OcupacionPct = Ocupacion(e.Estado, e.BicisDisponibles, e.CapacidadTotal)
		e.ActualizadoEn = e.ActualizadoEn.UTC()
		return e, err
	})
}

// Ocupacion = 100 × bicis_disponibles / capacidad_total con 2 decimales; 0 si
// la estación está fuera de servicio o sin capacidad (data-model, R9).
func Ocupacion(estado string, bicis, capacidad int) float64 {
	if estado != models.EstacionOperativa || capacidad <= 0 {
		return 0
	}
	return math.Round(10000*float64(bicis)/float64(capacidad)) / 100
}

func aCache(estados []EstacionEstado) []cache.EstadoEstacion {
	out := make([]cache.EstadoEstacion, len(estados))
	for i, e := range estados {
		out[i] = cache.EstadoEstacion{
			EstacionID: e.ID, BicisDisponibles: e.BicisDisponibles, AnclajesLibres: e.AnclajesLibres,
			OcupacionPct: e.OcupacionPct, Estado: e.Estado, ActualizadoEn: e.ActualizadoEn,
		}
	}
	return out
}

// RecalcularEstacion es el ÚNICO punto de entrada que las specs llaman tras el
// commit de una mutación de estaciones/bicicletas/anclajes. Recalcula desde
// PostgreSQL, reescribe estacion:{id}:estado en Redis y, al final, notifica a
// los suscriptores con hooks.EstacionCambiada (T012). No llamar al hook
// directamente ni desde un suscriptor.
//
// Si Redis falla, el error se registra y se devuelve envuelto en ErrCache: el
// llamador solo lo registra (el commit ya está hecho y el barrido lo repara).
func (s *Servicio) RecalcularEstacion(ctx context.Context, ids []int64) error {
	ids = normalizar(ids)
	if len(ids) == 0 {
		return nil
	}
	// Los suscriptores se notifican siempre: el cambio ya está confirmado en
	// PostgreSQL, aunque falle el cálculo o la escritura del estado en vivo.
	defer func() {
		hookCtx, cancelar := context.WithTimeout(context.WithoutCancel(ctx), timeoutHook)
		defer cancelar()
		hooks.EstacionCambiada(hookCtx, ids)
	}()

	estados, err := s.calcular(ctx, ids)
	if err != nil {
		s.log.Error("estacion: recálculo fallido; lo reparará el barrido", "ids", ids, "error", err)
		return err
	}
	if err := s.cache.GuardarEstados(ctx, aCache(estados)); err != nil {
		s.log.Warn("estacion: Redis no actualizado tras el commit; lo reparará el barrido", "ids", ids, "error", err)
		return fmt.Errorf("%w: %v", ErrCache, err)
	}
	return nil
}

// RecalcularTodas recalcula la red completa con una única consulta agregada y
// la escribe en Redis. La usa el worker refresco_estado (red de seguridad de
// frescura); no notifica el hook (las alertas tienen su propio barrido).
func (s *Servicio) RecalcularTodas(ctx context.Context) (int, error) {
	estados, err := s.calcular(ctx, nil)
	if err != nil {
		return 0, err
	}
	if err := s.cache.GuardarEstados(ctx, aCache(estados)); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCache, err)
	}
	return len(estados), nil
}

// Snapshot devuelve la red completa con el estado en vivo de Redis y la marca
// de frescura global (la más antigua de las estaciones). Las claves ausentes o
// expiradas se recalculan desde PostgreSQL y se repueblan (FR-005); si Redis
// no responde, se sirve el cálculo de PostgreSQL con su marca de tiempo.
func (s *Servicio) Snapshot(ctx context.Context) ([]EstacionEstado, time.Time, error) {
	estados, err := s.leer(ctx, nil)
	if err != nil {
		return nil, time.Time{}, err
	}
	var masAntigua time.Time
	for _, e := range estados {
		if masAntigua.IsZero() || e.ActualizadoEn.Before(masAntigua) {
			masAntigua = e.ActualizadoEn
		}
	}
	if masAntigua.IsZero() {
		masAntigua = time.Now().UTC()
	}
	return estados, masAntigua, nil
}

// Detalle devuelve una estación con sus anclajes.
func (s *Servicio) Detalle(ctx context.Context, id int64) (EstacionDetalle, error) {
	estados, err := s.leer(ctx, []int64{id})
	if err != nil {
		return EstacionDetalle{}, err
	}
	if len(estados) == 0 {
		return EstacionDetalle{}, ErrNoEncontrada
	}
	rows, err := s.pool.Query(ctx, `
		SELECT a.codigo_anclaje, a.estado, b.codigo, b.estado
		FROM anclajes a LEFT JOIN bicicletas b ON b.id = a.bicicleta_id
		WHERE a.estacion_id = $1
		ORDER BY a.codigo_anclaje`, id)
	if err != nil {
		return EstacionDetalle{}, fmt.Errorf("estacion: anclajes: %w", err)
	}
	anclajes, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (AnclajeDetalle, error) {
		var a AnclajeDetalle
		return a, r.Scan(&a.CodigoAnclaje, &a.Estado, &a.BicicletaCodigo, &a.BicicletaEstado)
	})
	if err != nil {
		return EstacionDetalle{}, fmt.Errorf("estacion: anclajes: %w", err)
	}
	return EstacionDetalle{EstacionEstado: estados[0], Anclajes: anclajes}, nil
}

// leer combina los datos fijos de PostgreSQL (nombre, coordenadas, capacidad)
// con los contadores y el estado de servicio en vivo de Redis.
func (s *Servicio) leer(ctx context.Context, ids []int64) ([]EstacionEstado, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, nombre, latitud::float8, longitud::float8, capacidad_total
		FROM estaciones
		WHERE $1::bigint[] IS NULL OR id = ANY($1::bigint[])
		ORDER BY id`, ids)
	if err != nil {
		return nil, fmt.Errorf("estacion: leer: %w", err)
	}
	base, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (EstacionEstado, error) {
		var e EstacionEstado
		return e, r.Scan(&e.ID, &e.Nombre, &e.Latitud, &e.Longitud, &e.CapacidadTotal)
	})
	if err != nil {
		return nil, fmt.Errorf("estacion: leer: %w", err)
	}
	if len(base) == 0 {
		return base, nil
	}

	todos := make([]int64, len(base))
	for i, e := range base {
		todos[i] = e.ID
	}
	vivos, err := s.cache.LeerEstados(ctx, todos)
	redisOK := err == nil
	if !redisOK {
		s.log.Warn("estacion: Redis no disponible; se sirve el cálculo de PostgreSQL", "error", err)
		vivos = map[int64]cache.EstadoEstacion{}
	}

	var faltan []int64
	for _, id := range todos {
		if _, ok := vivos[id]; !ok {
			faltan = append(faltan, id)
		}
	}
	recalculados := map[int64]EstacionEstado{}
	if len(faltan) > 0 {
		calc, err := s.calcular(ctx, faltan)
		if err != nil {
			return nil, err
		}
		for _, e := range calc {
			recalculados[e.ID] = e
		}
		// Solo se repuebla si la lectura funcionó: con Redis caído, escribir
		// añadiría otro timeout a la respuesta (el worker repuebla al volver).
		if redisOK {
			if err := s.cache.GuardarEstados(ctx, aCache(calc)); err != nil {
				s.log.Warn("estacion: no se pudo repoblar Redis", "ids", faltan, "error", err)
			}
		}
	}

	for i, e := range base {
		if v, ok := vivos[e.ID]; ok {
			base[i].Estado = v.Estado
			base[i].BicisDisponibles = v.BicisDisponibles
			base[i].AnclajesLibres = v.AnclajesLibres
			base[i].OcupacionPct = v.OcupacionPct
			base[i].ActualizadoEn = v.ActualizadoEn
		} else if r, ok := recalculados[e.ID]; ok {
			base[i] = r
		}
	}
	return base, nil
}

func normalizar(ids []int64) []int64 {
	out := slices.Clone(ids)
	slices.Sort(out)
	return slices.Compact(out)
}
