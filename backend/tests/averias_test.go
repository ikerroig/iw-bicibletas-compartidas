package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestAveriaBloqueaYEsIdempotente comprueba el recorrido principal y el
// reintento que puede hacer un usuario despues de perder la conexion.
func TestAveriaBloqueaYEsIdempotente(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var estacionID, bicicletaID int64
	insert(t, d, `INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		VALUES ('Averias test', 39.47, -0.37, 1) RETURNING id`)
	if err := d.Pool.QueryRow(ctx, `SELECT id FROM estaciones WHERE nombre = 'Averias test'`).Scan(&estacionID); err != nil {
		t.Fatal(err)
	}
	if err := d.Pool.QueryRow(ctx, `INSERT INTO bicicletas (codigo, estado)
		VALUES ('B-AVERIA-1', 'disponible') RETURNING id`).Scan(&bicicletaID); err != nil {
		t.Fatal(err)
	}
	insert(t, d, `INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
		VALUES ($1, 'AV-A01', 'ocupado', $2)`, estacionID, bicicletaID)
	_, token := crearUsuario(t, d, "averia-user@test.local", false)

	body := `{"bicicleta_codigo":"B-AVERIA-1","estacion_id":1,"tipo_fallo":"frenos","descripcion":"No frena"}`
	primera := postAveria(app, token, body)
	if primera.Code != http.StatusCreated {
		t.Fatalf("primer reporte: status %d, body %s", primera.Code, primera.Body.String())
	}

	var estado string
	var total int
	if err := d.Pool.QueryRow(ctx, `SELECT estado FROM bicicletas WHERE id = $1`, bicicletaID).Scan(&estado); err != nil {
		t.Fatal(err)
	}
	if estado != "averiada" {
		t.Fatalf("estado de bicicleta = %q, esperaba averiada", estado)
	}
	if err := d.Pool.QueryRow(ctx, `SELECT count(*) FROM incidencias_mantenimientos
		WHERE bicicleta_id = $1 AND estado <> 'resuelta'`, bicicletaID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("incidencias abiertas = %d, esperaba 1", total)
	}

	segunda := postAveria(app, token, body)
	if segunda.Code != http.StatusOK {
		t.Fatalf("reintento: status %d, body %s", segunda.Code, segunda.Body.String())
	}
	if err := d.Pool.QueryRow(ctx, `SELECT count(*) FROM incidencias_mantenimientos
		WHERE bicicleta_id = $1 AND estado <> 'resuelta'`, bicicletaID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("el reintento creo otra incidencia: %d", total)
	}
}

func TestAveriaPanelSoloTaller(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)
	_, usuario := crearUsuario(t, d, "panel-user@test.local", false)
	_, taller := crearUsuarioRol(t, d, "panel-taller@test.local", "operario_taller")

	usuarioResp := getIncidencias(app, usuario)
	if usuarioResp.Code != http.StatusForbidden {
		t.Fatalf("usuario en panel: status %d, esperaba 403", usuarioResp.Code)
	}
	tallerResp := getIncidencias(app, taller)
	if tallerResp.Code != http.StatusOK {
		t.Fatalf("taller en panel: status %d, body %s", tallerResp.Code, tallerResp.Body.String())
	}
	var respuesta struct {
		Items []any `json:"items"`
	}
	if err := json.Unmarshal(tallerResp.Body.Bytes(), &respuesta); err != nil {
		t.Fatal(err)
	}
	if respuesta.Items == nil {
		t.Fatal("el panel debe devolver items aunque este vacio")
	}
}

func TestAveriaRechazaEstacionDistinta(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
	               anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var estacionOrigen, estacionOtra, bicicletaID int64
	insert(t, d, `INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		VALUES ('Estacion origen', 39.47, -0.37, 1),
		       ('Estacion distinta', 39.48, -0.38, 1)`)
	if err := d.Pool.QueryRow(ctx, `SELECT id FROM estaciones WHERE nombre = 'Estacion origen'`).Scan(&estacionOrigen); err != nil {
		t.Fatal(err)
	}
	if err := d.Pool.QueryRow(ctx, `SELECT id FROM estaciones WHERE nombre = 'Estacion distinta'`).Scan(&estacionOtra); err != nil {
		t.Fatal(err)
	}
	if err := d.Pool.QueryRow(ctx, `INSERT INTO bicicletas (codigo, estado)
		VALUES ('B-AVERIA-OTRA', 'disponible') RETURNING id`).Scan(&bicicletaID); err != nil {
		t.Fatal(err)
	}
	insert(t, d, `INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
		VALUES ($1, 'AV-ORIGEN', 'ocupado', $2)`, estacionOrigen, bicicletaID)
	_, token := crearUsuario(t, d, "averia-estacion@test.local", false)

	body := fmt.Sprintf(`{"bicicleta_codigo":"B-AVERIA-OTRA","estacion_id":%d,"tipo_fallo":"cadena","descripcion":"La cadena se sale"}`, estacionOtra)
	respuesta := postAveria(app, token, body)
	if respuesta.Code != http.StatusConflict {
		t.Fatalf("estacion distinta: status %d, body %s", respuesta.Code, respuesta.Body.String())
	}
	var estado string
	if err := d.Pool.QueryRow(ctx, `SELECT estado FROM bicicletas WHERE id = $1`, bicicletaID).Scan(&estado); err != nil {
		t.Fatal(err)
	}
	if estado != "disponible" {
		t.Fatalf("la bicicleta cambio de estado tras el rechazo: %q", estado)
	}
}

func postAveria(app *gin.Engine, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/averias", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}

func getIncidencias(app *gin.Engine, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidencias", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}
