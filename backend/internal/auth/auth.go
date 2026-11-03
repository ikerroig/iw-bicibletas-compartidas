// Package auth implementa contraseñas bcrypt, JWT HS256 (exp 24 h, claims
// sub/rol/nombre) y los middlewares RequireAuth / RequireRole (research R5).
package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
)

// DuracionToken es la validez del JWT.
const DuracionToken = 24 * time.Hour

const claveContexto = "si16.sesion"

// Sesion es la identidad autenticada que los handlers obtienen con SesionDe.
type Sesion struct {
	UsuarioID int64
	Rol       models.Rol
	Nombre    string
}

type claims struct {
	Rol    models.Rol `json:"rol"`
	Nombre string     `json:"nombre"`
	jwt.RegisteredClaims
}

// HashPassword genera el hash bcrypt.
func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

// HashRelleno es un hash bcrypt válido (de una contraseña aleatoria descartada)
// con el coste por defecto. El login lo compara cuando el email no existe para
// que la respuesta tarde lo mismo que con una contraseña errónea.
var HashRelleno = func() string {
	h, err := bcrypt.GenerateFromPassword([]byte("relleno-sin-uso-"+time.Now().String()), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}()

// PasswordCorrecta compara en tiempo constante.
func PasswordCorrecta(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Emisor firma y valida tokens con el secreto configurado.
type Emisor struct {
	secreto []byte
	ahora   func() time.Time
}

// NuevoEmisor crea el emisor.
func NuevoEmisor(secreto string) *Emisor {
	return &Emisor{secreto: []byte(secreto), ahora: time.Now}
}

// Emitir genera un token para el usuario.
func (e *Emisor) Emitir(u models.Usuario) (string, error) {
	ahora := e.ahora()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Rol:    u.Rol,
		Nombre: u.Nombre,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(DuracionToken)),
		},
	})
	return t.SignedString(e.secreto)
}

// Validar devuelve la sesión de un token válido y no expirado.
func (e *Emisor) Validar(token string) (Sesion, error) {
	var cl claims
	_, err := jwt.ParseWithClaims(token, &cl, func(*jwt.Token) (any, error) { return e.secreto, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(e.ahora),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Sesion{}, err
	}
	id, err := strconv.ParseInt(cl.Subject, 10, 64)
	if err != nil || id <= 0 {
		return Sesion{}, errors.New("auth: sub inválido")
	}
	return Sesion{UsuarioID: id, Rol: cl.Rol, Nombre: cl.Nombre}, nil
}

// RequireAuth exige "Authorization: Bearer <JWT>" válido → 401 si no.
func (e *Emisor) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cab := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(cab, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			httperr.Responder(c, http.StatusUnauthorized, httperr.Unautenticado, "Inicia sesión para continuar")
			return
		}
		s, err := e.Validar(strings.TrimSpace(token))
		if err != nil {
			httperr.Responder(c, http.StatusUnauthorized, httperr.Unautenticado, "Sesión no válida o caducada")
			return
		}
		c.Set(claveContexto, s)
		c.Next()
	}
}

// RequireRole exige uno de los roles dados (usar tras RequireAuth) → 403 si no.
func RequireRole(roles ...models.Rol) gin.HandlerFunc {
	return func(c *gin.Context) {
		s, ok := SesionDe(c)
		if !ok {
			httperr.Responder(c, http.StatusUnauthorized, httperr.Unautenticado, "Inicia sesión para continuar")
			return
		}
		for _, r := range roles {
			if s.Rol == r {
				c.Next()
				return
			}
		}
		httperr.Responder(c, http.StatusForbidden, httperr.RolInsuficiente, "No tienes permiso para esta operación")
	}
}

// SesionDe devuelve la sesión que dejó RequireAuth.
func SesionDe(c *gin.Context) (Sesion, bool) {
	v, ok := c.Get(claveContexto)
	if !ok {
		return Sesion{}, false
	}
	s, ok := v.(Sesion)
	return s, ok
}
