// T024 (y test mínimo de T017) — Las respuestas de auth y de spec-001 cumplen
// specs/spec-001/contracts/openapi.yaml.
package contract

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

const rutaContrato = "../../../specs/spec-001/contracts/openapi.yaml"

var enrutador routers.Router

func contrato(t *testing.T) routers.Router {
	t.Helper()
	if enrutador != nil {
		return enrutador
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(rutaContrato)
	if err != nil {
		t.Fatalf("cargar contrato: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("contrato inválido: %v", err)
	}
	r, err := gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	enrutador = r
	return r
}

// llamar ejecuta la petición contra el router real y valida petición y
// respuesta contra el contrato.
func llamar(t *testing.T, e *tu.Entorno, metodo, ruta, token, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	url := "http://localhost:8080/api/v1" + ruta
	req := httptest.NewRequest(metodo, url, strings.NewReader(cuerpo))
	if cuerpo != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.Router.ServeHTTP(rec, req)

	ruteo, params, err := contrato(t).FindRoute(req)
	if err != nil {
		t.Fatalf("%s %s no está en el contrato: %v", metodo, ruta, err)
	}
	reqVal := &openapi3filter.RequestValidationInput{
		Request: req, PathParams: params, Route: ruteo,
		Options: &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc},
	}
	if cuerpo != "" {
		req.Body = io.NopCloser(strings.NewReader(cuerpo))
	}
	resVal := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqVal,
		Status:                 rec.Code,
		Header:                 rec.Header(),
		Options:                &openapi3filter.Options{IncludeResponseStatus: true},
	}
	resVal.SetBodyBytes(rec.Body.Bytes())
	if err := openapi3filter.ValidateResponse(context.Background(), resVal); err != nil {
		t.Errorf("%s %s → %d no cumple el contrato: %v\ncuerpo: %s", metodo, ruta, rec.Code, err, rec.Body)
	}
	return rec
}

func TestContratoEstaciones(t *testing.T) {
	e := tu.Iniciar(t, "contract", 14)
	id := e.CrearEstacion(t, "Contrato", models.EstacionOperativa, tu.ConBici, tu.ConAveriada, tu.Libre, tu.Roto)
	e.CrearEstacion(t, "Contrato cerrada", models.EstacionFueraDeServicio, tu.ConBici)
	tok := e.Token(t, models.RolUsuario)

	casos := []struct {
		ruta, token string
		quiero      int
	}{
		{"/estaciones", tok, http.StatusOK},
		{"/estaciones", "", http.StatusUnauthorized},
		{fmt.Sprintf("/estaciones/%d", id), tok, http.StatusOK},
		{"/estaciones/987654", tok, http.StatusNotFound},
		{fmt.Sprintf("/estaciones/%d", id), "", http.StatusUnauthorized},
	}
	for _, c := range casos {
		t.Run(fmt.Sprintf("%s_%d", c.ruta, c.quiero), func(t *testing.T) {
			rec := llamar(t, e, http.MethodGet, c.ruta, c.token, "")
			if rec.Code != c.quiero {
				t.Fatalf("status = %d; quiero %d (%s)", rec.Code, c.quiero, rec.Body)
			}
		})
	}
}

func TestContratoAuth(t *testing.T) {
	e := tu.Iniciar(t, "contract", 14)
	e.Token(t, models.RolGestor) // crea u<N>@si16.test con contraseña "password"
	var email string
	if err := e.Pool.QueryRow(e.Ctx, `SELECT email FROM usuarios LIMIT 1`).Scan(&email); err != nil {
		t.Fatal(err)
	}

	rec := llamar(t, e, http.MethodPost, "/auth/login", "", fmt.Sprintf(`{"email":%q,"password":"password"}`, email))
	if rec.Code != http.StatusOK {
		t.Fatalf("login = %d %s", rec.Code, rec.Body)
	}
	token := extraerToken(t, rec.Body.Bytes())

	if rec := llamar(t, e, http.MethodPost, "/auth/login", "", fmt.Sprintf(`{"email":%q,"password":"mala"}`, email)); rec.Code != http.StatusUnauthorized {
		t.Errorf("contraseña errónea = %d; quiero 401", rec.Code)
	}
	if rec := llamar(t, e, http.MethodGet, "/auth/me", token, ""); rec.Code != http.StatusOK {
		t.Errorf("/auth/me = %d %s", rec.Code, rec.Body)
	}
	if rec := llamar(t, e, http.MethodGet, "/auth/me", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("/auth/me sin token = %d; quiero 401", rec.Code)
	}
}

func extraerToken(t *testing.T, cuerpo []byte) string {
	t.Helper()
	i := bytes.Index(cuerpo, []byte(`"token":"`))
	if i < 0 {
		t.Fatalf("sin token en %s", cuerpo)
	}
	resto := cuerpo[i+len(`"token":"`):]
	return string(resto[:bytes.IndexByte(resto, '"')])
}
