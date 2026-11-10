// Tests unitarios de validación de traslado (HU-07 T006): 422 ENTRADA_INVALIDA
// sin BD. TDD: deben FALLAR hasta implementar traslados.go. La garantía de que
// "ninguna validación altera datos" se verifica además en T7g (integración,
// recuentos idénticos antes/después).
package balanceo

import (
	"testing"

	"si16/backend/internal/models"
)

func TestParseTrasladoEntradaInvalida(t *testing.T) {
	casos := []struct {
		nombre string
		body   string
	}{
		{"JSON malformado", `{"estacion_origen_id":`},
		{"unidades < 1 (cero)", `{"estacion_origen_id":4,"estacion_destino_id":3,"unidades":0}`},
		{"unidades negativas", `{"estacion_origen_id":4,"estacion_destino_id":3,"unidades":-5}`},
		{"origen == destino", `{"estacion_origen_id":4,"estacion_destino_id":4,"unidades":3}`},
		{"campo unidades ausente", `{"estacion_origen_id":4,"estacion_destino_id":3}`},
		{"campo origen ausente", `{"estacion_destino_id":3,"unidades":2}`},
		{"campo destino ausente", `{"estacion_origen_id":4,"unidades":2}`},
		{"unidades no numéricas", `{"estacion_origen_id":4,"estacion_destino_id":3,"unidades":"seis"}`},
		{"payload no es objeto", `[1,2,3]`},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req, e := ParseTraslado([]byte(c.body))
			if e == nil {
				t.Fatalf("ParseTraslado(%s) debió devolver 422, got req=%+v", c.body, req)
			}
			if e.Status != 422 {
				t.Errorf("status = %d, esperado 422", e.Status)
			}
			if e.Codigo != "ENTRADA_INVALIDA" {
				t.Errorf("codigo = %q, esperado ENTRADA_INVALIDA", e.Codigo)
			}
		})
	}
}

func TestParseTrasladoValido(t *testing.T) {
	req, e := ParseTraslado([]byte(`{"estacion_origen_id":4,"estacion_destino_id":3,"unidades":6}`))
	if e != nil {
		t.Fatalf("payload válido rechazado: %v", e)
	}
	esperado := models.TrasladoRequest{EstacionOrigenID: 4, EstacionDestinoID: 3, Unidades: 6}
	if req != esperado {
		t.Errorf("req = %+v, esperado %+v", req, esperado)
	}
}
