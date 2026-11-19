package devolucion

import "testing"

// T3a: viaje de 25 min → 0,00 € (tramo gratuito).
func TestCalcularCoste_ViajeCortoGratuito(t *testing.T) {
	coste := CalcularCoste(25*60, 0.20)
	if coste != 0.00 {
		t.Errorf("25 min: esperaba 0.00, obtuve %.2f", coste)
	}
}

// T3c: exactamente 30:00 → 0,00 € (borde incluido en el tramo gratuito).
func TestCalcularCoste_Borde30Minutos(t *testing.T) {
	coste := CalcularCoste(1800, 0.20)
	if coste != 0.00 {
		t.Errorf("1800 s: esperaba 0.00, obtuve %.2f", coste)
	}
}

// T3d: 30:01 → 1 minuto cobrado × 0,20 → 0,20 €.
func TestCalcularCoste_UnMinutoExcedente(t *testing.T) {
	coste := CalcularCoste(1801, 0.20)
	if coste != 0.20 {
		t.Errorf("1801 s: esperaba 0.20, obtuve %.2f", coste)
	}
}

// T3b: 45 min → 15 min excedentes × 0,20 → 3,00 € (sin cobrar los 30 iniciales).
func TestCalcularCoste_45Minutos(t *testing.T) {
	coste := CalcularCoste(45*60, 0.20)
	if coste != 3.00 {
		t.Errorf("45 min: esperaba 3.00, obtuve %.2f", coste)
	}
}

// Edge case (spec): duración negativa / reloj inconsistente → coste 0,00 (nunca negativo).
func TestCalcularCoste_SegundosNegativos(t *testing.T) {
	coste := CalcularCoste(-500, 0.20)
	if coste != 0.00 {
		t.Errorf("segundos negativos: esperaba 0.00, obtuve %.2f", coste)
	}
}

// Configuración "todo gratuito" (TARIFA_MINUTO=0): excedente gratis.
func TestCalcularCoste_TarifaCero(t *testing.T) {
	coste := CalcularCoste(60*60, 0)
	if coste != 0.00 {
		t.Errorf("tarifa 0: esperaba 0.00, obtuve %.2f", coste)
	}
}

// Gran viaje: 2 h → 90 min excedentes × 0,20 → 18,00 €.
func TestCalcularCoste_DosHoras(t *testing.T) {
	coste := CalcularCoste(2*60*60, 0.20)
	if coste != 18.00 {
		t.Errorf("2 h: esperaba 18.00, obtuve %.2f", coste)
	}
}

// duracion_minutos = ceil(segundos/60); con coste > 0 nunca figura ≤ 30 min (SC-001).
func TestDuracionMinutos_RedondeaHaciaArriba(t *testing.T) {
	casos := []struct {
		segundos int64
		minutos  int
	}{{-5, 0}, {0, 0}, {1, 1}, {60, 1}, {61, 2}, {23*60 + 47, 24}, {1800, 30}, {1801, 31}, {1830, 31}}
	for _, c := range casos {
		if got := DuracionMinutos(c.segundos); got != c.minutos {
			t.Errorf("%d s: esperaba %d min, obtuve %d", c.segundos, c.minutos, got)
		}
		if CalcularCoste(c.segundos, 0.20) > 0 && DuracionMinutos(c.segundos) <= 30 {
			t.Errorf("%d s: cobrado pero con duración ≤ 30 min", c.segundos)
		}
	}
}
