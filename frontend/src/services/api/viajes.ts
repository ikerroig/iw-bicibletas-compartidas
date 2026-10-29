/**
 * Cliente del endpoint HU-04: GET /api/v1/viajes
 * Contrato: specs/spec-004/contracts/viajes.md — schema Viaje + nombres de
 * estación (aditivos) dentro de { total, suma_coste, items }.
 */
import { api } from "../api";

export interface ViajeHistorial {
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
  estacion_origen_nombre: string;
  estacion_destino_nombre: string | null;
}

export interface HistorialResponse {
  total: number;
  suma_coste: number;
  items: ViajeHistorial[];
}

export interface FiltroHistorial {
  desde?: string; // YYYY-MM-DD inclusivo
  hasta?: string; // YYYY-MM-DD inclusivo
  page?: number; // ≥ 1 (default 1)
  pageSize?: number; // 1…200 (default 50)
}

/** Historial propio paginado; sin filtros devuelve el periodo completo. */
export function listarViajes(f: FiltroHistorial = {}): Promise<HistorialResponse> {
  const q = new URLSearchParams();
  if (f.desde) q.set("desde", f.desde);
  if (f.hasta) q.set("hasta", f.hasta);
  q.set("page", String(f.page ?? 1));
  q.set("page_size", String(f.pageSize ?? 50));
  return api.request<HistorialResponse>(`/viajes?${q.toString()}`);
}
