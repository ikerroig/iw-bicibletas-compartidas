// Traslado de lote entre estaciones (FR-006/FR-007 — research D4).
// Validación de entrada → 422; estado/lote → 409 con ROLLBACK íntegro;
// movimiento en una transacción FOR UPDATE + post-commit Redis y alertas inline.
package balanceo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"si16/backend/internal/apierr"
	"si16/backend/internal/models"
)

func errEntrada(mensaje string) *apierr.Error {
	return apierr.New(422, "ENTRADA_INVALIDA", mensaje)
}

func errCapacidad(mensaje string) *apierr.Error {
	return apierr.New(409, "ESTACION_SIN_CAPACIDAD", mensaje)
}

func errInterno() *apierr.Error {
	return apierr.New(500, "ERROR_INTERNO", "Error interno al confirmar el traslado")
}

// ParseTraslado interpreta el body del POST. 422 = no se puede interpretar la
// petición (JSON malformado, campos ausentes/no numéricos, unidades < 1,
// origen == destino): nunca llega a evaluarse contra el estado real.
func ParseTraslado(body []byte) (models.TrasladoRequest, *apierr.Error) {
	var campos map[string]json.RawMessage
	if err := json.Unmarshal(body, &campos); err != nil {
		return models.TrasladoRequest{}, errEntrada("JSON inválido: la petición no puede interpretarse")
	}
	for _, campo := range []string{"estacion_origen_id", "estacion_destino_id", "unidades"} {
		if _, ok := campos[campo]; !ok {
			return models.TrasladoRequest{}, errEntrada(fmt.Sprintf("campo obligatorio ausente: %s", campo))
		}
	}
	var req models.TrasladoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return models.TrasladoRequest{}, errEntrada("JSON inválido: tipos de campo incorrectos")
	}
	if e := validarEntrada(req); e != nil {
		return models.TrasladoRequest{}, e
	}
	return req, nil
}

// validarEntrada es pura: no toca datos (garantía verificada en T7g).
func validarEntrada(req models.TrasladoRequest) *apierr.Error {
	if req.Unidades < 1 {
		return errEntrada("unidades debe ser al menos 1")
	}
	if req.EstacionOrigenID == req.EstacionDestinoID {
		return errEntrada("origen y destino deben ser estaciones distintas")
	}
	return nil
}

// filaEstacion es la fila bloqueada de estaciones dentro de la transacción.
type filaEstacion struct {
	id             int64
	nombre         string
	latitud        float64
	longitud       float64
	capacidadTotal int
	estado         string
}

// ConfirmarTraslado ejecuta el traslado completo (data-model.md §4):
// bloqueo FOR UPDATE, validación 409, movimiento de anclajes, COMMIT y
// post-commit (Redis de ambas + evaluación inline de umbrales). Devuelve los
// EstacionEstado recalculados de la propia respuesta (FR-007, SC-003).
func (s *Servicio) ConfirmarTraslado(ctx context.Context, req models.TrasladoRequest) (models.EstacionEstado, models.EstacionEstado, *apierr.Error) {
	var cero models.EstacionEstado
	if s.estacion == nil {
		return cero, cero, errInterno()
	}
	if e := validarEntrada(req); e != nil {
		return cero, cero, e
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Printf("balanceo: begin: %v", err)
		return cero, cero, errInterno()
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op tras COMMIT

	// 1. Bloqueo de estaciones en orden de id (sin deadlocks).
	rows, err := tx.Query(ctx,
		`SELECT id, nombre, latitud::float8, longitud::float8, capacidad_total, estado
		   FROM estaciones
		  WHERE id = ANY($1::bigint[])
		  ORDER BY id FOR UPDATE`,
		[]int64{req.EstacionOrigenID, req.EstacionDestinoID})
	if err != nil {
		log.Printf("balanceo: bloqueo estaciones: %v", err)
		return cero, cero, errInterno()
	}
	filas := map[int64]filaEstacion{}
	for rows.Next() {
		var f filaEstacion
		if err := rows.Scan(&f.id, &f.nombre, &f.latitud, &f.longitud, &f.capacidadTotal, &f.estado); err != nil {
			rows.Close()
			log.Printf("balanceo: scan estación: %v", err)
			return cero, cero, errInterno()
		}
		filas[f.id] = f
	}
	rows.Close()
	if rows.Err() != nil {
		log.Printf("balanceo: filas estaciones: %v", rows.Err())
		return cero, cero, errInterno()
	}

	// 2. Validación de estado/lote → 409 (payload bien formado vs estado real).
	origen, ok := filas[req.EstacionOrigenID]
	if !ok {
		return cero, cero, errCapacidad("la estación origen no existe")
	}
	destino, ok := filas[req.EstacionDestinoID]
	if !ok {
		return cero, cero, errCapacidad("la estación destino no existe")
	}
	if origen.estado != "operativa" {
		return cero, cero, errCapacidad(fmt.Sprintf("la estación origen %q está fuera de servicio", origen.nombre))
	}
	if destino.estado != "operativa" {
		return cero, cero, errCapacidad(fmt.Sprintf("la estación destino %q está fuera de servicio", destino.nombre))
	}

	var bicisOrigen, libresDestino int
	if err := tx.QueryRow(ctx,
		`SELECT (SELECT COUNT(*)
		           FROM anclajes a
		           JOIN bicicletas b ON b.id = a.bicicleta_id
		          WHERE a.estacion_id = $1 AND b.estado = 'disponible'),
		        (SELECT COUNT(*)
		           FROM anclajes
		          WHERE estacion_id = $2 AND estado = 'libre')`,
		origen.id, destino.id).Scan(&bicisOrigen, &libresDestino); err != nil {
		log.Printf("balanceo: recuentos: %v", err)
		return cero, cero, errInterno()
	}
	if req.Unidades > bicisOrigen {
		return cero, cero, errCapacidad(fmt.Sprintf(
			"la estación origen no tiene bicicletas suficientes: disponibles %d, unidades pedidas %d",
			bicisOrigen, req.Unidades))
	}
	if req.Unidades > libresDestino {
		return cero, cero, errCapacidad(fmt.Sprintf(
			"la estación destino no tiene anclajes libres suficientes: libres %d, unidades pedidas %d",
			libresDestino, req.Unidades))
	}

	// 3. Selección del lote y de los huecos con FOR UPDATE (los recuentos
	// previos se re-verifican: caso límite de datos desactualizados / carreras).
	biciRows, err := tx.Query(ctx,
		`SELECT a.id, a.bicicleta_id
		   FROM anclajes a
		   JOIN bicicletas b ON b.id = a.bicicleta_id
		  WHERE a.estacion_id = $1 AND a.estado = 'ocupado' AND b.estado = 'disponible'
		  ORDER BY a.id LIMIT $2 FOR UPDATE OF a`, origen.id, req.Unidades)
	if err != nil {
		log.Printf("balanceo: selección lote: %v", err)
		return cero, cero, errInterno()
	}
	var anclajesOrigen []int64
	var bicisIDs []int64
	for biciRows.Next() {
		var anclajeID, biciID int64
		if err := biciRows.Scan(&anclajeID, &biciID); err != nil {
			biciRows.Close()
			log.Printf("balanceo: scan lote: %v", err)
			return cero, cero, errInterno()
		}
		anclajesOrigen = append(anclajesOrigen, anclajeID)
		bicisIDs = append(bicisIDs, biciID)
	}
	biciRows.Close()
	if biciRows.Err() != nil {
		log.Printf("balanceo: lote err: %v", biciRows.Err())
		return cero, cero, errInterno()
	}

	huecoRows, err := tx.Query(ctx,
		`SELECT id FROM anclajes
		  WHERE estacion_id = $1 AND estado = 'libre'
		  ORDER BY id LIMIT $2 FOR UPDATE`, destino.id, req.Unidades)
	if err != nil {
		log.Printf("balanceo: selección huecos: %v", err)
		return cero, cero, errInterno()
	}
	var anclajesDestino []int64
	for huecoRows.Next() {
		var id int64
		if err := huecoRows.Scan(&id); err != nil {
			huecoRows.Close()
			log.Printf("balanceo: scan huecos: %v", err)
			return cero, cero, errInterno()
		}
		anclajesDestino = append(anclajesDestino, id)
	}
	huecoRows.Close()
	if huecoRows.Err() != nil {
		log.Printf("balanceo: huecos err: %v", huecoRows.Err())
		return cero, cero, errInterno()
	}
	if len(anclajesOrigen) != req.Unidades || len(anclajesDestino) != req.Unidades {
		return cero, cero, errCapacidad("el estado de la red cambió durante la confirmación; vuelva a intentarlo")
	}

	// 4. Movimiento: origen → libre/sin bici; destino → ocupado con la bici
	// (las bicicletas conservan estado 'disponible').
	if _, err := tx.Exec(ctx,
		`UPDATE anclajes SET estado = 'libre', bicicleta_id = NULL
		  WHERE id = ANY($1::bigint[])`, anclajesOrigen); err != nil {
		log.Printf("balanceo: vaciar origen: %v", err)
		return cero, cero, errInterno()
	}
	for i, biciID := range bicisIDs {
		if _, err := tx.Exec(ctx,
			`UPDATE anclajes SET estado = 'ocupado', bicicleta_id = $1 WHERE id = $2`,
			biciID, anclajesDestino[i]); err != nil {
			log.Printf("balanceo: anclar destino: %v", err)
			return cero, cero, errInterno()
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("balanceo: commit: %v", err)
		return cero, cero, errInterno()
	}

	// 5. Post-commit best-effort (patrón de la devolución): el traslado ya está
	// aplicado; si algo falla, el worker lo corrige en < 60 s (SC-001).
	if err := s.estacion.RecalcularEstacion(ctx, []int64{origen.id, destino.id}); err != nil {
		log.Printf("balanceo: estado en vivo (traslado aplicado igualmente): %v", err)
	}
	for _, id := range []int64{origen.id, destino.id} {
		if err := s.EvaluarEstacion(ctx, id); err != nil {
			log.Printf("balanceo: evaluación inline estación %d: %v", id, err)
		}
	}

	origenEstado, oErr := s.estadoEstacion(ctx, origen.id)
	if oErr != nil {
		log.Printf("balanceo: estado origen: %v", oErr)
		return cero, cero, errInterno()
	}
	destinoEstado, dErr := s.estadoEstacion(ctx, destino.id)
	if dErr != nil {
		log.Printf("balanceo: estado destino: %v", dErr)
		return cero, cero, errInterno()
	}
	return origenEstado, destinoEstado, nil
}

// estadoEstacion monta el EstacionEstado de la respuesta con la fila de
// estaciones + los recuentos R9 recalculados tras el COMMIT (mismos valores
// que acaba de escribir en Redis).
func (s *Servicio) estadoEstacion(ctx context.Context, id int64) (models.EstacionEstado, error) {
	var e models.EstacionEstado
	err := s.pool.QueryRow(ctx,
		`SELECT id, nombre, latitud::float8, longitud::float8, capacidad_total, estado
		   FROM estaciones WHERE id = $1`, id).Scan(
		&e.ID, &e.Nombre, &e.Latitud, &e.Longitud, &e.CapacidadTotal, &e.Estado)
	if err != nil {
		return models.EstacionEstado{}, err
	}
	if err := s.pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*)
		           FROM anclajes a
		           JOIN bicicletas b ON b.id = a.bicicleta_id
		          WHERE a.estacion_id = $1 AND b.estado = 'disponible'),
		        (SELECT COUNT(*) FROM anclajes WHERE estacion_id = $1 AND estado = 'libre')`,
		id).Scan(&e.BicisDisponibles, &e.AnclajesLibres); err != nil {
		return models.EstacionEstado{}, err
	}
	e.OcupacionPct = ocupacionPct(e.BicisDisponibles, e.CapacidadTotal, e.Estado)
	e.ActualizadoEn = time.Now().UTC().Format(time.RFC3339)
	return e, nil
}
