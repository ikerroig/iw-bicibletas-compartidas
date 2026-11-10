// T034 — T2 (const. I): un solo viaje activo por usuario, también bajo
// concurrencia, y rechazos de anclaje sin efectos (HU-02 FR-003…FR-007).
package integration

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"si16/backend/internal/hooks"
	"si16/backend/internal/httperr"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func TestDesbloqueoValidoCreaViajeYActualizaEstado(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Válido", models.EstacionOperativa, tu.ConBici, tu.ConBici, tu.Libre)
	if err := e.Estacion.RecalcularEstacion(e.Ctx, []int64{id}); err != nil {
		t.Fatal(err)
	}
	uid, tok := e.Usuario(t, true, dias(30))
	notificadas := make(chan []int64, 1)
	hooks.Suscribir(func(_ context.Context, ids []int64) error { notificadas <- ids; return nil })

	if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
		t.Fatalf("respuesta = %d %s; quiero 201", status, cod)
	}

	var estadoViaje, estadoBici, estadoAnclaje string
	var conBici bool
	var origen int64
	var inicio time.Time
	if err := e.Pool.QueryRow(e.Ctx, `SELECT v.estado, v.estacion_origen_id, v.fecha_inicio, b.estado, a.estado, a.bicicleta_id IS NOT NULL
		FROM viajes v JOIN bicicletas b ON b.id = v.bicicleta_id, anclajes a
		WHERE v.usuario_id = $1 AND a.codigo_anclaje = $2`, uid, codigo(id, 1)).
		Scan(&estadoViaje, &origen, &inicio, &estadoBici, &estadoAnclaje, &conBici); err != nil {
		t.Fatal(err)
	}
	if estadoViaje != models.ViajeEnCurso || origen != id || time.Since(inicio) > time.Minute {
		t.Errorf("viaje = %s, origen %d, inicio %v", estadoViaje, origen, inicio)
	}
	if estadoBici != models.BicicletaEnUso || estadoAnclaje != models.AnclajeLibre || conBici {
		t.Errorf("bici = %s, anclaje = %s, con bici = %v", estadoBici, estadoAnclaje, conBici)
	}

	// El recálculo post-commit deja Redis al día y notifica el hook.
	h, ok := leerHash(t, e, id)
	if !ok || h.BicisDisponibles != 1 || h.AnclajesLibres != 2 {
		t.Errorf("hash = %+v; quiero 1 bici y 2 libres", h)
	}
	select {
	case ids := <-notificadas:
		if len(ids) != 1 || ids[0] != id {
			t.Errorf("hook con %v; quiero [%d]", ids, id)
		}
	default:
		t.Error("el hook de estación cambiada no se notificó")
	}
}

func TestSegundoDesbloqueoRechazado(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Segundo", models.EstacionOperativa, tu.ConBici, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))
	if status, cod := desbloquear(t, e, tok, codigo(id, 1)); status != http.StatusCreated {
		t.Fatalf("primer desbloqueo = %d %s", status, cod)
	}
	antes := fotografiar(t, e)

	status, cod := desbloquear(t, e, tok, codigo(id, 2))

	if status != http.StatusConflict || cod != httperr.ViajeActivo {
		t.Fatalf("respuesta = %d %s; quiero 409 %s", status, cod, httperr.ViajeActivo)
	}
	if d := fotografiar(t, e); d != antes {
		t.Errorf("el rechazo tuvo efectos: %+v → %+v", antes, d)
	}
}

// Anclajes distintos: el bloqueo de fila no los serializa; solo el índice
// ux_viajes_usuario_activo impide el segundo viaje.
func TestDesbloqueosConcurrentesMismoUsuario(t *testing.T) {
	const n = 8
	e := iniciar(t)
	id := e.CrearEstacion(t, "Concurrente", models.EstacionOperativa, tu.Repetir(n, tu.ConBici)...)
	uid, tok := e.Usuario(t, true, dias(30))

	var wg sync.WaitGroup
	salida := make(chan struct{})
	resultados := make([]int, n)
	codigos := make([]string, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-salida
			resultados[i], codigos[i] = desbloquear(t, e, tok, codigo(id, i+1))
		}()
	}
	close(salida)
	wg.Wait()

	creados := 0
	for i, s := range resultados {
		switch {
		case s == http.StatusCreated:
			creados++
		case s != http.StatusConflict || codigos[i] != httperr.ViajeActivo:
			t.Errorf("intento %d = %d %s; quiero 201 o 409 %s", i, s, codigos[i], httperr.ViajeActivo)
		}
	}
	var viajes int
	if err := e.Pool.QueryRow(e.Ctx, `SELECT count(*) FROM viajes WHERE usuario_id = $1 AND estado = 'en_curso'`, uid).Scan(&viajes); err != nil {
		t.Fatal(err)
	}
	if creados != 1 || viajes != 1 {
		t.Fatalf("creados = %d, viajes activos = %d; quiero exactamente 1", creados, viajes)
	}
	f := fotografiar(t, e)
	if f.EnUso != 1 || f.Disponibles != n-1 || f.Ocupados != n-1 {
		t.Errorf("hay efectos parciales: %+v", f)
	}
}

// Mismo anclaje, usuarios distintos: la bici solo sale una vez.
func TestDesbloqueosConcurrentesMismoAnclaje(t *testing.T) {
	const n = 6
	e := iniciar(t)
	id := e.CrearEstacion(t, "Disputado", models.EstacionOperativa, tu.ConBici)
	tokens := make([]string, n)
	for i := range tokens {
		_, tokens[i] = e.Usuario(t, true, dias(30))
	}

	var wg sync.WaitGroup
	salida := make(chan struct{})
	resultados := make([]int, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-salida
			resultados[i], _ = desbloquear(t, e, tokens[i], codigo(id, 1))
		}()
	}
	close(salida)
	wg.Wait()

	creados := 0
	for _, s := range resultados {
		if s == http.StatusCreated {
			creados++
		} else if s != http.StatusConflict {
			t.Errorf("respuesta = %d; quiero 201 o 409", s)
		}
	}
	if f := fotografiar(t, e); creados != 1 || f.Viajes != 1 {
		t.Fatalf("creados = %d, viajes = %d; quiero 1", creados, f.Viajes)
	}
}

func TestDesbloqueoRechazosDeAnclajeSinEfectos(t *testing.T) {
	conflicto := http.StatusConflict
	casos := []struct {
		nombre, estado string
		anclaje        tu.Anclaje
		status         int
		codigo         string
	}{
		{"estación fuera de servicio", models.EstacionFueraDeServicio, tu.ConBici, conflicto, httperr.AnclajeNoDisponible},
		{"anclaje vacío", models.EstacionOperativa, tu.Libre, conflicto, httperr.AnclajeNoDisponible},
		{"anclaje fuera de servicio", models.EstacionOperativa, tu.Roto, conflicto, httperr.AnclajeNoDisponible},
		{"bici en anclaje fuera de servicio", models.EstacionOperativa,
			tu.Anclaje{Estado: models.AnclajeFueraDeServicio, Bicicleta: models.BicicletaDisponible}, conflicto, httperr.AnclajeNoDisponible},
		{"bici averiada", models.EstacionOperativa, tu.ConAveriada, conflicto, httperr.BicicletaNoDisponible},
		{"bici en taller", models.EstacionOperativa, tu.ConEnTaller, conflicto, httperr.BicicletaNoDisponible},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := iniciar(t)
			id := e.CrearEstacion(t, "Rechazo", c.estado, c.anclaje)
			_, tok := e.Usuario(t, true, dias(30))
			antes := fotografiar(t, e)

			status, cod := desbloquear(t, e, tok, codigo(id, 1))

			if status != c.status || cod != c.codigo {
				t.Fatalf("respuesta = %d %s; quiero %d %s", status, cod, c.status, c.codigo)
			}
			if d := fotografiar(t, e); d != antes {
				t.Errorf("el rechazo tuvo efectos: %+v → %+v", antes, d)
			}
		})
	}
}

func TestDesbloqueoCodigoInexistente(t *testing.T) {
	e := iniciar(t)
	e.CrearEstacion(t, "Inexistente", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))
	antes := fotografiar(t, e)

	status, cod := desbloquear(t, e, tok, "NO-EXISTE")

	if status != http.StatusNotFound || cod != httperr.AnclajeNoExiste {
		t.Fatalf("respuesta = %d %s; quiero 404 %s", status, cod, httperr.AnclajeNoExiste)
	}
	if d := fotografiar(t, e); d != antes {
		t.Errorf("el rechazo tuvo efectos: %+v → %+v", antes, d)
	}
}

func TestDesbloqueoCodigoVacio(t *testing.T) {
	e := iniciar(t)
	_, tok := e.Usuario(t, true, dias(30))
	for _, cuerpo := range []string{`{}`, `{"codigo_anclaje":"   "}`, `no-json`} {
		if rec := e.Post(t, "/alquileres/desbloqueo", tok, cuerpo); rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s → %d; quiero 422", cuerpo, rec.Code)
		}
	}
}

func TestDesbloqueoRequiereRolUsuario(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Rol", models.EstacionOperativa, tu.ConBici)
	if status, cod := desbloquear(t, e, e.Token(t, models.RolGestor), codigo(id, 1)); status != http.StatusForbidden || cod != httperr.RolInsuficiente {
		t.Errorf("gestor = %d %s; quiero 403 %s", status, cod, httperr.RolInsuficiente)
	}
	if status, _ := desbloquear(t, e, "", codigo(id, 1)); status != http.StatusUnauthorized {
		t.Errorf("sin sesión = %d; quiero 401", status)
	}
}
