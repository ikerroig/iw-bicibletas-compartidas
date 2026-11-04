// T012 y T037 — FR-004 y FR-007: un cambio simultáneo de la estación o del abono
// no se ignora: el desbloqueo espera a que termine y decide con el dato nuevo.
package integration

import (
	"net/http"
	"testing"
	"time"

	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

// desbloqueoDuranteCambio ejecuta `cambio` en una transacción abierta, comprueba
// que el desbloqueo espera y lo deja continuar tras el commit.
func desbloqueoDuranteCambio(t *testing.T, e *tu.Entorno, tok, cod, cambio string, id int64) (int, string) {
	t.Helper()
	tx, err := e.Pool.Begin(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(e.Ctx) }()
	if _, err := tx.Exec(e.Ctx, cambio, id); err != nil {
		t.Fatal(err)
	}

	type resultado struct {
		status int
		codigo string
	}
	fin := make(chan resultado, 1)
	go func() {
		status, c := desbloquear(t, e, tok, cod)
		fin <- resultado{status, c}
	}()

	select {
	case r := <-fin:
		t.Fatalf("el desbloqueo no esperó al cambio: %d %s", r.status, r.codigo)
	case <-time.After(300 * time.Millisecond):
	}
	if err := tx.Commit(e.Ctx); err != nil {
		t.Fatal(err)
	}
	r := <-fin
	return r.status, r.codigo
}

func TestDesbloqueoEsperaAlCambioDeEstacion(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Cierre", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))

	status, cod := desbloqueoDuranteCambio(t, e, tok, codigo(id, 1),
		`UPDATE estaciones SET estado = 'fuera_de_servicio' WHERE id = $1`, id)

	if status != http.StatusConflict || cod != httperr.AnclajeNoDisponible {
		t.Fatalf("respuesta = %d %s; quiero 409 %s", status, cod, httperr.AnclajeNoDisponible)
	}
	if f := fotografiar(t, e); f.Viajes != 0 {
		t.Errorf("se creó un viaje en una estación fuera de servicio: %+v", f)
	}
}

func TestDesbloqueoEsperaALaRevocacionDelAbono(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Revocación", models.EstacionOperativa, tu.ConBici)
	uid, tok := e.Usuario(t, true, dias(30))

	status, cod := desbloqueoDuranteCambio(t, e, tok, codigo(id, 1),
		`UPDATE usuarios SET abono_activo = FALSE WHERE id = $1`, uid)

	if status != http.StatusForbidden || cod != httperr.AbonoVencido {
		t.Fatalf("respuesta = %d %s; quiero 403 %s", status, cod, httperr.AbonoVencido)
	}
	if f := fotografiar(t, e); f.Viajes != 0 {
		t.Errorf("se creó un viaje con el abono revocado: %+v", f)
	}
}
