import {
  createEffect,
  createSignal,
  For,
  Show,
  type Component,
} from "solid-js";
import {
  listarViajes,
  type ViajeHistorial,
} from "../../services/api/viajes";
import { FetchError } from "../../services/api";

/**
 * HU-04: "Mis Viajes" — historial cronológico de viajes finalizados con
 * filtro de mes/año (Europe/Madrid), sumatorio del periodo, etiqueta de
 * tramo gratuito tal cual del backend y paginación server-side (50/pág.).
 * Sin filtro se muestra el historial completo (FR-003).
 */

const PAGE_SIZE = 50;

const MESES = [
  "enero",
  "febrero",
  "marzo",
  "abril",
  "mayo",
  "junio",
  "julio",
  "agosto",
  "septiembre",
  "octubre",
  "noviembre",
  "diciembre",
];

/** dd/mm/aaaa HH:MM en Europe/Madrid (FR-001, FR-002). */
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

function costeEs(n: number): string {
  return `${n.toLocaleString("es-ES", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })} €`;
}

export const MisViajes: Component = () => {
  const [viajes, setViajes] = createSignal<ViajeHistorial[]>([]);
  const [total, setTotal] = createSignal(0);
  const [suma, setSuma] = createSignal(0);
  const [mes, setMes] = createSignal(""); // "" = sin filtro
  const [anio, setAnio] = createSignal(""); // "" = sin filtro
  const [pagina, setPagina] = createSignal(1);
  const [cargando, setCargando] = createSignal(true);
  const [error, setError] = createSignal<string | null>(null);

  const anios = () => {
    const a = new Date().getFullYear();
    return [a - 1, a, a + 1]; // research D7: actual−1 … actual+1
  };
  const filtrando = () => mes() !== "" && anio() !== "";
  const totalPaginas = () => Math.max(1, Math.ceil(total() / PAGE_SIZE));

  // Mes/año → desde = primer día, hasta = último día (ambos inclusivos).
  function rangoMes(): { desde?: string; hasta?: string } {
    if (!filtrando()) return {};
    const m = Number(mes());
    const a = Number(anio());
    const mm = String(m).padStart(2, "0");
    const ultimo = new Date(a, m, 0).getDate(); // día 0 del mes siguiente
    return { desde: `${a}-${mm}-01`, hasta: `${a}-${mm}-${String(ultimo).padStart(2, "0")}` };
  }

  createEffect(() => {
    const p = pagina();
    void (async () => {
      setCargando(true);
      setError(null);
      try {
        const r = await listarViajes({
          ...rangoMes(),
          page: p,
          pageSize: PAGE_SIZE,
        });
        setViajes(r.items);
        setTotal(r.total);
        setSuma(r.suma_coste);
      } catch (err) {
        setViajes([]);
        setTotal(0);
        setSuma(0);
        setError(
          err instanceof FetchError ? err.message : "No se pudo cargar el historial",
        );
      } finally {
        setCargando(false);
      }
    })();
  });

  function cambiarMes(e: Event) {
    setMes((e.currentTarget as HTMLSelectElement).value);
    setPagina(1);
  }
  function cambiarAnio(e: Event) {
    setAnio((e.currentTarget as HTMLSelectElement).value);
    setPagina(1);
  }

  return (
    <section class="mx-auto max-w-3xl">
      <div class="mb-4 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-lg font-bold text-gray-900">Mis viajes</h1>
          <p class="text-sm text-gray-500">
            Historial de viajes finalizados, más recientes primero
          </p>
        </div>
        <Show when={!error()}>
          <p
            data-testid="sumatorio"
            class="rounded-lg bg-blue-50 px-3 py-2 text-sm font-semibold text-blue-900"
          >
            Suma del periodo: {costeEs(suma())}
          </p>
        </Show>
      </div>

      <div class="mb-4 flex flex-wrap items-center gap-3 rounded-xl border border-gray-200 bg-white p-3">
        <label class="flex items-center gap-2 text-sm text-gray-700">
          Mes
          <select
            data-testid="select-mes"
            value={mes()}
            onChange={cambiarMes}
            class="rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          >
            <option value="">—</option>
            <For each={MESES}>
              {(nombre, i) => (
                <option value={String(i() + 1)}>{nombre}</option>
              )}
            </For>
          </select>
        </label>
        <label class="flex items-center gap-2 text-sm text-gray-700">
          Año
          <select
            data-testid="select-anio"
            value={anio()}
            onChange={cambiarAnio}
            class="rounded-lg border border-gray-300 px-2 py-1.5 focus:border-blue-500 focus:outline-none"
          >
            <option value="">—</option>
            <For each={anios()}>{(a) => <option value={String(a)}>{a}</option>}</For>
          </select>
        </label>
        <Show when={filtrando()}>
          <button
            type="button"
            data-testid="boton-quitar-filtro"
            onClick={() => {
              setMes("");
              setAnio("");
              setPagina(1);
            }}
            class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50"
          >
            Quitar filtro
          </button>
        </Show>
      </div>

      <Show when={!cargando()}>
        <Show
          when={error() === null}
          fallback={
            <p
              role="alert"
              data-testid="error-historial"
              class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800"
            >
              {error()}
            </p>
          }
        >
          <Show
            when={viajes().length > 0}
            fallback={
              <p
                data-testid="estado-vacio"
                class="rounded-xl border border-gray-200 bg-white p-6 text-center text-sm text-gray-500"
              >
                {filtrando()
                  ? "Sin viajes en el periodo seleccionado."
                  : "Aún no tienes viajes."}
              </p>
            }
          >
            <ul class="flex flex-col gap-2">
              <For each={viajes()}>
                {(v) => (
                  <li
                    data-testid="fila-viaje"
                    class="grid grid-cols-[auto_1fr_auto] items-center gap-x-4 gap-y-1 rounded-xl border border-gray-200 bg-white p-3 text-sm"
                  >
                    <span data-testid="fila-fecha" class="font-mono text-gray-900">
                      {formatearFecha(v.fecha_inicio)}
                    </span>
                    <span data-testid="fila-origen" class="text-gray-700">
                      {v.estacion_origen_nombre} → {v.estacion_destino_nombre ?? "—"}
                      <span data-testid="fila-destino" class="sr-only">
                        {v.estacion_destino_nombre ?? ""}
                      </span>
                    </span>
                    <span class="flex items-center gap-3">
                      <span data-testid="fila-duracion" class="text-gray-500">
                        {v.duracion_minutos ?? 0} min
                      </span>
                      <span data-testid="fila-coste" class="font-semibold text-gray-900">
                        {costeEs(v.coste_total)}
                      </span>
                    </span>
                    <Show when={v.etiqueta}>
                      <span
                        data-testid="fila-etiqueta"
                        class="col-span-3 rounded-md bg-green-100 px-2 py-1 text-center text-xs font-semibold text-green-800"
                      >
                        {v.etiqueta}
                      </span>
                    </Show>
                  </li>
                )}
              </For>
            </ul>

            <nav
              class="mt-4 flex items-center justify-between text-sm text-gray-700"
              aria-label="Paginación del historial"
            >
              <button
                type="button"
                data-testid="boton-anterior"
                disabled={pagina() <= 1}
                onClick={() => setPagina((p) => Math.max(1, p - 1))}
                class="rounded-lg border border-gray-300 px-3 py-1.5 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                Anterior
              </button>
              <span data-testid="pagina-info">
                Página {pagina()} de {totalPaginas()}
              </span>
              <button
                type="button"
                data-testid="boton-siguiente"
                disabled={pagina() >= totalPaginas()}
                onClick={() => setPagina((p) => p + 1)}
                class="rounded-lg border border-gray-300 px-3 py-1.5 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                Siguiente
              </button>
            </nav>
          </Show>
        </Show>
      </Show>

      <Show when={cargando()}>
        <p data-testid="cargando-historial" class="p-6 text-center text-sm text-gray-500">
          Cargando viajes…
        </p>
      </Show>
    </section>
  );
};

export default MisViajes;
