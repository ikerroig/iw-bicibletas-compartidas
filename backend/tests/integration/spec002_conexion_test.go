// T034 — Caso límite de la spec: perder la conexión a mitad del desbloqueo no
// deja efectos parciales, y el usuario puede reintentar.
package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"si16/backend/internal/models"
	"si16/backend/internal/services/alquiler"
	tu "si16/backend/internal/testutil"
)

func TestConexionPerdidaNoDejaViajeParcial(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Conexión", models.EstacionOperativa, tu.ConBici, tu.ConBici)
	uid, tok := e.Usuario(t, true, dias(30))

	// Un viaje sin confirmar del mismo usuario hace esperar al INSERT por el
	// índice único, cuando el desbloqueo ya ha escrito anclaje y bicicleta.
	var otraBici int64
	if err := e.Pool.QueryRow(e.Ctx, `SELECT bicicleta_id FROM anclajes WHERE codigo_anclaje = $1`, codigo(id, 2)).Scan(&otraBici); err != nil {
		t.Fatal(err)
	}
	tx, err := e.Pool.Begin(e.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(e.Ctx) }()
	if _, err := tx.Exec(e.Ctx, `INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id) VALUES ($1, $2, $3)`, uid, otraBici, id); err != nil {
		t.Fatal(err)
	}
	antes := fotografiar(t, e)

	conexion, cortar := context.WithCancel(e.Ctx)
	defer cortar()
	fin := make(chan error, 1)
	go func() {
		_, err := alquiler.Nuevo(e.Pool, e.Estacion, nil).Desbloquear(conexion, uid, codigo(id, 1))
		fin <- err
	}()

	enEspera := tu.Esperar(t, 3*time.Second, func() bool {
		var n int
		err := e.Pool.QueryRow(e.Ctx, `SELECT count(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'
			AND query LIKE '%INSERT INTO viajes AS v%'`).Scan(&n)
		return err == nil && n == 1
	})
	if !enEspera {
		t.Fatal("el desbloqueo no llegó a esperar en el INSERT del viaje")
	}
	cortar()
	if err := <-fin; err == nil {
		t.Fatal("con la conexión cortada debe devolver error")
	}
	if err := tx.Rollback(e.Ctx); err != nil {
		t.Fatal(err)
	}

	if d := fotografiar(t, e); d != antes {
		t.Fatalf("quedaron efectos parciales: %+v → %+v", antes, d)
	}
	if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
		t.Fatalf("reintento = %d %s; quiero 201", status, cod)
	}
}
