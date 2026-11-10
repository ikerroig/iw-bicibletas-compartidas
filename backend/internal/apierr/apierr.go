// Package apierr define el envelope de error JSON compartido por todos los
// endpoints (contrato: specs/spec-001/contracts/openapi.yaml → schema Error).
package apierr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope de error: { "error": { "codigo", "mensaje" } }.
type body struct {
	Error struct {
		Codigo  string `json:"codigo"`
		Mensaje string `json:"mensaje"`
	} `json:"error"`
}

// Error tipado que arrastra HTTP status + código de negocio + mensaje, para
// que los servicios puedan devolver errores ya mapeados al contrato.
type Error struct {
	Status  int    `json:"-"`
	Codigo  string `json:"codigo"`
	Mensaje string `json:"mensaje"`
}

func (e *Error) Error() string { return e.Codigo + ": " + e.Mensaje }

// New construye un *Error para propagar desde servicios hacia handlers.
func New(status int, codigo, mensaje string) *Error {
	return &Error{Status: status, Codigo: codigo, Mensaje: mensaje}
}

// Respond envía la respuesta de error estandarizada.
func Respond(c *gin.Context, status int, codigo, mensaje string) {
	var b body
	b.Error.Codigo = codigo
	b.Error.Mensaje = mensaje
	c.AbortWithStatusJSON(status, b)
}

// NoAutenticado aborta con 401 NO_AUTENTICADO.
func NoAutenticado(c *gin.Context) {
	Respond(c, http.StatusUnauthorized, "NO_AUTENTICADO", "Se requiere una sesión válida")
}

// RolNoPermitido aborta con 403 ROL_NO_PERMITIDO.
func RolNoPermitido(c *gin.Context) {
	Respond(c, http.StatusForbidden, "ROL_NO_PERMITIDO", "Tu rol no tiene permiso para esta operación")
}
