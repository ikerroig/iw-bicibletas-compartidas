// T031 — Caso límite de la spec: perder la conexión durante el desbloqueo no
// deja un viaje parcial, y el usuario puede reintentar.
package integration

import (
	"context"
	"net/http"
	"testing"

	"si16/backend/internal/models"
	"si16/backend/internal/services/alquiler"
	tu "si16/backend/internal/testutil"
)

func TestConexionPerdidaNoDejaViajeParcial(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Conexión", models.EstacionOperativa, tu.ConBici)
	uid, tok := e.Usuario(t, true, dias(30))
	antes := fotografiar(t, e)

	cancelado, cancelar := context.WithCancel(e.Ctx)
	cancelar()
	if _, err := alquiler.Nuevo(e.Pool, e.Estacion, nil).Desbloquear(cancelado, uid, codigo(id, 1)); err == nil {
		t.Fatal("con el contexto cancelado debe devolver error")
	}
	if d := fotografiar(t, e); d != antes {
		t.Fatalf("quedaron efectos parciales: %+v → %+v", antes, d)
	}

	if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
		t.Fatalf("reintento = %d %s; quiero 201", status, cod)
	}
}
