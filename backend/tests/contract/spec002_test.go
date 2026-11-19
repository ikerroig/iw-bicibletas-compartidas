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

type viajeActivo struct {
	models.Viaje
	BicicletaCodigo      string `json:"bicicleta_codigo"`
	EstacionOrigenNombre string `json:"estacion_origen_nombre"`
}

func TestContratoAlquileres(t *testing.T) {
	e := tu.Iniciar(t, "contract", 14)
	id := e.CrearEstacion(t, "Alquiler", models.EstacionOperativa, tu.ConBici, tu.ConBici, tu.Libre)
	venc := time.Now().UTC().AddDate(0, 0, 30)
	uid, tok := e.Usuario(t, true, &venc)
	cuerpo := func(n int) string { return fmt.Sprintf(`{"codigo_anclaje":"E%d-A%02d"}`, id, n) }

	// T047 (FR-010, D16): la respuesta incluye siempre el estado del abono.
	sinViaje := llamar(t, e, http.MethodGet, "/alquileres/activo", tok, "")
	if want := `{"viaje":null,"abono":{"vigente":true,"codigo":null,"mensaje":null}}`; sinViaje.Code != http.StatusOK || sinViaje.Body.String() != want {
		t.Fatalf("sin viaje = %d %s; quiero 200 %s", sinViaje.Code, sinViaje.Body, want)
	}

	creado := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", tok, cuerpo(1))
	if creado.Code != http.StatusCreated {
		t.Fatalf("desbloqueo = %d %s", creado.Code, creado.Body)
	}
	var viaje viajeActivo
	if err := json.Unmarshal(creado.Body.Bytes(), &viaje); err != nil {
		t.Fatal(err)
	}
	if viaje.UsuarioID != uid || viaje.EstacionOrigenID != id || viaje.Estado != models.ViajeEnCurso {
		t.Errorf("viaje = %+v", viaje)
	}
	codigoBici := fmt.Sprintf("B-%d-0", id)
	if viaje.BicicletaCodigo != codigoBici || viaje.EstacionOrigenNombre != "Alquiler" {
		t.Errorf("datos del viaje (FR-008) = %q, %q", viaje.BicicletaCodigo, viaje.EstacionOrigenNombre)
	}

	activo := llamar(t, e, http.MethodGet, "/alquileres/activo", tok, "")
	var r struct {
		Viaje *viajeActivo `json:"viaje"`
	}
	if err := json.Unmarshal(activo.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if activo.Code != http.StatusOK || r.Viaje == nil || r.Viaje.ID != viaje.ID {
		t.Fatalf("activo = %d %s", activo.Code, activo.Body)
	}
	if r.Viaje.BicicletaCodigo != codigoBici || r.Viaje.EstacionOrigenNombre != "Alquiler" {
		t.Errorf("datos del viaje activo (FR-008) = %q, %q", r.Viaje.BicicletaCodigo, r.Viaje.EstacionOrigenNombre)
	}

	if rec := llamar(t, e, http.MethodGet, "/alquileres/activo", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("activo sin token = %d; quiero 401", rec.Code)
	}
	if rec := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", "", cuerpo(2)); rec.Code != http.StatusUnauthorized {
		t.Errorf("desbloqueo sin token = %d; quiero 401", rec.Code)
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

	_, sinAbono := e.Usuario(t, false, nil)
	if rec := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", sinAbono, cuerpo(2)); rec.Code != http.StatusForbidden {
		t.Errorf("sin abono = %d; quiero 403", rec.Code)
	}

	// T047 (FR-010, D16): el aviso previo dice exactamente lo mismo que el 403.
	ayer := time.Now().UTC().AddDate(0, 0, -2)
	_, caducado := e.Usuario(t, true, &ayer)
	for _, tc := range []struct{ nombre, token, codigo string }{
		{"caducado", caducado, "ABONO_VENCIDO"},
		{"sin abono", sinAbono, "SIN_ABONO"},
	} {
		rechazo := llamar(t, e, http.MethodPost, "/alquileres/desbloqueo", tc.token, cuerpo(2))
		var r403 struct {
			Error struct{ Codigo, Mensaje string } `json:"error"`
		}
		if err := json.Unmarshal(rechazo.Body.Bytes(), &r403); err != nil {
			t.Fatal(err)
		}
		estado := llamar(t, e, http.MethodGet, "/alquileres/activo", tc.token, "")
		var r200 struct {
			Abono struct {
				Vigente bool    `json:"vigente"`
				Codigo  *string `json:"codigo"`
				Mensaje *string `json:"mensaje"`
			} `json:"abono"`
		}
		if err := json.Unmarshal(estado.Body.Bytes(), &r200); err != nil {
			t.Fatal(err)
		}
		a := r200.Abono
		switch {
		case estado.Code != http.StatusOK || a.Vigente || a.Codigo == nil || a.Mensaje == nil:
			t.Errorf("%s: activo = %d %s", tc.nombre, estado.Code, estado.Body)
		case *a.Codigo != tc.codigo || r403.Error.Codigo != tc.codigo:
			t.Errorf("%s: códigos %q (activo) y %q (403); quiero %q", tc.nombre, *a.Codigo, r403.Error.Codigo, tc.codigo)
		case *a.Mensaje != r403.Error.Mensaje:
			t.Errorf("%s: el aviso %q no coincide con el rechazo %q", tc.nombre, *a.Mensaje, r403.Error.Mensaje)
		}
	}
}
