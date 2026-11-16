// Tests de integracion de la HU-06 (ficha, reparacion y ubicacion en taller).
package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/models"
)

func TestTallerFichaReparacionYUbicacion(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
		anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var estacionID, bicicletaID, usuarioID int64
	insert(t, d, `INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		VALUES ('Taller test', 39.47, -0.37, 2)`)
	if err := d.Pool.QueryRow(ctx,
		`SELECT id FROM estaciones WHERE nombre = 'Taller test'`).Scan(&estacionID); err != nil {
		t.Fatal(err)
	}
	if err := d.Pool.QueryRow(ctx,
		`INSERT INTO bicicletas (codigo, estado)
		VALUES ('T-HU06-1', 'averiada') RETURNING id`).Scan(&bicicletaID); err != nil {
		t.Fatal(err)
	}
	insert(t, d, `INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
		VALUES ($1, 'T06-A01', 'ocupado', $2), ($1, 'T06-A02', 'libre', NULL)`,
		estacionID, bicicletaID)
	usuarioID, _ = crearUsuario(t, d, "hu06-user@test.local", false)
	insert(t, d, `INSERT INTO incidencias_mantenimientos
		(bicicleta_id, usuario_reporta_id, tipo_fallo, descripcion_reporte)
		VALUES ($1, $2, 'frenos', 'La maneta no frena')`, bicicletaID, usuarioID)
	_, tallerToken := crearUsuarioRol(t, d, "hu06-taller@test.local", "operario_taller")

	ficha := getFichaTaller(app, tallerToken, bicicletaID)
	if ficha.Code != http.StatusOK {
		t.Fatalf("ficha: status %d, body %s", ficha.Code, ficha.Body.String())
	}
	var cuerpoFicha map[string]any
	if err := json.Unmarshal(ficha.Body.Bytes(), &cuerpoFicha); err != nil {
		t.Fatal(err)
	}
	biciFicha := cuerpoFicha["bicicleta"].(map[string]any)
	incidenciasFicha := cuerpoFicha["incidencias"].([]any)
	if biciFicha["codigo"] != "T-HU06-1" || len(incidenciasFicha) != 1 ||
		incidenciasFicha[0].(map[string]any)["descripcion_reporte"] != "La maneta no frena" {
		t.Fatalf("ficha incompleta: %+v", cuerpoFicha)
	}

	parte := postIntervencionTaller(app, tallerToken, bicicletaID,
		"{\"piezas_sustituidas\":\"Maneta derecha\",\"notas_taller\":\"Frenos ajustados\",\"marcar_operativa\":true}")
	if parte.Code != http.StatusCreated {
		t.Fatalf("parte: status %d, body %s", parte.Code, parte.Body.String())
	}

	var estado, estadoIncidencia string
	var operarioID *int64
	var fechaReparacion interface{}
	if err := d.Pool.QueryRow(ctx,
		`SELECT estado FROM bicicletas WHERE id = $1`, bicicletaID).Scan(&estado); err != nil {
		t.Fatal(err)
	}
	if estado != models.BicicletaDisponible {
		t.Fatalf("estado de bicicleta = %q, esperaba disponible", estado)
	}
	if err := d.Pool.QueryRow(ctx, `SELECT estado, operario_id, fecha_reparacion
		FROM incidencias_mantenimientos WHERE bicicleta_id = $1`, bicicletaID).
		Scan(&estadoIncidencia, &operarioID, &fechaReparacion); err != nil {
		t.Fatal(err)
	}
	if estadoIncidencia != models.IncidenciaResuelta || operarioID == nil || fechaReparacion == nil {
		t.Fatalf("incidencia no firmada/cerrada: estado=%s operario=%v fecha=%v",
			estadoIncidencia, operarioID, fechaReparacion)
	}

	ubicacion := postUbicacionTaller(app, tallerToken, bicicletaID,
		"{\"codigo_anclaje\":\"T06-A02\"}")
	if ubicacion.Code != http.StatusOK {
		t.Fatalf("ubicacion: status %d, body %s", ubicacion.Code, ubicacion.Body.String())
	}
	var anclajeEstado string
	var anclajeBici int64
	if err := d.Pool.QueryRow(ctx, `SELECT estado, bicicleta_id
		FROM anclajes WHERE codigo_anclaje = 'T06-A02'`).
		Scan(&anclajeEstado, &anclajeBici); err != nil {
		t.Fatal(err)
	}
	if anclajeEstado != models.AnclajeOcupado || anclajeBici != bicicletaID {
		t.Fatalf("ubicacion no aplicada: estado=%s bici=%d", anclajeEstado, anclajeBici)
	}
}

func TestTallerMantenimientoPreventivo(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
		anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)

	var bicicletaID int64
	insert(t, d, `INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
		VALUES ('Preventivo test', 39.47, -0.37, 1)`)
	var estacionID int64
	if err := d.Pool.QueryRow(ctx,
		`SELECT id FROM estaciones WHERE nombre = 'Preventivo test'`).Scan(&estacionID); err != nil {
		t.Fatal(err)
	}
	if err := d.Pool.QueryRow(ctx, `INSERT INTO bicicletas (codigo, estado)
		VALUES ('T-HU06-P', 'disponible') RETURNING id`).Scan(&bicicletaID); err != nil {
		t.Fatal(err)
	}
	insert(t, d, `INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
		VALUES ($1, 'T06-P01', 'ocupado', $2)`, estacionID, bicicletaID)
	_, tallerToken := crearUsuarioRol(t, d, "hu06-preventivo@test.local", "operario_taller")

	parte := postIntervencionTaller(app, tallerToken, bicicletaID,
		"{\"notas_taller\":\"Revision preventiva\",\"marcar_operativa\":true}")
	if parte.Code != http.StatusCreated {
		t.Fatalf("preventivo: status %d, body %s", parte.Code, parte.Body.String())
	}
	var tipo, descripcion, estado string
	if err := d.Pool.QueryRow(ctx, `SELECT tipo_fallo, descripcion_reporte, estado
		FROM incidencias_mantenimientos WHERE bicicleta_id = $1`, bicicletaID).
		Scan(&tipo, &descripcion, &estado); err != nil {
		t.Fatal(err)
	}
	if tipo != "otro" || descripcion != "Mantenimiento preventivo" || estado != models.IncidenciaResuelta {
		t.Fatalf("preventivo incorrecto: tipo=%s descripcion=%s estado=%s", tipo, descripcion, estado)
	}
}

func TestTallerValidaRolYCampos(t *testing.T) {
	d := depsTest(t)
	app := nuevaApp(d)
	insert(t, d, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
		anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`)
	var bicicletaID int64
	if err := d.Pool.QueryRow(ctx, `INSERT INTO bicicletas (codigo, estado)
		VALUES ('T-HU06-E', 'averiada') RETURNING id`).Scan(&bicicletaID); err != nil {
		t.Fatal(err)
	}
	_, usuarioToken := crearUsuario(t, d, "hu06-no-taller@test.local", false)
	_, tallerToken := crearUsuarioRol(t, d, "hu06-error-taller@test.local", "operario_taller")

	fichaUsuario := getFichaTaller(app, usuarioToken, bicicletaID)
	if fichaUsuario.Code != http.StatusForbidden {
		t.Fatalf("rol usuario: status %d, esperaba 403", fichaUsuario.Code)
	}
	parteVacio := postIntervencionTaller(app, tallerToken, bicicletaID, "{}")
	if parteVacio.Code != http.StatusUnprocessableEntity {
		t.Fatalf("parte vacio: status %d, esperaba 422", parteVacio.Code)
	}
	ubicacion := postUbicacionTaller(app, tallerToken, bicicletaID,
		"{\"codigo_anclaje\":\"NO-EXISTE\"}")
	if ubicacion.Code != http.StatusConflict {
		t.Fatalf("ubicacion sin bicicleta operativa: status %d, esperaba 409", ubicacion.Code)
	}
}

func getFichaTaller(app *gin.Engine, token string, bicicletaID int64) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/v1/taller/bicicletas/%d", bicicletaID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}

func postIntervencionTaller(app *gin.Engine, token string, bicicletaID int64, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/taller/bicicletas/%d/intervenciones", bicicletaID),
		strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}

func postUbicacionTaller(app *gin.Engine, token string, bicicletaID int64, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/taller/bicicletas/%d/ubicacion", bicicletaID),
		strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}
