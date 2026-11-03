// Comando migrate: aplica las migraciones goose de backend/migrations.
//
//	go run ./cmd/migrate up      # aplica todas las pendientes
//	go run ./cmd/migrate down    # revierte la última
//	go run ./cmd/migrate status  # muestra el estado
package main

import (
	"context"
	"log"
	"os"

	"si16/backend/internal/config"
	"si16/backend/internal/db"
)

func main() {
	direccion := "up"
	if len(os.Args) > 1 {
		direccion = os.Args[1]
	}
	cfg, err := config.Cargar()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := db.Migrar(context.Background(), cfg.DatabaseURL, direccion); err != nil {
		log.Fatalf("migrate %s: %v", direccion, err)
	}
	log.Printf("migrate %s: OK", direccion)
}
