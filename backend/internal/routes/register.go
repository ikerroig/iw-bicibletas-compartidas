// Package routes monta el router /api/v1.
//
// ÚNICO archivo de rutas compartido entre specs (tarea T016): cada spec añade
// en NuevoRouter una sola línea que llama a su RegisterSpec00N, definida en
// su propio fichero routes/spec_00N_*.go. Así los merges son triviales.
package routes

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/auth"
	"si16/backend/internal/cache"
	"si16/backend/internal/config"
	"si16/backend/internal/deps"
	authhandler "si16/backend/internal/handlers/auth"
	"si16/backend/internal/services/estacion"
)

// Deps son las dependencias compartidas que recibe cada spec.
type Deps struct {
	Cfg      config.Config
	Pool     *pgxpool.Pool
	Cache    *cache.Cliente
	Auth     *auth.Emisor
	Estacion *estacion.Servicio // helper RecalcularEstacion (T020) para las specs que mutan
	Log      *slog.Logger
}

// NuevoRouter crea el motor Gin con todas las rutas registradas.
func NuevoRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	rg := r.Group("/api/v1")

	(&authhandler.Handler{Pool: d.Pool, Emisor: d.Auth, Log: d.Log}).Registrar(rg)

	// --- Una línea por spec (no reordenar; añadir al final) ---
	RegisterSpec001(rg, d)
	RegisterSpec002(rg, d)

	// specs que reciben deps.Deps (paquete deps): puente desde Deps global.
	sd := deps.Deps{
		Pool:         d.Pool,
		RDB:          d.Cache.Redis(),
		JWTSecret:    d.Cfg.JWTSecret,
		TarifaMinuto: d.Cfg.TarifaMinuto,
		Estacion:     d.Estacion,
	}
	RegisterSpec003(rg, sd)
	RegisterSpec004(rg, sd)
	RegisterSpec005(rg, sd)
	RegisterSpec006(rg, sd)
	RegisterSpec007(rg, sd)
	RegisterSpec008(rg, sd)

	return r
}

// Register monta las rutas base y las specs-003/004/005/007/008 a partir de
// deps.Deps (usado por los tests de esas specs; auth base incluida).
func Register(rg *gin.RouterGroup, d deps.Deps) {
	auth.RegisterAuth(rg, d) // base compartida: POST /auth/login, GET /auth/me
	RegisterSpec003(rg, d)   // spec-003 (Alex Morell): HU-03 devolución y tarificación
	RegisterSpec004(rg, d)   // spec-004 (Alex Morell): HU-04 historial de viajes
	RegisterSpec005(rg, d)
	RegisterSpec006(rg, d)
	RegisterSpec007(rg, d) // spec-007 (HU-07): alertas de balanceo y traslados
	RegisterSpec008(rg, d) // spec-008 (HU-08): analítica de uso y exportación CSV
}
