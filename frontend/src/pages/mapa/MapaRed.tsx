// Página principal de HU-01 (T031): mapa en vivo de la red, aviso de datos
// desactualizados y detalle de la estación seleccionada. El detalle se pinta
// con los datos del snapshot en vivo, así que también se actualiza solo.
import { createMemo, createSignal, For, Show } from 'solid-js';
import MapaBase, { type MarcadorMapa } from '../../components/MapaBase';
import type { EstacionEstado } from '../../services/api/estaciones';
import { crearRedEnVivo } from '../../stores/red';
import AvisoDesactualizado from './AvisoDesactualizado';
import DetalleEstacion from './DetalleEstacion';
import ListaEstaciones from './ListaEstaciones';
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
  // Vista «Mapa» o «Lista» (FR-011, R21). El mapa sigue montado aunque se vea la
  // lista, para no reconstruir Leaflet ni perder el encuadre al volver.
  const [vista, setVista] = createSignal<Vista>('mapa');

  return (
    <div class="flex h-full flex-col md:flex-row">
      <div class="flex min-h-[55vh] flex-1 flex-col">
        <ConmutadorVista vista={vista()} onCambiar={setVista} />

        <div class="relative flex-1">
          <div
            id="panel-mapa"
            role="tabpanel"
            aria-labelledby="pestana-mapa"
            class={`absolute inset-0 ${vista() === 'mapa' ? '' : 'invisible'}`}
          >
            <MapaBase
              marcadores={marcadores()}
              onSeleccionar={setSeleccionId}
              class="absolute inset-0"
            />
          </div>
          <Show when={vista() === 'lista'}>
            <div
              id="panel-lista"
              role="tabpanel"
              aria-labelledby="pestana-lista"
              class="absolute inset-0 overflow-y-auto bg-white pt-20"
            >
              <ListaEstaciones
                estaciones={red.estaciones()}
                seleccionId={seleccionId()}
                onSeleccionar={setSeleccionId}
              />
            </div>
          </Show>

          <div class="pointer-events-none absolute inset-x-0 top-2 z-[1000] flex flex-col items-center gap-2 px-2">
            <AvisoDesactualizado
              ultimaActualizacionOk={red.ultimaActualizacionOk()}
              desactualizado={red.desactualizado()}
            />
            {/* FR-010 (R20): primera carga fallida, sin ningún dato previo. */}
            <Show when={red.error() && !red.cargado()}>
              <div
                role="alert"
                class="pointer-events-auto flex flex-col items-center gap-2 rounded border-2 border-red-300 bg-red-50 px-4 py-3 text-center text-sm text-red-900 shadow"
              >
                <p class="font-semibold">No se pudo cargar la disponibilidad</p>
                <p class="text-xs">Lo volveremos a intentar automáticamente.</p>
                <button
                  type="button"
                  onClick={() => void red.refrescar()}
                  class="rounded bg-red-700 px-3 py-1 font-medium text-white hover:bg-red-800 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1"
                >
                  Reintentar
                </button>
              </div>
            </Show>
            {/* Con datos previos se mantienen los últimos contadores (R17). */}
            <Show when={red.error() && red.cargado()}>
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

type Vista = 'mapa' | 'lista';

const VISTAS: { id: Vista; texto: string }[] = [
  { id: 'mapa', texto: 'Mapa' },
  { id: 'lista', texto: 'Lista' },
];

// Pestañas accesibles (patrón WAI-ARIA «tabs»): solo la activa es tabulable y
// las flechas, Inicio y Fin cambian de vista.
function ConmutadorVista(props: { vista: Vista; onCambiar: (v: Vista) => void }) {
  const botones: HTMLButtonElement[] = [];

  const mover = (ev: KeyboardEvent, indice: number) => {
    const ultimo = VISTAS.length - 1;
    const destino =
      ev.key === 'ArrowRight'
        ? (indice + 1) % VISTAS.length
        : ev.key === 'ArrowLeft'
          ? (indice + ultimo) % VISTAS.length
          : ev.key === 'Home'
            ? 0
            : ev.key === 'End'
              ? ultimo
              : null;
    if (destino === null) return;
    ev.preventDefault();
    props.onCambiar(VISTAS[destino].id);
    botones[destino]?.focus();
  };

  return (
    <div
      role="tablist"
      aria-label="Vista de la disponibilidad"
      class="flex gap-1 border-b border-slate-200 bg-white px-2 pt-2"
    >
      <For each={VISTAS}>
        {(v, i) => (
          <button
            ref={(el) => (botones[i()] = el)}
            type="button"
            role="tab"
            id={`pestana-${v.id}`}
            aria-controls={`panel-${v.id}`}
            aria-selected={props.vista === v.id}
            tabIndex={props.vista === v.id ? 0 : -1}
            onClick={() => props.onCambiar(v.id)}
            onKeyDown={(ev) => mover(ev, i())}
            class={`rounded-t px-4 py-1.5 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-sky-500 ${
              props.vista === v.id
                ? 'border border-b-white border-slate-200 bg-white text-slate-900 -mb-px'
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            {v.texto}
          </button>
        )}
      </For>
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
