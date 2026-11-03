package routes

import (
	"github.com/gin-gonic/gin"

	estacionhandler "si16/backend/internal/handlers/estacion"
)

// RegisterSpec001 monta las rutas de HU-01 (Iker Roig): cualquier usuario
// autenticado puede consultar la disponibilidad (Assumptions de la spec).
func RegisterSpec001(rg *gin.RouterGroup, d Deps) {
	h := &estacionhandler.Handler{Svc: d.Estacion, Log: d.Log}
	g := rg.Group("/estaciones", d.Auth.RequireAuth())
	g.GET("", h.Listar)
	g.GET("/:id", h.Detalle)
}
