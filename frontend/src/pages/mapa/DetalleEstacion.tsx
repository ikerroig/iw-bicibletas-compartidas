// Detalle de una estación (T031; FR-004, FR-006, FR-008, FR-009).
// Muestra SOLO dos contadores (sin desglose de averiadas). Fuera de servicio:
// «Fuera de servicio» y «—», sin acciones de recogida ni devolución.
import { Show } from 'solid-js';
import type { EstacionEstado } from '../../services/api/estaciones';
import { estadoVisual, PRESENTACION } from './estadoVisual';

interface Props {
  estacion: EstacionEstado;
  onCerrar?: () => void;
}

export default function DetalleEstacion(props: Props) {
  const visual = () => estadoVisual(props.estacion);
  const fuera = () => visual() === 'fuera_de_servicio';
  const tituloId = () => `detalle-estacion-${props.estacion.id}`;

  const contador = (valor: number, textoCero: string) =>
    fuera() ? '—' : valor === 0 ? `0 · ${textoCero}` : String(valor);

  return (
    <section aria-labelledby={tituloId()} class="space-y-3 p-4">
      <header class="flex items-start gap-2">
        <h2 id={tituloId()} class="flex-1 text-lg font-semibold">
          {props.estacion.nombre}
        </h2>
        <Show when={props.onCerrar}>
          <button
            type="button"
            class="rounded px-2 text-slate-500 hover:bg-slate-100"
            aria-label="Cerrar detalle"
            onClick={() => props.onCerrar?.()}
          >
            ✕
          </button>
        </Show>
      </header>

      <p
        class={`inline-flex items-center gap-1 rounded border px-2 py-0.5 text-sm ${PRESENTACION[visual()].clase}`}
      >
        <span aria-hidden="true">{PRESENTACION[visual()].icono}</span>
        <span>{fuera() ? 'Fuera de servicio' : 'Operativa'}</span>
      </p>

      <dl class="grid grid-cols-2 gap-3">
        <div class="rounded bg-slate-100 p-3">
          <dt class="text-xs text-slate-600">Bicicletas disponibles</dt>
          <dd data-testid="contador-bicis" class="text-xl font-semibold">
            {contador(props.estacion.bicis_disponibles, 'Sin bicicletas')}
          </dd>
        </div>
        <div class="rounded bg-slate-100 p-3">
          <dt class="text-xs text-slate-600">Anclajes libres</dt>
          <dd data-testid="contador-anclajes" class="text-xl font-semibold">
            {contador(props.estacion.anclajes_libres, 'Sin anclajes libres')}
          </dd>
        </div>
      </dl>

      <Show when={fuera()}>
        <p class="text-sm text-slate-600">
          Esta estación no está disponible para recoger ni devolver bicicletas.
        </p>
      </Show>
    </section>
  );
}
