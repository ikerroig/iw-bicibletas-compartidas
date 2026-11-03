// Estado de la red en vivo (T029, research R16/R17).
//
// - Polling de GET /estaciones cada 15 s; si falla, se reintenta en el
//   siguiente ciclo SIN borrar el último estado conocido.
// - ultimaActualizacionOk: reloj del CLIENTE al recibir una respuesta 2xx
//   (evita falsos avisos por desfase de reloj con el servidor).
// - desactualizado: ahora − ultimaActualizacionOk > 30 s (estrictamente mayor),
//   recalculado cada segundo; vuelve a false con la primera respuesta correcta.
// - Pestaña oculta: se pausa el polling; al volver, petición inmediata y, durante
//   esa petición (máx. 2 s), no se muestra el aviso para evitar un parpadeo.
//
// Debe crearse dentro de un componente (usa onCleanup).
import { createMemo, createSignal, onCleanup, type Accessor } from 'solid-js';
import {
  listarEstaciones,
  type EstacionEstado,
  type SnapshotEstaciones,
} from '../services/api/estaciones';

export const INTERVALO_POLLING_MS = 15_000;
export const UMBRAL_DESACTUALIZADO_MS = 30_000;
export const GRACIA_REANUDAR_MS = 2_000;
const TICK_MS = 1_000;

export interface OpcionesRed {
  intervaloMs?: number;
  umbralDesactualizadoMs?: number;
  graciaReanudarMs?: number;
  obtener?: (signal: AbortSignal) => Promise<SnapshotEstaciones>;
}

export interface RedEnVivo {
  estaciones: Accessor<EstacionEstado[]>;
  /** true tras la primera respuesta correcta. */
  cargado: Accessor<boolean>;
  /** Mensaje del último fallo, o null si la última petición fue bien. */
  error: Accessor<string | null>;
  ultimaActualizacionOk: Accessor<number | null>;
  desactualizado: Accessor<boolean>;
  refrescar: () => Promise<void>;
}

export function crearRedEnVivo(opciones: OpcionesRed = {}): RedEnVivo {
  const intervalo = opciones.intervaloMs ?? INTERVALO_POLLING_MS;
  const umbral = opciones.umbralDesactualizadoMs ?? UMBRAL_DESACTUALIZADO_MS;
  const gracia = opciones.graciaReanudarMs ?? GRACIA_REANUDAR_MS;
  const obtener = opciones.obtener ?? listarEstaciones;
  const inicio = Date.now();

  const [estaciones, setEstaciones] = createSignal<EstacionEstado[]>([]);
  const [cargado, setCargado] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [ultimaOk, setUltimaOk] = createSignal<number | null>(null);
  const [reloj, setReloj] = createSignal(inicio);
  const [suprimirHasta, setSuprimirHasta] = createSignal(0);

  let enCurso: AbortController | null = null;
  let poll: ReturnType<typeof setInterval> | undefined;
  let destruido = false;

  async function refrescar(): Promise<void> {
    if (enCurso || destruido) return;
    const ctrl = new AbortController();
    enCurso = ctrl;
    // Una petición colgada no debe bloquear el polling indefinidamente.
    const limite = setTimeout(() => ctrl.abort(), Math.min(intervalo, 10_000));
    try {
      const s = await obtener(ctrl.signal);
      if (destruido) return;
      setEstaciones(s.estaciones);
      setUltimaOk(Date.now());
      setCargado(true);
      setError(null);
    } catch (e) {
      if (destruido) return;
      setError(
        ctrl.signal.aborted
          ? 'El servidor tarda demasiado en responder'
          : e instanceof Error
            ? e.message
            : 'No se pudo actualizar',
      );
    } finally {
      clearTimeout(limite);
      if (enCurso === ctrl) enCurso = null;
      if (!destruido) {
        setSuprimirHasta(0); // fin de la gracia: si falló, el aviso ya puede verse
        setReloj(Date.now());
      }
    }
  }

  const pararPolling = () => {
    if (poll !== undefined) clearInterval(poll);
    poll = undefined;
  };
  const arrancarPolling = () => {
    pararPolling();
    poll = setInterval(() => void refrescar(), intervalo);
  };

  const alCambiarVisibilidad = () => {
    if (document.visibilityState === 'hidden') {
      pararPolling();
    } else {
      setSuprimirHasta(Date.now() + gracia);
      void refrescar();
      arrancarPolling();
    }
  };

  const tick = setInterval(() => setReloj(Date.now()), TICK_MS);
  document.addEventListener('visibilitychange', alCambiarVisibilidad);
  void refrescar();
  if (document.visibilityState !== 'hidden') arrancarPolling();

  onCleanup(() => {
    destruido = true;
    pararPolling();
    clearInterval(tick);
    document.removeEventListener('visibilitychange', alCambiarVisibilidad);
    enCurso?.abort();
  });

  const desactualizado = createMemo(() => {
    const ahora = reloj();
    if (ahora < suprimirHasta()) return false;
    return ahora - (ultimaOk() ?? inicio) > umbral;
  });

  return { estaciones, cargado, error, ultimaActualizacionOk: ultimaOk, desactualizado, refrescar };
}
