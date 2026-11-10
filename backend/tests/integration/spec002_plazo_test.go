// T035 — SC-003 (research D11): un bloqueo retenido no deja colgado el desbloqueo.
package integration

import (
	"net/http"
	"testing"
	"time"

	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestDesbloqueoTienePlazoAnteBloqueoRetenido(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Bloqueo", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))
	antes := fotografiar(t, e)

	tx, err := e.Pool.Begin(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(e.Ctx) }()
	if _, err := tx.Exec(e.Ctx, `SELECT 1 FROM anclajes WHERE codigo_anclaje = $1 FOR UPDATE`, codigo(id, 1)); err != nil {
		t.Fatal(err)
	}

	type resultado struct {
		status   int
		duracion time.Duration
	}
	fin := make(chan resultado, 1)
	go func() {
		inicio := time.Now()
		status, _ := desbloquear(t, e, tok, codigo(id, 1))
		fin <- resultado{status, time.Since(inicio)}
	}()

	select {
	case r := <-fin:
		if r.status != http.StatusInternalServerError || r.duracion >= 5*time.Second {
			t.Errorf("respuesta = %d en %v; quiero 500 en menos de 5 s", r.status, r.duracion)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("el desbloqueo sigue esperando pasado el plazo")
	}
	if err := tx.Rollback(e.Ctx); err != nil {
		t.Fatal(err)
	}
	if d := fotografiar(t, e); d != antes {
		t.Errorf("hubo efectos: %+v → %+v", antes, d)
	}
}
