package refresco_estado

import (
	"testing"
	"time"
)

// R16: con polling de 15 s, el barrido debe ser de 10 s para que el peor caso
// extremo a extremo (10 + 15 + red + pintado) quede por debajo de 30 s.
func TestIntervaloDeProduccionEsDiezSegundos(t *testing.T) {
	if IntervaloBarrido != 10*time.Second {
		t.Fatalf("IntervaloBarrido = %v; quiero 10s (research R16)", IntervaloBarrido)
	}
	if w := Nuevo(nil, nil, nil, 0); w.intervalo != IntervaloBarrido {
		t.Errorf("intervalo por defecto = %v; quiero %v", w.intervalo, IntervaloBarrido)
	}
}
