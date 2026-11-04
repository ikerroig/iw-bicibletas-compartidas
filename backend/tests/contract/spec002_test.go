// T035 — POST /alquileres/desbloqueo y GET /alquileres/activo cumplen
// specs/spec-001/contracts/openapi.yaml.
package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestContratoAlquileres(t *testing.T) {
	e := tu.Iniciar(t, "contract", 14)
	id := e.CrearEstacion(t, "Alquiler", models.EstacionOperativa, tu.ConBici, tu.ConBici, tu.Libre)
	venc := time.Now().UTC().AddDate(0, 0, 30)
	uid, tok := e.Usuario(t, true, &venc)
	cuerpo := func(n int) string { return fmt.Sprintf(`{"codigo_anclaje":"E%d-A%02d"}`, id, n) }

	sinViaje := llamar(t, e, http.MethodGet, "/alquileres/activo", tok, "")
	if sinViaje.Code != http.StatusOK || sinViaje.Body.String() != `{"viaje":null}` {
		t.Fatalf("sin viaje = %d %s", sinViaje.Code, sinViaje.Body)
	}

	creado := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", tok, cuerpo(1))
	if creado.Code != http.StatusCreated {
		t.Fatalf("desbloqueo = %d %s", creado.Code, creado.Body)
	}
	var viaje models.Viaje
	if err := json.Unmarshal(creado.Body.Bytes(), &viaje); err != nil {
		t.Fatal(err)
	}
	if viaje.UsuarioID != uid || viaje.EstacionOrigenID != id || viaje.Estado != models.ViajeEnCurso {
		t.Errorf("viaje = %+v", viaje)
	}

	activo := llamar(t, e, http.MethodGet, "/alquileres/activo", tok, "")
	var r struct {
		Viaje *models.Viaje `json:"viaje"`
	}
	if err := json.Unmarshal(activo.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if activo.Code != http.StatusOK || r.Viaje == nil || r.Viaje.ID != viaje.ID {
		t.Fatalf("activo = %d %s", activo.Code, activo.Body)
	}

	_, libre := e.Usuario(t, true, &venc)
	for nombre, c := range map[string]struct {
		token, cuerpo string
		status        int
	}{
		"viaje activo":     {tok, cuerpo(2), http.StatusConflict},
		"código no existe": {libre, `{"codigo_anclaje":"X"}`, http.StatusNotFound},
		"código vacío":     {libre, `{"codigo_anclaje":""}`, http.StatusUnprocessableEntity},
		"anclaje vacío":    {libre, cuerpo(3), http.StatusConflict},
	} {
		if rec := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", c.token, c.cuerpo); rec.Code != c.status {
			t.Errorf("%s = %d; quiero %d", nombre, rec.Code, c.status)
		}
	}

	_, caducado := e.Usuario(t, false, nil)
	if rec := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", caducado, cuerpo(2)); rec.Code != http.StatusForbidden {
		t.Errorf("abono caducado = %d; quiero 403", rec.Code)
	}
}
