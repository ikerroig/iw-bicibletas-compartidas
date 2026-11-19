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

/** Estado del abono que calcula el backend (FR-010, research D16). */
export interface EstadoAbono {
  vigente: boolean;
  codigo: 'SIN_ABONO' | 'ABONO_VENCIDO' | null;
  mensaje: string | null;
}

/** Plazo del desbloqueo: pasado este tiempo el resultado es desconocido (FR-011, D17). */
export const PLAZO_DESBLOQUEO_MS = 10_000;

/** El plazo se agotó sin respuesta: no se sabe si el viaje se creó. */
export class PlazoAgotado extends Error {
  constructor() {
    super('El servidor no respondió a tiempo');
    this.name = 'PlazoAgotado';
  }
}

export async function desbloquear(codigoAnclaje: string): Promise<ViajeActivo> {
  const control = new AbortController();
  const temporizador = setTimeout(() => control.abort(), PLAZO_DESBLOQUEO_MS);
  try {
    return await apiFetch<ViajeActivo>('/alquileres/desbloqueo', {
      method: 'POST',
      body: JSON.stringify({ codigo_anclaje: codigoAnclaje }),
      signal: control.signal,
    });
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw new PlazoAgotado();
    throw e;
  } finally {
    clearTimeout(temporizador);
  }
}

/**
 * Viaje en curso y estado del abono (FR-008, FR-010). Con `plazoMs`, la consulta
 * se cancela si no responde a tiempo y rechaza con PlazoAgotado (FR-011, T061).
 */
export async function estadoAlquiler(
  plazoMs?: number,
): Promise<{ viaje: ViajeActivo | null; abono: EstadoAbono }> {
  if (plazoMs === undefined) {
    return apiFetch<{ viaje: ViajeActivo | null; abono: EstadoAbono }>('/alquileres/activo');
  }
  const control = new AbortController();
  const temporizador = setTimeout(() => control.abort(), plazoMs);
  try {
    return await apiFetch<{ viaje: ViajeActivo | null; abono: EstadoAbono }>('/alquileres/activo', {
      signal: control.signal,
    });
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw new PlazoAgotado();
    throw e;
  } finally {
    clearTimeout(temporizador);
  }
}
