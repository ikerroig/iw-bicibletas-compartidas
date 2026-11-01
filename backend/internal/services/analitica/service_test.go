// Unit tests del servicio de analítica HU-08 (T005): validación de entrada,
// zero-fill de buckets y mapeo de día de semana (lunes=0…domingo=6).
// Sin BD: solo funciones puras del paquete analitica.
package analitica

import (
	"testing"
	"time"
)

func TestValidarEntradaOK(t *testing.T) {
	p, aerr := ValidarEntrada("2026-10-01", "2026-10-31")
	if aerr != nil {
		t.Fatalf("rango válido devolvió error: %v", aerr)
	}
	if p.Desde.Format("2006-01-02") != "2026-10-01" || p.Hasta.Format("2006-01-02") != "2026-10-31" {
		t.Errorf("fechas parseadas incorrectamente: %v / %v", p.Desde, p.Hasta)
	}
}

func TestValidarEntradaMismoDia(t *testing.T) {
	if _, aerr := ValidarEntrada("2026-10-15", "2026-10-15"); aerr != nil {
		t.Errorf("desde == hasta es rango válido de un día: %v", aerr)
	}
}

func TestValidarEntradaInvalida(t *testing.T) {
	casos := []struct{ desde, hasta, motivo string }{
		{"", "2026-10-31", "desde ausente"},
		{"2026-10-01", "", "hasta ausente"},
		{"", "", "ambos ausentes"},
		{"01/10/2026", "2026-10-31", "formato dd/mm/yyyy"},
		{"2026-13-40", "2026-10-31", "mes/día inexistentes"},
		{"2026-10-31", "2026-10-01", "rango invertido"},
		{"2026-5-1", "2026-10-31", "dígito suelto (no estricto)"},
		{"not-a-date", "2026-10-31", "texto arbitrario"},
	}
	for _, c := range casos {
		_, aerr := ValidarEntrada(c.desde, c.hasta)
		if aerr == nil {
			t.Errorf("%s: se esperaba 422, se obtuvo nil", c.motivo)
			continue
		}
		if aerr.Status != 422 || aerr.Codigo != "ENTRADA_INVALIDA" {
			t.Errorf("%s: status/codigo = %d/%s, esperaba 422/ENTRADA_INVALIDA",
				c.motivo, aerr.Status, aerr.Codigo)
		}
	}
}

func TestBucketsCeroFill(t *testing.T) {
	r := nuevosBuckets()
	if len(r.PorHora) != 24 {
		t.Errorf("por_hora tiene %d posiciones, esperaba 24", len(r.PorHora))
	}
	if len(r.PorDiaSemana) != 7 {
		t.Errorf("por_dia_semana tiene %d posiciones, esperaba 7", len(r.PorDiaSemana))
	}
	for i, v := range r.PorHora {
		if v != 0 {
			t.Errorf("por_hora[%d] = %d, esperaba 0 sin datos", i, v)
		}
	}
	for i, v := range r.PorDiaSemana {
		if v != 0 {
			t.Errorf("por_dia_semana[%d] = %d, esperaba 0 sin datos", i, v)
		}
	}
	if r.TotalViajes != 0 {
		t.Errorf("total_viajes = %d, esperaba 0", r.TotalViajes)
	}
}

// TestMapeoDiaSemana: domingo → 6, lunes → 0, sábado → 5 (lunes=0 … domingo=6).
func TestMapeoDiaSemana(t *testing.T) {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatalf("Europe/Madrid: %v", err)
	}
	casos := []struct {
		fecha    time.Time
		esperado int
		nombre   string
	}{
		{time.Date(2026, 10, 26, 12, 0, 0, 0, madrid), 0, "lunes"},
		{time.Date(2026, 10, 27, 12, 0, 0, 0, madrid), 1, "martes"},
		{time.Date(2026, 10, 28, 12, 0, 0, 0, madrid), 2, "miércoles"},
		{time.Date(2026, 10, 29, 12, 0, 0, 0, madrid), 3, "jueves"},
		{time.Date(2026, 10, 30, 12, 0, 0, 0, madrid), 4, "viernes"},
		{time.Date(2026, 10, 31, 12, 0, 0, 0, madrid), 5, "sábado"},
		{time.Date(2026, 11, 1, 12, 0, 0, 0, madrid), 6, "domingo"},
	}
	for _, c := range casos {
		if got := indiceDiaSemana(c.fecha); got != c.esperado {
			t.Errorf("%s %v → índice %d, esperaba %d",
				c.nombre, c.fecha, got, c.esperado)
		}
	}
}

func TestNombreFichero(t *testing.T) {
	if got := NombreFichero("2026-10-01", "2026-10-31"); got != "viajes_20261001_20261031.csv" {
		t.Errorf("NombreFichero = %q", got)
	}
}
