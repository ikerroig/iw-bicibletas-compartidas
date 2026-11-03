// Package auth expone POST /auth/login y GET /auth/me (base compartida).
package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/auth"
	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
)

// Handler de autenticación.
type Handler struct {
	Pool   *pgxpool.Pool
	Emisor *auth.Emisor
	Log    *slog.Logger
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UsuarioDTO es el esquema Usuario del contrato.
type UsuarioDTO struct {
	ID                    int64      `json:"id"`
	Nombre                string     `json:"nombre"`
	Email                 string     `json:"email"`
	Rol                   models.Rol `json:"rol"`
	AbonoActivo           bool       `json:"abono_activo"`
	FechaVencimientoAbono *string    `json:"fecha_vencimiento_abono"`
}

// NuevoUsuarioDTO serializa la fecha como YYYY-MM-DD o null.
func NuevoUsuarioDTO(u models.Usuario) UsuarioDTO {
	d := UsuarioDTO{ID: u.ID, Nombre: u.Nombre, Email: u.Email, Rol: u.Rol, AbonoActivo: u.AbonoActivo}
	if u.FechaVencimientoAbono != nil {
		f := u.FechaVencimientoAbono.Format("2006-01-02")
		d.FechaVencimientoAbono = &f
	}
	return d
}

// Registrar monta las rutas en el grupo /api/v1.
func (h *Handler) Registrar(rg *gin.RouterGroup) {
	rg.POST("/auth/login", h.login)
	rg.GET("/auth/me", h.Emisor.RequireAuth(), h.me)
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.Responder(c, http.StatusUnauthorized, httperr.CredencialesInvalidas, "Email o contraseña incorrectos")
		return
	}
	u, err := auth.BuscarPorEmail(c.Request.Context(), h.Pool, req.Email)
	if err != nil && !errors.Is(err, auth.ErrUsuarioNoEncontrado) {
		h.Log.Error("auth: login", "error", err)
		httperr.ErrorInterno(c)
		return
	}
	// Mismo mensaje para email inexistente y contraseña errónea.
	if err != nil || !auth.PasswordCorrecta(u.PasswordHash, req.Password) {
		httperr.Responder(c, http.StatusUnauthorized, httperr.CredencialesInvalidas, "Email o contraseña incorrectos")
		return
	}
	token, err := h.Emisor.Emitir(u)
	if err != nil {
		h.Log.Error("auth: emitir token", "error", err)
		httperr.ErrorInterno(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "usuario": NuevoUsuarioDTO(u)})
}

func (h *Handler) me(c *gin.Context) {
	s, _ := auth.SesionDe(c)
	u, err := auth.BuscarPorID(c.Request.Context(), h.Pool, s.UsuarioID)
	if errors.Is(err, auth.ErrUsuarioNoEncontrado) {
		httperr.Responder(c, http.StatusUnauthorized, httperr.Unautenticado, "Sesión no válida")
		return
	}
	if err != nil {
		h.Log.Error("auth: me", "error", err)
		httperr.ErrorInterno(c)
		return
	}
	c.JSON(http.StatusOK, NuevoUsuarioDTO(u))
}
