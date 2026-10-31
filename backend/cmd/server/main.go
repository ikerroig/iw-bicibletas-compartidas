// Command server — punto de entrada HTTP del backend SI-16 (Go/Gin).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"si16/backend/internal/config"
	"si16/backend/internal/db"
	"si16/backend/internal/deps"
	"si16/backend/internal/routes"
	alertas "si16/backend/internal/workers/alertas"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "server:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	go alertas.Run(ctx, pool) // worker HU-07: evaluación inmediata + cada 55 s

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping a Redis: %w", err)
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api/v1")
	routes.Register(api, deps.Deps{
		Pool:         pool,
		RDB:          rdb,
		JWTSecret:    cfg.JWTSecret,
		TarifaMinuto: cfg.TarifaMinuto,
	})

	fmt.Println("SI-16 backend escuchando en", cfg.ServerAddr)
	return r.Run(cfg.ServerAddr)
}
