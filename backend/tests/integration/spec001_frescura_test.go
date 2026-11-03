// T025 — Frescura y cómputo de disponibilidad (HU-01 FR-005, FR-006, SC-003;
// research R9, R18, R19).
package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"si16/backend/internal/cache"
	"si16/backend/internal/hooks"
	"si16/backend/internal/models"
	"si16/backend/internal/services/estacion"
	tu "si16/backend/internal/testutil"
)

const redisDB = 15

func iniciar(t *testing.T) *tu.Entorno { return tu.Iniciar(t, "integration", redisDB) }

func anclajes(grupos ...[]tu.Anclaje) []tu.Anclaje {
	var out []tu.Anclaje
	for _, g := range grupos {
		out = append(out, g...)
	}
	return out
}

func leerHash(t *testing.T, e *tu.Entorno, id int64) (cache.EstadoEstacion, bool) {
	t.Helper()
	m, err := e.Cache.LeerEstados(e.Ctx, []int64{id})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := m[id]
	return v, ok
}

func TestMutacionActualizaHashTrasCommit(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Mutación", models.EstacionOperativa, anclajes(tu.Repetir(2, tu.ConBici), tu.Repetir(2, tu.Libre))...)
	if err := e.Estacion.RecalcularEstacion(e.Ctx, []int64{id}); err != nil {
		t.Fatal(err)
	}
	antes, _ := leerHash(t, e, id)
	if antes.BicisDisponibles != 2 || antes.AnclajesLibres != 2 {
		t.Fatalf("estado inicial = %+v; quiero 2 bicis y 2 libres", antes)
	}

	// Una "devolución": un anclaje libre pasa a ocupado con bici disponible.
	e.Exec(t, `WITH b AS (INSERT INTO bicicletas (codigo) VALUES ('NUEVA') RETURNING id)
		UPDATE anclajes SET estado='ocupado', bicicleta_id=(SELECT id FROM b)
		WHERE id = (SELECT min(id) FROM anclajes WHERE estacion_id=$1 AND estado='libre')`, id)
	if err := e.Estacion.RecalcularEstacion(e.Ctx, []int64{id}); err != nil {
		t.Fatal(err)
	}
	despues, ok := leerHash(t, e, id)
	if !ok || despues.BicisDisponibles != 3 || despues.AnclajesLibres != 1 {
		t.Fatalf("tras la mutación = %+v; quiero 3 bicis y 1 libre", despues)
	}
	if !despues.ActualizadoEn.After(antes.ActualizadoEn) {
		t.Errorf("actualizado_en no avanza: %v → %v", antes.ActualizadoEn, despues.ActualizadoEn)
	}
	ttl := e.Cache.Redis().TTL(e.Ctx, cache.ClaveEstado(id)).Val()
	if ttl <= 0 || ttl > cache.TTLEstado {
		t.Errorf("TTL = %v; quiero (0, %v]", ttl, cache.TTLEstado)
	}
}

func TestClaveExpiradaRecalculaDesdePostgresYRepuebla(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Expirada", models.EstacionOperativa, tu.ConBici, tu.Libre, tu.Libre)
	// No hay clave en Redis (nunca se calculó o expiró).
	if _, ok := leerHash(t, e, id); ok {
		t.Fatal("la clave no debería existir todavía")
	}
	estados, actualizado, err := e.Estacion.Snapshot(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(estados) != 1 || estados[0].BicisDisponibles != 1 || estados[0].AnclajesLibres != 2 {
		t.Fatalf("snapshot = %+v", estados)
	}
	if actualizado.IsZero() || time.Since(actualizado) > 5*time.Second {
		t.Errorf("actualizado_en global = %v; debe ser reciente (nunca datos sin marca de tiempo, FR-005)", actualizado)
	}
	if _, ok := leerHash(t, e, id); !ok {
		t.Error("el fallback no repobló estacion:{id}:estado")
	}
}

func TestAveriadasNoSumanEnNingunContador(t *testing.T) {
	e := iniciar(t)
	// 2 operativas, 1 averiada, 1 en taller, 2 libres, 1 anclaje roto → capacidad 7.
	id := e.CrearEstacion(t, "Mixta", models.EstacionOperativa,
		anclajes(tu.Repetir(2, tu.ConBici), []tu.Anclaje{tu.ConAveriada, tu.ConEnTaller, tu.Roto}, tu.Repetir(2, tu.Libre))...)
	estados, _, err := e.Estacion.Snapshot(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	s := estados[0]
	if s.BicisDisponibles != 2 {
		t.Errorf("bicis_disponibles = %d; quiero 2 (averiada y en_taller excluidas, FR-006)", s.BicisDisponibles)
	}
	if s.AnclajesLibres != 2 {
		t.Errorf("anclajes_libres = %d; quiero 2 (ni el de la averiada ni el roto son libres, SC-003)", s.AnclajesLibres)
	}
	if s.BicisDisponibles+s.AnclajesLibres >= s.CapacidadTotal {
		t.Errorf("se esperaba suma < capacidad (R19): %d + %d vs %d", s.BicisDisponibles, s.AnclajesLibres, s.CapacidadTotal)
	}
	if s.OcupacionPct != 28.57 {
		t.Errorf("ocupacion_pct = %v; quiero 28.57 (2/7)", s.OcupacionPct)
	}
	_ = id
}

func TestTodasAveriadasDaCeroYCero(t *testing.T) {
	e := iniciar(t)
	e.CrearEstacion(t, "Averiadas", models.EstacionOperativa, tu.Repetir(4, tu.ConAveriada)...)
	estados, _, err := e.Estacion.Snapshot(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	if s := estados[0]; s.BicisDisponibles != 0 || s.AnclajesLibres != 0 {
		t.Fatalf("= %d bicis, %d libres; quiero 0 y 0", s.BicisDisponibles, s.AnclajesLibres)
	}
}

func TestFueraDeServicioDevuelveEnterosEnLaAPI(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Cerrada", models.EstacionFueraDeServicio, tu.ConBici, tu.ConBici, tu.Libre)
	rec := e.Get(t, "/estaciones", e.Token(t, models.RolUsuario))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /estaciones = %d %s", rec.Code, rec.Body)
	}
	var cuerpo struct {
		Estaciones []map[string]any `json:"estaciones"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cuerpo); err != nil {
		t.Fatal(err)
	}
	s := cuerpo.Estaciones[0]
	if s["id"].(float64) != float64(id) || s["estado"] != "fuera_de_servicio" {
		t.Fatalf("estación = %v", s)
	}
	// El contrato EstacionEstado no cambia (R18): enteros, no null.
	if s["bicis_disponibles"] != float64(2) || s["anclajes_libres"] != float64(1) {
		t.Errorf("contadores = %v / %v; quiero 2 / 1", s["bicis_disponibles"], s["anclajes_libres"])
	}
	if s["ocupacion_pct"] != float64(0) {
		t.Errorf("ocupacion_pct = %v; quiero 0 en estación fuera de servicio", s["ocupacion_pct"])
	}
}

func TestDosCambiosSucesivosTerminanEnElEstadoFinal(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Sucesivos", models.EstacionOperativa, tu.ConBici, tu.ConBici, tu.Libre)
	e.Exec(t, `UPDATE bicicletas SET estado='averiada' WHERE id = (SELECT min(bicicleta_id) FROM anclajes WHERE estacion_id=$1)`, id)
	_ = e.Estacion.RecalcularEstacion(e.Ctx, []int64{id})
	e.Exec(t, `UPDATE bicicletas SET estado='averiada' WHERE id = (SELECT max(bicicleta_id) FROM anclajes WHERE estacion_id=$1)`, id)
	_ = e.Estacion.RecalcularEstacion(e.Ctx, []int64{id})
	if s, _ := leerHash(t, e, id); s.BicisDisponibles != 0 {
		t.Fatalf("bicis_disponibles = %d; quiero 0 (estado final)", s.BicisDisponibles)
	}

	// Una escritura tardía de un cálculo MÁS ANTIGUO no pisa el estado final.
	viejo := cache.EstadoEstacion{EstacionID: id, BicisDisponibles: 2, AnclajesLibres: 1, Estado: "operativa",
		ActualizadoEn: time.Now().Add(-time.Minute)}
	if err := e.Cache.GuardarEstados(e.Ctx, []cache.EstadoEstacion{viejo}); err != nil {
		t.Fatal(err)
	}
	if s, _ := leerHash(t, e, id); s.BicisDisponibles != 0 {
		t.Fatalf("un estado antiguo pisó al más reciente: %+v", s)
	}
}

func TestRecalcularNotificaUnaVezSinReentrar(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Hook", models.EstacionOperativa, tu.ConBici, tu.Libre)
	var llamadas atomic.Int32
	var recibidos []int64
	hooks.Suscribir(func(_ context.Context, ids []int64) error {
		llamadas.Add(1)
		recibidos = ids
		return nil
	})
	if err := e.Estacion.RecalcularEstacion(e.Ctx, []int64{id, id}); err != nil {
		t.Fatal(err)
	}
	if n := llamadas.Load(); n != 1 {
		t.Fatalf("el suscriptor recibió %d notificaciones; quiero 1", n)
	}
	if len(recibidos) != 1 || recibidos[0] != id {
		t.Errorf("ids notificados = %v; quiero [%d] sin duplicados", recibidos, id)
	}
	// El barrido del worker no notifica (las alertas tienen su propio barrido).
	if _, err := e.Estacion.RecalcularTodas(e.Ctx); err != nil {
		t.Fatal(err)
	}
	if n := llamadas.Load(); n != 1 {
		t.Errorf("RecalcularTodas notificó al hook (%d llamadas)", n)
	}
}

func TestDetalleIncluyeAnclajesY404(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Detalle", models.EstacionOperativa, tu.ConBici, tu.ConAveriada, tu.Libre)
	d, err := e.Estacion.Detalle(e.Ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Anclajes) != 3 || d.BicisDisponibles != 1 || d.AnclajesLibres != 1 {
		t.Fatalf("detalle = %+v", d)
	}
	if _, err := e.Estacion.Detalle(e.Ctx, id+999); err != estacion.ErrNoEncontrada {
		t.Errorf("estación inexistente: err = %v; quiero ErrNoEncontrada", err)
	}
	tok := e.Token(t, models.RolUsuario)
	if rec := e.Get(t, "/estaciones/99999", tok); rec.Code != http.StatusNotFound {
		t.Errorf("GET /estaciones/99999 = %d; quiero 404", rec.Code)
	}
	if rec := e.Get(t, "/estaciones/abc", tok); rec.Code != http.StatusNotFound {
		t.Errorf("GET /estaciones/abc = %d; quiero 404", rec.Code)
	}
}

func TestEstacionesExigeSesion(t *testing.T) {
	e := iniciar(t)
	if rec := e.Get(t, "/estaciones", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("sin token = %d; quiero 401", rec.Code)
	}
	if rec := e.Get(t, "/estaciones/1", "token-falso"); rec.Code != http.StatusUnauthorized {
		t.Errorf("token falso = %d; quiero 401", rec.Code)
	}
	// Cualquier rol autenticado puede consultar el mapa (Assumptions de la spec).
	for _, rol := range []models.Rol{models.RolUsuario, models.RolOperarioTaller, models.RolOperarioLogistica, models.RolGestor} {
		if rec := e.Get(t, "/estaciones", e.Token(t, rol)); rec.Code != http.StatusOK {
			t.Errorf("rol %s = %d; quiero 200", rol, rec.Code)
		}
	}
}
