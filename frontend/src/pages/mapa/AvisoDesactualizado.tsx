// T097 — Hora de la última actualización correcta y aviso de «datos
// desactualizados» (FR-007, SC-005). La región role="status" existe siempre
// para que los lectores de pantalla anuncien la aparición del aviso.
import { Show } from 'solid-js';

interface Props {
  ultimaActualizacionOk: number | null;
  desactualizado: boolean;
}

export function formatearHora(ms: number): string {
  return new Date(ms).toLocaleTimeString('es-ES', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export default function AvisoDesactualizado(props: Props) {
  return (
    <div role="status" aria-live="polite" class="flex flex-col items-center gap-1 text-sm">
      <p class="rounded bg-white/90 px-2 py-0.5 text-xs text-slate-700 shadow">
        {props.ultimaActualizacionOk === null
          ? 'Sin datos todavía'
          : `Actualizado a las ${formatearHora(props.ultimaActualizacionOk)}`}
      </p>
      {/* Solo con datos previos (converge F2, T112): sin ninguna carga correcta no
          hay «últimos datos conocidos» que mostrar; ese caso lo cubre FR-010. */}
      <Show when={props.desactualizado && props.ultimaActualizacionOk !== null}>
        <p class="flex items-center gap-2 rounded border-2 border-amber-600 bg-amber-50 px-3 py-1.5 text-amber-900 shadow">
          <span aria-hidden="true" class="text-lg">
            ⚠
          </span>
          <strong>Datos desactualizados</strong>
          <span class="hidden sm:inline">
            · Mostramos los últimos datos conocidos; reintentando…
          </span>
        </p>
      </Show>
    </div>
  );
}
