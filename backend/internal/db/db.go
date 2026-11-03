// Package db agrupa el pool de PostgreSQL, el helper de transacciones y la
// aplicación de migraciones (research R6: SQL explícito con pgx, sin ORM).
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"si16/backend/migrations"
)

// Conectar abre el pool y comprueba la conexión.
func Conectar(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("db: configuración inválida: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: sin conexión con PostgreSQL: %w", err)
	}
	return pool, nil
}

// EnTransaccion ejecuta fn dentro de una transacción: commit si fn devuelve
// nil, rollback en cualquier otro caso (incluido un panic). Las operaciones que
// necesiten bloqueo pesimista usan `SELECT … FOR UPDATE` dentro de fn
// (research R7; lo usan US2/US3/US5/US6/US7).
func EnTransaccion(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = errors.Join(err, rbErr)
			}
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Migrar aplica (direccion "up") o revierte la última ("down") migración
// embebida en backend/migrations.
func Migrar(ctx context.Context, url, direccion string) error {
	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		return err
	}
	sqlDB := stdlib.OpenDB(*cfg)
	defer func(c *sql.DB) { _ = c.Close() }(sqlDB)

	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	switch direccion {
	case "up":
		return goose.UpContext(ctx, sqlDB, ".")
	case "down":
		return goose.DownContext(ctx, sqlDB, ".")
	case "status":
		return goose.StatusContext(ctx, sqlDB, ".")
	default:
		return fmt.Errorf("dirección de migración desconocida %q (usa up, down o status)", direccion)
	}
}
