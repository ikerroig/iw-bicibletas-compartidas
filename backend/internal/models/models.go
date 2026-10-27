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
