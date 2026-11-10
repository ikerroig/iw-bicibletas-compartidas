// T036 — FR-007 (research D12): cuando fn termina bien, una cancelación de la
// petición ya no impide confirmar.
package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"si16/backend/internal/db"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestCommitSobreviveALaCancelacion(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Antes", models.EstacionOperativa, tu.Libre)

	ctx, cancelar := context.WithCancel(e.Ctx)
	defer cancelar()
	err := db.EnTransaccion(ctx, e.Pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE estaciones SET nombre = 'Confirmada' WHERE id = $1`, id)
		cancelar()
		return err
	})
	if err != nil {
		t.Fatalf("EnTransaccion = %v; quiero nil", err)
	}

	var nombre string
	if err := e.Pool.QueryRow(e.Ctx, `SELECT nombre FROM estaciones WHERE id = $1`, id).Scan(&nombre); err != nil {
		t.Fatal(err)
	}
	if nombre != "Confirmada" {
		t.Errorf("nombre = %q; el commit no se aplicó", nombre)
	}
}
