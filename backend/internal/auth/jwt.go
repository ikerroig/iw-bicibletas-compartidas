// Package auth implementa la autenticación JWT y los middlewares de rol
// (decisión R5 del plan global compartido).
package auth

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"si16/backend/internal/apierr"
)

// CrearToken genera un JWT HS256 con expiración de 24 h.
func CrearToken(secret string, usuarioID int64, nombre, rol string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":    usuarioID,
		"nombre": nombre,
		"rol":    rol,
		"iat":    now.Unix(),
		"exp":    now.Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// RequireAuth valida el Bearer JWT y deja `usuario_id` (int64) y `rol` (string)
// en el contexto Gin. 401 si falta o es inválido.
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			apierr.NoAutenticado(c)
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("método de firma no soportado")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			apierr.NoAutenticado(c)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			apierr.NoAutenticado(c)
			return
		}
		// `sub` puede ser numérico (CrearToken) o cadena (Emisor.Emitir):
		// se aceptan los dos formatos para que cualquier token válido lo sea
		// en ambos middlewares.
		var id int64
		switch v := claims["sub"].(type) {
		case float64:
			id = int64(v)
		case string:
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil || n <= 0 {
				apierr.NoAutenticado(c)
				return
			}
			id = n
		default:
			apierr.NoAutenticado(c)
			return
		}
		rol, _ := claims["rol"].(string)

		c.Set("usuario_id", id)
		c.Set("rol", rol)
		c.Next()
	}
}

// UsuarioID devuelve el id del usuario autenticado del contexto.
func UsuarioID(c *gin.Context) int64 {
	v, _ := c.Get("usuario_id")
	id, _ := v.(int64)
	return id
}
