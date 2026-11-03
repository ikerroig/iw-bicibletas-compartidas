// Package models contiene los structs y enums compartidos por las 8 specs,
// espejo de backend/migrations/0001_esquema_base.sql (data-model.md).
package models

import "time"

// Rol de usuario (research R5).
type Rol string

const (
	RolUsuario           Rol = "usuario"
	RolOperarioTaller    Rol = "operario_taller"
	RolOperarioLogistica Rol = "operario_logistica"
	RolGestor            Rol = "gestor"
)

// Estados de estación.
const (
	EstacionOperativa       = "operativa"
	EstacionFueraDeServicio = "fuera_de_servicio"
)

// Estados de bicicleta.
const (
	BicicletaDisponible = "disponible"
	BicicletaEnUso      = "en_uso"
	BicicletaAveriada   = "averiada"
	BicicletaEnTaller   = "en_taller"
)

// Estados de anclaje.
const (
	AnclajeLibre           = "libre"
	AnclajeOcupado         = "ocupado"
	AnclajeFueraDeServicio = "fuera_de_servicio"
)

// Estados de viaje.
const (
	ViajeEnCurso    = "en_curso"
	ViajeFinalizado = "finalizado"
)

// Estados de incidencia.
const (
	IncidenciaPendiente    = "pendiente"
	IncidenciaEnReparacion = "en_reparacion"
	IncidenciaResuelta     = "resuelta"
)

// Tipos de alerta de balanceo.
const (
	AlertaVaciaMenos10 = "vacia_menos_10"
	AlertaLlenaMas90   = "llena_mas_90"
)

// Usuario (tabla usuarios). PasswordHash nunca se serializa.
type Usuario struct {
	ID                    int64      `json:"id"`
	Nombre                string     `json:"nombre"`
	Email                 string     `json:"email"`
	PasswordHash          string     `json:"-"`
	Rol                   Rol        `json:"rol"`
	AbonoActivo           bool       `json:"abono_activo"`
	FechaVencimientoAbono *time.Time `json:"-"`
	CreadoEn              time.Time  `json:"-"`
}

// AbonoVigente aplica la regla crítica (constitución I, data-model §1):
// vigente ⇔ abono_activo AND fecha_vencimiento_abono >= hoy. Datos
// incoherentes (activo sin fecha) cuentan como caducado.
func AbonoVigente(u Usuario, hoy time.Time) bool {
	if !u.AbonoActivo || u.FechaVencimientoAbono == nil {
		return false
	}
	y, m, d := hoy.Date()
	inicioHoy := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	fy, fm, fd := u.FechaVencimientoAbono.Date()
	venc := time.Date(fy, fm, fd, 0, 0, 0, 0, time.UTC)
	return !venc.Before(inicioHoy)
}

// Estacion (tabla estaciones).
type Estacion struct {
	ID             int64   `json:"id"`
	Nombre         string  `json:"nombre"`
	Latitud        float64 `json:"latitud"`
	Longitud       float64 `json:"longitud"`
	CapacidadTotal int     `json:"capacidad_total"`
	Estado         string  `json:"estado"`
}

// Anclaje (tabla anclajes).
type Anclaje struct {
	ID            int64  `json:"id"`
	EstacionID    int64  `json:"estacion_id"`
	CodigoAnclaje string `json:"codigo_anclaje"`
	Estado        string `json:"estado"`
	BicicletaID   *int64 `json:"bicicleta_id"`
}

// Bicicleta (tabla bicicletas).
type Bicicleta struct {
	ID     int64  `json:"id"`
	Codigo string `json:"codigo"`
	Estado string `json:"estado"`
}

// Viaje (tabla viajes).
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
}

// Incidencia (tabla incidencias_mantenimientos).
type Incidencia struct {
	ID                 int64      `json:"id"`
	BicicletaID        int64      `json:"bicicleta_id"`
	UsuarioReportaID   *int64     `json:"usuario_reporta_id"`
	OperarioID         *int64     `json:"operario_id"`
	TipoFallo          string     `json:"tipo_fallo"`
	DescripcionReporte string     `json:"descripcion_reporte"`
	PiezasSustituidas  *string    `json:"piezas_sustituidas"`
	NotasTaller        *string    `json:"notas_taller"`
	FechaReporte       time.Time  `json:"fecha_reporte"`
	FechaReparacion    *time.Time `json:"fecha_reparacion"`
	Estado             string     `json:"estado"`
}

// AlertaBalanceo (tabla alertas_balanceo).
type AlertaBalanceo struct {
	ID                  int64     `json:"id"`
	EstacionID          int64     `json:"estacion_id"`
	TipoAlerta          string    `json:"tipo_alerta"`
	PorcentajeOcupacion float64   `json:"porcentaje_ocupacion"`
	Severidad           string    `json:"severidad"`
	FechaGeneracion     time.Time `json:"fecha_generacion"`
	Resuelta            bool      `json:"resuelta"`
}
