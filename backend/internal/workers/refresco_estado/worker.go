// Package refresco_estado es la red de seguridad de frescura (spec-001, T026):
// recalcula las ~60 estaciones con una única consulta agregada cada
// IntervaloBarrido y las escribe en Redis. Con el polling de 15 s del
// frontend deja el peor caso extremo a extremo en ≈ 28 s < 30 s, incluso si
// falla la escritura síncrona post-commit (research R16).
package refresco_estado

import (
	"context"
	"log/slog"
	"time"

	"si16/backend/internal/cache"
	"si16/backend/internal/services/estacion"
)

// IntervaloBarrido de producción (R16; antes 15 s).
const IntervaloBarrido = 10 * time.Second

// Worker de barrido.
type Worker struct {
	svc       *estacion.Servicio
	cache     *cache.Cliente
	log       *slog.Logger
	intervalo time.Duration
}

// Nuevo crea el worker; los tests pasan un intervalo corto (T099).
func Nuevo(svc *estacion.Servicio, c *cache.Cliente, log *slog.Logger, intervalo time.Duration) *Worker {
	if log == nil {
		log = slog.Default()
	}
	if intervalo <= 0 {
		intervalo = IntervaloBarrido
	}
	return &Worker{svc: svc, cache: c, log: log, intervalo: intervalo}
}

// Iniciar ejecuta un barrido inmediato y luego uno por intervalo hasta que se
// cancele ctx. Un barrido fallido se registra y se reintenta en el siguiente.
func (w *Worker) Iniciar(ctx context.Context) {
	w.barrer(ctx)
	t := time.NewTicker(w.intervalo)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.barrer(ctx)
		}
	}
}

func (w *Worker) barrer(ctx context.Context) {
	// Cada barrido debe acabar antes del siguiente para no acumularse.
	ctxB, cancelar := context.WithTimeout(ctx, w.intervalo)
	defer cancelar()
	n, err := w.svc.RecalcularTodas(ctxB)
	if err != nil {
		if ctx.Err() == nil {
			w.log.Warn("refresco_estado: barrido fallido", "error", err)
		}
		return
	}
	if err := w.cache.MarcarBarrido(ctxB, time.Now()); err != nil {
		w.log.Warn("refresco_estado: no se pudo marcar el barrido", "error", err)
	}
	w.log.Debug("refresco_estado: barrido", "estaciones", n)
}
