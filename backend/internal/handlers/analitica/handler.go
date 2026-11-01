// Handlers HU-08: agregados del panel (GET) y exportación CSV en streaming
// (GET). Contrato: specs/spec-008/contracts/analitica.md (envelope global
// apierr; la validación de la exportación ocurre ANTES de las cabeceras).
package analitica

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/apierr"
	"si16/backend/internal/deps"
	"si16/backend/internal/services/analitica"
)

// HandlerAgregados valida el rango (422), calcula los agregados server-side
// y responde 200 con models.AnaliticaViajes (eco de fechas incluido).
func HandlerAgregados(d deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, aerr := analitica.ValidarEntrada(c.Query("desde"), c.Query("hasta"))
		if aerr != nil {
			apierr.Respond(c, aerr.Status, aerr.Codigo, aerr.Mensaje)
			return
		}
		r, aerr := analitica.New(d.Pool).Agregados(c.Request.Context(), p)
		if aerr != nil {
			apierr.Respond(c, aerr.Status, aerr.Codigo, aerr.Mensaje)
			return
		}
		c.JSON(http.StatusOK, r)
	}
}

// HandlerExport valida el rango ANTES de escribir cabeceras de descarga y
// luego hace streaming del CSV (BOM, columnas R11, nombre del fichero).
func HandlerExport(d deps.Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		desde, hasta := c.Query("desde"), c.Query("hasta")
		p, aerr := analitica.ValidarEntrada(desde, hasta)
		if aerr != nil { // sin Content-Disposition: envelope JSON estándar
			apierr.Respond(c, aerr.Status, aerr.Codigo, aerr.Mensaje)
			return
		}
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition",
			"attachment; filename="+analitica.NombreFichero(desde, hasta))
		if err := analitica.New(d.Pool).
			ExportarCSV(c.Request.Context(), p, c.Writer); err != nil {
			// Cabeceras ya enviadas: no hay envelope posible; se aborta el
			// stream para que el cliente no consuma un fichero truncado
			// silenciosamente.
			c.Abort()
		}
	}
}
