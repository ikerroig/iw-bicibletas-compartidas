// Regresiones de integración entre specs detectadas al revisar main con las 8
// specs fusionadas (rama fix/integracion-post-merge).
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"si16/backend/internal/hooks"
	"si16/backend/internal/models"
	tu "si16/backend/internal/testutil"
)

func devolver(t *testing.T, e *tu.Entorno, token, cod string) (int, string) {
	t.Helper()
	rec := e.Post(t, "/devoluciones", token, fmt.Sprintf(`{"codigo_anclaje":%q}`, cod))
	var r struct {
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	return rec.Code, r.Error.Codigo
}

// conViajeEnCurso desbloquea una bici de una estación nueva y devuelve el token.
func conViajeEnCurso(t *testing.T, e *tu.Entorno) string {
	t.Helper()
	origen := e.CrearEstacion(t, fmt.Sprintf("Origen %s", t.Name()), models.EstacionOperativa, tu.ConBici)
	_, tok := e.Usuario(t, true, dias(30))
	if status, cod := desbloquear(t, e, tok, codigo(origen, 1)); status != http.StatusCreated {
		t.Fatalf("desbloqueo = %d %s", status, cod)
	}
	return tok
}

// El código impreso vale igual para desbloquear (HU-02 FR-001) que para devolver.
func TestDevolucionCodigoSinDistinguirMayusculas(t *testing.T) {
	e := iniciar(t)
	tok := conViajeEnCurso(t, e)
	destino := e.CrearEstacion(t, "Destino", models.EstacionOperativa, tu.Libre)

	if status, cod := devolver(t, e, tok, fmt.Sprintf("  e%d-a01 ", destino)); status != http.StatusOK {
		t.Fatalf("devolución en minúsculas = %d %s; quiero 200", status, cod)
	}
}

// anclajeRotoConBici no cuenta como disponible (HU-01 R9/R19).
var anclajeRotoConBici = tu.Anclaje{Estado: models.AnclajeFueraDeServicio, Bicicleta: models.BicicletaDisponible}

// Toda mutación actualiza el estado en vivo con la misma fórmula que el mapa y
// notifica el hook (T020).
func TestDevolucionUsaElRecalculoComun(t *testing.T) {
	e := iniciar(t)
	tok := conViajeEnCurso(t, e)
	destino := e.CrearEstacion(t, "Destino", models.EstacionOperativa, tu.Libre, anclajeRotoConBici)
	notificadas := escucharHook()

	if status, cod := devolver(t, e, tok, codigo(destino, 1)); status != http.StatusOK {
		t.Fatalf("devolución = %d %s; quiero 200", status, cod)
	}

	comprobarEstado(t, e, destino, 1, 0)
	comprobarNotificada(t, notificadas, destino)
}

func TestTrasladoUsaElRecalculoComun(t *testing.T) {
	e := iniciar(t)
	origen := e.CrearEstacion(t, "Origen traslado", models.EstacionOperativa, tu.ConBici, tu.ConBici)
	destino := e.CrearEstacion(t, "Destino traslado", models.EstacionOperativa, tu.Libre, anclajeRotoConBici)
	notificadas := escucharHook()

	rec := e.Post(t, "/balanceo/traslados", e.Token(t, models.RolOperarioLogistica),
		fmt.Sprintf(`{"estacion_origen_id":%d,"estacion_destino_id":%d,"unidades":1}`, origen, destino))
	if rec.Code != http.StatusOK {
		t.Fatalf("traslado = %d %s; quiero 200", rec.Code, rec.Body)
	}

	comprobarEstado(t, e, destino, 1, 0)
	comprobarNotificada(t, notificadas, destino)
}

func escucharHook() chan []int64 {
	c := make(chan []int64, 8)
	hooks.Suscribir(func(_ context.Context, ids []int64) error { c <- ids; return nil })
	return c
}

func comprobarEstado(t *testing.T, e *tu.Entorno, id int64, bicis, libres int) {
	t.Helper()
	h, ok := leerHash(t, e, id)
	if !ok || h.BicisDisponibles != bicis || h.AnclajesLibres != libres {
		t.Errorf("estado en vivo de %d = %+v; quiero %d bicis y %d libres", id, h, bicis, libres)
	}
}

func comprobarNotificada(t *testing.T, c chan []int64, id int64) {
	t.Helper()
	for {
		select {
		case ids := <-c:
			for _, x := range ids {
				if x == id {
					return
				}
			}
		default:
			t.Errorf("el hook no se notificó para la estación %d", id)
			return
		}
	}
}
