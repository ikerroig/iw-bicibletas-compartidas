/**
 * Cliente del endpoint HU-03: POST /api/v1/devoluciones
 * Contrato: specs/spec-003/contracts/devoluciones.md (Viaje + etiqueta opcional).
 */
import { api } from "../api";

export interface ViajeDevolucion {
  id: number;
  usuario_id: number;
  bicicleta_id: number;
  estacion_origen_id: number;
  estacion_destino_id: number | null;
  fecha_inicio: string;
  fecha_fin: string | null;
  duracion_minutos: number | null;
  coste_total: number;
  estado: string;
  etiqueta?: string;
}

/** Cierra el viaje activo anclando la bicicleta en `codigoAnclaje`. */
export function devolver(codigoAnclaje: string): Promise<ViajeDevolucion> {
  return api.request<ViajeDevolucion>("/devoluciones", {
    method: "POST",
    body: JSON.stringify({ codigo_anclaje: codigoAnclaje }),
  });
}
