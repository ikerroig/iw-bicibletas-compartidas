// Package devolucion implementa la HU-03: devolución de bicicleta y
// cálculo de tarificación. Contrato: specs/spec-003/contracts/devoluciones.md.
package devolucion

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/apierr"
	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	servicio "si16/backend/internal/services/devolucion"
)

// devolverRequest es el payload del contrato (D6).
type devolverRequest struct {
	CodigoAnclaje string `json:"codigo_anclaje"`
}

// Handler — POST /api/v1/devoluciones. Autenticación y rol `usuario` los
// aplica la ruta (D7); aquí se valida la entrada, se invoca al servicio y
// se traduce el resultado al contrato.
func Handler(d deps.Deps) gin.HandlerFunc {
	svc := servicio.New(d.Pool, d.Estacion, d.TarifaMinuto)
	return func(c *gin.Context) {
		var req devolverRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA",
				"El cuerpo debe ser JSON con el campo codigo_anclaje")
			return
		}
		// Mismo criterio que el desbloqueo (HU-02 FR-001): los códigos se guardan en mayúsculas.
		req.CodigoAnclaje = strings.ToUpper(strings.TrimSpace(req.CodigoAnclaje))
		if req.CodigoAnclaje == "" {
			apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA",
				"codigo_anclaje es obligatorio")
			return
		}
		// Contrato devoluciones.md: longitud ≤ 64 caracteres tras trim.
		if utf8.RuneCountInString(req.CodigoAnclaje) > 64 {
			apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA",
				"codigo_anclaje no puede superar 64 caracteres")
			return
		}

		viaje, aerr := svc.CerrarViaje(c.Request.Context(), auth.UsuarioID(c), req.CodigoAnclaje)
		if aerr != nil {
			apierr.Respond(c, aerr.Status, aerr.Codigo, aerr.Mensaje)
			return
		}
		c.JSON(http.StatusOK, viaje)
	}
}
