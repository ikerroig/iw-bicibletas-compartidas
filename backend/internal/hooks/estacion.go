// Package hooks publica el evento "estaciones cambiadas" para desacoplar las
// specs que mutan estaciones/bicicletas/anclajes de las que reaccionan
// (constitución V, tarea T012).
//
// Flujo único, sin ciclos:
//
//	spec que muta ──commit──▶ estacion.Servicio.RecalcularEstacion(ctx, ids)
//	                              ├─ reescribe Redis
//	                              └─ hooks.EstacionCambiada(ctx, ids) ──▶ suscriptores (p. ej. US7 umbrales)
//
// Las specs que mutan NO llaman a EstacionCambiada directamente, y ningún
// suscriptor debe volver a llamar a RecalcularEstacion.
package hooks

import (
	"context"
	"log/slog"
	"sync"
)

// Suscriptor reacciona a un cambio en las estaciones indicadas.
type Suscriptor func(ctx context.Context, ids []int64) error

var (
	mu           sync.RWMutex
	suscriptores []Suscriptor
)

// Suscribir registra fn (normalmente desde el init() del paquete de la spec).
func Suscribir(fn Suscriptor) {
	mu.Lock()
	defer mu.Unlock()
	suscriptores = append(suscriptores, fn)
}

// EstacionCambiada notifica a todos los suscriptores, en orden de registro. El
// error o el panic de un suscriptor se registra en el log y no corta a los
// demás. Solo la llama estacion.Servicio.RecalcularEstacion.
func EstacionCambiada(ctx context.Context, ids []int64) {
	if len(ids) == 0 {
		return
	}
	mu.RLock()
	lista := append([]Suscriptor(nil), suscriptores...)
	mu.RUnlock()
	for i, fn := range lista {
		notificar(ctx, i, fn, ids)
	}
}

func notificar(ctx context.Context, i int, fn Suscriptor, ids []int64) {
	defer func() {
		if p := recover(); p != nil {
			slog.Error("hooks: panic en suscriptor", "suscriptor", i, "panic", p)
		}
	}()
	if err := fn(ctx, ids); err != nil {
		slog.Error("hooks: error en suscriptor", "suscriptor", i, "error", err)
	}
}

// ReiniciarParaTests elimina todas las suscripciones. Solo para tests.
func ReiniciarParaTests() {
	mu.Lock()
	defer mu.Unlock()
	suscriptores = nil
}
