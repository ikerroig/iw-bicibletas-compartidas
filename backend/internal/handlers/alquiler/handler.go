// Package alquiler expone POST /alquileres/desbloqueo y GET /alquileres/activo
// (spec-002).
package alquiler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/httperr"
	svc "si16/backend/internal/services/alquiler"
)

// Handler HTTP de spec-002.
type Handler struct {
	Svc *svc.Servicio
	Log *slog.Logger
}

// rechazos traduce cada causa del servicio a su respuesta (FR-002…FR-004).
var rechazos = []struct {
	causa   error
	status  int
	codigo  string
	mensaje string
}{
	{svc.ErrAbonoVencido, http.StatusForbidden, httperr.AbonoVencido, "Tu abono anual no está vigente"},
	{svc.ErrViajeActivo, http.StatusConflict, httperr.ViajeActivo, "Ya tienes un viaje en curso"},
	{svc.ErrAnclajeNoExiste, http.StatusNotFound, httperr.AnclajeNoExiste, "No existe ningún anclaje con ese código"},
	{svc.ErrEstacionFueraServicio, http.StatusConflict, httperr.AnclajeNoDisponible, "La estación de este anclaje está fuera de servicio"},
	{svc.ErrAnclajeFueraServicio, http.StatusConflict, httperr.AnclajeNoDisponible, "Este anclaje está fuera de servicio"},
	{svc.ErrAnclajeVacio, http.StatusConflict, httperr.AnclajeNoDisponible, "Este anclaje no tiene ninguna bicicleta"},
	{svc.ErrBicicletaNoDisponible, http.StatusConflict, httperr.BicicletaNoDisponible, "La bicicleta de este anclaje no está disponible"},
}

// Desbloquear inicia un viaje con el código de un anclaje (FR-001, FR-005).
func (h *Handler) Desbloquear(c *gin.Context) {
	var cuerpo struct {
		CodigoAnclaje string `json:"codigo_anclaje"`
	}
	err := c.ShouldBindJSON(&cuerpo)
	codigo := strings.TrimSpace(cuerpo.CodigoAnclaje)
	if err != nil || codigo == "" {
		httperr.Responder(c, http.StatusUnprocessableEntity, httperr.Validacion, "Introduce el código del anclaje")
		return
	}
	sesion, ok := auth.SesionDe(c)
	if !ok {
		httperr.Responder(c, http.StatusUnauthorized, httperr.Unautenticado, "Inicia sesión para continuar")
		return
	}

	viaje, err := h.Svc.Desbloquear(c.Request.Context(), sesion.UsuarioID, codigo)
	if err != nil {
		for _, r := range rechazos {
			if errors.Is(err, r.causa) {
				httperr.Responder(c, r.status, r.codigo, r.mensaje)
				return
			}
		}
		h.Log.Error("POST /alquileres/desbloqueo", "usuario", sesion.UsuarioID, "error", err)
		httperr.ErrorInterno(c)
		return
	}
	c.JSON(http.StatusCreated, viaje)
}

// Activo devuelve el viaje en curso de la sesión, o null.
func (h *Handler) Activo(c *gin.Context) {
	sesion, ok := auth.SesionDe(c)
	if !ok {
		httperr.Responder(c, http.StatusUnauthorized, httperr.Unautenticado, "Inicia sesión para continuar")
		return
	}
	viaje, err := h.Svc.Activo(c.Request.Context(), sesion.UsuarioID)
	if err != nil {
		h.Log.Error("GET /alquileres/activo", "usuario", sesion.UsuarioID, "error", err)
		httperr.ErrorInterno(c)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"viaje": viaje})
}
