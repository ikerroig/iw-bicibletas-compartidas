// Aviso de viaje en curso (spec-002, FR-008; research D14): visible en toda la
// aplicación salvo en /alquiler, que ya muestra el viaje.
import { createEffect, Show } from 'solid-js';
import { A, useLocation } from '@solidjs/router';
import { sesion } from '../stores/sesion';
import { cargarViajeEnCurso, fijarViajeEnCurso, viajeEnCurso } from '../stores/viaje';

const hora = (iso: string) =>
  new Date(iso).toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' });

export default function AvisoViaje() {
  const location = useLocation();

  createEffect(() => {
    if (sesion.usuario()?.rol !== 'usuario') {
      fijarViajeEnCurso(null);
      return;
    }
    // Sin viaje conocido no hay nada que avisar: un fallo aquí no debe molestar.
    cargarViajeEnCurso().catch(() => fijarViajeEnCurso(null));
  });

  return (
    <Show when={location.pathname !== '/alquiler' ? viajeEnCurso() : null}>
      {(viaje) => (
        <p role="status" class="flex items-center gap-2 bg-amber-100 px-4 py-1 text-sm text-amber-900">
          <span>
            Viaje en curso desde las {hora(viaje().fecha_inicio)} · {viaje().bicicleta_codigo}
          </span>
          <A href="/alquiler" class="ml-auto font-medium underline">
            Ver
          </A>
        </p>
      )}
    </Show>
  );
}
