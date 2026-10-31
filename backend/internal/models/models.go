// Package models contiene los structs compartidos del modelo SI-16
// (espejo exacto de las tablas de specs/spec-001/data-model.md).
// Nota: coste_total (NUMERIC) se modela como float64 para JSON; en SQL se
// usan casts explícitos (::float8 / ::numeric) en ambos sentidos.
package models

import "time"

type Usuario struct {
	ID                    int64      `json:"id"`
	Nombre                string     `json:"nombre"`
	Email                 string     `json:"email"`
	Rol                   string     `json:"rol"`
	AbonoActivo           bool       `json:"abono_activo"`
	FechaVencimientoAbono *time.Time `json:"fecha_vencimiento_abono"`
}

type Estacion struct {
	ID             int64  `json:"id"`
	Nombre         string `json:"nombre"`
	Latitud        string `json:"latitud"`
	Longitud       string `json:"longitud"`
	CapacidadTotal int    `json:"capacidad_total"`
	Estado         string `json:"estado"`
}

type Bicicleta struct {
	ID     int64  `json:"id"`
	Codigo string `json:"codigo"`
	Estado string `json:"estado"`
}

type Anclaje struct {
	ID            int64  `json:"id"`
	EstacionID    int64  `json:"estacion_id"`
	CodigoAnclaje string `json:"codigo_anclaje"`
	Estado        string `json:"estado"`
	BicicletaID   *int64 `json:"bicicleta_id"`
}

// Viaje es la entidad que la HU-03 cierra y tarifica.
type Viaje struct {
	ID                int64      `json:"id"`
	UsuarioID         int64      `json:"usuario_id"`
	BicicletaID       int64      `json:"bicicleta_id"`
	EstacionOrigenID  int64      `json:"estacion_origen_id"`
	EstacionDestinoID *int64     `json:"estacion_destino_id"`
	FechaInicio       time.Time  `json:"fecha_inicio"`
	FechaFin          *time.Time `json:"fecha_fin"`
	DuracionMinutos   *int       `json:"duracion_minutos"`
	CosteTotal        float64    `json:"coste_total"`
	Estado            string     `json:"estado"`
	Etiqueta          *string    `json:"etiqueta,omitempty"`
}

// AlertaBalanceo es la fila del panel de balanceo (HU-07); espejo JSON del
// schema global AlertaBalanceo (openapi.yaml). estacion_nombre se resuelve
// en backend con JOIN; porcentaje_ocupacion (NUMERIC) → float64 (::float8).
type AlertaBalanceo struct {
	ID                  int64     `json:"id"`
	EstacionID          int64     `json:"estacion_id"`
	EstacionNombre      string    `json:"estacion_nombre"`
	TipoAlerta          string    `json:"tipo_alerta"`
	PorcentajeOcupacion float64   `json:"porcentaje_ocupacion"`
	Severidad           string    `json:"severidad"`
	FechaGeneracion     time.Time `json:"fecha_generacion"`
	Resuelta            bool      `json:"resuelta"`
}

// TrasladoRequest es el body de POST /balanceo/traslados (schema global;
// unidades tiene minimum: 1 validado server-side → 422 ENTRADA_INVALIDA).
type TrasladoRequest struct {
	EstacionOrigenID  int64 `json:"estacion_origen_id"`
	EstacionDestinoID int64 `json:"estacion_destino_id"`
	Unidades          int   `json:"unidades"`
}

// EstacionEstado es el estado recalculado de una estación (schema global;
// latitud/longitud NUMERIC(9,6) → float64 con cast ::float8 en SQL).
type EstacionEstado struct {
	ID               int64   `json:"id"`
	Nombre           string  `json:"nombre"`
	Latitud          float64 `json:"latitud"`
	Longitud         float64 `json:"longitud"`
	CapacidadTotal   int     `json:"capacidad_total"`
	Estado           string  `json:"estado"`
	BicisDisponibles int     `json:"bicis_disponibles"`
	AnclajesLibres   int     `json:"anclajes_libres"`
	OcupacionPct     float64 `json:"ocupacion_pct"`
	ActualizadoEn    string  `json:"actualizado_en"`
}
