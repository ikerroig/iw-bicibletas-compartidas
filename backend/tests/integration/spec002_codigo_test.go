// T011 — FR-001: normalización y validación del código de anclaje.
package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestCodigoSinDistinguirMayusculasNiEspacios(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Minúsculas", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))

	status, cod := desbloquear(t, e, tok, fmt.Sprintf("  e%d-a01 ", id))

	if status != http.StatusCreated {
		t.Fatalf("respuesta = %d %s; quiero 201", status, cod)
	}
}

func TestCodigoInvalidoEsErrorDeValidacion(t *testing.T) {
	e := iniciar(t)
	e.CrearEstacion(t, "Inválido", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))
	antes := fotografiar(t, e)

	for nombre, cuerpo := range map[string]string{
		"65 caracteres": fmt.Sprintf(`{"codigo_anclaje":%q}`, strings.Repeat("X", 65)),
		"byte NUL":      `{"codigo_anclaje":"E01\u0000"}`,
	} {
		rec := e.Post(t, "/alquileres/desbloqueo", tok, cuerpo)
		if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), httperr.Validacion) {
			t.Errorf("%s = %d %s; quiero 422 %s", nombre, rec.Code, rec.Body, httperr.Validacion)
		}
	}
	if d := fotografiar(t, e); d != antes {
		t.Errorf("el rechazo tuvo efectos: %+v → %+v", antes, d)
	}
}

func TestCodigoDeLongitudMaximaSeBusca(t *testing.T) {
	e := iniciar(t)
	_, tok := e.Usuario(t, true, dias(30))

	status, cod := desbloquear(t, e, tok, strings.Repeat("X", 64))

	if status != http.StatusNotFound || cod != httperr.AnclajeNoExiste {
		t.Fatalf("respuesta = %d %s; quiero 404 %s", status, cod, httperr.AnclajeNoExiste)
	}
}
