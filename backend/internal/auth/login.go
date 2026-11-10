package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"si16/backend/internal/apierr"
	"si16/backend/internal/deps"
	"si16/backend/internal/models"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token   string         `json:"token"`
	Usuario models.Usuario `json:"usuario"`
}

// Login: POST /auth/login — email + contraseña → JWT (R5).
func Login(pool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Password == "" {
			apierr.Respond(c, http.StatusUnprocessableEntity, "ENTRADA_INVALIDA", "email y password son obligatorios")
			return
		}

		var (
			u            models.Usuario
			passwordHash string
		)
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, nombre, email, rol, abono_activo, fecha_vencimiento_abono, password_hash
			   FROM usuarios WHERE email = $1`, req.Email,
		).Scan(&u.ID, &u.Nombre, &u.Email, &u.Rol, &u.AbonoActivo, &u.FechaVencimientoAbono, &passwordHash)
		if errors.Is(err, pgx.ErrNoRows) {
			apierr.Respond(c, http.StatusUnauthorized, "CREDENCIALES_INVALIDAS", "Email o contraseña incorrectos")
			return
		}
		if err != nil {
			apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error consultando el usuario")
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
			apierr.Respond(c, http.StatusUnauthorized, "CREDENCIALES_INVALIDAS", "Email o contraseña incorrectos")
			return
		}

		token, err := CrearToken(jwtSecret, u.ID, u.Nombre, string(u.Rol))
		if err != nil {
			apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "No se pudo generar el token")
			return
		}
		c.JSON(http.StatusOK, loginResponse{Token: token, Usuario: u})
	}
}

// Me: GET /auth/me — devuelve el usuario del token.
func Me(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u models.Usuario
		err := pool.QueryRow(c.Request.Context(),
			`SELECT id, nombre, email, rol, abono_activo, fecha_vencimiento_abono
			   FROM usuarios WHERE id = $1`, UsuarioID(c),
		).Scan(&u.ID, &u.Nombre, &u.Email, &u.Rol, &u.AbonoActivo, &u.FechaVencimientoAbono)
		if errors.Is(err, pgx.ErrNoRows) {
			apierr.Respond(c, http.StatusUnauthorized, "NO_AUTENTICADO", "Se requiere una sesión válida")
			return
		}
		if err != nil {
			apierr.Respond(c, http.StatusInternalServerError, "ERROR_INTERNO", "Error consultando el usuario")
			return
		}
		c.JSON(http.StatusOK, u)
	}
}

// RegisterAuth monta las rutas base de autenticación (compartidas).
func RegisterAuth(rg *gin.RouterGroup, d deps.Deps) {
	rg.POST("/auth/login", Login(d.Pool, d.JWTSecret))
	rg.GET("/auth/me", RequireAuth(d.JWTSecret), Me(d.Pool))
}
