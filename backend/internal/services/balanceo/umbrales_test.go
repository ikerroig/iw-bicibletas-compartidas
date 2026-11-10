// Tests unitarios de umbrales de balanceo (HU-07 T005): cómputo puro sin BD.
// TDD: deben FALLAR hasta implementar umbrales.go (regla T5 constitución I:
// umbrales estrictos <10 % / >90 %).
package balanceo

import "testing"

func TestDecidirUmbralesEstrictos(t *testing.T) {
	casos := []struct {
		nombre      string
		pct         float64
		estado      string
		crear       string // tipo de alerta a crear ("" = ninguna)
		contraria   string // umbral opuesto a resolver si estaba abierto
		resolverAll bool   // vuelve al rango [10, 90] → resolver todas
	}{
		{"9,99 % cruza vacía", 9.99, "operativa", "vacia_menos_10", "llena_mas_90", false},
		{"10,00 % exacto no alerta", 10, "operativa", "", "", true},
		{"90,00 % exacto no alerta", 90, "operativa", "", "", true},
		{"90,01 % cruza llena", 90.01, "operativa", "llena_mas_90", "vacia_menos_10", false},
		{"0 % crea vacía", 0, "operativa", "vacia_menos_10", "llena_mas_90", false},
		{"100 % crea llena", 100, "operativa", "llena_mas_90", "vacia_menos_10", false},
		{"5 % crea vacía", 5, "operativa", "vacia_menos_10", "llena_mas_90", false},
		{"95 % crea llena", 95, "operativa", "llena_mas_90", "vacia_menos_10", false},
		{"fuera_de_servicio excluida sin acción", 5, "fuera_de_servicio", "", "", false},
		{"fuera_de_servicio al 95 % excluida", 95, "fuera_de_servicio", "", "", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := decidir(c.pct, c.estado)
			if d.CrearTipo != c.crear {
				t.Errorf("decidir(%v, %q).CrearTipo = %q, esperado %q", c.pct, c.estado, d.CrearTipo, c.crear)
			}
			if d.ResolverContrario != c.contraria {
				t.Errorf("decidir(%v, %q).ResolverContrario = %q, esperado %q", c.pct, c.estado, d.ResolverContrario, c.contraria)
			}
			if d.ResolverTodas != c.resolverAll {
				t.Errorf("decidir(%v, %q).ResolverTodas = %v, esperado %v", c.pct, c.estado, d.ResolverTodas, c.resolverAll)
			}
		})
	}
}

func TestBiciCuenta(t *testing.T) {
	casos := []struct {
		nombre  string
		estado  string
		anclada bool
		cuenta  bool
	}{
		{"disponible anclada cuenta", "disponible", true, true},
		{"averiada anclada no cuenta", "averiada", true, false},
		{"en_taller anclada no cuenta", "en_taller", true, false},
		{"disponible sin anclaje no cuenta", "disponible", false, false},
	}
	for _, c := range casos {
		if got := biciCuenta(c.estado, c.anclada); got != c.cuenta {
			t.Errorf("biciCuenta(%q, anclada=%v) = %v, esperado %v — %s",
				c.estado, c.anclada, got, c.cuenta, c.nombre)
		}
	}
}

func TestOcupacionPctExclusiones(t *testing.T) {
	if pct := ocupacionPct(1, 20, "operativa"); pct != 5 {
		t.Errorf("ocupacionPct(1, 20, operativa) = %v, esperado 5", pct)
	}
	if pct := ocupacionPct(19, 20, "operativa"); pct != 95 {
		t.Errorf("ocupacionPct(19, 20, operativa) = %v, esperado 95", pct)
	}
	if pct := ocupacionPct(5, 20, "fuera_de_servicio"); pct != 0 {
		t.Errorf("estación fuera_de_servicio debe quedar excluida (0), got %v", pct)
	}
}

func TestScoreUrgenciaOrdenaExtremosPrimero(t *testing.T) {
	// 0 % es más urgente que 9 % (misma alerta vacía)…
	if scoreUrgencia("vacia_menos_10", 0) <= scoreUrgencia("vacia_menos_10", 9) {
		t.Error("0 % debe ordenarse antes que 9 %")
	}
	// …y 100 % más urgente que 91 % (misma alerta llena).
	if scoreUrgencia("llena_mas_90", 100) <= scoreUrgencia("llena_mas_90", 91) {
		t.Error("100 % debe ordenarse antes que 91 %")
	}
}
