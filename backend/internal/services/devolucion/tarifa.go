package devolucion

import "math"

// TramoGratuitoSegundos: los primeros 30 minutos del viaje no se cobran
// (regla tarifaria R8 / FR-004, decisión de investigación D1).
const TramoGratuitoSegundos int64 = 1800

// CalcularCoste aplica la regla tarifaria de la HU-03:
//   - duración ≤ 30 min → 0,00 € (tramo gratuito);
//   - en otro caso ceil((segundos − 1800) / 60) × tarifa por minuto,
//     redondeado a céntimos (NUMERIC(10,2)) y con piso 0,00 nunca negativo.
func CalcularCoste(segundos int64, tarifaMinuto float64) float64 {
	if segundos <= TramoGratuitoSegundos || tarifaMinuto <= 0 {
		return 0.00
	}
	minutosCobrados := math.Ceil(float64(segundos-TramoGratuitoSegundos) / 60.0)
	coste := minutosCobrados * tarifaMinuto
	if coste < 0 {
		return 0.00
	}
	// Piso a céntimos: evita binario-15×0.2=3.0000000000000004.
	return math.Round(coste*100) / 100
}

// DuracionMinutos devuelve ceil(segundos/60) (data-model R8): 23 min 47 s
// cuentan como 24 y 30 min 01 s como 31, coherente con CalcularCoste.
func DuracionMinutos(segundos int64) int {
	if segundos <= 0 {
		return 0
	}
	return int((segundos + 59) / 60)
}
