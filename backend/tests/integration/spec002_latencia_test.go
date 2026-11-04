// T030 — SC-003: un desbloqueo válido se confirma en menos de 10 s, también con
// Redis degradado (research D6, D10).
package integration

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"si16/backend/internal/cache"
	"si16/backend/internal/config"
	"si16/backend/internal/models"
	"si16/backend/internal/routes"
	"si16/backend/internal/services/estacion"
	tu "si16/backend/internal/testutil"
)

const limiteDesbloqueo = 10 * time.Second

func TestDesbloqueoEnMenosDe10s(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Latencia", models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))

	inicio := time.Now()
	status, cod := desbloquear(t, e, tok, codigo(id, 1))

	if status != http.StatusCreated {
		t.Fatalf("respuesta = %d %s; quiero 201", status, cod)
	}
	if d := time.Since(inicio); d >= limiteDesbloqueo {
		t.Errorf("tardó %v; quiero < %v", d, limiteDesbloqueo)
	}
}

// redisMudo acepta conexiones y nunca responde: el peor caso de un Redis
// degradado, porque el cliente espera en vez de fallar enseguida.
func redisMudo(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { _ = c.Close() })
		}
	}()
	return l.Addr().String()
}

func TestDesbloqueoConRedisDegradado(t *testing.T) {
	e := iniciar(t)
	id := e.CrearEstacion(t, "Degradado", models.EstacionOperativa, tu.ConBici)
	uid, tok := e.Usuario(t, true, dias(30))

	mudo := cache.Nuevo(redisMudo(t), redisDB)
	t.Cleanup(func() { _ = mudo.Cerrar() })
	router := routes.NuevoRouter(routes.Deps{
		Cfg: config.Config{TarifaMinuto: 0.20}, Pool: e.Pool, Cache: mudo, Auth: e.Emisor,
		Estacion: estacion.Nuevo(e.Pool, mudo, nil),
	})

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/alquileres/desbloqueo",
		strings.NewReader(fmt.Sprintf(`{"codigo_anclaje":%q}`, codigo(id, 1))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()

	inicio := time.Now()
	router.ServeHTTP(rec, req)
	d := time.Since(inicio)

	if rec.Code != http.StatusCreated {
		t.Fatalf("respuesta = %d %s; quiero 201", rec.Code, rec.Body)
	}
	if d >= limiteDesbloqueo {
		t.Errorf("tardó %v; quiero < %v", d, limiteDesbloqueo)
	}
	var viajes int
	if err := e.Pool.QueryRow(e.Ctx, `SELECT count(*) FROM viajes WHERE usuario_id = $1 AND estado = 'en_curso'`, uid).Scan(&viajes); err != nil {
		t.Fatal(err)
	}
	if viajes != 1 {
		t.Errorf("viajes activos = %d; quiero 1", viajes)
	}
}
