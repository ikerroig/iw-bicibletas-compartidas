// Cliente de spec-002 (T038). Viaje = esquema Viaje de
// specs/spec-001/contracts/openapi.yaml.
import { apiFetch } from '../api';

export interface Viaje {
  id: number;
  usuario_id: number;
  bicicleta_id: number;
  estacion_origen_id: number;
  estacion_destino_id: number | null;
  fecha_inicio: string;
  fecha_fin: string | null;
  duracion_minutos: number | null;
  coste_total: number;
  estado: 'en_curso' | 'finalizado';
}

export function desbloquear(codigoAnclaje: string): Promise<Viaje> {
  return apiFetch<Viaje>('/alquileres/desbloqueo', {
    method: 'POST',
    body: JSON.stringify({ codigo_anclaje: codigoAnclaje }),
  });
}

export async function viajeActivo(signal?: AbortSignal): Promise<Viaje | null> {
  const r = await apiFetch<{ viaje: Viaje | null }>('/alquileres/activo', { signal });
  return r.viaje;
}
