// T033 — T1 (const. I): sin abono vigente no hay desbloqueo, y el rechazo no
// deja ningún efecto (HU-02 FR-002, SC-001).
// T045 — la causa distingue «sin abono» de «caducado», con la fecha (D15).
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
	status, codigo, _ := desbloquearConMensaje(t, e, token, cod)
	return status, codigo
}

// desbloquearConMensaje devuelve también el mensaje del envelope de error.
func desbloquearConMensaje(t *testing.T, e *tu.Entorno, token, cod string) (int, string, string) {
	t.Helper()
	rec := e.Post(t, "/alquileres/desbloqueo", token, fmt.Sprintf(`{"codigo_anclaje":%q}`, cod))
	var r struct {
		Error struct {
			Codigo  string `json:"codigo"`
			Mensaje string `json:"mensaje"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	return rec.Code, r.Error.Codigo, r.Error.Mensaje
}

// caducoEl es el mensaje de ABONO_VENCIDO para una fecha (D15).
func caducoEl(v *time.Time) string {
	return "Tu abono anual caducó el " + v.Format("02/01/2006")
}

const sinAbono = "No tienes abono anual"

// dias devuelve la fecha de hoy en España más n días: la misma que usa el servicio.
func dias(n int) *time.Time {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		panic(err)
	}
	a, m, d := time.Now().In(madrid).Date()
	v := time.Date(a, m, d+n, 0, 0, 0, 0, time.UTC)
	return &v
}

func TestDesbloqueoRechazaAbonoNoVigente(t *testing.T) {
	ayer := dias(-1)
	casos := []struct {
		nombre          string
		activo          bool
		venc            *time.Time
		codigo, mensaje string
	}{
		{"caducado ayer", true, ayer, httperr.AbonoVencido, caducoEl(ayer)},
		{"inactivo con fecha futura", false, dias(30), httperr.SinAbono, sinAbono},
		{"incoherente: activo sin fecha", true, nil, httperr.SinAbono, sinAbono},
		{"inactivo sin fecha", false, nil, httperr.SinAbono, sinAbono},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := iniciar(t)
			id := e.CrearEstacion(t, "Abono", models.EstacionOperativa, tu.ConBici, tu.Libre)
			_, tok := e.Usuario(t, c.activo, c.venc)
			antes := fotografiar(t, e)

			status, cod, msg := desbloquearConMensaje(t, e, tok, codigo(id, 1))

			if status != http.StatusForbidden || cod != c.codigo || msg != c.mensaje {
				t.Fatalf("respuesta = %d %s %q; quiero 403 %s %q", status, cod, msg, c.codigo, c.mensaje)
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

	status, cod, msg := desbloquearConMensaje(t, e, tok, codigo(id, 1))
	if want := caducoEl(dias(-1)); status != http.StatusForbidden || cod != httperr.AbonoVencido || msg != want {
		t.Fatalf("respuesta = %d %s %q; quiero 403 %s %q", status, cod, msg, httperr.AbonoVencido, want)
	}
}
