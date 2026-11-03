// Cliente de spec-001 (T028). Tipos = esquemas EstacionEstado / EstacionDetalle
// de specs/spec-001/contracts/openapi.yaml.
import { apiFetch } from '../api';

export type EstadoServicio = 'operativa' | 'fuera_de_servicio';

export interface EstacionEstado {
  id: number;
  nombre: string;
  latitud: number;
  longitud: number;
  capacidad_total: number;
  estado: EstadoServicio;
  bicis_disponibles: number;
  anclajes_libres: number;
  ocupacion_pct?: number;
  actualizado_en: string;
}

export interface AnclajeDetalle {
  codigo_anclaje: string;
  estado: 'libre' | 'ocupado' | 'fuera_de_servicio';
  bicicleta_codigo: string | null;
  bicicleta_estado: string | null;
}

export interface EstacionDetalle extends EstacionEstado {
  anclajes?: AnclajeDetalle[];
}

export interface SnapshotEstaciones {
  actualizado_en: string;
  estaciones: EstacionEstado[];
}

export function listarEstaciones(signal?: AbortSignal): Promise<SnapshotEstaciones> {
  return apiFetch<SnapshotEstaciones>('/estaciones', { signal });
}

export function detalleEstacion(id: number, signal?: AbortSignal): Promise<EstacionDetalle> {
  return apiFetch<EstacionDetalle>(`/estaciones/${id}`, { signal });
}
