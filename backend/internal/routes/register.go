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

	return r
}
