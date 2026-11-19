// Viaje en curso del usuario (spec-002, D14), compartido entre la página de
// desbloqueo y el aviso del Layout. La devolución (spec-003) debe llamar a
// fijarViajeEnCurso(null) al cerrar el viaje. El estado del abono (D16) va en
// una señal aparte para no cambiar la firma de fijarViajeEnCurso/viajeEnCurso.
import { createSignal } from 'solid-js';
import { estadoAlquiler, type EstadoAbono, type ViajeActivo } from '../services/api/alquiler';

const [viaje, setViaje] = createSignal<ViajeActivo | null>(null);
const [abono, setAbono] = createSignal<EstadoAbono | null>(null);

export const viajeEnCurso = viaje;

/** Último estado del abono recibido del backend, o null si aún no se ha consultado. */
export const estadoAbono = abono;

export function fijarViajeEnCurso(v: ViajeActivo | null): void {
  setViaje(v);
}

/** Consulta el viaje y el abono; `plazoMs` acota la espera (FR-011, T061). */
export async function cargarViajeEnCurso(plazoMs?: number): Promise<ViajeActivo | null> {
  const r = await estadoAlquiler(plazoMs);
  setViaje(r.viaje);
  setAbono(r.abono);
  return r.viaje;
}
