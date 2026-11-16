import { createResource, For, onCleanup, onMount, Show } from 'solid-js';
import { listarIncidencias } from '../../services/api/averias';

const fechaLocal = (fecha: string) =>
  new Date(fecha).toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' });

export function PanelIncidencias() {
  const [incidencias, { refetch }] = createResource(() => listarIncidencias());

  // Actualiza el panel sin obligar al operario a recargar manualmente.
  // El intervalo queda por debajo del límite de 30 segundos de la HU-05.
  onMount(() => {
    const intervalo = setInterval(() => void refetch(), 10_000);
    onCleanup(() => clearInterval(intervalo));
  });

  return (
    <div class="mx-auto max-w-5xl space-y-4 p-4">
      <header class="flex items-center justify-between">
        <div>
          <h1 class="text-xl font-semibold">Incidencias pendientes</h1>
          <p class="text-sm text-slate-600">Reportes de averias que debe revisar el taller.</p>
        </div>
        <button type="button" onClick={() => refetch()} class="rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50">
          Actualizar
        </button>
      </header>
      <Show when={incidencias.error}>
        <p role="alert" class="rounded bg-red-50 p-3 text-sm text-red-800">No se pudieron cargar las incidencias.</p>
      </Show>
      <Show when={!incidencias.loading && incidencias()?.length === 0}>
        <p class="rounded bg-white p-6 text-sm text-slate-600 shadow">No hay incidencias pendientes.</p>
      </Show>
      <Show when={incidencias()?.length}>
        <div class="overflow-x-auto rounded-xl bg-white shadow">
          <table class="min-w-full text-left text-sm">
            <thead class="border-b bg-slate-50 text-slate-600"><tr>
              <th class="px-4 py-3">Bicicleta</th><th class="px-4 py-3">Estacion</th>
              <th class="px-4 py-3">Tipo</th><th class="px-4 py-3">Descripcion</th>
              <th class="px-4 py-3">Fecha</th>
            </tr></thead>
            <tbody>
              <For each={incidencias()}>{(item) => <tr class="border-b last:border-0">
                <td class="px-4 py-3 font-medium">{item.bicicleta_codigo}</td>
                <td class="px-4 py-3">{item.estacion_nombre}</td>
                <td class="px-4 py-3">{item.tipo_fallo}</td>
                <td class="max-w-md px-4 py-3">{item.descripcion_reporte}</td>
                <td class="whitespace-nowrap px-4 py-3">{fechaLocal(item.fecha_reporte)}</td>
              </tr>}</For>
            </tbody>
          </table>
        </div>
      </Show>
    </div>
  );
}
