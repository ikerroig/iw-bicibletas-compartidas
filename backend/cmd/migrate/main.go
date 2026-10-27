// Command migrate aplica las migraciones goose (esquema compartido SI-16).
//
// Uso (desde backend/ o desde la raíz):
//
//	go run ./cmd/migrate up
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"si16/backend/internal/config"
	"si16/backend/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
	fmt.Println("migraciones aplicadas correctamente")
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("abrir conexión: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.Up(db, ""); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
