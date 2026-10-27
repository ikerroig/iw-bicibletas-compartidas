import { createSignal, Show, type Component } from "solid-js";
import { devolver, type ViajeDevolucion } from "../../services/api/devolucion";
import { FetchError } from "../../services/api";

/**
 * HU-03: devolución de bicicleta. Muestra el resultado del cierre
 * (duración, coste y etiqueta de tramo gratuito) o el mensaje de error
 * del contrato (409 conflictos / 422 entrada inválida).
 */
export const Devolucion: Component = () => {
  const [codigo, setCodigo] = createSignal("");
  const [cargando, setCargando] = createSignal(false);
  const [viaje, setViaje] = createSignal<ViajeDevolucion | null>(null);
  const [error, setError] = createSignal<string | null>(null);

  async function enviar(e: Event) {
    e.preventDefault();
    setViaje(null);
    setError(null);
    const cod = codigo().trim();
    if (!cod) {
      setError("Introduce el código del anclaje");
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
    } finally {
      setCargando(false);
    }
  }

  return (
    <section class="mx-auto max-w-md">
      <h1 class="mb-4 text-xl font-bold">Devolver bicicleta</h1>

      <form onSubmit={enviar} class="mb-6 flex flex-col gap-3">
        <label class="flex flex-col gap-1">
          <span class="text-sm font-medium">Código del anclaje</span>
          <input
            type="text"
            value={codigo()}
            onInput={(e) => setCodigo(e.currentTarget.value)}
            placeholder="AN-00010"
            data-testid="input-anclaje"
            class="rounded border px-3 py-2"
            disabled={cargando()}
          />
        </label>
        <button
          type="submit"
          disabled={cargando()}
          data-testid="boton-devolver"
          class="rounded bg-blue-600 px-4 py-2 font-semibold text-white disabled:opacity-50"
        >
          {cargando() ? "Cerrando viaje…" : "Confirmar devolución"}
        </button>
      </form>

      <Show when={error()}>
        <p
          role="alert"
          data-testid="error-devolucion"
          class="rounded border border-red-300 bg-red-50 p-3 text-red-800"
        >
          {error()}
        </p>
      </Show>

      <Show when={viaje()}>
        <div
          data-testid="exito-devolucion"
          class="rounded border border-green-300 bg-green-50 p-4 text-green-900"
        >
          <h2 class="mb-2 font-bold">¡Viaje finalizado!</h2>
          <ul class="flex flex-col gap-1 text-sm">
            <li>
              Duración:{" "}
              <strong data-testid="duracion">
                {viaje()!.duracion_minutos ?? 0} min
              </strong>
            </li>
            <li>
              Coste:{" "}
              <strong data-testid="coste">
                {viaje()!.coste_total.toFixed(2)} €
              </strong>
            </li>
            <Show when={viaje()!.etiqueta}>
              <li data-testid="etiqueta" class="font-semibold">
                {viaje()!.etiqueta}
              </li>
            </Show>
          </ul>
        </div>
      </Show>
    </section>
  );
};

export default Devolucion;
