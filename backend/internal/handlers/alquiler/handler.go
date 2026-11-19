// Package alquiler expone POST /alquileres/desbloqueo y GET /alquileres/activo
// (spec-002).
package alquiler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

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
	{svc.ErrViajeActivo, http.StatusConflict, httperr.ViajeActivo, "Ya tienes un viaje en curso"},
	{svc.ErrAnclajeNoExiste, http.StatusNotFound, httperr.AnclajeNoExiste, "No existe ningún anclaje con ese código"},
	{svc.ErrEstacionFueraServicio, http.StatusConflict, httperr.AnclajeNoDisponible, "La estación de este anclaje está fuera de servicio"},
	{svc.ErrAnclajeFueraServicio, http.StatusConflict, httperr.AnclajeNoDisponible, "Este anclaje está fuera de servicio"},
	{svc.ErrAnclajeVacio, http.StatusConflict, httperr.AnclajeNoDisponible, "Este anclaje no tiene ninguna bicicleta"},
	{svc.ErrBicicletaNoDisponible, http.StatusConflict, httperr.BicicletaNoDisponible, "La bicicleta de este anclaje no está disponible"},
}

// codigoAbono traduce la causa del abono a su código del contrato (D15).
var codigoAbono = map[error]string{
	svc.ErrSinAbono:      httperr.SinAbono,
	svc.ErrAbonoCaducado: httperr.AbonoVencido,
}

// maxCodigo coincide con el límite de POST /devoluciones.
const maxCodigo = 64

// Mensajes del 422 (research D2, hallazgo U2 del analyze).
const (
	mensajeFaltaCodigo    = "Introduce el código del anclaje"
	mensajeCodigoInvalido = "El código no es válido"
)

// normalizarCodigo aplica FR-001: los códigos se guardan en mayúsculas. Los
// caracteres de control se rechazan aquí porque PostgreSQL no admite NUL en
// texto. Devuelve el mensaje del 422 si el código no sirve.
func normalizarCodigo(crudo string) (codigo, problema string) {
	codigo = strings.ToUpper(strings.TrimSpace(crudo))
	switch {
	case codigo == "":
		return "", mensajeFaltaCodigo
	case utf8.RuneCountInString(codigo) > maxCodigo || strings.ContainsFunc(codigo, unicode.IsControl):
		return "", mensajeCodigoInvalido
	}
	return codigo, ""
}

// Desbloquear inicia un viaje con el código de un anclaje (FR-001, FR-005).
func (h *Handler) Desbloquear(c *gin.Context) {
	var cuerpo struct {
		CodigoAnclaje string `json:"codigo_anclaje"`
	}
	if err := c.ShouldBindJSON(&cuerpo); err != nil {
		httperr.Responder(c, http.StatusUnprocessableEntity, httperr.Validacion, mensajeFaltaCodigo)
		return
	}
	codigo, problema := normalizarCodigo(cuerpo.CodigoAnclaje)
	if problema != "" {
		httperr.Responder(c, http.StatusUnprocessableEntity, httperr.Validacion, problema)
		return
	}
	sesion, _ := auth.SesionDe(c)

	viaje, err := h.Svc.Desbloquear(c.Request.Context(), sesion.UsuarioID, codigo)
	var abono *svc.RechazoAbono
	if errors.As(err, &abono) {
		httperr.Responder(c, http.StatusForbidden, codigoAbono[abono.Causa], abono.Mensaje)
		return
	}
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

// estadoAbono es el objeto `abono` de GET /alquileres/activo (FR-010, D16).
type estadoAbono struct {
	Vigente bool    `json:"vigente"`
	Codigo  *string `json:"codigo"`
	Mensaje *string `json:"mensaje"`
}

func nuevoEstadoAbono(a svc.CausaAbono) estadoAbono {
	if a.Vigente {
		return estadoAbono{Vigente: true}
	}
	codigo, mensaje := codigoAbono[a.Causa], a.Mensaje
	return estadoAbono{Codigo: &codigo, Mensaje: &mensaje}
}

// Activo devuelve el viaje en curso de la sesión, o null, y el estado del
// abono para avisar antes de desbloquear (FR-010).
func (h *Handler) Activo(c *gin.Context) {
	sesion, _ := auth.SesionDe(c)
	ctx := c.Request.Context()
	viaje, err := h.Svc.Activo(ctx, sesion.UsuarioID)
	if err != nil {
		h.Log.Error("GET /alquileres/activo", "usuario", sesion.UsuarioID, "error", err)
		httperr.ErrorInterno(c)
		return
	}
	abono, err := h.Svc.Abono(ctx, sesion.UsuarioID)
	if err != nil {
		h.Log.Error("GET /alquileres/activo: abono", "usuario", sesion.UsuarioID, "error", err)
		httperr.ErrorInterno(c)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, struct {
		Viaje *svc.ViajeActivo `json:"viaje"`
		Abono estadoAbono      `json:"abono"`
	}{viaje, nuevoEstadoAbono(abono)})
}
