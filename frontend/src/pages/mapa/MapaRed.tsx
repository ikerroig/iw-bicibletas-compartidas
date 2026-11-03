// Página principal de HU-01 (T031): mapa en vivo de la red, aviso de datos
// desactualizados y detalle de la estación seleccionada. El detalle se pinta
// con los datos del snapshot en vivo, así que también se actualiza solo.
import { createMemo, createSignal, For, Show } from 'solid-js';
import MapaBase, { type MarcadorMapa } from '../../components/MapaBase';
import type { EstacionEstado } from '../../services/api/estaciones';
import { crearRedEnVivo } from '../../stores/red';
import AvisoDesactualizado from './AvisoDesactualizado';
import DetalleEstacion from './DetalleEstacion';
import {
  descripcionAccesible,
  estadoVisual,
  PRESENTACION,
  type EstadoVisual,
} from './estadoVisual';

interface Props {
  /** Solo para tests; en producción, 15 s (R16). */
  intervaloMs?: number;
}

export default function MapaRed(props: Props) {
  // El intervalo se fija al montar la página; cambiarlo después no tiene efecto.
  // eslint-disable-next-line solid/reactivity
  const red = crearRedEnVivo({ intervaloMs: props.intervaloMs });
  const [seleccionId, setSeleccionId] = createSignal<number | null>(null);

  const seleccionada = createMemo(
    () => red.estaciones().find((e) => e.id === seleccionId()) ?? null,
  );
  const marcadores = createMemo(() =>
    red.estaciones().map((e) => marcadorDe(e, e.id === seleccionId())),
  );

  return (
    <div class="flex h-full flex-col md:flex-row">
      <div class="relative min-h-[55vh] flex-1">
        <MapaBase
          marcadores={marcadores()}
          onSeleccionar={setSeleccionId}
          class="absolute inset-0"
        />

        <div class="pointer-events-none absolute inset-x-0 top-2 z-[1000] flex flex-col items-center gap-2 px-2">
          <AvisoDesactualizado
            ultimaActualizacionOk={red.ultimaActualizacionOk()}
            desactualizado={red.desactualizado()}
          />
          <Show when={red.error()}>
            <p role="alert" class="rounded bg-red-50 px-3 py-1 text-sm text-red-800 shadow">
              {red.error()}. Reintentando…
            </p>
          </Show>
          <Show when={!red.cargado() && !red.error()}>
            <p class="rounded bg-white/90 px-3 py-1 text-sm shadow">Cargando estaciones…</p>
          </Show>
          <Show when={red.cargado() && red.estaciones().length === 0}>
            <p class="rounded bg-white/90 px-3 py-1 text-sm shadow">
              No hay estaciones para mostrar.
            </p>
          </Show>
        </div>
      </div>

      <aside class="max-h-[45vh] overflow-y-auto border-t border-slate-200 bg-white md:max-h-none md:w-80 md:border-l md:border-t-0">
        <Show
          when={seleccionada()}
          fallback={
            <p class="p-4 text-sm text-slate-600">
              Selecciona una estación para ver su disponibilidad.
            </p>
          }
        >
          {(e) => <DetalleEstacion estacion={e()} onCerrar={() => setSeleccionId(null)} />}
        </Show>
        <Leyenda />
      </aside>
    </div>
  );
}

function Leyenda() {
  const orden: EstadoVisual[] = [
    'normal',
    'sin_bicis',
    'sin_anclajes',
    'sin_bicis_y_sin_anclajes',
    'fuera_de_servicio',
  ];
  return (
    <section aria-label="Leyenda" class="border-t border-slate-200 p-4">
      <h3 class="mb-2 text-xs font-semibold uppercase text-slate-500">Leyenda</h3>
      <ul class="space-y-1 text-sm">
        <For each={orden}>
          {(v) => (
            <li class="flex items-center gap-2">
              <span
                class={`inline-flex min-w-7 justify-center rounded border px-1 ${PRESENTACION[v].clase}`}
                aria-hidden="true"
              >
                {PRESENTACION[v].icono}
              </span>
              {PRESENTACION[v].etiqueta}
            </li>
          )}
        </For>
      </ul>
    </section>
  );
}

function escapar(s: string): string {
  return s.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);
}

function marcadorDe(e: EstacionEstado, seleccionado: boolean): MarcadorMapa {
  const v = estadoVisual(e);
  const p = PRESENTACION[v];
  const texto = p.corta ?? String(e.bicis_disponibles);
  const anillo = seleccionado ? ' ring-2 ring-sky-500 ring-offset-1' : '';
  return {
    id: e.id,
    lat: e.latitud,
    lng: e.longitud,
    ariaLabel: descripcionAccesible(e),
    html:
      `<div class="absolute -translate-x-1/2 -translate-y-full whitespace-nowrap rounded-full border-2 px-2 py-0.5 text-xs font-semibold shadow ${p.clase}${anillo}">` +
      `<span aria-hidden="true">${escapar(p.icono)}</span> <span>${escapar(texto)}</span></div>`,
  };
}
