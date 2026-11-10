// Worker alertas (HU-07): recorre la red cada periodoBarrido (55 s — research
// D6, desviación mínima de R10 documentada) creando/resolviendo alertas.
// Cubre SC-001 (< 60 s estricto = periodo + T_barrido) para cualquier mutación
// de la red sin modificar el código de otras specs (const. V).
package alertas

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/services/balanceo"
)

// PeriodoBarrido es el intervalo del ticker. Inyectable en tests (T7h):
// se lee al crear el ticker dentro de Run.
var PeriodoBarrido = 55 * time.Second

// presupuestoBarrido: T_barrido < 5 s estricto para que periodo + T < 60 s
// (D6). Se verifica por test y se observa con el warning de Barrido.
const presupuestoBarrido = 5 * time.Second

// Run evalúa la red inmediatamente al arrancar y después cada PeriodoBarrido
// hasta cancelar el context (apagado del servidor).
func Run(ctx context.Context, pool *pgxpool.Pool) {
	barridoInicial := func() {
		if _, err := Barrido(ctx, pool); err != nil && ctx.Err() == nil {
			log.Printf("alertas: barrido: %v", err)
		}
	}
	barridoInicial()

	ticker := time.NewTicker(PeriodoBarrido)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			barridoInicial()
		}
	}
}

// Barrido ejecuta UNA pasada completa sobre las estaciones operativas y
// devuelve su duración medida (canal de lectura del presupuesto — L1).
// Una estación con error no detiene el barrido del resto.
func Barrido(ctx context.Context, pool *pgxpool.Pool) (time.Duration, error) {
	inicio := time.Now()
	svc := balanceo.New(pool)

	rows, err := pool.Query(ctx,
		`SELECT id FROM estaciones WHERE estado = 'operativa' ORDER BY id`)
	if err != nil {
		return time.Since(inicio), err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return time.Since(inicio), err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return time.Since(inicio), err
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return time.Since(inicio), ctx.Err()
		}
		if err := svc.EvaluarEstacion(ctx, id); err != nil {
			log.Printf("alertas: estación %d: %v", id, err) // no detiene el resto
		}
	}

	dur := time.Since(inicio)
	if dur >= presupuestoBarrido {
		log.Printf("alertas: WARN barrido %v ≥ presupuesto %v (SC-001: revisar periodo)", dur, presupuestoBarrido)
	}
	return dur, nil
}
