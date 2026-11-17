// Rutas de spec-006 (HU-06: ficha, reparacion y ubicacion en taller).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/handlers/taller"
	servicio "si16/backend/internal/services/taller"
)

func RegisterSpec006(rg *gin.RouterGroup, d deps.Deps) {
	h := &taller.Handler{Svc: servicio.Nuevo(d.Pool, d.Estacion, nil)}
	grupo := rg.Group("/taller/bicicletas/:id",
		auth.RequireAuth(d.JWTSecret),
		auth.RequireRole("operario_taller"),
	)
	grupo.GET("", h.Ficha)
	grupo.POST("/intervenciones", h.RegistrarIntervencion)
	grupo.POST("/ubicacion", h.ConfirmarUbicacion)
}
