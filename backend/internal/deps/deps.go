// Package deps agrupa las dependencias compartidas que los registros de
// rutas de cada spec reciben (evita ciclos de importación routes ↔ auth).
package deps

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"si16/backend/internal/services/estacion"
)

// Deps son las dependencias del backend inyectadas en los registros de rutas.
type Deps struct {
	Pool         *pgxpool.Pool
	RDB          *redis.Client
	JWTSecret    string
	TarifaMinuto float64
	// Estacion mantiene Redis actualizado despues de una mutacion de red.
	Estacion *estacion.Servicio
}
