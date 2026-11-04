// Viaje en curso del usuario (spec-002, D14), compartido entre la página de
// desbloqueo y el aviso del Layout. La devolución (spec-003) debe llamar a
// fijarViajeEnCurso(null) al cerrar el viaje.
import { createSignal } from 'solid-js';
import { viajeActivo, type ViajeActivo } from '../services/api/alquiler';

const [viaje, setViaje] = createSignal<ViajeActivo | null>(null);

export const viajeEnCurso = viaje;

export function fijarViajeEnCurso(v: ViajeActivo | null): void {
  setViaje(v);
}

export async function cargarViajeEnCurso(): Promise<ViajeActivo | null> {
  const v = await viajeActivo();
  setViaje(v);
  return v;
}
