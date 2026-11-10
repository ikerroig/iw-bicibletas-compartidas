import { createSignal, Show, type Component } from "solid-js";
import { devolver, type ViajeDevolucion } from "../../services/api/devolucion";
import { FetchError } from "../../services/api";

/**
 * HU-03: devolución de bicicleta. Muestra el resultado del cierre (duración,
 * coste y etiqueta de tramo gratuito) o el mensaje de error del contrato
 * (409 conflictos / 422 entrada inválida). UX: código en mayúsculas, spinner
 * durante el envío, foco devuelto al input y botón de repetir tras el éxito.
 */
export const Devolucion: Component = () => {
  const [codigo, setCodigo] = createSignal("");
  const [cargando, setCargando] = createSignal(false);
  const [viaje, setViaje] = createSignal<ViajeDevolucion | null>(null);
  const [error, setError] = createSignal<string | null>(null);
  let inputRef: HTMLInputElement | undefined;

  async function enviar(e: Event) {
    e.preventDefault();
    setViaje(null);
    setError(null);
    const cod = codigo().trim();
    if (!cod) {
      setError("Introduce el código del anclaje");
      inputRef?.focus();
      return;
    }
    setCargando(true);
    try {
      const v = await devolver(cod);
      setViaje(v);
      setCodigo("");
    } catch (err) {
      // 409/422/401: mostramos el mensaje del envelope de error.
      setError(err instanceof FetchError ? err.message : "Error inesperado");
      inputRef?.focus();
    } finally {
      setCargando(false);
    }
  }

  function otraVez() {
    setViaje(null);
    setError(null);
    setCodigo("");
    inputRef?.focus();
  }

  return (
    <section class="mx-auto flex max-w-md flex-col items-center">
      <div class="w-full rounded-2xl border border-gray-200 bg-white p-6 shadow-sm">
        <div class="mb-5 flex items-center gap-3">
          <span
            class="flex h-11 w-11 items-center justify-center rounded-full bg-blue-100 text-2xl"
            aria-hidden="true"
          >
            🚲
          </span>
          <div>
            <h1 class="text-lg font-bold text-gray-900">Devolver bicicleta</h1>
            <p class="text-sm text-gray-500">
              Ancla la bicicleta y cierra tu viaje
            </p>
          </div>
        </div>

        <form onSubmit={enviar} class="flex flex-col gap-4">
          <label class="flex flex-col gap-1">
            <span class="text-sm font-medium text-gray-700">
              Código del anclaje
            </span>
            <input
              ref={inputRef}
              type="text"
              value={codigo()}
              onInput={(e) => {
                setCodigo(e.currentTarget.value.toUpperCase());
                if (error()) setError(null);
              }}
              placeholder="AN-00010"
              maxlength="10"
              autocomplete="off"
              autocapitalize="characters"
              spellcheck={false}
              data-testid="input-anclaje"
              class="rounded-lg border border-gray-300 px-3 py-2 font-mono tracking-wider uppercase placeholder:font-sans placeholder:normal-case placeholder:tracking-normal focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200 disabled:opacity-60"
              disabled={cargando()}
            />
            <span class="text-xs text-gray-500">
              Está impreso en la placa del anclaje de la estación.
            </span>
          </label>

          <button
            type="submit"
            disabled={cargando()}
            data-testid="boton-devolver"
            class="flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <Show
              when={cargando()}
              fallback={<span>Confirmar devolución</span>}
            >
              <span
                class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
                aria-hidden="true"
              />
              <span>Cerrando viaje…</span>
            </Show>
          </button>
        </form>

        <Show when={error()}>
          <p
            role="alert"
            data-testid="error-devolucion"
            class="mt-4 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
          >
            <span aria-hidden="true">⚠️</span>
            <span>{error()}</span>
          </p>
        </Show>

        <Show when={viaje()}>
          <div
            data-testid="exito-devolucion"
            class="mt-4 rounded-xl border border-green-200 bg-green-50 p-4"
          >
            <div class="mb-3 flex items-center gap-2">
              <span
                class="flex h-7 w-7 items-center justify-center rounded-full bg-green-600 text-sm font-bold text-white"
                aria-hidden="true"
              >
                ✓
              </span>
              <h2 class="font-bold text-green-900">Viaje finalizado</h2>
            </div>
            <dl class="grid grid-cols-[1fr_auto] gap-y-1.5 text-sm text-green-900">
              <dt class="text-green-700">Duración</dt>
              <dd data-testid="duracion" class="text-right font-semibold">
                {viaje()!.duracion_minutos ?? 0} min
              </dd>
              <dt class="text-green-700">Coste</dt>
              <dd data-testid="coste" class="text-right font-semibold">
                {viaje()!.coste_total.toFixed(2)} €
              </dd>
              <Show when={viaje()!.etiqueta}>
                <dt class="sr-only">Concepto</dt>
                <dd
                  data-testid="etiqueta"
                  class="col-span-2 mt-1 rounded-md bg-green-100 px-2 py-1 text-center text-xs font-semibold text-green-800"
                >
                  {viaje()!.etiqueta}
                </dd>
              </Show>
            </dl>
            <button
              type="button"
              onClick={otraVez}
              data-testid="boton-otra"
              class="mt-4 w-full rounded-lg border border-green-300 bg-white px-4 py-2 text-sm font-semibold text-green-800 transition hover:bg-green-100"
            >
              Devolver otra bicicleta
            </button>
          </div>
        </Show>
      </div>

      <p class="mt-3 text-center text-xs text-gray-400">
        Los primeros 30 minutos son gratuitos.
      </p>
    </section>
  );
};

export default Devolucion;
