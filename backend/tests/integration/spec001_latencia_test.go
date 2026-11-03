// T099 — Camino de respaldo de la latencia extremo a extremo (HU-01 FR-003,
// SC-002, SC-006, SC-007; research R16).
package integration

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"si16/backend/internal/cache"
	"si16/backend/internal/models"
	"si16/backend/internal/services/estacion"
	tu "si16/backend/internal/testutil"
	"si16/backend/internal/workers/refresco_estado"
)

// Con intervalo de 1 s el test es determinista: espera 1 s + 2 s de margen.
const (
	intervaloTest = time.Second
	margen        = 2 * time.Second
)

func arrancarWorker(t *testing.T, svc *estacion.Servicio, c *cache.Cliente) {
	t.Helper()
	ctx, cancelar := context.WithCancel(context.Background())
	t.Cleanup(cancelar)
	w := refresco_estado.Nuevo(svc, c, slog.New(slog.NewTextHandler(io.Discard, nil)), intervaloTest)
	go w.Iniciar(ctx)
}

func TestBarridoReflejaMutacionesSinRecalculoSincrono(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Respaldo", models.EstacionOperativa, tu.ConBici, tu.ConBici, tu.Libre)
	if err := e.Estacion.RecalcularEstacion(e.Ctx, []int64{id}); err != nil {
		t.Fatal(err)
	}
	arrancarWorker(t, e.Estacion, e.Cache)

	casos := []struct {
		nombre string
		sql    string
		ok     func(cache.EstadoEstacion) bool
	}{
		{"bici averiada (anclajes/bicicletas)",
			`UPDATE bicicletas SET estado='averiada' WHERE id IN (SELECT bicicleta_id FROM anclajes WHERE estacion_id=$1)`,
			func(s cache.EstadoEstacion) bool { return s.BicisDisponibles == 0 }},
		{"vuelve a haber disponibilidad (SC-006)",
			`UPDATE bicicletas SET estado='disponible' WHERE id IN (SELECT bicicleta_id FROM anclajes WHERE estacion_id=$1)`,
			func(s cache.EstadoEstacion) bool { return s.BicisDisponibles == 2 }},
		{"pasa a fuera de servicio (SC-007)",
			`UPDATE estaciones SET estado='fuera_de_servicio' WHERE id=$1`,
			func(s cache.EstadoEstacion) bool {
				return s.Estado == models.EstacionFueraDeServicio && s.OcupacionPct == 0
			}},
		{"vuelve a operativa (SC-007)",
			`UPDATE estaciones SET estado='operativa' WHERE id=$1`,
			func(s cache.EstadoEstacion) bool { return s.Estado == models.EstacionOperativa }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e.Exec(t, c.sql, id) // mutación directa: NO se llama a RecalcularEstacion
			if !tu.Esperar(t, intervaloTest+margen, func() bool {
				s, ok := leerHash(t, e, id)
				return ok && c.ok(s)
			}) {
				s, _ := leerHash(t, e, id)
				t.Fatalf("el barrido no reflejó el cambio en %v: %+v", intervaloTest+margen, s)
			}
		})
	}
	if v := e.Cache.Redis().Get(e.Ctx, cache.ClaveUltimoBarrido).Val(); v == "" {
		t.Errorf("el worker no registró %s", cache.ClaveUltimoBarrido)
	}
}

func TestRedisCaidoNoRevierteElCommitYElBarridoRepara(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "SinRedis", models.EstacionOperativa, tu.ConBici, tu.Libre)
	if err := e.Estacion.RecalcularEstacion(e.Ctx, []int64{id}); err != nil {
		t.Fatal(err)
	}

	// Servicio cuyo Redis no responde (puerto cerrado).
	caido := cache.Nuevo("127.0.0.1:1", e.RedisDB)
	t.Cleanup(func() { _ = caido.Cerrar() })
	svcCaido := estacion.Nuevo(e.Pool, caido, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// "Desbloqueo": la bici sale del anclaje; commit en PostgreSQL.
	e.Exec(t, `UPDATE anclajes SET estado='libre', bicicleta_id=NULL WHERE estacion_id=$1 AND estado='ocupado'`, id)
	ctx, cancelar := context.WithTimeout(e.Ctx, 3*time.Second)
	defer cancelar()
	err := svcCaido.RecalcularEstacion(ctx, []int64{id})
	if !errors.Is(err, estacion.ErrCache) {
		t.Fatalf("err = %v; quiero ErrCache", err)
	}
	// El commit sigue ahí (no se revierte por fallar Redis).
	var libres int
	if err := e.Pool.QueryRow(e.Ctx, `SELECT count(*) FROM anclajes WHERE estacion_id=$1 AND estado='libre'`, id).Scan(&libres); err != nil || libres != 2 {
		t.Fatalf("libres en PostgreSQL = %d (%v); quiero 2", libres, err)
	}
	// Redis sigue con el dato viejo hasta que el barrido (con Redis ya sano) lo repara.
	if s, _ := leerHash(t, e, id); s.AnclajesLibres != 1 {
		t.Fatalf("hash = %+v; esperaba aún el dato antiguo", s)
	}
	arrancarWorker(t, e.Estacion, e.Cache)
	if !tu.Esperar(t, intervaloTest+margen, func() bool {
		s, ok := leerHash(t, e, id)
		return ok && s.AnclajesLibres == 2 && s.BicisDisponibles == 0
	}) {
		t.Fatal("el barrido no reparó el hash tras volver Redis")
	}
}

func TestSnapshotFuncionaConRedisCaido(t *testing.T) {
	e := iniciar(t)
	e.CrearEstacion(t, "Fallback", models.EstacionOperativa, tu.ConBici, tu.Libre)
	caido := cache.Nuevo("127.0.0.1:1", e.RedisDB)
	t.Cleanup(func() { _ = caido.Cerrar() })
	svc := estacion.Nuevo(e.Pool, caido, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancelar := context.WithTimeout(e.Ctx, 5*time.Second)
	defer cancelar()
	estados, actualizado, err := svc.Snapshot(ctx)
	if err != nil {
		t.Fatalf("Snapshot con Redis caído: %v", err)
	}
	if len(estados) != 1 || estados[0].BicisDisponibles != 1 || actualizado.IsZero() {
		t.Fatalf("snapshot = %+v (%v)", estados, actualizado)
	}
}
