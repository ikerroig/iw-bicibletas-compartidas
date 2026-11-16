// Rutas de spec-005 (HU-05: reporte e inhabilitacion por averia).
package routes

import (
	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/cache"
	"si16/backend/internal/deps"
	incidenciahandler "si16/backend/internal/handlers/incidencias"
	"si16/backend/internal/services/estacion"
	"si16/backend/internal/services/incidencias"
)

// RegisterSpec005 monta el formulario de usuario y el panel de taller.
func RegisterSpec005(rg *gin.RouterGroup, d deps.Deps) {
	est := d.Estacion
	if est == nil && d.RDB != nil {
		est = estacion.Nuevo(d.Pool, cache.NuevoDesdeRedis(d.RDB), nil)
	}
	h := &incidenciahandler.Handler{Svc: incidencias.Nuevo(d.Pool, est)}

	usuario := rg.Group("/averias", auth.RequireAuth(d.JWTSecret), auth.RequireRole("usuario"))
	usuario.POST("", h.Reportar)
	// Alias con el nombre de la entidad, util para clientes que usan el panel
	// de incidencias como punto de entrada del reporte.
	reporte := rg.Group("/incidencias", auth.RequireAuth(d.JWTSecret), auth.RequireRole("usuario"))
	reporte.POST("", h.Reportar)

	taller := rg.Group("/incidencias", auth.RequireAuth(d.JWTSecret), auth.RequireRole("operario_taller"))
	taller.GET("", h.Pendientes)
	taller.GET("/pendientes", h.Pendientes)
}
