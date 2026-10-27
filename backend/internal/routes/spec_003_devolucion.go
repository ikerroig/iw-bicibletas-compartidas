// Rutas de spec-003 (HU-03: devolución de bicicleta y cálculo de tarificación).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/handlers/devolucion"
)

// RegisterSpec003 monta POST /devoluciones (rol: usuario).
func RegisterSpec003(rg *gin.RouterGroup, d deps.Deps) {
	rg.POST("/devoluciones",
		auth.RequireAuth(d.JWTSecret),
		auth.RequireRole("usuario"),
		devolucion.Handler(d),
	)
}
