// Rutas de spec-007 (HU-07: panel de alertas de balanceo y traslados).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/handlers/balanceo"
)

// RegisterSpec007 monta los endpoints de logística: ambos exigen el rol
// operario_logistica (research D5; sin token → 401, otro rol → 403).
func RegisterSpec007(rg *gin.RouterGroup, d deps.Deps) {
	rg.GET("/balanceo/alertas",
		auth.RequireAuth(d.JWTSecret),
		auth.RequireRole("operario_logistica"),
		balanceo.HandlerPanel(d),
	)
	rg.POST("/balanceo/traslados",
		auth.RequireAuth(d.JWTSecret),
		auth.RequireRole("operario_logistica"),
		balanceo.HandlerTraslado(d),
	)
}
