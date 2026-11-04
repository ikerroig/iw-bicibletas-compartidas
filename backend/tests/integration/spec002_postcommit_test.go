// T013 — SC-003 (research D6): los suscriptores reciben un contexto vigente
// aunque el de quien recalcula ya no lo esté.
package integration

import (
	"context"
	"testing"

	"si16/backend/internal/hooks"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestHookRecibeContextoVigenteTrasCancelacion(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Hook", models.EstacionOperativa, tu.ConBici)

	var errHook error
	notificado := false
	hooks.Suscribir(func(ctx context.Context, _ []int64) error {
		notificado, errHook = true, ctx.Err()
		return nil
	})

	cancelado, cancelar := context.WithCancel(e.Ctx)
	cancelar()
	_ = e.Estacion.RecalcularEstacion(cancelado, []int64{id})

	if !notificado {
		t.Fatal("el suscriptor no fue notificado")
	}
	if errHook != nil {
		t.Errorf("contexto del suscriptor = %v; quiero vigente", errHook)
	}
}
