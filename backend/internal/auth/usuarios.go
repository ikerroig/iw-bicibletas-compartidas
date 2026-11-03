package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/models"
)

// ErrUsuarioNoEncontrado se devuelve si no existe el usuario.
var ErrUsuarioNoEncontrado = errors.New("auth: usuario no encontrado")

const sqlUsuario = `
SELECT id, nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono, creado_en
FROM usuarios `

func escanear(row pgx.Row) (models.Usuario, error) {
	var u models.Usuario
	err := row.Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.Rol, &u.AbonoActivo,
		&u.FechaVencimientoAbono, &u.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrUsuarioNoEncontrado
	}
	return u, err
}

// BuscarPorEmail busca sin distinguir mayúsculas.
func BuscarPorEmail(ctx context.Context, pool *pgxpool.Pool, email string) (models.Usuario, error) {
	return escanear(pool.QueryRow(ctx, sqlUsuario+`WHERE lower(email) = $1`, strings.ToLower(strings.TrimSpace(email))))
}

// BuscarPorID busca por clave primaria.
func BuscarPorID(ctx context.Context, pool *pgxpool.Pool, id int64) (models.Usuario, error) {
	return escanear(pool.QueryRow(ctx, sqlUsuario+`WHERE id = $1`, id))
}
