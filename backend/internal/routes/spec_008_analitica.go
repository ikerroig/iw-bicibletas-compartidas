// Rutas de spec-008 (HU-08: panel de analítica y exportación CSV).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/handlers/analitica"
)

// RegisterSpec008 monta GET /analitica/viajes y GET /analitica/export,
// restringidos al rol gestor (FR-005, SC-003; research D5).
func RegisterSpec008(rg *gin.RouterGroup, d deps.Deps) {
	rg.GET("/analitica/viajes",
		auth.RequireAuth(d.JWTSecret),
		auth.RequireRole("gestor"),
		analitica.HandlerAgregados(d),
	)
	rg.GET("/analitica/export",
		auth.RequireAuth(d.JWTSecret),
		auth.RequireRole("gestor"),
		analitica.HandlerExport(d),
	)
}
