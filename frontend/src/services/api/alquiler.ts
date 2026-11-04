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

/** Viaje con los datos que el usuario ve durante el trayecto (FR-008). */
export interface ViajeActivo extends Viaje {
  bicicleta_codigo: string;
  estacion_origen_nombre: string;
}

export function desbloquear(codigoAnclaje: string): Promise<ViajeActivo> {
  return apiFetch<ViajeActivo>('/alquileres/desbloqueo', {
    method: 'POST',
    body: JSON.stringify({ codigo_anclaje: codigoAnclaje }),
  });
}

export async function viajeActivo(): Promise<ViajeActivo | null> {
  const r = await apiFetch<{ viaje: ViajeActivo | null }>('/alquileres/activo');
  return r.viaje;
}
