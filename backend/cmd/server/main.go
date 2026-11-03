// Comando server: API REST /api/v1 del SI-16 (Go + Gin).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"si16/backend/internal/auth"
	"si16/backend/internal/cache"
	"si16/backend/internal/config"
	"si16/backend/internal/db"
	"si16/backend/internal/routes"
	"si16/backend/internal/services/estacion"
	"si16/backend/internal/workers/refresco_estado"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Cargar()
	if err != nil {
		log.Error("config", "error", err)
		os.Exit(1)
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	pool, err := db.Conectar(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	rc := cache.Nuevo(cfg.RedisAddr, 0)
	defer func() { _ = rc.Cerrar() }()
	if err := rc.Ping(ctx); err != nil {
		// No es fatal: GET /estaciones sirve el cálculo de PostgreSQL (FR-005)
		// y el worker repoblará Redis cuando vuelva.
		log.Warn("redis no disponible al arrancar", "addr", cfg.RedisAddr, "error", err)
	}

	deps := routes.Deps{
		Cfg:      cfg,
		Pool:     pool,
		Cache:    rc,
		Auth:     auth.NuevoEmisor(cfg.JWTSecret),
		Estacion: estacion.Nuevo(pool, rc, log),
		Log:      log,
	}

	// --- Workers: una línea por spec ---
	go refresco_estado.Nuevo(deps.Estacion, rc, log, refresco_estado.IntervaloBarrido).Iniciar(ctx) // spec-001

	gin.SetMode(gin.ReleaseMode)
	srv := &http.Server{
		Addr:              ":" + cfg.Puerto,
		Handler:           routes.NuevoRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Info("API escuchando", "url", "http://localhost:"+cfg.Puerto+"/api/v1")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http", "error", err)
			parar()
		}
	}()

	<-ctx.Done()
	apagado, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	_ = srv.Shutdown(apagado)
	log.Info("servidor detenido")
}
