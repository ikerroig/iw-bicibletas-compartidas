package cache

import (
	"strconv"
	"time"
)

// Contrato de claves compartido (specs/spec-001/contracts/redis-state.md).
// Cualquier clave nueva o cambio de formato requiere acuerdo del equipo
// (constitución, principio III).
const (
	// TTLEstado es la red de seguridad de expiración del hash de estación.
	TTLEstado = 120 * time.Second

	// ClaveUltimoBarrido es interna del worker refresco_estado (observabilidad).
	ClaveUltimoBarrido = "si16:worker:refresco:ultimo"

	// Campos del HASH estacion:{id}:estado.
	CampoBicisDisponibles = "bicis_disponibles"
	CampoAnclajesLibres   = "anclajes_libres"
	CampoOcupacionPct     = "ocupacion_pct"
	CampoEstado           = "estado"
	CampoActualizadoEn    = "actualizado_en"

	// FormatoActualizadoEn es RFC 3339 en UTC con nanosegundos de ancho fijo.
	// Cualquier escritor puede usar RFC 3339 UTC con o sin fracción (contrato):
	// el script de escritura normaliza ambos antes de comparar.
	FormatoActualizadoEn = "2006-01-02T15:04:05.000000000Z07:00"
)

// ClaveEstado devuelve "estacion:{id}:estado".
func ClaveEstado(id int64) string {
	return "estacion:" + strconv.FormatInt(id, 10) + ":estado"
}
