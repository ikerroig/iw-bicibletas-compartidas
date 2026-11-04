// T012 — FR-004: una estación que pasa a fuera de servicio durante el
// desbloqueo no admite el viaje.
package integration

import (
	"net/http"
	"testing"
	"time"

	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestDesbloqueoEsperaAlCambioDeEstacion(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Cierre", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))

	tx, err := e.Pool.Begin(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(e.Ctx) }()
	if _, err := tx.Exec(e.Ctx, `UPDATE estaciones SET estado = 'fuera_de_servicio' WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}

	type resultado struct {
		status int
		codigo string
	}
	fin := make(chan resultado, 1)
	go func() {
		status, cod := desbloquear(t, e, tok, codigo(id, 1))
		fin <- resultado{status, cod}
	}()

	select {
	case r := <-fin:
		t.Fatalf("el desbloqueo no esperó al cambio de la estación: %d %s", r.status, r.codigo)
	case <-time.After(300 * time.Millisecond):
	}
	if err := tx.Commit(e.Ctx); err != nil {
		t.Fatal(err)
	}

	r := <-fin
	if r.status != http.StatusConflict || r.codigo != httperr.AnclajeNoDisponible {
		t.Fatalf("respuesta = %d %s; quiero 409 %s", r.status, r.codigo, httperr.AnclajeNoDisponible)
	}
	if f := fotografiar(t, e); f.Viajes != 0 {
		t.Errorf("se creó un viaje en una estación fuera de servicio: %+v", f)
	}
}
