// Package cache encapsula el acceso a Redis y el formato del estado en vivo de
// las estaciones (constitución IV, research R3/R4). El cálculo de los
// contadores y el fallback a PostgreSQL viven en services/estacion; este
// paquete solo lee y escribe el contrato de claves.
package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// EstadoEstacion es el contenido del HASH estacion:{id}:estado.
type EstadoEstacion struct {
	EstacionID       int64
	BicisDisponibles int
	AnclajesLibres   int
	OcupacionPct     float64
	Estado           string // operativa | fuera_de_servicio
	ActualizadoEn    time.Time
}

// Cliente envuelve go-redis con las operaciones del contrato.
type Cliente struct {
	rdb *redis.Client
}

// Nuevo crea el cliente (db = base lógica de Redis; 0 en desarrollo).
func Nuevo(addr string, db int) *Cliente {
	return &Cliente{rdb: redis.NewClient(&redis.Options{Addr: addr, DB: db, ContextTimeoutEnabled: true})}
}

// NuevoDesdeRedis reutiliza un cliente ya creado por el proceso. Lo usan las
// rutas registradas desde deps.Deps, que reciben Redis como dependencia comun.
func NuevoDesdeRedis(rdb *redis.Client) *Cliente {
	return &Cliente{rdb: rdb}
}

// Ping comprueba la conexión.
func (c *Cliente) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }

// Cerrar libera las conexiones.
func (c *Cliente) Cerrar() error { return c.rdb.Close() }

// Redis expone el cliente subyacente (tests y workers).
func (c *Cliente) Redis() *redis.Client { return c.rdb }

// scriptGuardar escribe el hash completo con su TTL de forma atómica, salvo que
// el estado ya guardado sea MÁS reciente que el nuevo. Así dos recálculos
// concurrentes nunca dejan en Redis el más antiguo (caso límite de HU-01:
// "dos cambios sucesivos → la vista termina con el estado final").
//
// Los instantes se comparan normalizados (fecha-hora de 19 caracteres + fracción
// rellenada a 9 dígitos), no como texto crudo: así funciona igual con el
// formato del contrato ("…T10:15:32Z") que con fracciones de cualquier ancho.
// Requiere UTC con sufijo Z, como fija el contrato.
var scriptGuardar = redis.NewScript(`
local function norm(t)
  local frac = string.match(t, '^.-%.(%d+)') or ''
  return string.sub(t, 1, 19) .. string.sub(frac .. '000000000', 1, 9)
end
local actual = redis.call('HGET', KEYS[1], 'actualizado_en')
if actual and norm(actual) > norm(ARGV[5]) then
  return 0
end
redis.call('HSET', KEYS[1],
  'bicis_disponibles', ARGV[1],
  'anclajes_libres',   ARGV[2],
  'ocupacion_pct',     ARGV[3],
  'estado',            ARGV[4],
  'actualizado_en',    ARGV[5])
redis.call('EXPIRE', KEYS[1], ARGV[6])
return 1
`)

// GuardarEstados escribe varios estados en un único pipeline.
func (c *Cliente) GuardarEstados(ctx context.Context, estados []EstadoEstacion) error {
	if len(estados) == 0 {
		return nil
	}
	err := c.ejecutarGuardar(ctx, estados)
	if err != nil && redis.HasErrorPrefix(err, "NOSCRIPT") {
		// Primer uso del script en este Redis: cargarlo y reintentar UNA vez.
		if err = scriptGuardar.Load(ctx, c.rdb).Err(); err == nil {
			err = c.ejecutarGuardar(ctx, estados)
		}
	}
	if err != nil {
		return fmt.Errorf("cache: guardar estados: %w", err)
	}
	return nil
}

func (c *Cliente) ejecutarGuardar(ctx context.Context, estados []EstadoEstacion) error {
	pipe := c.rdb.Pipeline()
	for _, e := range estados {
		scriptGuardar.EvalSha(ctx, pipe, []string{ClaveEstado(e.EstacionID)},
			strconv.Itoa(e.BicisDisponibles),
			strconv.Itoa(e.AnclajesLibres),
			strconv.FormatFloat(e.OcupacionPct, 'f', 2, 64),
			e.Estado,
			e.ActualizadoEn.UTC().Format(FormatoActualizadoEn),
			int(TTLEstado.Seconds()),
		)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// LeerEstados devuelve los estados presentes en Redis para los ids dados (un
// HGETALL por id en pipeline). Los ids ausentes, expirados o corruptos no
// aparecen en el mapa: el llamador debe recalcularlos (fallback, FR-005).
func (c *Cliente) LeerEstados(ctx context.Context, ids []int64) (map[int64]EstadoEstacion, error) {
	res := make(map[int64]EstadoEstacion, len(ids))
	if len(ids) == 0 {
		return res, nil
	}
	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, ClaveEstado(id))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("cache: leer estados: %w", err)
	}
	for i, id := range ids {
		campos, err := cmds[i].Result()
		if err != nil || len(campos) == 0 {
			continue
		}
		if e, ok := parsear(id, campos); ok {
			res[id] = e
		}
	}
	return res, nil
}

// MarcarBarrido registra el instante del último barrido completo del worker.
func (c *Cliente) MarcarBarrido(ctx context.Context, t time.Time) error {
	return c.rdb.Set(ctx, ClaveUltimoBarrido, t.UTC().Format(time.RFC3339Nano), 0).Err()
}

func parsear(id int64, m map[string]string) (EstadoEstacion, bool) {
	bicis, err1 := strconv.Atoi(m[CampoBicisDisponibles])
	libres, err2 := strconv.Atoi(m[CampoAnclajesLibres])
	ocup, err3 := strconv.ParseFloat(m[CampoOcupacionPct], 64)
	t, err4 := time.Parse(time.RFC3339Nano, m[CampoActualizadoEn])
	estado := m[CampoEstado]
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil ||
		(estado != "operativa" && estado != "fuera_de_servicio") {
		return EstadoEstacion{}, false
	}
	return EstadoEstacion{
		EstacionID: id, BicisDisponibles: bicis, AnclajesLibres: libres,
		OcupacionPct: ocup, Estado: estado, ActualizadoEn: t.UTC(),
	}, true
}
