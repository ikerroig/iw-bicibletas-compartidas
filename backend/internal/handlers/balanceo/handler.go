// Handlers HU-07: panel de alertas (GET) y confirmación de traslado (POST).
// Contrato: specs/spec-007/contracts/balanceo.md (envelope global apierr).
package balanceo

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/apierr"
	"si16/backend/internal/auth"
	"si16/backend/internal/deps"
	"si16/backend/internal/models"
	"si16/backend/internal/services/balanceo"
)

// HandlerPanel responde 200 { items: [...] } con las alertas abiertas
// ordenadas por urgencia (CASE score DESC, fecha_generacion ASC) y el
// estacion_nombre resuelto por JOIN (data-model.md §3). items vacío → [].
func HandlerPanel(d deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := d.Pool.Query(c.Request.Context(),
			`SELECT a.id, a.estacion_id, e.nombre AS estacion_nombre,
			        a.tipo_alerta, a.porcentaje_ocupacion::float8, a.severidad,
			        a.fecha_generacion, a.resuelta
			   FROM alertas_balanceo a
			   JOIN estaciones e ON e.id = a.estacion_id
			  WHERE NOT a.resuelta
			  ORDER BY CASE WHEN a.tipo_alerta = 'vacia_menos_10'
			                THEN 10 - a.porcentaje_ocupacion
			                ELSE     a.porcentaje_ocupacion - 90 END DESC,
			          a.fecha_generacion ASC`)
		if err != nil {
			apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error al leer las alertas de balanceo")
			return
		}
		defer rows.Close()

		items := make([]models.AlertaBalanceo, 0)
		for rows.Next() {
			var a models.AlertaBalanceo
			if err := rows.Scan(&a.ID, &a.EstacionID, &a.EstacionNombre, &a.TipoAlerta,
				&a.PorcentajeOcupacion, &a.Severidad, &a.FechaGeneracion, &a.Resuelta); err != nil {
				apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error al leer las alertas de balanceo")
				return
			}
			items = append(items, a)
		}
		if err := rows.Err(); err != nil {
			apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error al leer las alertas de balanceo")
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

// HandlerTraslado parsea TrasladoRequest (422 = payload ilegible), llama al
// servicio (409 = estado actual incompatible) y responde 200 { origen, destino }
// con ambas estaciones recalculadas (FR-007, SC-003).
func HandlerTraslado(d deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA",
				"JSON inválido: la petición no puede interpretarse")
			return
		}
		req, e := balanceo.ParseTraslado(body)
		if e != nil {
			apierr.Respond(c, e.Status, e.Codigo, e.Mensaje)
			return
		}
		origen, destino, e := balanceo.NewConEstacion(d.Pool, d.Estacion).
			ConfirmarTraslado(c.Request.Context(), auth.UsuarioID(c), req)
		if e != nil {
			apierr.Respond(c, e.Status, e.Codigo, e.Mensaje)
			return
		}
		c.JSON(http.StatusOK, gin.H{"origen": origen, "destino": destino})
	}
}
