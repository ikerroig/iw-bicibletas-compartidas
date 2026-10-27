// Package deps agrupa las dependencias compartidas que los registros de
// rutas de cada spec reciben (evita ciclos de importación routes ↔ auth).
package deps

import (
	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps son las dependencias del backend inyectadas en los registros de rutas.
type Deps struct {
	Pool         *pgxpool.Pool
	RDB          *redis.Client
	JWTSecret    string
	TarifaMinuto float64
}
