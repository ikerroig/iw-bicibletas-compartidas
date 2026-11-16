// Package incidencias expone los endpoints de HU-05.
package incidencias

import (
	"errors"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/apierr"
	"si16/backend/internal/auth"
	svc "si16/backend/internal/services/incidencias"
)

type Handler struct {
	Svc *svc.Servicio
}

type reporteRequest struct {
	BicicletaID     int64  `json:"bicicleta_id"`
	BicicletaCodigo string `json:"bicicleta_codigo"`
	EstacionID      int64  `json:"estacion_id"`
	TipoFallo       string `json:"tipo_fallo"`
	Descripcion     string `json:"descripcion"`
}

// Reportar crea la incidencia y confirma el registro al usuario.
func (h *Handler) Reportar(c *gin.Context) {
	var req reporteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responderValidacion(c, "El cuerpo del reporte no es valido")
		return
	}
	req.BicicletaCodigo = strings.TrimSpace(req.BicicletaCodigo)
	req.Descripcion = strings.TrimSpace(req.Descripcion)
	req.TipoFallo = normalizarTipo(req.TipoFallo)
	if !reporteValido(req) {
		responderValidacion(c, "Indica bicicleta, estacion, tipo de fallo y una descripcion")
		return
	}

	sesionID := auth.UsuarioID(c)
	resultado, err := h.Svc.Reportar(c.Request.Context(), sesionID, svc.Reporte{
		BicicletaID: req.BicicletaID, BicicletaCodigo: req.BicicletaCodigo,
		EstacionID: req.EstacionID, TipoFallo: req.TipoFallo, Descripcion: req.Descripcion,
	})
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrBicicletaNoEncontrada):
			apierr.Respond(c, http.StatusNotFound, "BICICLETA_NO_ENCONTRADA", "No existe esa bicicleta")
		case errors.Is(err, svc.ErrEstacionNoEncontrada):
			apierr.Respond(c, http.StatusNotFound, "ESTACION_NO_ENCONTRADA", "No existe esa estacion")
		case errors.Is(err, svc.ErrBicicletaOtraEstacion):
			apierr.Respond(c, http.StatusConflict, "BICICLETA_EN_OTRA_ESTACION", "La bicicleta no esta en la estacion indicada")
		case errors.Is(err, svc.ErrBicicletaEnTaller):
			apierr.Respond(c, http.StatusConflict, "BICICLETA_EN_TALLER", "La bicicleta ya esta en el taller")
		default:
			apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error interno del servidor")
		}
		return
	}

	status := http.StatusCreated
	mensaje := "Averia registrada correctamente"
	if resultado.Repetido {
		status = http.StatusOK
		mensaje = "La averia ya estaba registrada"
	}
	c.JSON(status, gin.H{
		"incidencia": resultado.Incidencia,
		"mensaje":    mensaje,
		"repetido":   resultado.Repetido,
	})
}

// Pendientes devuelve el panel de trabajo del equipo de taller.
func (h *Handler) Pendientes(c *gin.Context) {
	items, err := h.Svc.Pendientes(c.Request.Context())
	if err != nil {
		apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error interno del servidor")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func responderValidacion(c *gin.Context, mensaje string) {
	apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA", mensaje)
}

func reporteValido(req reporteRequest) bool {
	return (req.BicicletaID > 0 || req.BicicletaCodigo != "") && req.EstacionID > 0 &&
		esTipoValido(req.TipoFallo) && req.Descripcion != "" &&
		utf8.RuneCountInString(req.Descripcion) <= 2000 &&
		!strings.ContainsFunc(req.Descripcion, unicode.IsControl)
}

func esTipoValido(tipo string) bool {
	switch tipo {
	case "frenos", "cadena", "neumatico", "otro":
		return true
	default:
		return false
	}
}

func normalizarTipo(tipo string) string {
	tipo = strings.ToLower(strings.TrimSpace(tipo))
	if tipo == "neumatico" || tipo == "neumático" {
		return "neumatico"
	}
	return tipo
}
