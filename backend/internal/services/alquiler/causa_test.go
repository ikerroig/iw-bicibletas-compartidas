// T044 — FR-002, SC-001 (research D15): la causa del abono no vigente
// distingue «sin abono» de «caducado», con la fecha en el segundo caso.
package alquiler

import (
	"errors"
	"testing"
	"time"

	"si16/backend/internal/models"
)

func TestCausaAbono(t *testing.T) {
	hoy := time.Date(2026, 11, 19, 12, 0, 0, 0, zonaNegocio)
	fecha := func(a int, m time.Month, d int) *time.Time {
		v := time.Date(a, m, d, 0, 0, 0, 0, time.UTC)
		return &v
	}
	casos := []struct {
		nombre  string
		u       models.Usuario
		vigente bool
		causa   error
		mensaje string
	}{
		{"inactivo", models.Usuario{AbonoActivo: false, FechaVencimientoAbono: fecha(2027, 1, 1)},
			false, ErrSinAbono, "No tienes abono anual"},
		{"activo sin fecha", models.Usuario{AbonoActivo: true},
			false, ErrSinAbono, "No tienes abono anual"},
		{"inactivo y con fecha pasada", models.Usuario{AbonoActivo: false, FechaVencimientoAbono: fecha(2026, 11, 9)},
			false, ErrSinAbono, "No tienes abono anual"},
		{"caducado", models.Usuario{AbonoActivo: true, FechaVencimientoAbono: fecha(2026, 11, 9)},
			false, ErrAbonoCaducado, "Tu abono anual caducó el 09/11/2026"},
		{"vence hoy", models.Usuario{AbonoActivo: true, FechaVencimientoAbono: fecha(2026, 11, 19)},
			true, nil, ""},
		{"vigente", models.Usuario{AbonoActivo: true, FechaVencimientoAbono: fecha(2027, 6, 7)},
			true, nil, ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := causaAbono(c.u, hoy)
			if got.Vigente != c.vigente || !errors.Is(got.Causa, c.causa) || got.Mensaje != c.mensaje {
				t.Errorf("causaAbono = {%v, %v, %q}; quiero {%v, %v, %q}",
					got.Vigente, got.Causa, got.Mensaje, c.vigente, c.causa, c.mensaje)
			}
		})
	}
}

// El error del rechazo lleva el mensaje y se reconoce con errors.Is.
func TestRechazoAbonoEnvuelveLaCausa(t *testing.T) {
	err := error(&RechazoAbono{Causa: ErrAbonoCaducado, Mensaje: "Tu abono anual caducó el 09/11/2026"})
	if !errors.Is(err, ErrAbonoCaducado) || errors.Is(err, ErrSinAbono) {
		t.Fatalf("errors.Is no reconoce la causa: %v", err)
	}
	var r *RechazoAbono
	if !errors.As(err, &r) || r.Mensaje != "Tu abono anual caducó el 09/11/2026" {
		t.Fatalf("errors.As = %+v", r)
	}
}
