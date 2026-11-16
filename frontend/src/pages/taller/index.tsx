import { createSignal, For, Show, type Component } from 'solid-js';
import { ApiError } from '../../services/api';
import {
  confirmarUbicacion,
  obtenerFicha,
  registrarIntervencion,
  type FichaBicicleta,
} from '../../services/api/taller';

const fechaLocal = (fecha: string) =>
  new Date(fecha).toLocaleString('es-ES', { dateStyle: 'short', timeStyle: 'short' });

export const Taller: Component = () => {
  const [idTexto, setIdTexto] = createSignal('');
  const [ficha, setFicha] = createSignal<FichaBicicleta | null>(null);
  const [piezas, setPiezas] = createSignal('');
  const [notas, setNotas] = createSignal('');
  const [marcarOperativa, setMarcarOperativa] = createSignal(true);
  const [codigoAnclaje, setCodigoAnclaje] = createSignal('');
  const [cargando, setCargando] = createSignal(false);
  const [guardando, setGuardando] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [mensaje, setMensaje] = createSignal<string | null>(null);

  const mostrarError = (err: unknown, generico: string) => {
    setError(err instanceof ApiError ? err.message : generico);
  };

  const cargarFicha = async (evento: Event) => {
    evento.preventDefault();
    const id = Number(idTexto());
    if (!Number.isInteger(id) || id <= 0) {
      setError('Introduce un ID de bicicleta valido');
      return;
    }
    setCargando(true);
    setError(null);
    setMensaje(null);
    try {
      setFicha(await obtenerFicha(id));
    } catch (err) {
      setFicha(null);
      mostrarError(err, 'No se pudo cargar la ficha');
    } finally {
      setCargando(false);
    }
  };

  const guardarParte = async (evento: Event) => {
    evento.preventDefault();
    const actual = ficha();
    if (!actual) return;
    if (!piezas().trim() && !notas().trim()) {
      setError('Indica las piezas sustituidas o las notas del taller');
      return;
    }
    setGuardando(true);
    setError(null);
    setMensaje(null);
    try {
      await registrarIntervencion(actual.bicicleta.id, {
        piezas_sustituidas: piezas().trim() || null,
        notas_taller: notas().trim() || null,
        marcar_operativa: marcarOperativa(),
      });
      setMensaje('Parte de reparación guardado correctamente');
      setPiezas('');
      setNotas('');
      setFicha(await obtenerFicha(actual.bicicleta.id));
    } catch (err) {
      mostrarError(err, 'No se pudo guardar el parte');
    } finally {
      setGuardando(false);
    }
  };

  const ubicarBicicleta = async (evento: Event) => {
    evento.preventDefault();
    const actual = ficha();
    if (!actual || !codigoAnclaje().trim()) {
      setError('Indica el código del anclaje');
      return;
    }
    setGuardando(true);
    setError(null);
    setMensaje(null);
    try {
      const estado = await confirmarUbicacion(actual.bicicleta.id, codigoAnclaje().trim());
      setMensaje('Bicicleta ubicada en ' + estado.nombre);
      setCodigoAnclaje('');
      setFicha(await obtenerFicha(actual.bicicleta.id));
    } catch (err) {
      mostrarError(err, 'No se pudo confirmar la ubicación');
    } finally {
      setGuardando(false);
    }
  };

  return (
    <section class="mx-auto max-w-4xl space-y-4 p-4">
      <header>
        <h1 class="text-xl font-semibold">Taller</h1>
        <p class="text-sm text-slate-600">
          Consulta una bicicleta, registra su reparación y confirma su ubicación.
        </p>
      </header>

      <form onSubmit={cargarFicha} class="flex flex-wrap items-end gap-3 rounded-xl bg-white p-4 shadow">
        <label class="flex-1 text-sm">
          ID de bicicleta
          <input
            required
            type="number"
            min="1"
            value={idTexto()}
            onInput={(e) => setIdTexto(e.currentTarget.value)}
            data-testid="input-bicicleta-id"
            class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
          />
        </label>
        <button
          type="submit"
          disabled={cargando()}
          data-testid="boton-buscar-ficha"
          class="rounded bg-slate-800 px-4 py-2 font-medium text-white disabled:opacity-60"
        >
          {cargando() ? 'Buscando...' : 'Consultar ficha'}
        </button>
      </form>

      <Show when={error()}>
        <p role="alert" data-testid="error-taller" class="rounded bg-red-50 p-3 text-sm text-red-800">
          {error()}
        </p>
      </Show>
      <Show when={mensaje()}>
        <p role="status" data-testid="mensaje-taller" class="rounded bg-emerald-50 p-3 text-sm text-emerald-800">
          {mensaje()}
        </p>
      </Show>

      <Show when={ficha()}>
        {(resultado) => (
          <div class="space-y-4">
            <article class="rounded-xl bg-white p-4 shadow">
              <h2 class="text-lg font-semibold">{resultado().bicicleta.codigo}</h2>
              <p data-testid="estado-bicicleta" class="text-sm text-slate-600">
                Estado: {resultado().bicicleta.estado}
              </p>
            </article>

            <article class="rounded-xl bg-white p-4 shadow">
              <h2 class="mb-3 font-semibold">Historial de averías e intervenciones</h2>
              <Show
                when={resultado().incidencias.length > 0}
                fallback={<p class="text-sm text-slate-600">No hay intervenciones registradas.</p>}
              >
                <ul class="space-y-3">
                  <For each={resultado().incidencias}>
                    {(incidencia) => (
                      <li class="rounded border border-slate-200 p-3 text-sm">
                        <div class="flex flex-wrap justify-between gap-2">
                          <strong>{incidencia.tipo_fallo}</strong>
                          <span class="text-slate-500">{incidencia.estado}</span>
                        </div>
                        <p data-testid="descripcion-incidencia">{incidencia.descripcion_reporte}</p>
                        <p class="text-xs text-slate-500">
                          Reportada: {fechaLocal(incidencia.fecha_reporte)}
                        </p>
                        <Show when={incidencia.piezas_sustituidas}>
                          <p>Piezas: {incidencia.piezas_sustituidas}</p>
                        </Show>
                        <Show when={incidencia.notas_taller}>
                          <p>Notas: {incidencia.notas_taller}</p>
                        </Show>
                      </li>
                    )}
                  </For>
                </ul>
              </Show>
            </article>

            <form onSubmit={guardarParte} class="space-y-3 rounded-xl bg-white p-4 shadow">
              <h2 class="font-semibold">Registrar parte de reparación</h2>
              <label class="block text-sm">
                Piezas sustituidas
                <textarea
                  rows="2"
                  value={piezas()}
                  onInput={(e) => setPiezas(e.currentTarget.value)}
                  data-testid="input-piezas"
                  class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
                />
              </label>
              <label class="block text-sm">
                Notas del técnico
                <textarea
                  rows="3"
                  value={notas()}
                  onInput={(e) => setNotas(e.currentTarget.value)}
                  data-testid="input-notas"
                  class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
                />
              </label>
              <label class="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={marcarOperativa()}
                  onChange={(e) => setMarcarOperativa(e.currentTarget.checked)}
                  data-testid="checkbox-operativa"
                />
                Marcar bicicleta como operativa al guardar
              </label>
              <button
                type="submit"
                disabled={guardando()}
                data-testid="boton-guardar-parte"
                class="rounded bg-emerald-700 px-4 py-2 font-medium text-white disabled:opacity-60"
              >
                Guardar parte
              </button>
            </form>

            <Show when={resultado().bicicleta.estado === 'disponible'}>
              <form onSubmit={ubicarBicicleta} class="flex flex-wrap items-end gap-3 rounded-xl bg-white p-4 shadow">
                <label class="flex-1 text-sm">
                  Código del anclaje libre
                  <input
                    required
                    value={codigoAnclaje()}
                    onInput={(e) => setCodigoAnclaje(e.currentTarget.value)}
                    data-testid="input-codigo-anclaje"
                    class="mt-1 w-full rounded border border-slate-300 px-3 py-2"
                  />
                </label>
                <button
                  type="submit"
                  disabled={guardando()}
                  data-testid="boton-ubicar"
                  class="rounded bg-blue-700 px-4 py-2 font-medium text-white disabled:opacity-60"
                >
                  Confirmar ubicación
                </button>
              </form>
            </Show>
          </div>
        )}
      </Show>
    </section>
  );
};

export default Taller;
