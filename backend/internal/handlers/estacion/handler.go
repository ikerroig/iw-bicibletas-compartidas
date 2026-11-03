// Package estacion expone GET /estaciones y GET /estaciones/{id} (spec-001).
package estacion

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/httperr"
	svc "si16/backend/internal/services/estacion"
)

// Handler HTTP de spec-001.
type Handler struct {
	Svc *svc.Servicio
	Log *slog.Logger
}

type respuestaSnapshot struct {
	ActualizadoEn time.Time            `json:"actualizado_en"`
	Estaciones    []svc.EstacionEstado `json:"estaciones"`
}

// Listar: snapshot de la red con marca de frescura (FR-001, FR-002, FR-005).
func (h *Handler) Listar(c *gin.Context) {
	estaciones, actualizado, err := h.Svc.Snapshot(c.Request.Context())
	if err != nil {
		h.Log.Error("GET /estaciones", "error", err)
		httperr.ErrorInterno(c)
		return
	}
	if estaciones == nil {
		estaciones = []svc.EstacionEstado{}
	}
	c.Header("Cache-Control", "no-store") // estado en vivo: nunca desde caché HTTP
	c.JSON(http.StatusOK, respuestaSnapshot{ActualizadoEn: actualizado, Estaciones: estaciones})
}

// Detalle de una estación (FR-004).
func (h *Handler) Detalle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httperr.Responder(c, http.StatusNotFound, httperr.NoEncontrado, "Estación no encontrada")
		return
	}
	d, err := h.Svc.Detalle(c.Request.Context(), id)
	if errors.Is(err, svc.ErrNoEncontrada) {
		httperr.Responder(c, http.StatusNotFound, httperr.NoEncontrado, "Estación no encontrada")
		return
	}
	if err != nil {
		h.Log.Error("GET /estaciones/:id", "id", id, "error", err)
		httperr.ErrorInterno(c)
		return
	}
	if d.Anclajes == nil {
		d.Anclajes = []svc.AnclajeDetalle{}
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, d)
}
