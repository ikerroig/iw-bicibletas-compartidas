/**
 * Cliente del endpoint HU-07: GET /balanceo/alertas y POST /balanceo/traslados.
 * Contrato: specs/spec-007/contracts/balanceo.md (envelope global apierr para
 * los errores: 422 ENTRADA_INVALIDA / 409 ESTACION_SIN_CAPACIDAD).
 */
import { api } from "../api";

export interface AlertaBalanceo {
  id: number;
  estacion_id: number;
  estacion_nombre: string;
  tipo_alerta: string; // "vacia_menos_10" | "llena_mas_90"
  porcentaje_ocupacion: number;
  severidad: string;
  fecha_generacion: string;
  resuelta: boolean;
}

export interface PanelBalanceo {
  items: AlertaBalanceo[];
}

export interface EstacionEstado {
  id: number;
  nombre: string;
  latitud: number;
  longitud: number;
  capacidad_total: number;
  estado: string;
  bicis_disponibles: number;
  anclajes_libres: number;
  ocupacion_pct: number;
  actualizado_en: string;
}

export interface TrasladoRequest {
  estacion_origen_id: number;
  estacion_destino_id: number;
  unidades: number;
}

export interface TrasladoResponse {
  origen: EstacionEstado;
  destino: EstacionEstado;
}

/** Alertas abiertas, ordenadas por urgencia (orden del backend, const. I). */
export function listarAlertas(): Promise<PanelBalanceo> {
  return api.request<PanelBalanceo>("/balanceo/alertas");
}

/** Confirma el traslado de lote (FR-006) y devuelve ambas estaciones (FR-007). */
export function confirmarTraslado(req: TrasladoRequest): Promise<TrasladoResponse> {
  return api.request<TrasladoResponse>("/balanceo/traslados", {
    method: "POST",
    body: JSON.stringify(req),
  });
}
