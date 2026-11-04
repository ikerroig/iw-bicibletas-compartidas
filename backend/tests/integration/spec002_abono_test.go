// T033 — T1 (const. I): sin abono vigente no hay desbloqueo, y el rechazo no
// deja ningún efecto (HU-02 FR-002, SC-001).
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

// foto resume todo lo que un desbloqueo puede modificar.
type foto struct {
	Viajes, Libres, Ocupados, Disponibles, EnUso int
}

func fotografiar(t *testing.T, e *tu.Entorno) foto {
	t.Helper()
	var f foto
	err := e.Pool.QueryRow(e.Ctx, `SELECT
		(SELECT count(*) FROM viajes),
		(SELECT count(*) FROM anclajes WHERE estado = 'libre'),
		(SELECT count(*) FROM anclajes WHERE estado = 'ocupado'),
		(SELECT count(*) FROM bicicletas WHERE estado = 'disponible'),
		(SELECT count(*) FROM bicicletas WHERE estado = 'en_uso')`).
		Scan(&f.Viajes, &f.Libres, &f.Ocupados, &f.Disponibles, &f.EnUso)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func codigo(id int64, n int) string { return fmt.Sprintf("E%d-A%02d", id, n) }

func desbloquear(t *testing.T, e *tu.Entorno, token, cod string) (int, string) {
	t.Helper()
	rec := e.Post(t, "/alquileres/desbloqueo", token, fmt.Sprintf(`{"codigo_anclaje":%q}`, cod))
	var r struct {
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	return rec.Code, r.Error.Codigo
}

func dias(n int) *time.Time {
	v := time.Now().UTC().AddDate(0, 0, n)
	return &v
}

func TestDesbloqueoRechazaAbonoNoVigente(t *testing.T) {
	casos := []struct {
		nombre string
		activo bool
		venc   *time.Time
	}{
		{"caducado ayer", true, dias(-1)},
		{"inactivo con fecha futura", false, dias(30)},
		{"incoherente: activo sin fecha", true, nil},
		{"inactivo sin fecha", false, nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := iniciar(t)
			id := e.CrearEstacion(t, "Abono", models.EstacionOperativa, tu.ConBici, tu.Libre)
			_, tok := e.Usuario(t, c.activo, c.venc)
			antes := fotografiar(t, e)

			status, cod := desbloquear(t, e, tok, codigo(id, 1))

			if status != http.StatusForbidden || cod != httperr.AbonoVencido {
				t.Fatalf("respuesta = %d %s; quiero 403 %s", status, cod, httperr.AbonoVencido)
			}
			if d := fotografiar(t, e); d != antes {
				t.Errorf("el rechazo tuvo efectos: %+v → %+v", antes, d)
			}
		})
	}
}

func TestDesbloqueoConAbonoQueVenceHoy(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Vence hoy", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(0))
	if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
		t.Fatalf("respuesta = %d %s; el día de vencimiento el abono sigue vigente", status, cod)
	}
}

// El abono se comprueba en cada solicitud, no al iniciar sesión (caso límite
// de la spec): un JWT emitido con el abono vigente no sirve tras caducar.
func TestAbonoCaducadoTrasIniciarSesion(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Caduca", models.EstacionOperativa, tu.ConBici)
	uid, tok := e.Usuario(t, true, dias(10))
	e.Exec(t, `UPDATE usuarios SET fecha_vencimiento_abono = CURRENT_DATE - 1 WHERE id = $1`, uid)

	if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusForbidden || cod != httperr.AbonoVencido {
		t.Fatalf("respuesta = %d %s; quiero 403 %s", status, cod, httperr.AbonoVencido)
	}
}
