// Desbloqueo por código de anclaje (T039; HU-02). Las reglas las decide el
// backend (const. I): esta página solo envía el código y muestra la causa que
// devuelve. Con un viaje en curso enseña ese viaje en lugar del formulario.
import { createResource, createSignal, Match, Show, Switch } from 'solid-js';
import { ApiError } from '../../services/api';
import { desbloquear, viajeActivo, type ViajeActivo } from '../../services/api/alquiler';

const horaLocal = (iso: string) =>
  new Date(iso).toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' });

function ViajeEnCurso(props: { viaje: ViajeActivo }) {
  return (
    <section aria-labelledby="titulo-viaje" class="space-y-2 rounded-xl bg-white p-6 shadow">
      <h1 id="titulo-viaje" class="text-xl font-semibold">
        Viaje en curso
      </h1>
      <dl class="space-y-1 text-sm">
        <div class="flex gap-2">
          <dt class="text-slate-600">Inicio:</dt>
          <dd data-testid="viaje-inicio">{horaLocal(props.viaje.fecha_inicio)}</dd>
        </div>
        <div class="flex gap-2">
          <dt class="text-slate-600">Bicicleta:</dt>
          <dd>{props.viaje.bicicleta_codigo}</dd>
        </div>
        <div class="flex gap-2">
          <dt class="text-slate-600">Estación de origen:</dt>
          <dd>{props.viaje.estacion_origen_nombre}</dd>
        </div>
      </dl>
      <p class="text-sm text-slate-600">
        No puedes desbloquear otra bicicleta hasta devolver esta.
      </p>
    </section>
  );
}

export default function Desbloqueo() {
  const [activo, { mutate }] = createResource(() => viajeActivo());
  const [codigo, setCodigo] = createSignal('');
  const [error, setError] = createSignal<string | null>(null);
  const [enviando, setEnviando] = createSignal(false);

  const enviar = async (ev: SubmitEvent) => {
    ev.preventDefault();
    setError(null);
    setEnviando(true);
    try {
      mutate(await desbloquear(codigo().trim()));
    } catch (e) {
      if (e instanceof ApiError && e.codigo === 'VIAJE_ACTIVO') {
        // El viaje se inició en otra pestaña o dispositivo: se muestra ese.
        try {
          mutate(await viajeActivo());
        } catch {
          /* se queda el mensaje de abajo */
        }
      }
      setError(e instanceof ApiError ? e.message : 'No se pudo desbloquear la bicicleta');
    } finally {
      setEnviando(false);
    }
  };

  return (
    <div class="mx-auto max-w-md p-4">
      <Switch>
        <Match when={activo.error}>
          <p role="alert" class="text-sm text-red-700">
            No se pudo consultar tu viaje actual. Recarga la página.
          </p>
        </Match>
        <Match when={activo.loading}>
          <p class="text-sm text-slate-600">Cargando…</p>
        </Match>
        <Match when={activo()}>{(viaje) => <ViajeEnCurso viaje={viaje()} />}</Match>
        <Match when={true}>
          <form
            onSubmit={enviar}
            class="space-y-4 rounded-xl bg-white p-6 shadow"
            aria-labelledby="titulo-desbloqueo"
          >
            <h1 id="titulo-desbloqueo" class="text-xl font-semibold">
              Desbloquear bicicleta
            </h1>
            <label class="block text-sm">
              Código de anclaje
              <input
                type="text"
                required
                autocomplete="off"
                placeholder="E01-A01"
                class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
                value={codigo()}
                onInput={(e) => setCodigo(e.currentTarget.value)}
              />
            </label>
            <Show when={error()}>
              <p role="alert" class="text-sm text-red-700">
                {error()}
              </p>
            </Show>
            <button
              type="submit"
              disabled={enviando()}
              class="w-full rounded bg-emerald-700 py-2 font-medium text-white hover:bg-emerald-800 disabled:opacity-60"
            >
              {enviando() ? 'Desbloqueando…' : 'Desbloquear'}
            </button>
          </form>
        </Match>
      </Switch>
    </div>
  );
}
