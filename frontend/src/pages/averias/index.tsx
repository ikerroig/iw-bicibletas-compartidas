import { createResource, createSignal, For, Show } from 'solid-js';
import { ApiError } from '../../services/api';
import { reportarAveria, type TipoFallo } from '../../services/api/averias';
import { listarEstaciones } from '../../services/api/estaciones';

const tipos: { value: TipoFallo; label: string }[] = [
  { value: 'frenos', label: 'Frenos' },
  { value: 'cadena', label: 'Cadena' },
  { value: 'neumatico', label: 'Neumatico' },
  { value: 'otro', label: 'Otro' },
];

export function ReportarAveria() {
  const [estaciones] = createResource(() => listarEstaciones());
  const [bicicleta, setBicicleta] = createSignal('');
  const [estacion, setEstacion] = createSignal('');
  const [tipo, setTipo] = createSignal<TipoFallo>('frenos');
  const [descripcion, setDescripcion] = createSignal('');
  const [mensaje, setMensaje] = createSignal<string | null>(null);
  const [error, setError] = createSignal<string | null>(null);
  const [enviando, setEnviando] = createSignal(false);

  const enviar = async (evento: SubmitEvent) => {
    evento.preventDefault();
    setMensaje(null);
    setError(null);
    setEnviando(true);
    try {
      const respuesta = await reportarAveria({
        bicicleta_codigo: bicicleta().trim(),
        estacion_id: Number(estacion()),
        tipo_fallo: tipo(),
        descripcion: descripcion().trim(),
      });
      setMensaje(respuesta.mensaje);
      setDescripcion('');
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'No se pudo registrar la averia');
    } finally {
      setEnviando(false);
    }
  };

  return (
    <div class="mx-auto max-w-lg p-4">
      <form
        onSubmit={enviar}
        class="space-y-4 rounded-xl bg-white p-6 shadow"
        aria-labelledby="titulo-averia"
      >
        <h1 id="titulo-averia" class="text-xl font-semibold">Reportar averia</h1>
        <p class="text-sm text-slate-600">
          Indica la bicicleta y la estacion donde la has encontrado. La bicicleta quedara bloqueada
          para el alquiler.
        </p>

        <label class="block text-sm">
          Codigo de bicicleta
          <input
            required
            type="text"
            autocomplete="off"
            placeholder="BICI-0001"
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
            value={bicicleta()}
            onInput={(e) => setBicicleta(e.currentTarget.value)}
          />
        </label>

        <label class="block text-sm">
          Estacion
          <select
            required
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
            value={estacion()}
            onChange={(e) => setEstacion(e.currentTarget.value)}
          >
            <option value="">Selecciona una estacion</option>
            <For each={estaciones()?.estaciones ?? []}>
              {(item) => <option value={item.id}>{item.nombre}</option>}
            </For>
          </select>
        </label>

        <label class="block text-sm">
          Tipo de fallo
          <select
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
            value={tipo()}
            onChange={(e) => setTipo(e.currentTarget.value as TipoFallo)}
          >
            <For each={tipos}>{(item) => <option value={item.value}>{item.label}</option>}</For>
          </select>
        </label>

        <label class="block text-sm">
          Descripcion del problema
          <textarea
            required
            rows="4"
            maxlength="2000"
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
            value={descripcion()}
            onInput={(e) => setDescripcion(e.currentTarget.value)}
          />
        </label>

        <Show when={mensaje()}>
          <p role="status" class="rounded bg-emerald-50 p-3 text-sm text-emerald-800">{mensaje()}</p>
        </Show>
        <Show when={error()}>
          <p role="alert" class="rounded bg-red-50 p-3 text-sm text-red-800">{error()}</p>
        </Show>
        <Show when={estaciones.error}>
          <p role="alert" class="text-sm text-red-700">No se pudieron cargar las estaciones.</p>
        </Show>

        <button
          type="submit"
          disabled={enviando()}
          class="w-full rounded bg-emerald-700 py-2 font-medium text-white hover:bg-emerald-800 disabled:opacity-60"
        >
          {enviando() ? 'Enviando...' : 'Registrar averia'}
        </button>
      </form>
    </div>
  );
}
