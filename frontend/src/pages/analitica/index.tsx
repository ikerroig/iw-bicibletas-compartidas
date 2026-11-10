import {
  createEffect,
  createSignal,
  Show,
  type Component,
} from "solid-js";
import {
  consultarAnalitica,
  descargarCSV,
  type AnaliticaViajes,
} from "../../services/api/analitica";
import { FetchError } from "../../services/api";
import { GraficoBarras } from "./GraficoBarras";

/**
 * HU-08: panel de analítica del gestor — rango de fechas inclusivo
 * (Europe/Madrid, validado server-side → 422), dos gráficos de barras con
 * los arrays que calcula el backend (const. I: el frontend solo pinta),
 * totales, aviso "sin datos en el periodo" y descarga autenticada del CSV
 * del mismo rango (JWT en header, nunca en la URL).
 */

const HORAS = Array.from({ length: 24 }, (_, i) =>
  String(i).padStart(2, "0"),
);
// Orden lunes → domingo (lunes = índice 0 — clarificación 2).
const DIAS = ["lun", "mar", "mié", "jue", "vie", "sáb", "dom"];

function isoLocal(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const dia = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${dia}`;
}

function haceDias(n: number): string {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return isoLocal(d);
}

function costeEs(n: number): string {
  return `${n.toLocaleString("es-ES", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })} €`;
}

function mensajeError(err: unknown): string {
  if (err instanceof FetchError) {
    if (err.status === 403) {
      return "No tienes permiso para acceder a la analítica";
    }
    if (err.status === 422) {
      return `Rango de fechas inválido: ${err.message}`;
    }
    return err.message;
  }
  return "No se pudo cargar la analítica";
}

export const PanelAnalitica: Component = () => {
  const [desde, setDesde] = createSignal(haceDias(30));
  const [hasta, setHasta] = createSignal(isoLocal(new Date()));
  const [datos, setDatos] = createSignal<AnaliticaViajes | null>(null);
  const [cargando, setCargando] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [descargando, setDescargando] = createSignal(false);

  createEffect(() => {
    const d = desde();
    const h = hasta();
    void (async () => {
      setCargando(true);
      setError(null);
      // Validación cliente (el backend manda: su 422 se muestra igualmente).
      if (d > h) {
        setDatos(null);
        setError("Rango inválido: el inicio no puede ser posterior al fin");
        setCargando(false);
        return;
      }
      try {
        setDatos(await consultarAnalitica(d, h));
      } catch (err) {
        setDatos(null);
        setError(mensajeError(err));
      } finally {
        setCargando(false);
      }
    })();
  });

  async function manejarDescarga() {
    if (descargando()) return; // edge del doble clic: un archivo por acción
    setDescargando(true);
    setError(null);
    try {
      await descargarCSV(desde(), hasta());
    } catch (err) {
      setError(mensajeError(err));
    } finally {
      setDescargando(false);
    }
  }

  return (
    <section class="mx-auto max-w-4xl">
      <div class="mb-4">
        <h1 class="text-lg font-bold text-gray-900">Analítica de la red</h1>
        <p class="text-sm text-gray-500">
          Viajes finalizados del rango seleccionado, por hora del día y día de
          la semana (Europe/Madrid)
        </p>
      </div>

      <div class="mb-4 flex flex-wrap items-center gap-3 rounded-xl border border-gray-200 bg-white p-3">
        <label class="flex items-center gap-2 text-sm text-gray-700">
          Desde
          <input
            type="date"
            data-testid="input-desde"
            value={desde()}
            onInput={(e) => setDesde(e.currentTarget.value)}
            class="rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          />
        </label>
        <label class="flex items-center gap-2 text-sm text-gray-700">
          Hasta
          <input
            type="date"
            data-testid="input-hasta"
            value={hasta()}
            onInput={(e) => setHasta(e.currentTarget.value)}
            class="rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          />
        </label>
        <button
          type="button"
          data-testid="boton-descargar"
          disabled={descargando()}
          onClick={() => void manejarDescarga()}
          class="rounded-lg bg-blue-600 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
        >
          {descargando() ? "Descargando…" : "Descargar CSV"}
        </button>
      </div>

      <Show when={error() === null}>
        <Show when={cargando()}>
          <p
            data-testid="cargando-analitica"
            class="p-6 text-center text-sm text-gray-500"
          >
            Cargando analítica…
          </p>
        </Show>
        <Show when={!cargando() && datos()}>
          <p
            data-testid="totales"
            class="mb-4 flex flex-wrap gap-3 text-sm font-semibold text-gray-900"
          >
            <span
              data-testid="total-viajes"
              class="rounded-lg bg-blue-50 px-3 py-2"
            >
              Viajes: {datos()!.total_viajes}
            </span>
            <span
              data-testid="total-coste"
              class="rounded-lg bg-blue-50 px-3 py-2"
            >
              Coste total: {costeEs(datos()!.total_coste)}
            </span>
          </p>
          <Show when={datos()!.total_viajes > 0}>
            <div class="grid gap-4 md:grid-cols-2">
              <GraficoBarras
                testid="grafico-hora"
                titulo="Viajes por hora del día"
                etiquetas={HORAS}
                datos={datos()!.por_hora}
              />
              <GraficoBarras
                testid="grafico-dia"
                titulo="Viajes por día de la semana"
                etiquetas={DIAS}
                datos={datos()!.por_dia_semana}
              />
            </div>
          </Show>
          <Show when={datos()!.total_viajes === 0}>
            <p
              data-testid="sin-datos"
              class="rounded-xl border border-gray-200 bg-white p-6 text-center text-sm text-gray-500"
            >
              sin datos en el periodo
            </p>
          </Show>
        </Show>
      </Show>

      <Show when={error() !== null}>
        <p
          role="alert"
          data-testid="error-analitica"
          class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
        >
          {error()}
        </p>
      </Show>
    </section>
  );
};

export default PanelAnalitica;
