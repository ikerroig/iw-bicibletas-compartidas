// Tests de integración de la HU-07 (alertas de balanceo y traslados).
// Ejecutan contra la BD y Redis del docker-compose (ver quickstart.md §2: T7a–T7h).
// TDD: deben FALLAR hasta implementar services/balanceo + workers/alertas +
// handlers + rutas. Patrón: tests/devolucion_test.go e historial_test.go.
package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/cache"
	"si16/backend/internal/deps"
	"si16/backend/internal/models"
	"si16/backend/internal/services/balanceo"
	"si16/backend/internal/workers/alertas"
)

// ---- helpers de este fichero -------------------------------------------------

func crearUsuarioRol(t *testing.T, d deps.Deps, email, rol string) (int64, string) {
	t.Helper()
	hash := nuevoHash(t)
	var id int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
		 VALUES ('BalTester', $1, $2, $3, true, CURRENT_DATE + 365)
		 RETURNING id`, email, hash, rol).Scan(&id); err != nil {
		t.Fatalf("crear usuario %s: %v", rol, err)
	}
	token, err := auth.CrearToken(d.JWTSecret, id, "BalTester", rol)
	if err != nil {
		t.Fatalf("token %s: %v", rol, err)
	}
	return id, token
}

func crearEstacionBal(t *testing.T, d deps.Deps, nombre string, capacidad int) int64 {
	t.Helper()
	var id int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		 VALUES ($1, 39.47, -0.38, $2) RETURNING id`, nombre, capacidad).Scan(&id); err != nil {
		t.Fatalf("estación %s: %v", nombre, err)
	}
	return id
}

// poblar crea `ocupados` bicis ancladas (estado estadoBici) y `libres` anclajes
// libres en la estación, con códigos derivados de prefijo (espacios únicos
// por test; cumple los CHECKs de anclajes).
func poblar(t *testing.T, d deps.Deps, estacion int64, prefijo string, ocupados, libres int, estadoBici string) {
	t.Helper()
	if estadoBici == "" {
		estadoBici = "disponible"
	}
	if ocupados > 0 {
		insert(t, d,
			`INSERT INTO bicicletas (codigo, estado)
			 SELECT $1 || 'BICI' || lpad(g::text, 5, '0'), $2
			   FROM generate_series(1, $3) g`, prefijo, estadoBici, ocupados)
		insert(t, d,
			`INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
			 SELECT $1, $2 || 'A' || lpad(g::text, 5, '0'), 'ocupado', b.id
			   FROM generate_series(1, $3) g
			   JOIN bicicletas b ON b.codigo = $2 || 'BICI' || lpad(g::text, 5, '0')`,
			estacion, prefijo, ocupados)
	}
	if libres > 0 {
		insert(t, d,
			`INSERT INTO anclajes (estacion_id, codigo_anclaje)
			 SELECT $1, $2 || 'L' || lpad(g::text, 5, '0')
			   FROM generate_series(1, $3) g`, estacion, prefijo, libres)
	}
}

type escenarioBalanceo struct {
	tokenLog     string // operario_logistica (rol válido)
	operarioID   int64
	tokenUsuario string
	tokenGestor  string
	tokenTaller  string
	vacia        int64 // 20 capacidad, 1 bici → 5 %
	llena        int64 // 20 capacidad, 19 bicis → 95 %
}

// prepararBalanceo limpia la BD (fixtures propios) y crea el operario, los
// distractores de rol y las estaciones base en extremos.
func prepararBalanceo(t *testing.T, d deps.Deps) escenarioBalanceo {
	t.Helper()
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	_, tokenUsuario := crearUsuario(t, d, "bal-usuario@test.local", false)
	_, tokenGestor := crearUsuarioRol(t, d, "bal-gestor@test.local", "gestor")
	_, tokenTaller := crearUsuarioRol(t, d, "bal-taller@test.local", "operario_taller")
	operarioID, tokenLog := crearUsuarioRol(t, d, "bal-log@test.local", "operario_logistica")

	vacia := crearEstacionBal(t, d, "B-Vacia", 20)
	poblar(t, d, vacia, "V", 1, 19, "disponible")
	llena := crearEstacionBal(t, d, "B-Llena", 20)
	poblar(t, d, llena, "L", 19, 1, "disponible")

	return escenarioBalanceo{
		tokenLog: tokenLog, operarioID: operarioID, tokenUsuario: tokenUsuario,
		tokenGestor: tokenGestor, tokenTaller: tokenTaller,
		vacia: vacia, llena: llena,
	}
}

func evaluar(t *testing.T, d deps.Deps, estacion int64) {
	t.Helper()
	if err := balanceo.New(d.Pool).EvaluarEstacion(ctx, estacion); err != nil {
		t.Fatalf("EvaluarEstacion(%d): %v", estacion, err)
	}
}

func contarAlerta(t *testing.T, d deps.Deps, estacion int64, resuelta bool) int {
	t.Helper()
	var n int
	if err := d.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM alertas_balanceo WHERE estacion_id = $1 AND resuelta = $2`,
		estacion, resuelta).Scan(&n); err != nil {
		t.Fatalf("contar alertas: %v", err)
	}
	return n
}

type respuestaPanel struct {
	Items []models.AlertaBalanceo `json:"items"`
}

func getPanel(app *gin.Engine, token string) (respuesta, respuestaPanel) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/balanceo/alertas", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	r := respuesta{status: w.Code, raw: w.Body.String()}
	_ = json.Unmarshal(w.Body.Bytes(), &r.body)
	var p respuestaPanel
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	return r, p
}

func postTraslado(app *gin.Engine, token, body string) respuesta {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/balanceo/traslados", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	r := respuesta{status: w.Code, raw: w.Body.String()}
	_ = json.Unmarshal(w.Body.Bytes(), &r.body)
	return r
}

// recuentos de una estación para demostrar "sin alterar nada" (T7g).
type recuentoEstacion struct {
	bicis        int
	anclajesOcup int
	alertas      int
}

func recuento(t *testing.T, d deps.Deps, estacion int64) recuentoEstacion {
	t.Helper()
	var r recuentoEstacion
	if err := d.Pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*) FROM anclajes a JOIN bicicletas b ON b.id = a.bicicleta_id
		          WHERE a.estacion_id = $1 AND b.estado = 'disponible'),
		        (SELECT COUNT(*) FROM anclajes a WHERE a.estacion_id = $1 AND a.estado = 'ocupado'),
		        (SELECT COUNT(*) FROM alertas_balanceo WHERE estacion_id = $1)`,
		estacion).Scan(&r.bicis, &r.anclajesOcup, &r.alertas); err != nil {
		t.Fatalf("recuento estación %d: %v", estacion, err)
	}
	return r
}

// ---- T7a: umbrales estrictos (gate T5 constitución I) -----------------------

func TestBalanceoT7aUmbralesEstrictos(t *testing.T) {
	d := depsTest(t)
	prepararBalanceo(t, d)

	// 9,99 % (999/10000) y 90,01 % (9001/10000) son los únicos valores
	// exactos alcanzables con esos porcentajes (capacidad 10000).
	e999 := crearEstacionBal(t, d, "B-9,99", 10000)
	poblar(t, d, e999, "E999", 999, 9001, "disponible")
	e10 := crearEstacionBal(t, d, "B-10,00", 20)
	poblar(t, d, e10, "E10", 2, 18, "disponible")
	e90 := crearEstacionBal(t, d, "B-90,00", 10)
	poblar(t, d, e90, "E90", 9, 1, "disponible")
	e9001 := crearEstacionBal(t, d, "B-90,01", 10000)
	poblar(t, d, e9001, "E9001", 9001, 999, "disponible")

	for _, id := range []int64{e999, e10, e90, e9001} {
		evaluar(t, d, id)
	}

	if n := contarAlerta(t, d, e999, false); n != 1 {
		t.Errorf("9,99 %% → %d alertas abiertas, esperado 1 (vacia_menos_10)", n)
	}
	if n := contarAlerta(t, d, e10, false); n != 0 {
		t.Errorf("10,00 %% exacto → %d alertas, esperado 0 (umbral estricto)", n)
	}
	if n := contarAlerta(t, d, e90, false); n != 0 {
		t.Errorf("90,00 %% exacto → %d alertas, esperado 0 (umbral estricto)", n)
	}
	if n := contarAlerta(t, d, e9001, false); n != 1 {
		t.Errorf("90,01 %% → %d alertas abiertas, esperado 1 (llena_mas_90)", n)
	}

	var tipo, severidad string
	if err := d.Pool.QueryRow(ctx,
		`SELECT tipo_alerta, severidad FROM alertas_balanceo
		  WHERE estacion_id = $1 AND NOT resuelta`, e999).Scan(&tipo, &severidad); err != nil {
		t.Fatalf("alerta 9,99: %v", err)
	}
	if tipo != "vacia_menos_10" {
		t.Errorf("tipo = %q, esperado vacia_menos_10", tipo)
	}
	if severidad != "alta" {
		t.Errorf("severidad = %q, esperado alta", severidad)
	}
}

// ---- T7b: unicidad y resolución (FR-003/FR-004) -----------------------------

func TestBalanceoT7bUnicidadYResolucion(t *testing.T) {
	d := depsTest(t)
	bal := prepararBalanceo(t, d)

	// La estación LLENA base está al 95 %.
	evaluar(t, d, bal.llena)
	evaluar(t, d, bal.llena) // segunda pasada → sin duplicados
	if n := contarAlerta(t, d, bal.llena, false); n != 1 {
		t.Fatalf("evaluar 2× → %d alertas abiertas, esperado 1 (unicidad FR-003)", n)
	}

	// Volver a rango: dejar exactamente 2/20 = 10,00 % (dentro de [10, 90]).
	insert(t, d,
		`UPDATE anclajes SET estado = 'libre', bicicleta_id = NULL
		  WHERE id IN (SELECT id FROM anclajes
		               WHERE estacion_id = $1 AND estado = 'ocupado'
		               ORDER BY id DESC LIMIT 17)`, bal.llena)
	evaluar(t, d, bal.llena)
	if n := contarAlerta(t, d, bal.llena, false); n != 0 {
		t.Errorf("al volver a 10 %% → %d alertas abiertas, esperado 0 (FR-004)", n)
	}
	if n := contarAlerta(t, d, bal.llena, true); n != 1 {
		t.Errorf("al volver a rango → %d resueltas, esperado 1", n)
	}

	// Re-cruzar el umbral (17 bicis más → 19/20 = 95 %) → fila NUEVA abierta.
	poblar(t, d, bal.llena, "L2", 17, 0, "disponible")
	evaluar(t, d, bal.llena)
	if n := contarAlerta(t, d, bal.llena, false); n != 1 {
		t.Errorf("re-cruzar → %d alertas abiertas, esperado 1 (fila nueva)", n)
	}
	if n := contarAlerta(t, d, bal.llena, true); n != 1 {
		t.Errorf("la fila resuelta debe conservarse como histórico: %d, esperado 1", n)
	}
	var total int
	if err := d.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM alertas_balanceo WHERE estacion_id = $1`, bal.llena).Scan(&total); err != nil {
		t.Fatalf("total alertas: %v", err)
	}
	if total != 2 {
		t.Errorf("total de filas = %d, esperado 2 (1 resuelta + 1 abierta, 0 duplicados)", total)
	}
}

// ---- T7c: exclusiones (FR-008 + caso límite) --------------------------------

func TestBalanceoT7cExclusiones(t *testing.T) {
	d := depsTest(t)
	prepararBalanceo(t, d)

	// Estación fuera_de_servicio en extremo (5 %) → excluida, sin alertas.
	fds := crearEstacionBal(t, d, "B-FDS", 20)
	poblar(t, d, fds, "F", 1, 19, "disponible")
	insert(t, d, `UPDATE estaciones SET estado = 'fuera_de_servicio' WHERE id = $1`, fds)
	evaluar(t, d, fds)
	if n := contarAlerta(t, d, fds, false); n != 0 {
		t.Errorf("fuera_de_servicio con 5 %% → %d alertas, esperado 0 (excluida)", n)
	}

	// Solo bicis averiadas ancladas → no cuentan → ocupación 0 % → alerta vacía.
	averiadas := crearEstacionBal(t, d, "B-Averiadas", 20)
	poblar(t, d, averiadas, "A", 1, 19, "averiada")
	evaluar(t, d, averiadas)
	if n := contarAlerta(t, d, averiadas, false); n != 1 {
		t.Errorf("estación con solo averiadas → %d alertas, esperado 1 (ocupación 0 %%)", n)
	}

	// Bici disponible SIN anclaje → no cuenta en ninguna estación:
	// la base queda exactamente en 10,00 % aunque existan bicis sueltas.
	insert(t, d,
		`INSERT INTO bicicletas (codigo, estado) VALUES ('SUELTA-1', 'disponible')`)
	suelta := crearEstacionBal(t, d, "B-10,00-suelta", 20)
	poblar(t, d, suelta, "S", 2, 18, "disponible")
	evaluar(t, d, suelta)
	if n := contarAlerta(t, d, suelta, false); n != 0 {
		t.Errorf("10 %% con bici suelta → %d alertas, esperado 0 (sin anclaje no cuenta)", n)
	}
}

// ---- T7d: panel (FR-005, SC-002) --------------------------------------------

func TestBalanceoT7dPanel(t *testing.T) {
	d := depsTest(t)
	bal := prepararBalanceo(t, d)
	app := nuevaApp(d)

	// Red todavía no evaluada → items: [] (no es error).
	r, panel := getPanel(app, bal.tokenLog)
	if r.status != http.StatusOK {
		t.Fatalf("GET panel status = %d, esperado 200 (body: %s)", r.status, r.raw)
	}
	if len(panel.Items) != 0 || !strings.Contains(r.raw, `"items":[]`) {
		t.Fatalf("red equilibrada → items [] esperado, got %s", r.raw)
	}

	// Cuatro estaciones en extremos con urgencias 10, 10, 1 y 1.
	e0 := crearEstacionBal(t, d, "B-0", 20) // 0 %    → score 10
	poblar(t, d, e0, "P0", 0, 20, "disponible")
	e9 := crearEstacionBal(t, d, "B-9", 100) // 9 %    → score 1
	poblar(t, d, e9, "P9", 9, 91, "disponible")
	e91 := crearEstacionBal(t, d, "B-91", 100) // 91 %   → score 1
	poblar(t, d, e91, "P91", 91, 9, "disponible")
	e100 := crearEstacionBal(t, d, "B-100", 20) // 100 % → score 10
	poblar(t, d, e100, "P100", 20, 0, "disponible")

	for _, id := range []int64{e0, e9, e91, e100} {
		evaluar(t, d, id)
	}
	// Empates de score → fecha_generacion ASC: fijamos fechas conocidas.
	insert(t, d, `UPDATE alertas_balanceo SET fecha_generacion = now() - interval '4 minutes' WHERE estacion_id = $1`, e0)
	insert(t, d, `UPDATE alertas_balanceo SET fecha_generacion = now() - interval '3 minutes' WHERE estacion_id = $1`, e100)
	insert(t, d, `UPDATE alertas_balanceo SET fecha_generacion = now() - interval '2 minutes' WHERE estacion_id = $1`, e9)
	insert(t, d, `UPDATE alertas_balanceo SET fecha_generacion = now() - interval '1 minute'  WHERE estacion_id = $1`, e91)

	r, panel = getPanel(app, bal.tokenLog)
	if r.status != http.StatusOK {
		t.Fatalf("GET panel status = %d, esperado 200", r.status)
	}
	if len(panel.Items) != 4 {
		t.Fatalf("items = %d, esperado 4 (solo abiertas)", len(panel.Items))
	}
	var tieneIndiceUrgencia bool
	if err := d.Pool.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM pg_indexes
		    WHERE schemaname = current_schema()
		      AND indexname = 'ix_alertas_balanceo_urgencia'
		 )`).Scan(&tieneIndiceUrgencia); err != nil {
		t.Fatalf("consultar índice de urgencia: %v", err)
	}
	if !tieneIndiceUrgencia {
		t.Fatal("falta el índice de expresión de urgencia de T022")
	}
	esperado := []int64{e0, e100, e9, e91} // urgencia DESC, fecha ASC
	for i, id := range esperado {
		if panel.Items[i].EstacionID != id {
			t.Errorf("items[%d].estacion_id = %d, esperado %d (orden por urgencia)",
				i, panel.Items[i].EstacionID, id)
		}
		if panel.Items[i].EstacionNombre == "" {
			t.Errorf("items[%d] sin estacion_nombre (JOIN)", i)
		}
		if panel.Items[i].Resuelta {
			t.Errorf("items[%d] resuelta aparece en el panel (SC-004)", i)
		}
	}
	// Porcentajes en el orden de urgencia: 0, 100 (score 10) y 9, 91 (score 1).
	esperadoPct := []float64{0, 100, 9, 91}
	for i, pct := range esperadoPct {
		if panel.Items[i].PorcentajeOcupacion != pct {
			t.Errorf("items[%d].porcentaje_ocupacion = %v, esperado %v", i, panel.Items[i].PorcentajeOcupacion, pct)
		}
	}
}

// ---- T7e: roles (const. I) --------------------------------------------------

func TestBalanceoT7eRoles(t *testing.T) {
	d := depsTest(t)
	bal := prepararBalanceo(t, d)
	app := nuevaApp(d)

	body := fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":1}`,
		bal.llena, bal.vacia)

	// Sin token → 401 en ambos endpoints.
	r, _ := getPanel(app, "")
	if r.status != http.StatusUnauthorized || errorCodigo(r) != "NO_AUTENTICADO" {
		t.Errorf("GET sin token → status %d codigo %q, esperado 401 NO_AUTENTICADO", r.status, errorCodigo(r))
	}
	r = postTraslado(app, "", body)
	if r.status != http.StatusUnauthorized || errorCodigo(r) != "NO_AUTENTICADO" {
		t.Errorf("POST sin token → status %d codigo %q, esperado 401 NO_AUTENTICADO", r.status, errorCodigo(r))
	}

	// Otros roles → 403 en ambos endpoints.
	otros := map[string]string{
		"usuario":         bal.tokenUsuario,
		"gestor":          bal.tokenGestor,
		"operario_taller": bal.tokenTaller,
	}
	for rol, token := range otros {
		r, _ := getPanel(app, token)
		if r.status != http.StatusForbidden || errorCodigo(r) != "ROL_NO_PERMITIDO" {
			t.Errorf("GET rol %s → status %d codigo %q, esperado 403 ROL_NO_PERMITIDO", rol, r.status, errorCodigo(r))
		}
		r = postTraslado(app, token, body)
		if r.status != http.StatusForbidden || errorCodigo(r) != "ROL_NO_PERMITIDO" {
			t.Errorf("POST rol %s → status %d codigo %q, esperado 403 ROL_NO_PERMITIDO", rol, r.status, errorCodigo(r))
		}
	}
}

// ---- T7f: traslado válido (FR-006/FR-007, SC-003) ---------------------------

func TestBalanceoT7fTrasladoValido(t *testing.T) {
	d := depsTest(t)
	bal := prepararBalanceo(t, d)
	app := nuevaApp(d)

	// Origen LLENA (95 %) con alerta abierta pre-evaluada; destino al 85 %
	// (sin alerta): con 3 unidades libres, recibe 20/20 = 100 % → alerta inline.
	destino := crearEstacionBal(t, d, "B-Destino", 20)
	poblar(t, d, destino, "D", 17, 3, "disponible")
	evaluar(t, d, bal.llena)
	evaluar(t, d, destino)
	if n := contarAlerta(t, d, destino, false); n != 0 {
		t.Fatalf("destino al 85 %% no debe tener alertas, got %d", n)
	}

	body := fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":3}`,
		bal.llena, destino)
	r := postTraslado(app, bal.tokenLog, body)
	if r.status != http.StatusOK {
		t.Fatalf("traslado → status %d, esperado 200 (body: %s)", r.status, r.raw)
	}
	var auditOrigen, auditDestino, auditUnidades, auditOperario int64
	var auditFecha time.Time
	if err := d.Pool.QueryRow(ctx,
		`SELECT estacion_origen_id, estacion_destino_id, unidades, operario_id, fecha
		   FROM traslados_balanceo
		  ORDER BY id DESC LIMIT 1`).Scan(
		&auditOrigen, &auditDestino, &auditUnidades, &auditOperario, &auditFecha); err != nil {
		t.Fatalf("leer auditoría de traslado: %v", err)
	}
	if auditOrigen != bal.llena || auditDestino != destino || auditUnidades != 3 ||
		auditOperario != bal.operarioID || auditFecha.IsZero() {
		t.Errorf("auditoría incorrecta: origen=%d destino=%d unidades=%d operario=%d fecha=%v",
			auditOrigen, auditDestino, auditUnidades, auditOperario, auditFecha)
	}

	origenEstado, _ := r.body["origen"].(map[string]any)
	destinoEstado, _ := r.body["destino"].(map[string]any)
	if origenEstado == nil || destinoEstado == nil {
		t.Fatalf("respuesta sin origen/destino: %s", r.raw)
	}
	if v := origenEstado["bicis_disponibles"].(float64); v != 16 {
		t.Errorf("origen.bicis_disponibles = %v, esperado 16 (19-3)", v)
	}
	if v := origenEstado["anclajes_libres"].(float64); v != 4 {
		t.Errorf("origen.anclajes_libres = %v, esperado 4", v)
	}
	if v := origenEstado["ocupacion_pct"].(float64); v != 80 {
		t.Errorf("origen.ocupacion_pct = %v, esperado 80", v)
	}
	if v := destinoEstado["bicis_disponibles"].(float64); v != 20 {
		t.Errorf("destino.bicis_disponibles = %v, esperado 20 (17+3)", v)
	}
	if v := destinoEstado["anclajes_libres"].(float64); v != 0 {
		t.Errorf("destino.anclajes_libres = %v, esperado 0", v)
	}
	if v := destinoEstado["ocupacion_pct"].(float64); v != 100 {
		t.Errorf("destino.ocupacion_pct = %v, esperado 100", v)
	}

	// Redis post-commit de ambas (FR-007 / constitución IV).
	for _, par := range []struct {
		id            int64
		bicis, libres string
	}{{bal.llena, "16", "4"}, {destino, "20", "0"}} {
		clave := cache.ClaveEstado(par.id)
		if v, err := d.RDB.HGet(ctx, clave, "bicis_disponibles").Result(); err != nil || v != par.bicis {
			t.Errorf("Redis %s bicis_disponibles = %q (err %v), esperado %s", clave, v, err, par.bicis)
		}
		if v, err := d.RDB.HGet(ctx, clave, "anclajes_libres").Result(); err != nil || v != par.libres {
			t.Errorf("Redis %s anclajes_libres = %q (err %v), esperado %s", clave, v, err, par.libres)
		}
		if v, err := d.RDB.HGet(ctx, clave, "actualizado_en").Result(); err != nil || v == "" {
			t.Errorf("Redis %s sin actualizado_en: %v", clave, err)
		}
	}

	// Evaluación inline: la alerta del origen (95 % → 80 %) queda resuelta…
	if n := contarAlerta(t, d, bal.llena, false); n != 0 {
		t.Errorf("alerta abierta del origen tras volver a rango = %d, esperado 0 (inline)", n)
	}
	if n := contarAlerta(t, d, bal.llena, true); n != 1 {
		t.Errorf("alerta del origen debe quedar resuelta = %d, esperado 1", n)
	}
	// …y el destino que cruza > 90 % genera la suya en la siguiente lectura.
	_, panel := getPanel(app, bal.tokenLog)
	encontrada := false
	for _, it := range panel.Items {
		if it.EstacionID == destino && it.TipoAlerta == "llena_mas_90" && !it.Resuelta {
			encontrada = true
		}
	}
	if !encontrada {
		t.Errorf("destino al 100 %% debe aparecer en el panel con llena_mas_90; items=%+v", panel.Items)
	}
}

// ---- T7g: traslado inconsistente (caso límite) ------------------------------

func TestBalanceoT7gTrasladoInconsistente(t *testing.T) {
	d := depsTest(t)
	bal := prepararBalanceo(t, d)
	app := nuevaApp(d)

	// Escenario adicional: destino SIN huecos y origen fuera de servicio.
	sinHuecos := crearEstacionBal(t, d, "B-SinHuecos", 4)
	poblar(t, d, sinHuecos, "SH", 4, 0, "disponible")
	fds := crearEstacionBal(t, d, "B-FDS-Origen", 20)
	poblar(t, d, fds, "FO", 5, 15, "disponible")
	insert(t, d, `UPDATE estaciones SET estado = 'fuera_de_servicio' WHERE id = $1`, fds)

	anteriores := func() string {
		a, b := recuento(t, d, bal.llena), recuento(t, d, bal.vacia)
		return fmt.Sprintf("%+v/%+v", a, b)
	}
	iguales := func(antes, ahora string) bool { return antes == ahora }

	casos := []struct {
		nombre string
		body   string
		status int
		codigo string
	}{
		{"unidades > disponibles en origen",
			fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":999}`, bal.llena, bal.vacia),
			409, "ESTACION_SIN_CAPACIDAD"},
		{"destino sin anclajes libres",
			fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":1}`, bal.llena, sinHuecos),
			409, "ESTACION_SIN_CAPACIDAD"},
		{"estación inexistente",
			fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":99999,"unidades":1}`, bal.llena),
			409, "ESTACION_SIN_CAPACIDAD"},
		{"origen fuera de servicio",
			fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":1}`, fds, bal.vacia),
			409, "ESTACION_SIN_CAPACIDAD"},
		{"unidades < 1",
			fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":0}`, bal.llena, bal.vacia),
			422, "ENTRADA_INVALIDA"},
		{"origen == destino",
			fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":3}`, bal.vacia, bal.vacia),
			422, "ENTRADA_INVALIDA"},
		{"JSON malformado",
			`{"estacion_origen_id":`,
			422, "ENTRADA_INVALIDA"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			antes := anteriores()
			var auditoriasAntes int
			if err := d.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM traslados_balanceo`).Scan(&auditoriasAntes); err != nil {
				t.Fatalf("contar auditorías antes: %v", err)
			}
			r := postTraslado(app, bal.tokenLog, c.body)
			if r.status != c.status {
				t.Errorf("status = %d, esperado %d (body: %s)", r.status, c.status, r.raw)
			}
			if cod := errorCodigo(r); cod != c.codigo {
				t.Errorf("codigo = %q, esperado %q", cod, c.codigo)
			}
			if ahora := anteriores(); !iguales(antes, ahora) {
				t.Errorf("recuentos alterados: antes %s, después %s", antes, ahora)
			}
			var auditoriasDespues int
			if err := d.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM traslados_balanceo`).Scan(&auditoriasDespues); err != nil {
				t.Fatalf("contar auditorías después: %v", err)
			}
			if auditoriasDespues != auditoriasAntes {
				t.Errorf("se registró un traslado rechazado: auditorías antes=%d después=%d",
					auditoriasAntes, auditoriasDespues)
			}
		})
	}
}

// ---- T7h: latencia del worker (SC-001) --------------------------------------

func TestBalanceoT7hLatenciaWorker(t *testing.T) {
	d := depsTest(t)
	prepararBalanceo(t, d)

	// Volumen de estaciones definido para el proyecto (~60, data-model.md §3).
	for i := 0; i < 58; i++ {
		e := crearEstacionBal(t, d, fmt.Sprintf("B-Masiva-%02d", i), 20)
		ocupados := i % 21 // 0..20 → variopinto
		poblar(t, d, e, fmt.Sprintf("M%02d", i), ocupados, 20-ocupados, "disponible")
	}

	// (i) Barrido completo cronometrado < 5 s (presupuesto verificable D6).
	dur, err := alertas.Barrido(ctx, d.Pool)
	if err != nil {
		t.Fatalf("Barrido: %v", err)
	}
	if dur >= 5*time.Second {
		t.Errorf("T_barrido = %v, esperado < 5 s (presupuesto de SC-001)", dur)
	}

	// (ii) Con periodo de test inyectado: mutación → alerta en < periodo + 5 s.
	periodoOriginal := alertas.PeriodoBarrido
	alertas.PeriodoBarrido = 300 * time.Millisecond
	wctx, cancel := context.WithCancel(ctx)
	go alertas.Run(wctx, d.Pool)
	t.Cleanup(func() {
		cancel()
		time.Sleep(100 * time.Millisecond)
		alertas.PeriodoBarrido = periodoOriginal
	})
	time.Sleep(200 * time.Millisecond) // arranque + barrido inmediato

	// Estación equilibrada exactamente en 10 % → mutación a 0 % (sin anclajes).
	objetivo := crearEstacionBal(t, d, "B-Mutada", 20)
	poblar(t, d, objetivo, "MU", 2, 18, "disponible")
	insert(t, d,
		`UPDATE anclajes SET estado = 'libre', bicicleta_id = NULL
		  WHERE estacion_id = $1 AND estado = 'ocupado'`, objetivo)

	limite := time.Now().Add(alertas.PeriodoBarrido + 5*time.Second)
	for {
		n := 0
		if err := d.Pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM alertas_balanceo WHERE estacion_id = $1 AND NOT resuelta`,
			objetivo).Scan(&n); err != nil {
			t.Fatalf("poll alerta: %v", err)
		}
		if n == 1 {
			break
		}
		if time.Now().After(limite) {
			t.Fatalf("mutación → alerta no generada en < periodo (%v) + 5 s", alertas.PeriodoBarrido)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
