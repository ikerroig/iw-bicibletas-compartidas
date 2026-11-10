// Rutas de spec-004 (HU-04: historial de viajes y desglose mensual).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/handlers/viajes"
)

// RegisterSpec004 monta GET /viajes (cualquier rol autenticado; siempre
// es el historial propio del token — FR-005).
func RegisterSpec004(rg *gin.RouterGroup, d deps.Deps) {
	rg.GET("/viajes",
		auth.RequireAuth(d.JWTSecret),
		viajes.Handler(d),
	)
}
