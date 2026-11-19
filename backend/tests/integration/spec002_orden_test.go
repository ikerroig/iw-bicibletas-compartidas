// T046 — FR-012 (research D4): si fallan varias reglas a la vez, se devuelve
// una sola causa, la primera en este orden: abono → viaje activo → anclaje
// existe → estación en servicio → anclaje en servicio → anclaje con bicicleta
// → bicicleta disponible. Test de caracterización: el orden ya existía (P3.5).
package integration

import (
	"net/http"
	"testing"

	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestOrdenDeCausasDelUsuario(t *testing.T) {
	t.Run("abono caducado y código inexistente → abono", func(t *testing.T) {
		e := iniciar(t)
		ayer := dias(-1)
		_, tok := e.Usuario(t, true, ayer)

		status, cod, msg := desbloquearConMensaje(t, e, tok, "NO-EXISTE")

		if status != http.StatusForbidden || cod != httperr.AbonoVencido || msg != caducoEl(ayer) {
			t.Fatalf("respuesta = %d %s %q; quiero 403 %s", status, cod, msg, httperr.AbonoVencido)
		}
	})

	t.Run("sin abono y con viaje activo → abono", func(t *testing.T) {
		e := iniciar(t)
		id := e.CrearEstacion(t, "Orden", models.EstacionOperativa, tu.ConBici, tu.ConBici)
		uid, tok := e.Usuario(t, true, dias(30))
		if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
			t.Fatalf("preparación: desbloqueo = %d %s", status, cod)
		}
		e.Exec(t, `UPDATE usuarios SET abono_activo = FALSE WHERE id = $1`, uid)

		status, cod := desbloquear(t, e, tok, codigo(id, 2))

		if status != http.StatusForbidden || cod != httperr.SinAbono {
			t.Fatalf("respuesta = %d %s; quiero 403 %s", status, cod, httperr.SinAbono)
		}
	})

	t.Run("viaje activo y código inexistente → viaje activo", func(t *testing.T) {
		e := iniciar(t)
		id := e.CrearEstacion(t, "Orden", models.EstacionOperativa, tu.ConBici)
		_, tok := e.Usuario(t, true, dias(30))
		if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
			t.Fatalf("preparación: desbloqueo = %d %s", status, cod)
		}

		status, cod := desbloquear(t, e, tok, "NO-EXISTE")

		if status != http.StatusConflict || cod != httperr.ViajeActivo {
			t.Fatalf("respuesta = %d %s; quiero 409 %s", status, cod, httperr.ViajeActivo)
		}
	})
}

func TestOrdenDeCausasDelAnclaje(t *testing.T) {
	const (
		estacionFuera = "La estación de este anclaje está fuera de servicio"
		anclajeFuera  = "Este anclaje está fuera de servicio"
		sinBici       = "Este anclaje no tiene ninguna bicicleta"
	)
	averiadaEnRoto := tu.Anclaje{Estado: models.AnclajeFueraDeServicio, Bicicleta: models.BicicletaAveriada}
	casos := []struct {
		nombre, estacion string
		anclaje          tu.Anclaje
		codigo, mensaje  string
	}{
		{"estación fuera de servicio y anclaje vacío", models.EstacionFueraDeServicio, tu.Libre, httperr.AnclajeNoDisponible, estacionFuera},
		{"estación fuera de servicio y anclaje roto", models.EstacionFueraDeServicio, tu.Roto, httperr.AnclajeNoDisponible, estacionFuera},
		{"estación fuera de servicio y bici averiada", models.EstacionFueraDeServicio, tu.ConAveriada, httperr.AnclajeNoDisponible, estacionFuera},
		{"anclaje roto y vacío", models.EstacionOperativa, tu.Roto, httperr.AnclajeNoDisponible, anclajeFuera},
		{"anclaje roto con bici averiada", models.EstacionOperativa, averiadaEnRoto, httperr.AnclajeNoDisponible, anclajeFuera},
		{"anclaje vacío en estación operativa", models.EstacionOperativa, tu.Libre, httperr.AnclajeNoDisponible, sinBici},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := iniciar(t)
			id := e.CrearEstacion(t, "Orden", c.estacion, c.anclaje)
			_, tok := e.Usuario(t, true, dias(30))

			status, cod, msg := desbloquearConMensaje(t, e, tok, codigo(id, 1))

			if status != http.StatusConflict || cod != c.codigo || msg != c.mensaje {
				t.Fatalf("respuesta = %d %s %q; quiero 409 %s %q", status, cod, msg, c.codigo, c.mensaje)
			}
		})
	}
}
