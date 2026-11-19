// Desbloqueo por código de anclaje (T039; HU-02). Las reglas las decide el
// backend (const. I): esta página solo envía el código y muestra la causa que
// devuelve. Con un viaje en curso enseña ese viaje en lugar del formulario.
// Refinamiento 2026-11-19: guardia por rol (FR-009, D9), aviso del abono al
// abrir (FR-010, D16) y resultado desconocido (FR-011, D17).
import { Navigate } from '@solidjs/router';
import { createResource, createSignal, Match, Show, Switch } from 'solid-js';
import { ApiError } from '../../services/api';
import {
  desbloquear,
  PLAZO_DESBLOQUEO_MS,
  PlazoAgotado,
  type ViajeActivo,
} from '../../services/api/alquiler';
import { sesion } from '../../stores/sesion';
import { cargarViajeEnCurso, estadoAbono, fijarViajeEnCurso } from '../../stores/viaje';

const COMPROBANDO = 'No sabemos si se ha desbloqueado; comprobando…';
const NO_DESBLOQUEADO = 'No se ha desbloqueado, inténtalo de nuevo';

const horaLocal = (iso: string) =>
  new Date(iso).toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' });

/**
 * Sin respuesta, plazo agotado o 5xx: el viaje pudo crearse o no (D17). Cualquier
 * otro error que no sea un ApiError (p. ej., un 201 cuyo cuerpo no se puede leer)
 * también es desconocido: solo provoca una consulta de lectura (T063).
 */
function resultadoDesconocido(e: unknown): boolean {
  if (e instanceof PlazoAgotado) return true;
  if (e instanceof ApiError) return e.status === 0 || e.status >= 500;
  return true;
}

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

function PantallaDesbloqueo() {
  const [activo, { mutate }] = createResource(() => cargarViajeEnCurso());
  const [codigo, setCodigo] = createSignal('');
  const [error, setError] = createSignal<string | null>(null);
  const [aviso, setAviso] = createSignal<string | null>(null);
  const [enviando, setEnviando] = createSignal(false);
  // FR-011: 'comprobando' mientras se consulta el viaje; 'sin_comprobar' si la
  // consulta también falla. En ambos casos no se puede volver a enviar.
  const [fase, setFase] = createSignal<'formulario' | 'comprobando' | 'sin_comprobar'>('formulario');

  const sinAbono = () => {
    const a = estadoAbono();
    return a !== null && !a.vigente ? (a.mensaje ?? 'Tu abono anual no está vigente') : null;
  };
  const bloqueado = () => enviando() || fase() !== 'formulario' || sinAbono() !== null;

  /** Consulta el viaje activo una sola vez; nunca reenvía el desbloqueo (D17). */
  const comprobar = async () => {
    setFase('comprobando');
    try {
      // Con plazo: una consulta que no responde cuenta como fallo (T061).
      const viaje = await cargarViajeEnCurso(PLAZO_DESBLOQUEO_MS);
      if (viaje) {
        mutate(viaje);
      } else {
        setAviso(NO_DESBLOQUEADO);
      }
      setFase('formulario');
    } catch {
      setFase('sin_comprobar');
    }
  };

  const enviar = async (ev: SubmitEvent) => {
    ev.preventDefault();
    if (bloqueado()) return;
    setError(null);
    setAviso(null);
    setEnviando(true);
    try {
      const viaje = await desbloquear(codigo().trim());
      fijarViajeEnCurso(viaje);
      mutate(viaje);
    } catch (e) {
      if (resultadoDesconocido(e)) {
        setEnviando(false);
        await comprobar();
        return;
      }
      if (e instanceof ApiError && e.codigo === 'VIAJE_ACTIVO') {
        // El viaje se inició en otra pestaña o dispositivo: se muestra ese.
        try {
          mutate(await cargarViajeEnCurso());
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
            <Show when={sinAbono()}>
              {(mensaje) => (
                <p role="alert" class="rounded bg-amber-50 p-3 text-sm text-amber-900">
                  {mensaje()}
                </p>
              )}
            </Show>
            <label class="block text-sm">
              Código de anclaje
              <input
                type="text"
                required
                autocomplete="off"
                placeholder="E01-A01"
                class="mt-1 w-full rounded border border-slate-300 px-3 py-2 disabled:bg-slate-100"
                value={codigo()}
                disabled={sinAbono() !== null}
                onInput={(e) => setCodigo(e.currentTarget.value)}
              />
            </label>
            <Show when={error()}>
              <p role="alert" class="text-sm text-red-700">
                {error()}
              </p>
            </Show>
            <Show when={fase() === 'comprobando'}>
              <p role="status" class="text-sm text-slate-700">
                {COMPROBANDO}
              </p>
            </Show>
            <Show when={fase() === 'sin_comprobar'}>
              <div role="alert" class="space-y-2 text-sm text-amber-900">
                <p>No sabemos si se ha desbloqueado: no hemos podido comprobarlo. Revisa tu conexión.</p>
                <button
                  type="button"
                  class="rounded border border-amber-700 px-3 py-1 font-medium"
                  onClick={() => void comprobar()}
                >
                  Volver a comprobar
                </button>
              </div>
            </Show>
            <Show when={aviso()}>
              <p role="status" class="text-sm text-slate-700">
                {aviso()}
              </p>
            </Show>
            <button
              type="submit"
              disabled={bloqueado()}
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

export default function Desbloqueo() {
  // FR-009 (D9): solo el rol `usuario`; el resto vuelve al mapa. Sin usuario
  // guardado tampoco se conoce el rol (mismo criterio que el menú, T029).
  return (
    <Show when={sesion.usuario()?.rol === 'usuario'} fallback={<Navigate href="/" />}>
      <PantallaDesbloqueo />
    </Show>
  );
}
