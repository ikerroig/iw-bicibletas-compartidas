// Package httperr define el envelope de error uniforme
// {"error":{"codigo","mensaje"}} y los códigos de contracts/README.md.
package httperr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Códigos de error del contrato.
const (
	Unautenticado         = "UNAUTENTICADO"
	RolInsuficiente       = "ROL_INSUFICIENTE"
	CredencialesInvalidas = "CREDENCIALES_INVALIDAS"
	NoEncontrado          = "NO_ENCONTRADO"
	Interno               = "INTERNO"

	AbonoVencido          = "ABONO_VENCIDO"
	SinAbono              = "SIN_ABONO" // HU-02 D15: abono inactivo o sin fecha
	ViajeActivo           = "VIAJE_ACTIVO"
	SinViajeActivo        = "SIN_VIAJE_ACTIVO"
	AnclajeNoExiste       = "ANCLAJE_NO_EXISTE"
	AnclajeNoDisponible   = "ANCLAJE_NO_DISPONIBLE"
	AnclajeOcupado        = "ANCLAJE_OCUPADO"
	BicicletaNoDisponible = "BICICLETA_NO_DISPONIBLE"
	EstacionSinCapacidad  = "ESTACION_SIN_CAPACIDAD"
	Validacion            = "VALIDACION"
)

// Cuerpo es el envelope serializado.
type Cuerpo struct {
	Error Detalle `json:"error"`
}

// Detalle del error.
type Detalle struct {
	Codigo  string `json:"codigo"`
	Mensaje string `json:"mensaje"`
}

// Responder escribe el envelope y aborta la cadena de middlewares.
func Responder(c *gin.Context, status int, codigo, mensaje string) {
	c.AbortWithStatusJSON(status, Cuerpo{Error: Detalle{Codigo: codigo, Mensaje: mensaje}})
}

// Interno responde 500 sin filtrar detalles internos al cliente.
func ErrorInterno(c *gin.Context) {
	Responder(c, http.StatusInternalServerError, Interno, "Error interno del servidor")
}
