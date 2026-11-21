// Tests de integración de la HU-04 (historial de viajes y desglose mensual).
// Ejecutan contra la BD y Redis del docker-compose (ver quickstart.md).
// Deben FALLAR antes de implementar el servicio/handler/ruta (TDD).
// Patrón: tests/devolucion_test.go (httptest + fixtures propias).
package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
)

// itemHistorial es el item del contrato (schema Viaje + nombres aditivos).
type itemHistorial struct {
	ID                    int64    `json:"id"`
	UsuarioID             int64    `json:"usuario_id"`
	EstacionOrigenID      int64    `json:"estacion_origen_id"`
	EstacionDestinoID     *int64   `json:"estacion_destino_id"`
	FechaInicio           string   `json:"fecha_inicio"`
	FechaFin              *string  `json:"fecha_fin"`
	DuracionMinutos       *int     `json:"duracion_minutos"`
	CosteTotal            float64  `json:"coste_total"`
	Estado                string   `json:"estado"`
	Etiqueta              *string  `json:"etiqueta"`
	EstacionOrigenNombre  string   `json:"estacion_origen_nombre"`
	EstacionDestinoNombre *string  `json:"estacion_destino_nombre"`
}

type respuestaHistorial struct {
	Total     int             `json:"total"`
	SumaCoste float64         `json:"suma_coste"`
	Items     []itemHistorial `json:"items"`
}

// escenarioHistorial: ana con 5 viajes finalizados (3 en el mes actual,
// 1 en el anterior, 1 hace 2 meses; uno con coste 0,00) + 1 en_curso;
// carla con 1 viaje propio; tzuser con 2 viajes en el límite de mes.
type escenarioHistorial struct {
	tokenAna    string
	tokenCarla  string
	tokenTZ     string
	anaID       int64
	sumaAna     float64 // 0 + 2.50 + 1.25 + 3.00 + 4.00 = 10.75
	totalAna    int     // 5 (el en_curso no cuenta)
}

func prepararHistorial(t *testing.T, d deps.Deps) escenarioHistorial {
	t.Helper()
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var eOrigen, eDestino int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		 VALUES ('Origen H', 39.47, -0.38, 5), ('Destino H', 39.48, -0.37, 5)
		 RETURNING id`).Scan(&eOrigen); err != nil {
		t.Fatalf("estación origen: %v", err)
	}
	if err := d.Pool.QueryRow(ctx,
		`SELECT id FROM estaciones WHERE nombre = 'Destino H'`).Scan(&eDestino); err != nil {
		t.Fatalf("estación destino: %v", err)
	}

	anaID, tokenAna := crearUsuario(t, d, "ana-hist@test.local", false)
	_, tokenCarla := crearUsuario(t, d, "carla-hist@test.local", false)
	_, tokenTZ := crearUsuario(t, d, "tz-hist@test.local", false)

	// Bicicletas de los viajes (los finalizados pueden compartir).
	var biciHist, biciActiva int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO bicicletas (codigo, estado) VALUES ('T-H-1', 'disponible') RETURNING id`).Scan(&biciHist); err != nil {
		t.Fatalf("bici hist: %v", err)
	}
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO bicicletas (codigo, estado) VALUES ('T-H-2', 'en_uso') RETURNING id`).Scan(&biciActiva); err != nil {
		t.Fatalf("bici activa: %v", err)
	}

	// Viaje finalizado: día 15 mediodía UTC → día 15 tarde en Madrid
	// (sin bordes de mes ni de zona horaria).
	insertViajeFinalizado := func(usuario, bici, origen, destino int64, inicio time.Time, minutos int, coste string) {
		t.Helper()
		fin := inicio.Add(time.Duration(minutos) * time.Minute)
		insert(t, d,
			`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id,
			                     estacion_destino_id, fecha_inicio, fecha_fin,
			                     duracion_minutos, coste_total, estado)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::numeric, 'finalizado')`,
			usuario, bici, origen, destino, inicio, fin, minutos, coste)
	}

	ahora := time.Now()
	// Anclar al día 1 antes de restar meses: si hoy es día 31, restar un mes
	// normaliza (31-sept → 1-oct) y colapsaría viajes de meses distintos.
	diaMes := func(offsetMes int, dia int) time.Time {
		base := time.Date(ahora.Year(), ahora.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, offsetMes, 0)
		return time.Date(base.Year(), base.Month(), dia, 12, 0, 0, 0, time.UTC)
	}

	// Ana: 3 en el mes actual (0,00 + 2,50 + 1,25), 1 anterior, 1 anteactual.
	insertViajeFinalizado(anaID, biciHist, eOrigen, eDestino, diaMes(0, 15), 12, "0.00")
	insertViajeFinalizado(anaID, biciHist, eOrigen, eDestino, diaMes(0, 10), 25, "2.50")
	insertViajeFinalizado(anaID, biciHist, eOrigen, eDestino, diaMes(0, 5), 40, "1.25")
	insertViajeFinalizado(anaID, biciHist, eOrigen, eDestino, diaMes(-1, 15), 30, "3.00")
	insertViajeFinalizado(anaID, biciHist, eOrigen, eDestino, diaMes(-2, 15), 55, "4.00")

	// Viaje ACTIVO de Ana: jamás debe listarse (FR-005, caso límite).
	insert(t, d,
		`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id, fecha_inicio)
		 VALUES ($1, $2, $3, now() - interval '1 hour')`, anaID, biciActiva, eOrigen)

	// Carla: un viaje propio (aislamiento entre usuarios).
	var carlaID int64
	if err := d.Pool.QueryRow(ctx, `SELECT id FROM usuarios WHERE email = 'carla-hist@test.local'`).Scan(&carlaID); err != nil {
		t.Fatalf("id carla: %v", err)
	}
	insertViajeFinalizado(carlaID, biciHist, eOrigen, eDestino, diaMes(0, 15), 20, "9.99")

	// tzuser: dos viajes en el límite UTC→Madrid de octubre/noviembre 2026.
	// El 31-oct-2026 Madrid ya está en CET (+1, el DST terminó el 25-oct):
	// 21:30 UTC = 22:30 CET (31-oct) → octubre; 23:30 UTC = 00:30 CET
	// (1-nov) → noviembre. Ambas horas caen en el día esperado también
	// con CEST, el test es robusto al cambio de hora (FR-002, T4g).
	var tzID int64
	if err := d.Pool.QueryRow(ctx, `SELECT id FROM usuarios WHERE email = 'tz-hist@test.local'`).Scan(&tzID); err != nil {
		t.Fatalf("id tz: %v", err)
	}
	insertViajeFinalizado(tzID, biciHist, eOrigen, eDestino, time.Date(2026, 10, 31, 21, 30, 0, 0, time.UTC), 10, "1.00")
	insertViajeFinalizado(tzID, biciHist, eOrigen, eDestino, time.Date(2026, 10, 31, 23, 30, 0, 0, time.UTC), 10, "2.00")

	return escenarioHistorial{
		tokenAna:   tokenAna,
		tokenCarla: tokenCarla,
		tokenTZ:    tokenTZ,
		anaID:      anaID,
		sumaAna:    10.75,
		totalAna:   5,
	}
}

func getHistorial(t *testing.T, app *gin.Engine, token, query string) (*httptest.ResponseRecorder, respuestaHistorial) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/viajes"+query, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var r respuestaHistorial
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
			t.Fatalf("unmarshal response (%s): %v", w.Body.String(), err)
		}
	}
	return w, r
}

// codigoError extrae error.codigo del envelope global (contrato).
func codigoError(body []byte) string {
	var e struct {
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &e)
	return e.Error.Codigo
}

func rangoMesActual(t *testing.T) (desde, hasta string) {
	t.Helper()
	ahora := time.Now()
	sig := ahora.AddDate(0, 1, 0)
	desde = time.Date(ahora.Year(), ahora.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	hasta = time.Date(sig.Year(), sig.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1).Format("2006-01-02")
	return desde, hasta
}

func TestHistorialT4a(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	w, r := getHistorial(t, app, esc.tokenAna, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperaba 200 (body %s)", w.Code, w.Body.String())
	}
	if r.Total != esc.totalAna {
		t.Errorf("total = %d, esperaba %d (el viaje en_curso no cuenta)", r.Total, esc.totalAna)
	}
	if len(r.Items) != esc.totalAna {
		t.Fatalf("items = %d, esperaba %d", len(r.Items), esc.totalAna)
	}

	// Orden fecha_inicio DESC estricto.
	for i := 1; i < len(r.Items); i++ {
		if !(r.Items[i-1].FechaInicio > r.Items[i].FechaInicio) {
			t.Errorf("orden no DESC en posiciones %d/%d: %s <= %s",
				i-1, i, r.Items[i-1].FechaInicio, r.Items[i].FechaInicio)
		}
	}

	// Campos y nombres de estación por JOIN (FR-001/FR-006).
	for _, it := range r.Items {
		if it.Estado != "finalizado" {
			t.Errorf("viaje %d con estado %q, solo se admiten finalizado", it.ID, it.Estado)
		}
		if it.UsuarioID != esc.anaID {
			t.Errorf("viaje %d pertenece a usuario %d, debía ser de %d", it.ID, it.UsuarioID, esc.anaID)
		}
		if it.EstacionOrigenNombre == "" {
			t.Errorf("viaje %d sin estacion_origen_nombre", it.ID)
		}
		if it.EstacionDestinoNombre == nil || *it.EstacionDestinoNombre == "" {
			t.Errorf("viaje %d sin estacion_destino_nombre", it.ID)
		}
		if it.DuracionMinutos == nil || it.FechaFin == nil {
			t.Errorf("viaje %d finalizado con fecha_fin/duracion nulos", it.ID)
		}
	}
}

func TestHistorialT4bFiltroMes(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)
	desde, hasta := rangoMesActual(t)

	w, r := getHistorial(t, app, esc.tokenAna, fmt.Sprintf("?desde=%s&hasta=%s", desde, hasta))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d (body %s)", w.Code, w.Body.String())
	}
	if r.Total != 3 {
		t.Errorf("total mes actual = %d, esperaba 3", r.Total)
	}
	// 0,00 + 2,50 + 1,25 = 3,75 (suma exacta del periodo, FR-003).
	if diff := r.SumaCoste - 3.75; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("suma_coste = %v, esperaba 3.75", r.SumaCoste)
	}
	// Solo viajes del mes: ninguno de los meses anteriores.
	for _, it := range r.Items {
		f, err := time.Parse(time.RFC3339, it.FechaInicio)
		if err != nil {
			t.Fatalf("parse fecha %s: %v", it.FechaInicio, err)
		}
		// El día 15 mediodía UTC cae siempre en el mes local esperado.
		if f.Day() < 5 {
			t.Errorf("viaje fuera del mes filtrado: %s", it.FechaInicio)
		}
	}
}

func TestHistorialT4cEtiqueta(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	_, r := getHistorial(t, app, esc.tokenAna, "")
	var conEtiqueta, sinEtiqueta int
	for _, it := range r.Items {
		if it.CosteTotal == 0 {
			if it.Etiqueta == nil || *it.Etiqueta != "Tramo gratuito (0,00 €)" {
				t.Errorf("viaje %d coste 0 sin etiqueta correcta: %v", it.ID, it.Etiqueta)
			} else {
				conEtiqueta++
			}
		} else if it.Etiqueta != nil {
			t.Errorf("viaje %d coste %v no debía llevar etiqueta: %q", it.ID, it.CosteTotal, *it.Etiqueta)
		} else {
			sinEtiqueta++
		}
	}
	if conEtiqueta != 1 || sinEtiqueta != 4 {
		t.Errorf("etiquetas: %d con, %d sin; esperaba 1/4", conEtiqueta, sinEtiqueta)
	}
}

func TestHistorialT4dAislamientoY401(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	// Carla no ve viajes de ana.
	_, r := getHistorial(t, app, esc.tokenCarla, "")
	if r.Total != 1 {
		t.Fatalf("carla total = %d, esperaba 1", r.Total)
	}
	if r.Items[0].UsuarioID == esc.anaID || r.SumaCoste != 9.99 {
		t.Errorf("carla ve datos ajenos: usuario %d suma %v", r.Items[0].UsuarioID, r.SumaCoste)
	}

	// Sin token → 401 NO_AUTENTICADO (contrato).
	w, _ := getHistorial(t, app, "", "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("sin token: status = %d, esperaba 401", w.Code)
	}
	if cod := codigoError(w.Body.Bytes()); cod != "NO_AUTENTICADO" {
		t.Errorf("sin token: codigo = %q, esperaba NO_AUTENTICADO (body %s)", cod, w.Body.String())
	}
}

func TestHistorialT4ePaginacion(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	w1, r1 := getHistorial(t, app, esc.tokenAna, "?page=1&page_size=1")
	w2, r2 := getHistorial(t, app, esc.tokenAna, "?page=2&page_size=1")
	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Fatalf("status página 1/2: %d/%d", w1.Code, w2.Code)
	}
	if len(r1.Items) != 1 || len(r2.Items) != 1 {
		t.Fatalf("page_size=1 devolvió %d/%d items", len(r1.Items), len(r2.Items))
	}
	if r1.Items[0].ID == r2.Items[0].ID {
		t.Error("página 1 y 2 devuelven el mismo viaje")
	}
	// El sumatorio es del periodo completo: idéntico en cualquier página.
	if r1.Total != esc.totalAna || r2.Total != esc.totalAna {
		t.Errorf("total distinto entre páginas: %d / %d", r1.Total, r2.Total)
	}
	if r1.SumaCoste != esc.sumaAna || r2.SumaCoste != esc.sumaAna {
		t.Errorf("suma_coste distinto entre páginas: %v / %v (esperaba %v)",
			r1.SumaCoste, r2.SumaCoste, esc.sumaAna)
	}
}

func TestHistorialT4fPaginaFueraDeRango(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	w, r := getHistorial(t, app, esc.tokenAna, "?page=999")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperaba 200 sin error", w.Code)
	}
	if len(r.Items) != 0 {
		t.Errorf("items = %d, esperaba []", len(r.Items))
	}
	if r.Total != esc.totalAna || r.SumaCoste != esc.sumaAna {
		t.Errorf("fuera de rango debe conservar total/suma: %d/%v, esperaba %d/%v",
			r.Total, r.SumaCoste, esc.totalAna, esc.sumaAna)
	}
}

func TestHistorialT4gZonaHoraria(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	// 31-oct 21:30 UTC = 31-oct 22:30 CET (Madrid) → octubre.
	_, oct := getHistorial(t, app, esc.tokenTZ, "?desde=2026-10-01&hasta=2026-10-31")
	if oct.Total != 1 || oct.SumaCoste != 1.00 {
		t.Errorf("octubre: total %d suma %v, esperaba 1 / 1.00", oct.Total, oct.SumaCoste)
	}
	// 31-oct 23:30 UTC = 1-nov 00:30 CET (Madrid) → noviembre.
	_, nov := getHistorial(t, app, esc.tokenTZ, "?desde=2026-11-01&hasta=2026-11-30")
	if nov.Total != 1 || nov.SumaCoste != 2.00 {
		t.Errorf("noviembre: total %d suma %v, esperaba 1 / 2.00", nov.Total, nov.SumaCoste)
	}
}

func TestHistorialT4hPeriodoVacio(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	w, r := getHistorial(t, app, esc.tokenAna, "?desde=2099-01-01&hasta=2099-01-31")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, un periodo vacío no es error", w.Code)
	}
	if r.Total != 0 || r.SumaCoste != 0 || len(r.Items) != 0 {
		t.Errorf("periodo vacío: total %d suma %v items %d, esperaba 0/0/[]",
			r.Total, r.SumaCoste, len(r.Items))
	}
}

func TestHistorialT4iEntradaInvalida(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	consultas := []string{
		"?page=0",
		"?page_size=500",
		"?desde=2026-05-10&hasta=2026-05-01", // desde > hasta
		"?desde=2026-5-1",                     // formato no estricto
		"?hasta=not-a-date",
		"?page_size=0",
	}
	for _, q := range consultas {
		w, _ := getHistorial(t, app, esc.tokenAna, q)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("consulta %q: status = %d, esperaba 422 (body %s)", q, w.Code, w.Body.String())
			continue
		}
		if cod := codigoError(w.Body.Bytes()); cod != "ENTRADA_INVALIDA" {
			t.Errorf("consulta %q: codigo = %q, esperaba ENTRADA_INVALIDA (body %s)", q, cod, w.Body.String())
		}
	}
}

func TestHistorialT4jLatencia(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	// SC-004 / quickstart T4j: un año completo de uso habitual (~365 viajes
	// finalizados) ANTES de cronometrar (hallazgo A2).
	var bici, eOrigen, eDestino int64
	if err := d.Pool.QueryRow(ctx,
		`SELECT b.id, o.id, d.id FROM bicicletas b
		   CROSS JOIN estaciones o, estaciones d
		  WHERE b.codigo = 'T-H-1' AND o.nombre = 'Origen H' AND d.nombre = 'Destino H'`).
		Scan(&bici, &eOrigen, &eDestino); err != nil {
		t.Fatalf("fixture año: %v", err)
	}
	if _, err := d.Pool.Exec(ctx,
		`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id,
		                     estacion_destino_id, fecha_inicio, fecha_fin,
		                     duracion_minutos, coste_total, estado)
		 SELECT $1, $2, $3, $4,
		        now() - (g || ' days')::interval,
		        now() - (g || ' days')::interval + interval '20 minutes',
		        20, (g % 5) * 0.50, 'finalizado'
		   FROM generate_series(1, 365) g`,
		esc.anaID, bici, eOrigen, eDestino); err != nil {
		t.Fatalf("insertar año de viajes: %v", err)
	}

	inicio := time.Now()
	w, r := getHistorial(t, app, esc.tokenAna, "")
	elapsed := time.Since(inicio)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if r.Total < 365 {
		t.Fatalf("fixture del año incompleta: total = %d (esperaba ≥ 365)", r.Total)
	}
	if elapsed > 3*time.Second {
		t.Errorf("historial tardó %v con %d viajes, objetivo < 3 s (SC-004)", elapsed, r.Total)
	}
}

// T4k (T016): AC4 del clarify 2026-11-19 — cualquier rol autenticado
// consulta su propio historial (el propietario se resuelve desde el JWT);
// un parámetro de usuario ajeno se ignora; sin sesión → 401 (T4d).
func TestHistorialT4kRoles(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	esc := prepararHistorial(t, d)

	hash := nuevoHash(t)
	crearRol := func(email, rol string) string {
		t.Helper()
		var id int64
		if err := d.Pool.QueryRow(ctx,
			`INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
			 VALUES ($1, $2, $3, $4, true, CURRENT_DATE + 365)
			 RETURNING id`, rol, email, hash, rol).Scan(&id); err != nil {
			t.Fatalf("crear %s: %v", rol, err)
		}
		tok, err := auth.CrearToken(d.JWTSecret, id, rol, rol)
		if err != nil {
			t.Fatalf("token %s: %v", rol, err)
		}
		return tok
	}

	for _, caso := range []struct{ nombre, token string }{
		{"operario_taller", crearRol("operario-hist@test.local", "operario_taller")},
		{"gestor", crearRol("gestor-hist@test.local", "gestor")},
	} {
		w, r := getHistorial(t, app, caso.token, "")
		if w.Code != http.StatusOK {
			t.Errorf("%s: status = %d, esperaba 200 (body %s)", caso.nombre, w.Code, w.Body.String())
			continue
		}
		if r.Total != 0 || len(r.Items) != 0 || r.SumaCoste != 0 {
			t.Errorf("%s: debe ver su propio historial (vacío aquí): total %d items %d suma %v",
				caso.nombre, r.Total, len(r.Items), r.SumaCoste)
		}
	}

	// El propietario se resuelve desde el JWT: un parámetro de usuario
	// ajeno no altera el resultado (AC4: nunca por parámetro de entrada).
	w, r := getHistorial(t, app, esc.tokenAna, fmt.Sprintf("?usuario_id=%d", esc.anaID+999999))
	if w.Code != http.StatusOK {
		t.Fatalf("parámetro ajeno: status = %d (body %s)", w.Code, w.Body.String())
	}
	if r.Total != esc.totalAna || r.SumaCoste != esc.sumaAna {
		t.Errorf("parámetro alteró el historial: total %d suma %v, esperaba %d/%v",
			r.Total, r.SumaCoste, esc.totalAna, esc.sumaAna)
	}
}
