// Tests unitarios de la lógica pura del servicio de historial (HU-04):
// derivación de etiqueta, offset de paginación y validación de entrada
// (research D6). Sin BD: solo funciones puras del paquete viajes.
package viajes

import (
	"testing"
	"time"
)

func TestEtiquetaParaCoste(t *testing.T) {
	casos := []struct {
		nombre       string
		coste        float64
		esperaEtique bool
	}{
		{"coste 0 exacto → etiqueta", 0, true},
		{"coste 0.00 vía float", 0.00, true},
		{"coste positivo → sin etiqueta", 2.5, false},
		{"coste muy pequeño pero > 0 → sin etiqueta", 0.01, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			et := EtiquetaParaCoste(c.coste)
			if c.esperaEtique {
				if et == nil {
					t.Fatalf("coste %v: esperaba etiqueta, llegó nil", c.coste)
				}
				if *et != "Tramo gratuito (0,00 €)" {
					t.Errorf("etiqueta = %q, esperaba %q", *et, "Tramo gratuito (0,00 €)")
				}
			} else if et != nil {
				t.Errorf("coste %v: no debía llevar etiqueta, llegó %q", c.coste, *et)
			}
		})
	}
}

func TestOffset(t *testing.T) {
	casos := []struct {
		page, pageSize, esperado int
	}{
		{1, 50, 0},
		{2, 50, 50},
		{3, 1, 2},
		{7, 200, 1200},
	}
	for _, c := range casos {
		if got := Offset(c.page, c.pageSize); got != c.esperado {
			t.Errorf("Offset(%d,%d) = %d, esperaba %d", c.page, c.pageSize, got, c.esperado)
		}
	}
}

func TestValidarEntrada(t *testing.T) {
	t.Run("vacío: defaults page=1 page_size=50 sin fechas", func(t *testing.T) {
		p, aerr := ValidarEntrada("", "", "", "")
		if aerr != nil {
			t.Fatalf("error inesperado: %v", aerr)
		}
		if p.Page != 1 || p.PageSize != 50 {
			t.Errorf("defaults = page %d pageSize %d, esperaba 1/50", p.Page, p.PageSize)
		}
		if p.Desde != nil || p.Hasta != nil {
			t.Error("sin fechas: desde y hasta deben ser nil")
		}
	})

	t.Run("fechas válidas y desde ≤ hasta", func(t *testing.T) {
		p, aerr := ValidarEntrada("2026-10-01", "2026-10-31", "", "")
		if aerr != nil {
			t.Fatalf("error inesperado: %v", aerr)
		}
		if p.Desde == nil || p.Desde.Format("2006-01-02") != "2026-10-01" {
			t.Errorf("desde mal parseado: %v", p.Desde)
		}
		if p.Hasta == nil || p.Hasta.Format("2006-01-02") != "2026-10-31" {
			t.Errorf("hasta mal parseado: %v", p.Hasta)
		}
	})

	t.Run("desde solo (rango abierto por arriba)", func(t *testing.T) {
		p, aerr := ValidarEntrada("2026-10-01", "", "", "")
		if aerr != nil {
			t.Fatalf("error inesperado: %v", aerr)
		}
		if p.Desde == nil || p.Hasta != nil {
			t.Error("solo desde: hasta debe ser nil")
		}
	})

	t.Run("hasta solo (rango abierto por abajo)", func(t *testing.T) {
		p, aerr := ValidarEntrada("", "2026-10-31", "", "")
		if aerr != nil {
			t.Fatalf("error inesperado: %v", aerr)
		}
		if p.Desde != nil || p.Hasta == nil {
			t.Error("solo hasta: desde debe ser nil")
		}
	})

	errores := []struct {
		nombre                string
		desde, hasta, pg, pgs string
	}{
		{"page=0 → 422", "", "", "0", ""},
		{"page negativo → 422", "", "", "-3", ""},
		{"page no numérica → 422", "", "", "abc", ""},
		{"page_size=0 → 422", "", "", "", "0"},
		{"page_size=500 → 422", "", "", "", "500"},
		{"page_size=201 → 422", "", "", "", "201"},
		{"desde > hasta → 422", "2026-10-31", "2026-10-01", "", ""},
		{"desde malformada → 422", "2026-10-32", "", "", ""},        // día inexistente
		{"desde no estricta → 422", "2026-1-01", "", "", ""},         // un dígito
		{"hasta con hora → 422", "", "2026-10-31T00:00:00Z", "", ""}, // no es date plano
	}
	for _, c := range errores {
		t.Run(c.nombre, func(t *testing.T) {
			_, aerr := ValidarEntrada(c.desde, c.hasta, c.pg, c.pgs)
			if aerr == nil {
				t.Fatal("esperaba 422 ENTRADA_INVALIDA, llegó sin error")
			}
			if aerr.Status != 422 || aerr.Codigo != "ENTRADA_INVALIDA" {
				t.Errorf("error = %d %s, esperaba 422 ENTRADA_INVALIDA", aerr.Status, aerr.Codigo)
			}
		})
	}

	t.Run("page_size en extremos válidos 1 y 200", func(t *testing.T) {
		for _, pgs := range []string{"1", "200"} {
			if _, aerr := ValidarEntrada("", "", "", pgs); aerr != nil {
				t.Errorf("page_size=%s debía ser válido: %v", pgs, aerr)
			}
		}
	})
}

func TestParsearFechaParam(t *testing.T) {
	t.Run("formato estricto YYYY-MM-DD", func(t *testing.T) {
		f, aerr := ParsearFechaParam("2026-02-28")
		if aerr != nil {
			t.Fatalf("error inesperado: %v", aerr)
		}
		esperada := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
		if !f.Equal(esperada) {
			t.Errorf("fecha = %v, esperaba %v", f, esperada)
		}
	})

	t.Run("29 de febrero en año no bisiesto → 422", func(t *testing.T) {
		if _, aerr := ParsearFechaParam("2026-02-29"); aerr == nil {
			t.Error("esperaba 422 para 2026-02-29")
		}
	})
}
