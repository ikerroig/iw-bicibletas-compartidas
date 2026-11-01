// Registro compartido de rutas del backend.
//
// ARCHIVO COMPARTIDO ENTRE RAMAS: cada spec añade UNA línea con su
// RegisterSpec00N (principio V — conflictos de merge triviales).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
)

// Register monta todas las rutas bajo el grupo /api/v1.
func Register(rg *gin.RouterGroup, d deps.Deps) {
	auth.RegisterAuth(rg, d) // base compartida: POST /auth/login, GET /auth/me
	RegisterSpec003(rg, d)   // spec-003 (Alex Morell): HU-03 devolución y tarificación
	RegisterSpec004(rg, d)   // spec-004 (Alex Morell): HU-04 historial de viajes
	RegisterSpec007(rg, d)   // spec-007 (HU-07): alertas de balanceo y traslados
	RegisterSpec008(rg, d)   // spec-008 (HU-08): analítica de uso y exportación CSV
}
