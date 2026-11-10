// Unit tests del escritor CSV de la HU-08 (T006): BOM UTF-8, cabecera,
// CRLF, comillas RFC 4180, decimal con punto y rango vacío (solo cabecera).
// Sin BD: filas simuladas a bytes.Buffer.
package analitica

import (
	"bytes"
	"strings"
	"testing"
)

func TestCSVVacioSoloCabecera(t *testing.T) {
	var buf bytes.Buffer
	if err := escribirCSV(&buf, nil); err != nil {
		t.Fatalf("escribirCSV vacío: %v", err)
	}
	b := buf.Bytes()
	if len(b) < 3 || b[0] != 0xEF || b[1] != 0xBB || b[2] != 0xBF {
		t.Errorf("primeros bytes = % X, esperaba BOM EF BB BF", b[:min(3, len(b))])
	}
	cuerpo := string(b[3:])
	esperada := "id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur\r\n"
	if cuerpo != esperada {
		t.Errorf("cuerpo = %q, esperaba solo cabecera %q", cuerpo, esperada)
	}
}

func TestCSVFilasCompletas(t *testing.T) {
	var buf bytes.Buffer
	filas := []FilaCSV{
		{ID: 1201, FechaHora: "2026-10-05 08:15", Origen: "Estació Universitat", Destino: "Estació Mercat", Duracion: 12, Coste: 0},
		{ID: 1202, FechaHora: "2026-10-05 17:40", Origen: "Estació Mercat", Destino: "Estació Nord", Duracion: 45, Coste: 3.00},
	}
	if err := escribirCSV(&buf, filas); err != nil {
		t.Fatalf("escribirCSV: %v", err)
	}
	b := buf.Bytes()
	if b[0] != 0xEF || b[1] != 0xBB || b[2] != 0xBF {
		t.Error("falta el BOM UTF-8")
	}
	cuerpo := string(b[3:])

	if !strings.HasPrefix(cuerpo, "id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur\r\n") {
		t.Errorf("cabecera incorrecta: %q", cuerpo[:80])
	}
	// CRLF en todas las líneas y coma como separador.
	if !strings.Contains(cuerpo, "1201,2026-10-05 08:15,Estació Universitat,Estació Mercat,12,0.00\r\n") {
		t.Errorf("fila 1201 malformada o sin CRLF:\n%s", cuerpo)
	}
	if !strings.Contains(cuerpo, "1202,2026-10-05 17:40,Estació Mercat,Estació Nord,45,3.00\r\n") {
		t.Errorf("fila 1202 malformada:\n%s", cuerpo)
	}
	// Tildes/ñ intactos.
	if !strings.Contains(cuerpo, "Estació Universitat") || !strings.Contains(cuerpo, "Estació Mercat") {
		t.Error("nombres con tildes dañados")
	}
}

func TestCSVComillaRFC4180(t *testing.T) {
	var buf bytes.Buffer
	filas := []FilaCSV{
		{ID: 1, FechaHora: "2026-10-05 08:15", Origen: "Plaça, Major", Destino: "Estació \"Nord\"", Duracion: 10, Coste: 0},
	}
	if err := escribirCSV(&buf, filas); err != nil {
		t.Fatalf("escribirCSV: %v", err)
	}
	cuerpo := string(buf.Bytes()[3:])
	if !strings.Contains(cuerpo, `"Plaça, Major"`) {
		t.Errorf("campo con coma sin comillas RFC 4180:\n%s", cuerpo)
	}
	if !strings.Contains(cuerpo, `"Estació ""Nord"""`) {
		t.Errorf("comillas internas sin escapar RFC 4180:\n%s", cuerpo)
	}
}

func TestCSVCosteDecimalPunto(t *testing.T) {
	var buf bytes.Buffer
	filas := []FilaCSV{
		{ID: 1, FechaHora: "2026-10-05 08:15", Origen: "A", Destino: "B", Duracion: 45, Coste: 3.00},
	}
	if err := escribirCSV(&buf, filas); err != nil {
		t.Fatalf("escribirCSV: %v", err)
	}
	cuerpo := string(buf.Bytes()[3:])
	if !strings.Contains(cuerpo, ",3.00\r\n") {
		t.Errorf("coste_eur sin punto decimal y 2 decimales:\n%s", cuerpo)
	}
	if strings.Contains(cuerpo, ",3,00") {
		t.Errorf("coste_eur con coma decimal (prohibido con separador coma):\n%s", cuerpo)
	}
}
