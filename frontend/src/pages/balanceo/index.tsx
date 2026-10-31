import {
  createEffect,
  createSignal,
  For,
  Show,
  type Component,
} from "solid-js";
import {
  confirmarTraslado,
  listarAlertas,
  type AlertaBalanceo,
  type EstacionEstado,
  type TrasladoResponse,
} from "../../services/api/balanceo";
import { FetchError } from "../../services/api";

/**
 * HU-07: panel de alertas de balanceo para el operario de calle. Pinta las
 * alertas TAL CUAL llegan del backend (orden por urgencia — const. I), con la
 * etiqueta de umbral, el porcentaje en es-ES y la fecha en Europe/Madrid; y
 * permite confirmar traslados de lote que devuelven ambas estaciones
 * recalculadas (FR-006/FR-007).
 */

const ETIQUETAS: Record<string, string> = {
  vacia_menos_10: "Vacía (< 10 %)",
  llena_mas_90: "Llena (> 90 %)",
};

/** dd/mm/aaaa HH:MM en Europe/Madrid (fecha de la alerta). */
function formatearFecha(iso: string): string {
  const partes = new Intl.DateTimeFormat("es-ES", {
    timeZone: "Europe/Madrid",
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).formatToParts(new Date(iso));
  const g = (t: string) => partes.find((p) => p.type === t)?.value ?? "";
  return `${g("day")}/${g("month")}/${g("year")} ${g("hour")}:${g("minute")}`;
}

function pctEs(n: number): string {
  return `${n.toLocaleString("es-ES", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })} %`;
}

function TarjetaEstado(props: { id: string; titulo: string; estado: EstacionEstado }) {
  return (
    <div
      data-testid={props.id}
      class="rounded-xl border border-gray-200 bg-white p-3 text-sm text-gray-700"
    >
      <p class="mb-1 font-semibold text-gray-900">
        {props.titulo}: {props.estado.nombre}
      </p>
      <p>Bicis disponibles {props.estado.bicis_disponibles}</p>
      <p>Anclajes libres {props.estado.anclajes_libres}</p>
      <p>Ocupación {pctEs(props.estado.ocupacion_pct)}</p>
      <p class="text-xs text-gray-500">Estado: {props.estado.estado}</p>
    </div>
  );
}

export const Balanceo: Component = () => {
  const [alertas, setAlertas] = createSignal<AlertaBalanceo[]>([]);
  const [cargando, setCargando] = createSignal(true);
  const [error, setError] = createSignal<string | null>(null);
  const [origen, setOrigen] = createSignal("");
  const [destino, setDestino] = createSignal("");
  const [unidades, setUnidades] = createSignal("");
  const [resultado, setResultado] = createSignal<TrasladoResponse | null>(null);
  const [enviando, setEnviando] = createSignal(false);
  const [errorTraslado, setErrorTraslado] = createSignal<string | null>(null);
  const [recarga, setRecarga] = createSignal(0);

  // Una única fuente de disparo de la lectura del panel (arranque, botón
  // "Actualizar" y recarga tras un traslado confirmado).
  createEffect(() => {
    recarga();
    void (async () => {
      setCargando(true);
      setError(null);
      try {
        const r = await listarAlertas();
        setAlertas(r.items);
      } catch (err) {
        setAlertas([]);
        setError(
          err instanceof FetchError
            ? err.message
            : "No se pudo cargar el panel de balanceo",
        );
      } finally {
        setCargando(false);
      }
    })();
  });

  function cambiar(setter: (v: string) => void) {
    return (e: Event) => {
      setter((e.currentTarget as HTMLInputElement).value);
      if (errorTraslado()) setErrorTraslado(null);
    };
  }

  async function enviar(e: Event) {
    e.preventDefault();
    setErrorTraslado(null);
    const o = Number(origen());
    const d = Number(destino());
    const u = Number(unidades());
    if (!Number.isInteger(o) || !Number.isInteger(d) || !Number.isInteger(u) ||
        o <= 0 || d <= 0 || u < 1 || o === d) {
      setErrorTraslado(
        "Introduce dos estaciones distintas y un número de unidades ≥ 1",
      );
      return;
    }
    setEnviando(true);
    try {
      const r = await confirmarTraslado({
        estacion_origen_id: o,
        estacion_destino_id: d,
        unidades: u,
      });
      setResultado(r);
      setAlertas([]); // el panel se refresca a continuación con el estado real
      setRecarga((x) => x + 1);
    } catch (err) {
      setErrorTraslado(
        err instanceof FetchError
          ? err.message
          : "Error inesperado al confirmar el traslado",
      );
    } finally {
      setEnviando(false);
    }
  }

  return (
    <section class="mx-auto max-w-3xl">
      <div class="mb-4 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-lg font-bold text-gray-900">Balanceo de flota</h1>
          <p class="text-sm text-gray-500">
            Estaciones fuera de rango (&lt; 10 % o &gt; 90 %), más urgentes
            primero
          </p>
        </div>
        <button
          type="button"
          data-testid="boton-actualizar"
          onClick={() => setRecarga((x) => x + 1)}
          disabled={cargando()}
          class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50 disabled:opacity-60"
        >
          Actualizar
        </button>
      </div>

      <Show
        when={!cargando()}
        fallback={
          <p data-testid="cargando-panel" class="p-6 text-center text-sm text-gray-500">
            Cargando alertas…
          </p>
        }
      >
        <Show
          when={error() === null}
          fallback={
            <p
              role="alert"
              data-testid="error-panel"
              class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
            >
              {error()}
            </p>
          }
        >
          <Show
            when={alertas().length > 0}
            fallback={
              <p
                data-testid="estado-vacio"
                class="rounded-xl border border-gray-200 bg-white p-6 text-center text-sm text-gray-500"
              >
                Sin estaciones descompensadas — la red está equilibrada
              </p>
            }
          >
            <ul class="flex flex-col gap-2">
              <For each={alertas()}>
                {(a) => (
                  <li
                    data-testid="fila-alerta"
                    class="grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-1 rounded-xl border border-gray-200 bg-white p-3 text-sm"
                  >
                    <span data-testid="fila-nombre" class="font-semibold text-gray-900">
                      {a.estacion_nombre}
                    </span>
                    <span
                      data-testid="fila-fecha"
                      class="font-mono text-xs text-gray-500"
                    >
                      {formatearFecha(a.fecha_generacion)}
                    </span>
                    <span
                      data-testid="fila-tipo"
                      class="rounded-md bg-amber-100 px-2 py-1 text-xs font-semibold text-amber-900"
                    >
                      {ETIQUETAS[a.tipo_alerta] ?? a.tipo_alerta}
                    </span>
                    <span class="flex items-center gap-3 text-xs text-gray-600">
                      <span data-testid="fila-porcentaje" class="font-semibold text-gray-900">
                        {pctEs(a.porcentaje_ocupacion)}
                      </span>
                      <span
                        data-testid="fila-severidad"
                        class="rounded-md bg-red-100 px-2 py-1 font-semibold uppercase text-red-800"
                      >
                        {a.severidad}
                      </span>
                    </span>
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </Show>
      </Show>

      <Show when={resultado()}>
        <div
          data-testid="resultado-traslado"
          class="mt-4 rounded-xl border border-green-200 bg-green-50 p-4"
        >
          <h2 class="mb-3 font-bold text-green-900">Traslado confirmado</h2>
          <div class="grid gap-3 sm:grid-cols-2">
            <TarjetaEstado id="tarjeta-origen" titulo="Origen" estado={resultado()!.origen} />
            <TarjetaEstado id="tarjeta-destino" titulo="Destino" estado={resultado()!.destino} />
          </div>
        </div>
      </Show>

      <form
        onSubmit={enviar}
        class="mt-4 flex flex-wrap items-end gap-3 rounded-xl border border-gray-200 bg-white p-3"
      >
        <h2 class="w-full text-sm font-semibold text-gray-900">
          Confirmar traslado de lote
        </h2>
        <label class="flex flex-col gap-1 text-sm text-gray-700">
          Estación origen (id)
          <input
            type="number"
            min="1"
            value={origen()}
            onInput={cambiar(setOrigen)}
            onChange={cambiar(setOrigen)}
            data-testid="input-origen"
            class="w-28 rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          />
        </label>
        <label class="flex flex-col gap-1 text-sm text-gray-700">
          Estación destino (id)
          <input
            type="number"
            min="1"
            value={destino()}
            onInput={cambiar(setDestino)}
            onChange={cambiar(setDestino)}
            data-testid="input-destino"
            class="w-28 rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          />
        </label>
        <label class="flex flex-col gap-1 text-sm text-gray-700">
          Unidades
          <input
            type="number"
            min="1"
            value={unidades()}
            onInput={cambiar(setUnidades)}
            onChange={cambiar(setUnidades)}
            data-testid="input-unidades"
            class="w-24 rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          />
        </label>
        <button
          type="submit"
          disabled={enviando()}
          data-testid="boton-traslado"
          class="flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2 font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <Show
            when={!enviando()}
            fallback={
              <>
                <span
                  class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
                  aria-hidden="true"
                />
                <span>Confirmando…</span>
              </>
            }
          >
            <span>Confirmar traslado</span>
          </Show>
        </button>

        <Show when={errorTraslado()}>
          <p
            role="alert"
            data-testid="error-traslado"
            class="w-full rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
          >
            {errorTraslado()}
          </p>
        </Show>
      </form>
    </section>
  );
};

export default Balanceo;
