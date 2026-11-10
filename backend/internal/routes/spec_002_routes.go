package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	alquilerhandler "si16/backend/internal/handlers/alquiler"
	"si16/backend/internal/models"
	"si16/backend/internal/services/alquiler"
)

// RegisterSpec002 monta las rutas de HU-02 (Iker Roig). Desbloquear es solo del
// rol usuario (data-model §1); cada sesión consulta únicamente su propio viaje.
func RegisterSpec002(rg *gin.RouterGroup, d Deps) {
	h := &alquilerhandler.Handler{Svc: alquiler.Nuevo(d.Pool, d.Estacion, d.Log), Log: d.Log}
	g := rg.Group("/alquileres", d.Auth.RequireAuth())
	g.POST("/desbloqueo", auth.RequireRole(string(models.RolUsuario)), h.Desbloquear)
	g.GET("/activo", h.Activo)
}
