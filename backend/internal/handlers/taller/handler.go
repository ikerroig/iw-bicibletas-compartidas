// Package taller expone los endpoints de HU-06.
package taller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/apierr"
	"si16/backend/internal/auth"
	svc "si16/backend/internal/services/taller"
)

type Handler struct {
	Svc *svc.Servicio
}

type intervencionRequest struct {
	PiezasSustituidas *string `json:"piezas_sustituidas"`
	NotasTaller       *string `json:"notas_taller"`
	MarcarOperativa   bool    `json:"marcar_operativa"`
}

type ubicacionRequest struct {
	CodigoAnclaje string `json:"codigo_anclaje"`
}

func (h *Handler) Ficha(c *gin.Context) {
	id, ok := bicicletaID(c)
	if !ok {
		apierr.Respond(c, http.StatusNotFound, "BICICLETA_NO_ENCONTRADA", "No existe esa bicicleta")
		return
	}
	ficha, err := h.Svc.Ficha(c.Request.Context(), id)
	if errors.Is(err, svc.ErrBicicletaNoEncontrada) {
		apierr.Respond(c, http.StatusNotFound, "BICICLETA_NO_ENCONTRADA", "No existe esa bicicleta")
		return
	}
	if err != nil {
		apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error interno del servidor")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, ficha)
}

func (h *Handler) RegistrarIntervencion(c *gin.Context) {
	id, ok := bicicletaID(c)
	if !ok {
		apierr.Respond(c, http.StatusNotFound, "BICICLETA_NO_ENCONTRADA", "No existe esa bicicleta")
		return
	}
	var req intervencionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA", "El cuerpo del parte no es valido")
		return
	}
	req.PiezasSustituidas = limpiarTexto(req.PiezasSustituidas)
	req.NotasTaller = limpiarTexto(req.NotasTaller)
	if req.PiezasSustituidas == nil && req.NotasTaller == nil {
		apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA", "Indica las piezas sustituidas o las notas del taller")
		return
	}

	incidencia, err := h.Svc.RegistrarIntervencion(c.Request.Context(), auth.UsuarioID(c), id, svc.Intervencion{
		PiezasSustituidas: req.PiezasSustituidas,
		NotasTaller:       req.NotasTaller,
		MarcarOperativa:   req.MarcarOperativa,
	})
	if err != nil {
		responderErrorTaller(c, err)
		return
	}
	c.JSON(http.StatusCreated, incidencia)
}

func (h *Handler) ConfirmarUbicacion(c *gin.Context) {
	id, ok := bicicletaID(c)
	if !ok {
		apierr.Respond(c, http.StatusNotFound, "BICICLETA_NO_ENCONTRADA", "No existe esa bicicleta")
		return
	}
	var req ubicacionRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.CodigoAnclaje) == "" {
		apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA", "Indica el codigo del anclaje")
		return
	}
	estado, err := h.Svc.ConfirmarUbicacion(c.Request.Context(), id, req.CodigoAnclaje)
	if err != nil {
		responderErrorTaller(c, err)
		return
	}
	c.JSON(http.StatusOK, estado)
}

func bicicletaID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func limpiarTexto(valor *string) *string {
	if valor == nil {
		return nil
	}
	texto := strings.TrimSpace(*valor)
	if texto == "" || utf8.RuneCountInString(texto) > 2000 ||
		strings.ContainsFunc(texto, unicode.IsControl) {
		return nil
	}
	return &texto
}

func responderErrorTaller(c *gin.Context, err error) {
	switch {
	case errors.Is(err, svc.ErrBicicletaNoEncontrada):
		apierr.Respond(c, http.StatusNotFound, "BICICLETA_NO_ENCONTRADA", "No existe esa bicicleta")
	case errors.Is(err, svc.ErrAnclajeNoEncontrado):
		apierr.Respond(c, http.StatusNotFound, "ANCLAJE_NO_ENCONTRADO", "No existe ese anclaje")
	case errors.Is(err, svc.ErrAnclajeOcupado):
		apierr.Respond(c, http.StatusConflict, "ANCLAJE_OCUPADO", "El anclaje esta ocupado")
	case errors.Is(err, svc.ErrAnclajeFueraServicio):
		apierr.Respond(c, http.StatusConflict, "ANCLAJE_FUERA_DE_SERVICIO", "El anclaje esta fuera de servicio")
	case errors.Is(err, svc.ErrEstacionFueraServicio):
		apierr.Respond(c, http.StatusConflict, "ESTACION_FUERA_DE_SERVICIO", "La estacion esta fuera de servicio")
	case errors.Is(err, svc.ErrBicicletaEnUso):
		apierr.Respond(c, http.StatusConflict, "BICICLETA_EN_USO", "La bicicleta tiene un viaje activo")
	case errors.Is(err, svc.ErrBicicletaNoDisponible):
		apierr.Respond(c, http.StatusConflict, "BICICLETA_NO_DISPONIBLE", "La bicicleta no esta reparada")
	case errors.Is(err, svc.ErrBicicletaYaUbicada):
		apierr.Respond(c, http.StatusConflict, "BICICLETA_YA_UBICADA", "La bicicleta ya esta ubicada")
	default:
		apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error interno del servidor")
	}
}
