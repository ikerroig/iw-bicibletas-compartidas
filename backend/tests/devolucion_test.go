// Tests de integración de la HU-03 (devolución y tarificación).
// Ejecutan contra la BD y Redis del docker-compose (ver quickstart.md).
// Deben FALLAR antes de implementar el servicio (TDD).
package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/routes"
	tu "si16/backend/internal/testutil"
)

var ctx = context.Background()

// depsTest usa la BD y la base de Redis propias de este paquete (testutil):
// nunca DATABASE_URL, porque los tests vacían las tablas.
func depsTest(t *testing.T) deps.Deps {
	t.Helper()
	e := tu.Iniciar(t, "tests", 13)
	return deps.Deps{
		Pool:         e.Pool,
		RDB:          e.Cache.Redis(),
		JWTSecret:    tu.SecretoTests,
		TarifaMinuto: 0.20,
		Estacion:     e.Estacion,
	}
}

func nuevaApp(d deps.Deps) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	routes.Register(api, d)
	return r
}

// escenario: datos mínimos para probar la devolución.
type escenario struct {
	usuarioID       int64
	token           string
	usuarioSinViaje string // email de otro usuario sin viaje activo
	tokenSinViaje   string
	estacionDestino int64
	anclajeLibre    string // libre (destino válido)
	anclajeLibre2   string // segunda opción libre (concurrencia)
	anclajeOcupado  string // ocupado
	anclajeRetirado string // fuera_de_servicio
	viajeID         int64
	biciID          int64
}

func nuevoHash(t *testing.T) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(h)
}

// crearUsuario inserta un usuario y devuelve su id y token.
func crearUsuario(t *testing.T, d deps.Deps, email string, conViaje bool) (int64, string) {
	t.Helper()
	hash := nuevoHash(t)
	var id int64
	err := d.Pool.QueryRow(ctx,
		`INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
		 VALUES ('Tester', $1, $2, 'usuario', true, CURRENT_DATE + 365)
		 RETURNING id`, email, hash).Scan(&id)
	if err != nil {
		t.Fatalf("crear usuario: %v", err)
	}
	token, err := auth.CrearToken(d.JWTSecret, id, "Tester", "usuario")
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	_ = conViaje
	return id, token
}

func insert(t *testing.T, d deps.Deps, sql string, args ...any) {
	t.Helper()
	if _, err := d.Pool.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("fixture %s: %v", sql, err)
	}
}

// prepararEscenario limpia la BD y crea estaciones, anclajes y un viaje activo.
func prepararEscenario(t *testing.T, d deps.Deps) escenario {
	t.Helper()
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var estOrigen, estDestino int64
	err := d.Pool.QueryRow(ctx,
		`INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		 VALUES ('Origen T', 39.47, -0.38, 5), ('Destino T', 39.48, -0.37, 5)
		 RETURNING id`).Scan(&estOrigen)
	if err != nil {
		t.Fatalf("estación origen: %v", err)
	}
	if err := d.Pool.QueryRow(ctx,
		`SELECT id FROM estaciones WHERE nombre = 'Destino T'`).Scan(&estDestino); err != nil {
		t.Fatalf("estación destino: %v", err)
	}

	crearAnclaje := func(estacion int64, codigo, estado string, biciID *int64) {
		t.Helper()
		var bici any
		if biciID != nil {
			bici = *biciID
		}
		insert(t, d,
			`INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
			 VALUES ($1, $2, $3, $4)`, estacion, codigo, estado, bici)
	}

	// Bicicleta ya anclada en el anclaje ocupado.
	var biciOcupada int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO bicicletas (codigo, estado) VALUES ('T-BICI-3', 'disponible') RETURNING id`).Scan(&biciOcupada); err != nil {
		t.Fatalf("bici ocupada: %v", err)
	}

	crearAnclaje(estDestino, "AN-90001", "libre", nil)
	crearAnclaje(estDestino, "AN-90002", "libre", nil)
	crearAnclaje(estDestino, "AN-90003", "ocupado", &biciOcupada)
	crearAnclaje(estDestino, "AN-90004", "fuera_de_servicio", nil)

	// Usuario CON viaje activo.
	usuarioID, token := crearUsuario(t, d, "viajero@test.local", true)
	// Usuario SIN viaje activo.
	_, tokenSinViaje := crearUsuario(t, d, "sinviaje@test.local", false)

	// Bicicleta en uso con viaje activo (15 min de trayecto → tramo gratuito).
	var biciID int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO bicicletas (codigo, estado) VALUES ('T-BICI-1', 'en_uso') RETURNING id`).Scan(&biciID); err != nil {
		t.Fatalf("bici viaje: %v", err)
	}
	var viajeID int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id, fecha_inicio)
		 VALUES ($1, $2, $3, now() - interval '15 minutes') RETURNING id`,
		usuarioID, biciID, estOrigen).Scan(&viajeID); err != nil {
		t.Fatalf("viaje activo: %v", err)
	}

	return escenario{
		usuarioID:       usuarioID,
		token:           token,
		tokenSinViaje:   tokenSinViaje,
		estacionDestino: estDestino,
		anclajeLibre:    "AN-90001",
		anclajeLibre2:   "AN-90002",
		anclajeOcupado:  "AN-90003",
		anclajeRetirado: "AN-90004",
		viajeID:         viajeID,
		biciID:          biciID,
	}
}

type respuesta struct {
	status int
	body   map[string]any
	raw    string
}

func postDevolucion(app *gin.Engine, token, codigoAnclaje string) respuesta {
	body := fmt.Sprintf(`{"codigo_anclaje":%q}`, codigoAnclaje)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devoluciones", strings.NewReader(body))
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

func errorCodigo(r respuesta) string {
	if e, ok := r.body["error"].(map[string]any); ok {
		c, _ := e["codigo"].(string)
		return c
	}
	return ""
}

func viajeEstado(t *testing.T, d deps.Deps, id int64) (estado string, coste float64) {
	t.Helper()
	err := d.Pool.QueryRow(ctx,
		`SELECT estado, coste_total::float8 FROM viajes WHERE id = $1`, id).Scan(&estado, &coste)
	if err != nil {
		t.Fatalf("consultar viaje: %v", err)
	}
	return estado, coste
}

func anclajeEstado(t *testing.T, d deps.Deps, codigo string) (estado string, biciID *int64) {
	t.Helper()
	err := d.Pool.QueryRow(ctx,
		`SELECT estado, bicicleta_id FROM anclajes WHERE codigo_anclaje = $1`, codigo).Scan(&estado, &biciID)
	if err != nil {
		t.Fatalf("consultar anclaje %s: %v", codigo, err)
	}
	return estado, biciID
}

// T3a/e2e feliz + T3j: devolución válida → 200, viaje finalizado, 0,00 €,
// anclaje ocupado, usuario libre, Redis actualizado y latencia < 10 s (SC-003).
func TestDevolucion_ViajeCorto_FinalizaYGratuito(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	inicio := time.Now()
	r := postDevolucion(app, e.token, e.anclajeLibre)
	latencia := time.Since(inicio)

	if r.status != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", r.status, r.raw)
	}
	if latencia > 10*time.Second {
		t.Errorf("SC-003: la devolución tardó %v (> 10 s)", latencia)
	}

	// Viaje finalizado con coste 0,00 y etiqueta de tramo gratuito.
	estado, coste := viajeEstado(t, d, e.viajeID)
	if estado != "finalizado" {
		t.Errorf("estado del viaje: esperaba finalizado, obtuve %q", estado)
	}
	if coste != 0.00 {
		t.Errorf("coste: esperaba 0.00 (25 min ≤ 30), obtuve %.2f", coste)
	}
	if et, _ := r.body["etiqueta"].(string); et != "Tramo gratuito (0,00 €)" {
		t.Errorf("etiqueta: esperaba 'Tramo gratuito (0,00 €)', obtuve %q", et)
	}
	if est, _ := r.body["estado"].(string); est != "finalizado" {
		t.Errorf("respuesta.estado: esperaba finalizado, obtuve %q", est)
	}

	// Anclaje ocupado con la bicicleta del viaje.
	estadoAnclaje, biciEnAnclaje := anclajeEstado(t, d, e.anclajeLibre)
	if estadoAnclaje != "ocupado" {
		t.Errorf("anclaje: esperaba ocupado, obtuve %q", estadoAnclaje)
	}
	if biciEnAnclaje == nil || *biciEnAnclaje != e.biciID {
		t.Errorf("anclaje no aloja la bicicleta del viaje")
	}

	// Bicicleta disponible.
	var estadoBici string
	if err := d.Pool.QueryRow(ctx, `SELECT estado FROM bicicletas WHERE id = $1`, e.biciID).Scan(&estadoBici); err != nil {
		t.Fatalf("bici: %v", err)
	}
	if estadoBici != "disponible" {
		t.Errorf("bicicleta: esperaba disponible, obtuve %q", estadoBici)
	}

	// Usuario libre: sin viajes en_curso (SC-004).
	var activos int
	if err := d.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM viajes WHERE usuario_id = $1 AND estado = 'en_curso'`, e.usuarioID).Scan(&activos); err != nil {
		t.Fatalf("viajes activos: %v", err)
	}
	if activos != 0 {
		t.Errorf("viajes activos tras devolución: esperaba 0, obtuve %d", activos)
	}

	// Frescura Redis (constitución IV): clave de la estación destino con marca nueva.
	clave := fmt.Sprintf("estacion:%d:estado", e.estacionDestino)
	if err := d.RDB.HGet(ctx, clave, "actualizado_en").Err(); err != nil {
		t.Errorf("clave Redis %s no escrita post-commit: %v", clave, err)
	}
}

// T3e: sin viaje activo → 409 SIN_VIAJE_ACTIVO y el anclaje queda intacto (FR-007).
func TestDevolucion_SinViajeActivo_RechazaSinEfectos(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	r := postDevolucion(app, e.tokenSinViaje, e.anclajeLibre)
	if r.status != http.StatusConflict {
		t.Fatalf("esperaba 409, obtuve %d: %s", r.status, r.raw)
	}
	if cod := errorCodigo(r); cod != "SIN_VIAJE_ACTIVO" {
		t.Errorf("código: esperaba SIN_VIAJE_ACTIVO, obtuve %q", cod)
	}
	estadoAnclaje, bici := anclajeEstado(t, d, e.anclajeLibre)
	if estadoAnclaje != "libre" || bici != nil {
		t.Errorf("el anclaje fue alterado: estado=%q bici=%v", estadoAnclaje, bici)
	}
}

// T3g: anclaje ocupado / inexistente / retirado → 409 y viaje sigue activo.
func TestDevolucion_AnclajesInvalidos_Rechazan(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	casos := []struct {
		codigo, esperado string
	}{
		{e.anclajeOcupado, "ANCLAJE_OCUPADO"},
		{"AN-NOEXISTE", "ANCLAJE_NO_ENCONTRADO"},
		{e.anclajeRetirado, "ANCLAJE_FUERA_DE_SERVICIO"},
	}
	for _, c := range casos {
		r := postDevolucion(app, e.token, c.codigo)
		if r.status != http.StatusConflict {
			t.Errorf("%s: esperaba 409, obtuve %d: %s", c.codigo, r.status, r.raw)
			continue
		}
		if cod := errorCodigo(r); cod != c.esperado {
			t.Errorf("%s: esperaba %s, obtuve %q", c.codigo, c.esperado, cod)
		}
	}

	// El viaje sigue activo tras los rechazos y el anclaje ocupado sin tocar.
	if estado, _ := viajeEstado(t, d, e.viajeID); estado != "en_curso" {
		t.Errorf("el viaje debía seguir en_curso, está %q", estado)
	}
	if estadoAnclaje, _ := anclajeEstado(t, d, e.anclajeOcupado); estadoAnclaje != "ocupado" {
		t.Errorf("anclaje ocupado alterado: %q", estadoAnclaje)
	}
}

// T3f: dos devoluciones casi simultáneas del mismo viaje → exactamente una
// tiene efecto; la otra recibe 409 (FR-006, edge case del spec).
func TestDevolucion_DobleConfirmacion_SoloUnaEfecto(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	var wg sync.WaitGroup
	responses := make([]respuesta, 2)
	for i, codigo := range []string{e.anclajeLibre, e.anclajeLibre2} {
		wg.Add(1)
		go func(i int, codigo string) {
			defer wg.Done()
			responses[i] = postDevolucion(app, e.token, codigo)
		}(i, codigo)
	}
	wg.Wait()

	oks := 0
	for i, r := range responses {
		switch r.status {
		case http.StatusOK:
			oks++
		case http.StatusConflict:
			cod := errorCodigo(r)
			if cod != "VIAJE_YA_CERRADO" && cod != "SIN_VIAJE_ACTIVO" && cod != "ANCLAJE_OCUPADO" {
				t.Errorf("respuesta %d: código 409 inesperado %q", i, cod)
			}
		default:
			t.Errorf("respuesta %d: status %d inesperado: %s", i, r.status, r.raw)
		}
	}
	if oks != 1 {
		t.Errorf("esperaba exactamente 1 × 200, obtuve %d", oks)
	}

	// Estado final: un solo viaje, finalizado exactamente una vez.
	estado, _ := viajeEstado(t, d, e.viajeID)
	if estado != "finalizado" {
		t.Errorf("viaje: esperaba finalizado, obtuve %q", estado)
	}
	var total int
	if err := d.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM viajes`).Scan(&total); err != nil {
		t.Fatalf("count viajes: %v", err)
	}
	if total != 1 {
		t.Errorf("viajes en BD: esperaba 1, obtuve %d", total)
	}
}

// T3h: bici reportada averiada DURANTE el viaje → la devolución la deja
// anclada y AVERIADA (edge case HU-05 del spec), con la tarifa intacta.
func TestDevolucion_BiciAveriadaDuranteViaje_QuedaAncladaYaveriada(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	insert(t, d, `UPDATE bicicletas SET estado = 'averiada' WHERE id = $1`, e.biciID)

	r := postDevolucion(app, e.token, e.anclajeLibre)
	if r.status != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", r.status, r.raw)
	}

	var estadoBici string
	if err := d.Pool.QueryRow(ctx, `SELECT estado FROM bicicletas WHERE id = $1`, e.biciID).Scan(&estadoBici); err != nil {
		t.Fatalf("bici: %v", err)
	}
	if estadoBici != "averiada" {
		t.Errorf("bicicleta: esperaba que siguiera averiada, obtuve %q", estadoBici)
	}
	estadoAnclaje, bici := anclajeEstado(t, d, e.anclajeLibre)
	if estadoAnclaje != "ocupado" || bici == nil {
		t.Errorf("la bici averiada debía quedar anclada: estado=%q bici=%v", estadoAnclaje, bici)
	}
	// La avería no altera la tarificación: viaje finalizado y coste 0 (15 min).
	estado, coste := viajeEstado(t, d, e.viajeID)
	if estado != "finalizado" || coste != 0.00 {
		t.Errorf("viaje: estado=%q coste=%.2f (esperaba finalizado / 0.00)", estado, coste)
	}
}

// T3i: reloj inconsistente (fecha_inicio en el futuro → duración negativa)
// → se usa hora validada y el coste nunca es negativo.
func TestDevolucion_RelojInconsistente_CosteNoNegativo(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	insert(t, d, `UPDATE viajes SET fecha_inicio = now() + interval '2 hours' WHERE id = $1`, e.viajeID)

	r := postDevolucion(app, e.token, e.anclajeLibre)
	if r.status != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", r.status, r.raw)
	}

	var (
		estado   string
		coste    float64
		duracion *int
		fechaFin *time.Time
	)
	if err := d.Pool.QueryRow(ctx,
		`SELECT estado, coste_total::float8, duracion_minutos, fecha_fin
		   FROM viajes WHERE id = $1`, e.viajeID).Scan(&estado, &coste, &duracion, &fechaFin); err != nil {
		t.Fatalf("viaje: %v", err)
	}
	if estado != "finalizado" {
		t.Errorf("estado: esperaba finalizado, obtuve %q", estado)
	}
	if coste < 0 {
		t.Errorf("coste negativo: %.2f", coste)
	}
	if duracion == nil || *duracion < 0 {
		t.Errorf("duración inválida: %v", duracion)
	}
	if fechaFin == nil {
		t.Errorf("fecha_fin no registrada")
	}
}

// T3k (T028): permisos de cierre (FR-008, AC4) — sin sesión → 401;
// con sesión de rol no-usuario → 403 ROL_INSUFICIENTE; en ambos casos el
// viaje permanece en_curso y el anclaje libre intacto (nada se cierra).
func TestDevolucion_Permisos_SinSesionYRolNoUsuario(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	// Usuario con rol operario_taller (no puede cerrar viajes).
	hash := nuevoHash(t)
	var operarioID int64
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
		 VALUES ('Operario', 'operario@test.local', $1, 'operario_taller', true, CURRENT_DATE + 365)
		 RETURNING id`, hash).Scan(&operarioID); err != nil {
		t.Fatalf("crear operario: %v", err)
	}
	tokenOperario, err := auth.CrearToken(d.JWTSecret, operarioID, "Operario", "operario_taller")
	if err != nil {
		t.Fatalf("token operario: %v", err)
	}

	casos := []struct {
		nombre, token, codigo string
		status                int
	}{
		{"sin sesión", "", "NO_AUTENTICADO", http.StatusUnauthorized},
		{"rol operario_taller", tokenOperario, "ROL_NO_PERMITIDO", http.StatusForbidden},
	}
	for _, c := range casos {
		r := postDevolucion(app, c.token, e.anclajeLibre)
		if r.status != c.status {
			t.Errorf("%s: esperaba %d, obtuve %d: %s", c.nombre, c.status, r.status, r.raw)
			continue
		}
		if cod := errorCodigo(r); cod != c.codigo {
			t.Errorf("%s: esperaba código %s, obtuve %q", c.nombre, c.codigo, cod)
		}
	}

	// Sin efectos: el viaje sigue abierto y el anclaje libre intacto.
	if estado, _ := viajeEstado(t, d, e.viajeID); estado != "en_curso" {
		t.Errorf("viaje: esperaba en_curso, obtuve %q", estado)
	}
	if estadoAnclaje, bici := anclajeEstado(t, d, e.anclajeLibre); estadoAnclaje != "libre" || bici != nil {
		t.Errorf("anclaje alterado: estado=%q bici=%v", estadoAnclaje, bici)
	}
}

// T3l (T029): abono caducado (FR-009, AC5) — el cierre NO exige abono
// vigente: usuario con abono_activo = false y fecha vencida devuelve la
// bici con 200, viaje finalizado y coste registrado.
func TestDevolucion_AbonoCaducado_PuedeDevolver(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	insert(t, d,
		`UPDATE usuarios SET abono_activo = false, fecha_vencimiento_abono = CURRENT_DATE - 30
		 WHERE id = $1`, e.usuarioID)

	r := postDevolucion(app, e.token, e.anclajeLibre)
	if r.status != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", r.status, r.raw)
	}

	estado, coste := viajeEstado(t, d, e.viajeID)
	if estado != "finalizado" {
		t.Errorf("viaje: esperaba finalizado, obtuve %q", estado)
	}
	if coste != 0.00 {
		t.Errorf("coste: esperaba 0.00 (15 min ≤ 30), obtuve %.2f", coste)
	}
	estadoAnclaje, bici := anclajeEstado(t, d, e.anclajeLibre)
	if estadoAnclaje != "ocupado" || bici == nil || *bici != e.biciID {
		t.Errorf("anclaje: estado=%q bici=%v (esperaba ocupado con la bici del viaje)", estadoAnclaje, bici)
	}
}

// T032: contrato devoluciones.md — `codigo_anclaje` ≤ 64 chars.
// 65 → 422 ENTRADA_INVALIDA sin efectos; 64 → pasa la validación (409 por
// código inexistente, no 422).
func TestDevolucion_CodigoLargo_ValidacionContrato(t *testing.T) {
	d := depsTest(t)
	e := prepararEscenario(t, d)
	app := nuevaApp(d)

	r := postDevolucion(app, e.token, strings.Repeat("A", 65))
	if r.status != http.StatusUnprocessableEntity {
		t.Fatalf("65 chars: esperaba 422, obtuve %d: %s", r.status, r.raw)
	}
	if cod := errorCodigo(r); cod != "ENTRADA_INVALIDA" {
		t.Errorf("65 chars: esperaba ENTRADA_INVALIDA, obtuve %q", cod)
	}
	if estado, _ := viajeEstado(t, d, e.viajeID); estado != "en_curso" {
		t.Errorf("viaje alterado por rechazo de entrada: %q", estado)
	}

	r64 := postDevolucion(app, e.token, strings.Repeat("A", 64))
	if r64.status != http.StatusConflict || errorCodigo(r64) != "ANCLAJE_NO_ENCONTRADO" {
		t.Errorf("64 chars: esperaba 409 ANCLAJE_NO_ENCONTRADO (validación superada), obtuve %d %q: %s",
			r64.status, errorCodigo(r64), r64.raw)
	}
}
