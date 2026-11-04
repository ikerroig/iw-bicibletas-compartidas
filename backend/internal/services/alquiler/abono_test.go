// T010 — FR-002: la vigencia del abono se decide con la fecha de España.
package alquiler

import (
	"testing"
	"time"

	"si16/backend/internal/models"
)

func TestFechaNegocio(t *testing.T) {
	casos := []struct {
		nombre, instante string
		dia              int
	}{
		{"invierno: 23:59:59 en España", "2026-11-04T22:59:59Z", 4},
		{"invierno: 00:00:00 en España", "2026-11-04T23:00:00Z", 5},
		{"verano: 23:59:59 en España", "2026-07-14T21:59:59Z", 14},
		{"verano: 00:00:00 en España", "2026-07-14T22:00:00Z", 15},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			instante, err := time.Parse(time.RFC3339, c.instante)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, dia := fechaNegocio(instante).Date(); dia != c.dia {
				t.Errorf("día = %d; quiero %d", dia, c.dia)
			}
		})
	}
}

func TestAbonoVenceALaMedianocheDeEspana(t *testing.T) {
	venc := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	u := models.Usuario{AbonoActivo: true, FechaVencimientoAbono: &venc}

	ultimoInstante := time.Date(2026, 11, 4, 22, 59, 59, 0, time.UTC)
	if !models.AbonoVigente(u, fechaNegocio(ultimoInstante)) {
		t.Error("a las 23:59:59 del día de vencimiento debe seguir vigente")
	}
	medianoche := time.Date(2026, 11, 4, 23, 0, 0, 0, time.UTC)
	if models.AbonoVigente(u, fechaNegocio(medianoche)) {
		t.Error("a las 00:00:00 del día siguiente debe estar caducado")
	}
}
