// Package viajes implementa el endpoint HU-04: GET /api/v1/viajes
// (historial propio con filtro de periodo, paginación y sumatorio).
// Contrato: specs/spec-004/contracts/viajes.md.
package viajes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/apierr"
	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	servicio "si16/backend/internal/services/viajes"
)

// Handler — GET /api/v1/viajes. La autenticación la aplica la ruta
// (spec_004_viajes.go); aquí se valida la query, se invoca al servicio y
// se responde { total, suma_coste, items } (contrato). El usuario del
// historial NUNCA es entrada del request: sale del JWT (FR-005, V1).
func Handler(d deps.Deps) gin.HandlerFunc {
	svc := servicio.New(d.Pool)
	return func(c *gin.Context) {
		params, aerr := servicio.ValidarEntrada(
			c.Query("desde"), c.Query("hasta"),
			c.Query("page"), c.Query("page_size"),
		)
		if aerr != nil {
			apierr.Respond(c, aerr.Status, aerr.Codigo, aerr.Mensaje)
			return
		}

		r, aerr := svc.ListarHistorial(c.Request.Context(), auth.UsuarioID(c), params)
		if aerr != nil {
			apierr.Respond(c, aerr.Status, aerr.Codigo, aerr.Mensaje)
			return
		}
		c.JSON(http.StatusOK, r)
	}
}
