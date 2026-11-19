// T011 — FR-001: normalización y validación del código de anclaje.
package integration

import (
	"encoding/json"
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

	// T059 (D2, hallazgo U2): mensaje distinto si falta el código o si no es válido.
	const (
		falta    = "Introduce el código del anclaje"
		noValido = "El código no es válido"
	)
	for nombre, c := range map[string]struct{ cuerpo, mensaje string }{
		"vacío":         {`{"codigo_anclaje":""}`, falta},
		"solo espacios": {`{"codigo_anclaje":"   "}`, falta},
		"sin el campo":  {`{}`, falta},
		"65 caracteres": {fmt.Sprintf(`{"codigo_anclaje":%q}`, strings.Repeat("X", 65)), noValido},
		"byte NUL":      {`{"codigo_anclaje":"E01\u0000"}`, noValido},
	} {
		rec := e.Post(t, "/alquileres/desbloqueo", tok, c.cuerpo)
		var r struct {
			Error struct{ Codigo, Mensaje string } `json:"error"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &r)
		if rec.Code != http.StatusUnprocessableEntity || r.Error.Codigo != httperr.Validacion || r.Error.Mensaje != c.mensaje {
			t.Errorf("%s = %d %s; quiero 422 %s %q", nombre, rec.Code, rec.Body, httperr.Validacion, c.mensaje)
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
