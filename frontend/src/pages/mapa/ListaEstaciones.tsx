// T104 — Listado textual de estaciones (HU-01 FR-011, SC-008; research R21).
// Alternativa al mapa para quien no puede usarlo (lector de pantalla) o cuando
// el mapa de fondo no carga. Recibe las MISMAS estaciones del store en vivo y
// usa la MISMA función estadoVisual que los marcadores, así que estado,
// contadores y frescura coinciden siempre con el mapa. No depende de Leaflet.
import { createMemo, For, Show } from 'solid-js';
import type { EstacionEstado } from '../../services/api/estaciones';
import { descripcionAccesible, estadoVisual, PRESENTACION } from './estadoVisual';

interface Props {
  estaciones: EstacionEstado[];
  seleccionId: number | null;
  onSeleccionar: (id: number) => void;
}

export default function ListaEstaciones(props: Props) {
  const ordenadas = createMemo(() =>
    [...props.estaciones].sort((a, b) => a.nombre.localeCompare(b.nombre, 'es')),
  );

  return (
    <ul aria-label="Listado de estaciones" class="divide-y divide-slate-200">
      <For each={ordenadas()}>
        {(e) => {
          const v = () => estadoVisual(e);
          const fuera = () => v() === 'fuera_de_servicio';
          return (
            <li>
              <button
                type="button"
                aria-label={descripcionAccesible(e)}
                aria-pressed={props.seleccionId === e.id}
                onClick={() => props.onSeleccionar(e.id)}
                class={`flex w-full items-center gap-3 px-4 py-2 text-left hover:bg-slate-50 focus:bg-slate-100 focus:outline-none ${
                  props.seleccionId === e.id ? 'bg-sky-50' : ''
                }`}
              >
                <span
                  aria-hidden="true"
                  class={`inline-flex min-w-7 justify-center rounded border px-1 ${PRESENTACION[v()].clase}`}
                >
                  {PRESENTACION[v()].icono}
                </span>
                <span class="flex-1">
                  <span data-campo="nombre" class="block font-medium">
                    {e.nombre}
                  </span>
                  <span class="block text-xs text-slate-600">{PRESENTACION[v()].etiqueta}</span>
                </span>
                <Show when={!fuera()} fallback={<span class="text-sm text-slate-500">—</span>}>
                  <span class="text-right text-sm tabular-nums">
                    <span class="block">{e.bicis_disponibles} bicis</span>
                    <span class="block text-xs text-slate-600">{e.anclajes_libres} libres</span>
                  </span>
                </Show>
              </button>
            </li>
          );
        }}
      </For>
    </ul>
  );
}
