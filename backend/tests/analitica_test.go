// Tests de integración de la HU-08 (analítica y exportación CSV).
// Ejecutan contra la BD del docker-compose (ver quickstart.md).
// Escenarios T8a–T8i; fixtures propias (patrón historial_test.go).
package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
)

type respuestaAnalitica struct {
	Desde        string  `json:"desde"`
	Hasta        string  `json:"hasta"`
	TotalViajes  int     `json:"total_viajes"`
	TotalCoste   float64 `json:"total_coste"`
	PorHora      []int   `json:"por_hora"`
	PorDiaSemana []int   `json:"por_dia_semana"`
}

// escenarioAnalitica: gestor + usuarios de otros roles; viajes finalizados
// con horas y días conocidos en la semana del 2026-10-12 (lunes) al
// 2026-10-18 (domingo), todos convertidos a Europe/Madrid:
//
//	lunes 12-oct 09:00 local (07:00 UTC), 45 min, 3.00
//	jueves 15-oct 23:30 local (21:30 UTC), 20 min, 0.00  — borde por_hasta
//	sábado 17-oct 12:00 local (10:00 UTC), 60 min, 6.00
//	domingo 18-oct 00:30 local (17-oct 22:30 UTC), 15 min, 0.00 — borde por_desde
//	+ 1 viaje en_curso (excluido, T8d)
type escenarioAnalitica struct {
	tokenGestor  string
	tokenUsuario string
	tokenTaller  string
	tokenLog     string
}

func crearTokenRol(t *testing.T, d deps.Deps, email, rol string) string {
	t.Helper()
	var id int64
	err := d.Pool.QueryRow(ctx,
		`INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo)
		 VALUES ('Tester', $1, $2, $3, false) RETURNING id`,
		email, nuevoHash(t), rol).Scan(&id)
	if err != nil {
		t.Fatalf("crear usuario %s: %v", rol, err)
	}
	token, err := auth.CrearToken(d.JWTSecret, id, "Tester", rol)
	if err != nil {
		t.Fatalf("token %s: %v", rol, err)
	}
	return token
}

func prepararAnalitica(t *testing.T, d deps.Deps) escenarioAnalitica {
	t.Helper()
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var eOrigen, eDestino int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		 VALUES ('Estació Universitat', 39.47, -0.38, 5), ('Estació Mercat', 39.48, -0.37, 5)
		 RETURNING id`).Scan(&eOrigen); err != nil {
		t.Fatalf("estación: %v", err)
	}
	if err := d.Pool.QueryRow(ctx,
		`SELECT id FROM estaciones WHERE nombre = 'Estació Mercat'`).Scan(&eDestino); err != nil {
		t.Fatalf("estación destino: %v", err)
	}

	esc := escenarioAnalitica{
		tokenGestor:  crearTokenRol(t, d, "gestor-anal@test.local", "gestor"),
		tokenUsuario: crearTokenRol(t, d, "usuario-anal@test.local", "usuario"),
		tokenTaller:  crearTokenRol(t, d, "taller-anal@test.local", "operario_taller"),
		tokenLog:     crearTokenRol(t, d, "log-anal@test.local", "operario_logistica"),
	}

	var anaID int64
	if err := d.Pool.QueryRow(ctx,
		`SELECT id FROM usuarios WHERE email = 'gestor-anal@test.local'`).Scan(&anaID); err != nil {
		t.Fatalf("id usuario: %v", err)
	}
	var biciID int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO bicicletas (codigo, estado) VALUES ('T-A-1', 'disponible') RETURNING id`).Scan(&biciID); err != nil {
		t.Fatalf("bici: %v", err)
	}

	insertFinalizado := func(inicio time.Time, minutos int, coste string) {
		t.Helper()
		fin := inicio.Add(time.Duration(minutos) * time.Minute)
		insert(t, d,
			`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id,
			                     estacion_destino_id, fecha_inicio, fecha_fin,
			                     duracion_minutos, coste_total, estado)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::numeric, 'finalizado')`,
			anaID, biciID, eOrigen, eDestino, inicio, fin, minutos, coste)
	}

	// Octubre 2026: Madrid en CEST (+2) hasta el día 25.
	insertFinalizado(time.Date(2026, 10, 12, 7, 0, 0, 0, time.UTC), 45, "3.00")   // lun 09:00 local
	insertFinalizado(time.Date(2026, 10, 15, 21, 30, 0, 0, time.UTC), 20, "0.00") // jue 23:30 local
	insertFinalizado(time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC), 60, "6.00")  // sáb 12:00 local
	insertFinalizado(time.Date(2026, 10, 17, 22, 30, 0, 0, time.UTC), 15, "0.00") // dom 00:30 local

	// Viaje ACTIVO: jamás debe contarse (T8d).
	insert(t, d,
		`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id, fecha_inicio)
		 VALUES ($1, $2, $3, now() - interval '1 hour')`, anaID, biciID, eOrigen)

	return esc
}

func getAnalitica(t *testing.T, app *gin.Engine, token, query string) (*httptest.ResponseRecorder, respuestaAnalitica) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analitica/viajes"+query, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	var r respuestaAnalitica
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
			t.Fatalf("unmarshal analitica (%s): %v", w.Body.String(), err)
		}
	}
	return w, r
}

func getExport(t *testing.T, app *gin.Engine, token, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analitica/export"+query, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w
}

// filasCSV devuelve las líneas de datos del body (sin BOM ni cabecera),
// separadas en campos por coma (RFC 4180 simple: los fixtures no usan
// comas dentro de campos — el quoting se prueba en el unitario).
func filasCSV(t *testing.T, body string) [][]string {
	t.Helper()
	if !strings.HasPrefix(body, "\ufeff") {
		t.Errorf("falta el BOM UTF-8 (primeros bytes %q)", body[:min(3, len(body))])
	}
	cuerpo := strings.TrimPrefix(body, "\ufeff")
	lineas := strings.Split(strings.TrimRight(cuerpo, "\r\n"), "\r\n")
	returned := make([][]string, 0, len(lineas))
	for i, l := range lineas {
		if i == 0 {
			continue // cabecera
		}
		returned = append(returned, strings.Split(l, ","))
	}
	return returned
}

const semanaAnalitica = "?desde=2026-10-12&hasta=2026-10-18"

// T8a: consistencia gráficos ↔ CSV del mismo rango (FR-006, SC-002).
func TestAnaliticaT8aConsistencia(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	_, r := getAnalitica(t, app, esc.tokenGestor, semanaAnalitica)
	if r.TotalViajes != 4 {
		t.Fatalf("total_viajes = %d, esperaba 4", r.TotalViajes)
	}
	if diff := r.TotalCoste - 9.00; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("total_coste = %v, esperaba 9.00", r.TotalCoste)
	}

	w := getExport(t, app, esc.tokenGestor, semanaAnalitica)
	if w.Code != http.StatusOK {
		t.Fatalf("export status = %d", w.Code)
	}
	filas := filasCSV(t, w.Body.String())
	if len(filas) != r.TotalViajes {
		t.Errorf("filas CSV = %d, total_viajes = %d", len(filas), r.TotalViajes)
	}
	var suma float64
	for _, f := range filas {
		var c float64
		if _, err := fmt.Sscanf(f[len(f)-1], "%f", &c); err != nil {
			t.Fatalf("parse coste %q: %v", f[len(f)-1], err)
		}
		suma += c
	}
	if diff := suma - r.TotalCoste; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("suma CSV = %v, total_coste = %v", suma, r.TotalCoste)
	}
	if r.PorHora[9] != 1 || r.PorHora[23] != 1 || r.PorHora[12] != 1 || r.PorHora[0] != 1 {
		t.Errorf("por_hora = %v, esperaba 1 en índices 0/9/12/23", r.PorHora)
	}
}

// T8b: bordes inclusivos del rango (Europe/Madrid).
func TestAnaliticaT8bBordesInclusivos(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	// Viaje del jueves 15 a las 23:30 local: cuenta con hasta=15…
	_, r := getAnalitica(t, app, esc.tokenGestor, "?desde=2026-10-01&hasta=2026-10-15")
	if r.TotalViajes != 2 { // día 12 (09:00) + día 15 (23:30)
		t.Errorf("hasta=15: total = %d, esperaba 2 (incluye viaje 23:30 local)", r.TotalViajes)
	}
	// …y NO cuenta con hasta=14.
	_, r = getAnalitica(t, app, esc.tokenGestor, "?desde=2026-10-01&hasta=2026-10-14")
	if r.TotalViajes != 1 { // solo el día 12; el 23:30 del 15 queda fuera
		t.Errorf("hasta=14: total = %d, esperaba 1 (viaje 23:30 del 15 excluido)", r.TotalViajes)
	}
	// Viaje del domingo 18 a las 00:30 local: cuenta con desde=18…
	_, r = getAnalitica(t, app, esc.tokenGestor, "?desde=2026-10-18&hasta=2026-10-31")
	if r.TotalViajes != 1 {
		t.Errorf("desde=18: total = %d, esperaba 1 (viaje 00:30 local)", r.TotalViajes)
	}
	// …y NO cuenta con desde=19.
	_, r = getAnalitica(t, app, esc.tokenGestor, "?desde=2026-10-19&hasta=2026-10-31")
	if r.TotalViajes != 0 {
		t.Errorf("desde=19: total = %d, esperaba 0", r.TotalViajes)
	}
}

// T8c: agrupación por hora local y día de la semana (lunes=0…domingo=6).
func TestAnaliticaT8cAgrupacion(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	_, r := getAnalitica(t, app, esc.tokenGestor, semanaAnalitica)
	if len(r.PorHora) != 24 {
		t.Fatalf("por_hora tiene %d posiciones, esperaba 24", len(r.PorHora))
	}
	if len(r.PorDiaSemana) != 7 {
		t.Fatalf("por_dia_semana tiene %d posiciones, esperaba 7", len(r.PorDiaSemana))
	}
	// Días: lunes(12)=0, jueves(15)=3, sábado(17)=5, domingo(18)=6.
	for i, v := range r.PorDiaSemana {
		esperado := 0
		switch i {
		case 0, 3, 5, 6:
			esperado = 1
		}
		if v != esperado {
			t.Errorf("por_dia_semana[%d] = %d, esperaba %d", i, v, esperado)
		}
	}
	// total_viajes == suma de por_hora (por construcción).
	suma := 0
	for _, v := range r.PorHora {
		suma += v
	}
	if suma != r.TotalViajes {
		t.Errorf("suma por_hora = %d, total_viajes = %d", suma, r.TotalViajes)
	}
}

// T8d: el viaje en_curso no aparece en barras ni en CSV.
func TestAnaliticaT8dExcluyeEnCurso(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	// Rango que cubre "ahora" (el viaje activo empezó hace 1 h).
	hoy := time.Now().UTC().Format("2006-01-02")
	_, r := getAnalitica(t, app, esc.tokenGestor, "?desde="+hoy+"&hasta="+hoy)
	if r.TotalViajes != 0 {
		t.Errorf("viaje en_curso contado: total = %d", r.TotalViajes)
	}
	w := getExport(t, app, esc.tokenGestor, "?desde="+hoy+"&hasta="+hoy)
	if w.Code != http.StatusOK {
		t.Fatalf("export status = %d", w.Code)
	}
	if filas := filasCSV(t, w.Body.String()); len(filas) != 0 {
		t.Errorf("viaje en_curso en CSV: %d filas", len(filas))
	}
}

// T8e: rango vacío y rango futuro → 200 con ceros y CSV solo cabecera.
func TestAnaliticaT8eRangosSinDatos(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	for _, q := range []string{
		"?desde=2099-01-01&hasta=2099-01-31",
		"?desde=2099-06-01&hasta=2099-06-30",
	} {
		w, r := getAnalitica(t, app, esc.tokenGestor, q)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status = %d", q, w.Code)
		}
		if r.TotalViajes != 0 || r.TotalCoste != 0 {
			t.Errorf("%s: total_viajes=%d total_coste=%v, esperaba 0/0", q, r.TotalViajes, r.TotalCoste)
		}
		for i, v := range r.PorHora {
			if v != 0 {
				t.Errorf("%s: por_hora[%d] = %d", q, i, v)
			}
		}
		wExp := getExport(t, app, esc.tokenGestor, q)
		if wExp.Code != http.StatusOK {
			t.Fatalf("%s export: status = %d", q, wExp.Code)
		}
		if filas := filasCSV(t, wExp.Body.String()); len(filas) != 0 {
			t.Errorf("%s: CSV con %d filas, esperaba solo cabecera", q, len(filas))
		}
	}
}

// T8f: 422 ENTRADA_INVALIDA en ambos endpoints.
func TestAnaliticaT8fEntradaInvalida(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	consultas := []string{
		"",                                   // ambos ausentes
		"?desde=2026-10-01",                  // hasta ausente
		"?hasta=2026-10-31",                  // desde ausente
		"?desde=01/10/2026&hasta=2026-10-31", // formato no estricto
		"?desde=2026-13-40&hasta=2026-10-31", // día/mes inexistentes
		"?desde=2026-10-31&hasta=2026-10-01", // invertido
		"?desde=not-a-date&hasta=2026-10-31",
	}
	for _, q := range consultas {
		w, _ := getAnalitica(t, app, esc.tokenGestor, q)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("agregados %q: status = %d, esperaba 422 (body %s)", q, w.Code, w.Body.String())
		}
		wExp := getExport(t, app, esc.tokenGestor, q)
		if wExp.Code != http.StatusUnprocessableEntity {
			t.Errorf("export %q: status = %d, esperaba 422", q, wExp.Code)
		}
		// Sin cabeceras de descarga en error.
		if cd := wExp.Header().Get("Content-Disposition"); cd != "" {
			t.Errorf("export %q: Content-Disposition en error: %q", q, cd)
		}
	}
}

// T8g: roles — 401 sin token, 403 otros roles, 200 gestor.
func TestAnaliticaT8gRoles(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	for _, ep := range []string{"/api/v1/analitica/viajes" + semanaAnalitica,
		"/api/v1/analitica/export" + semanaAnalitica} {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s sin token: status = %d, esperaba 401", ep, w.Code)
		}
		for _, tok := range []string{esc.tokenUsuario, esc.tokenTaller, esc.tokenLog} {
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			req.Header.Set("Authorization", "Bearer "+tok)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Errorf("%s rol no gestor: status = %d, esperaba 403", ep, w.Code)
			}
		}
		req = httptest.NewRequest(http.MethodGet, ep, nil)
		req.Header.Set("Authorization", "Bearer "+esc.tokenGestor)
		w = httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%s gestor: status = %d, esperaba 200 (body %s)", ep, w.Code, w.Body.String())
		}
	}
}

// T8h: formato completo del CSV (BOM, columnas, CRLF, orden, nombre, tildes).
func TestAnaliticaT8hFormatoCSV(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	w := getExport(t, app, esc.tokenGestor, semanaAnalitica)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q, esperaba text/csv", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "filename=viajes_20261012_20261018.csv") {
		t.Errorf("Content-Disposition = %q, esperaba filename=viajes_20261012_20261018.csv", cd)
	}
	b := w.Body.Bytes()
	if len(b) < 3 || b[0] != 0xEF || b[1] != 0xBB || b[2] != 0xBF {
		t.Errorf("primeros bytes = % X, esperaba BOM EF BB BF", b[:min(3, len(b))])
	}
	cuerpo := strings.TrimPrefix(w.Body.String(), "\ufeff")
	if !strings.HasPrefix(cuerpo, "id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur\r\n") {
		t.Errorf("cabecera incorrecta: %q", cuerpo[:min(100, len(cuerpo))])
	}
	// CRLF: todo \n va precedido de \r (y hay al menos un CRLF).
	if !strings.Contains(cuerpo, "\r\n") {
		t.Error("el cuerpo no contiene finales de línea CRLF")
	}
	for i := 1; i < len(cuerpo); i++ {
		if cuerpo[i] == '\n' && cuerpo[i-1] != '\r' {
			t.Errorf("salto de línea sin CR en la posición %d", i)
			break
		}
	}
	// Orden fecha_inicio ASC.
	filas := filasCSV(t, w.Body.String())
	var fechaAnterior string
	for _, f := range filas {
		fecha := f[1]
		if fechaAnterior != "" && fecha < fechaAnterior {
			t.Errorf("orden no ASC: %s < %s", fecha, fechaAnterior)
		}
		fechaAnterior = fecha
	}
	// Tildes intactas y coste con punto decimal.
	cuerpo2 := w.Body.String()
	if !strings.Contains(cuerpo2, "Estació Universitat") || !strings.Contains(cuerpo2, "Estació Mercat") {
		t.Error("tildes dañadas en los nombres de estación")
	}
	if strings.Contains(cuerpo2, ",3,00") || !strings.Contains(cuerpo2, ",3.00") {
		t.Error("coste_eur sin punto decimal")
	}
	// fecha_hora en formato AAAA-MM-DD HH:MM Europe/Madrid.
	if !strings.Contains(cuerpo2, "2026-10-12 09:00") {
		t.Errorf("fecha_hora local incorrecta (esperaba 2026-10-12 09:00):\n%s", cuerpo2)
	}
}

// T8i: latencia de agregados con rango de un año < 5 s (SC-001).
func TestAnaliticaT8iLatencia(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararAnalitica(t, d)

	inicio := time.Now()
	w, _ := getAnalitica(t, app, esc.tokenGestor, "?desde=2026-01-01&hasta=2026-12-31")
	elapsed := time.Since(inicio)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if elapsed > 5*time.Second {
		t.Errorf("agregados tardaron %v, objetivo < 5 s (SC-001)", elapsed)
	}
}
