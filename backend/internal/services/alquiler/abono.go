// Causa concreta del abono no vigente (spec-002, research D15 y D16): una sola
// función decide la vigencia y el texto, para que el aviso previo de
// GET /alquileres/activo y el rechazo del desbloqueo digan lo mismo.
package alquiler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"si16/backend/internal/models"
)

// Causas del abono no vigente (FR-002).
var (
	ErrSinAbono      = errors.New("alquiler: sin abono anual")
	ErrAbonoCaducado = errors.New("alquiler: abono anual caducado")
)

const mensajeSinAbono = "No tienes abono anual"

// CausaAbono es el estado del abono en un instante: si no es vigente, Causa es
// ErrSinAbono o ErrAbonoCaducado y Mensaje es el texto que ve el usuario.
type CausaAbono struct {
	Vigente bool
	Causa   error
	Mensaje string
}

// RechazoAbono es el error del desbloqueo por abono: envuelve la causa (para
// errors.Is) y lleva el mensaje con la fecha.
type RechazoAbono struct {
	Causa   error
	Mensaje string
}

func (r *RechazoAbono) Error() string { return r.Causa.Error() + ": " + r.Mensaje }
func (r *RechazoAbono) Unwrap() error { return r.Causa }

// causaAbono aplica FR-002 con la fecha de negocio `hoy` (Europe/Madrid, D1).
// Inactivo o sin fecha es «sin abono»; activo con la fecha pasada, «caducado».
// La fecha viene de una columna DATE, así que se formatea sin cambiar de zona.
func causaAbono(u models.Usuario, hoy time.Time) CausaAbono {
	switch {
	case models.AbonoVigente(u, hoy):
		return CausaAbono{Vigente: true}
	case !u.AbonoActivo || u.FechaVencimientoAbono == nil:
		return CausaAbono{Causa: ErrSinAbono, Mensaje: mensajeSinAbono}
	default:
		return CausaAbono{
			Causa:   ErrAbonoCaducado,
			Mensaje: "Tu abono anual caducó el " + u.FechaVencimientoAbono.Format("02/01/2006"),
		}
	}
}

// rechazo convierte una causa no vigente en el error del desbloqueo.
func (c CausaAbono) rechazo() error {
	return &RechazoAbono{Causa: c.Causa, Mensaje: c.Mensaje}
}

// Abono devuelve el estado del abono del usuario en este momento (FR-010). Es
// informativo: el desbloqueo vuelve a comprobarlo en su transacción.
func (s *Servicio) Abono(ctx context.Context, usuarioID int64) (CausaAbono, error) {
	var u models.Usuario
	err := s.pool.QueryRow(ctx, `SELECT abono_activo, fecha_vencimiento_abono FROM usuarios WHERE id = $1`, usuarioID).
		Scan(&u.AbonoActivo, &u.FechaVencimientoAbono)
	if errors.Is(err, pgx.ErrNoRows) {
		return CausaAbono{Causa: ErrSinAbono, Mensaje: mensajeSinAbono}, nil
	}
	if err != nil {
		return CausaAbono{}, fmt.Errorf("alquiler: leer abono: %w", err)
	}
	return causaAbono(u, fechaNegocio(time.Now())), nil
}
