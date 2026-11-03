// Estado visual de una estación (T095, research R18). Es lógica de
// PRESENTACIÓN derivada de los contadores del backend (constitución I): no
// decide nada de negocio.
import type { EstacionEstado } from '../../services/api/estaciones';

export type EstadoVisual =
  'fuera_de_servicio' | 'sin_bicis_y_sin_anclajes' | 'sin_bicis' | 'sin_anclajes' | 'normal';

/** Prioridad R18: fuera de servicio > ambos a 0 > sin bicis > sin anclajes > normal. */
export function estadoVisual(
  e: Pick<EstacionEstado, 'estado' | 'bicis_disponibles' | 'anclajes_libres'>,
): EstadoVisual {
  if (e.estado === 'fuera_de_servicio') return 'fuera_de_servicio';
  const sinBicis = e.bicis_disponibles === 0;
  const sinAnclajes = e.anclajes_libres === 0;
  if (sinBicis && sinAnclajes) return 'sin_bicis_y_sin_anclajes';
  if (sinBicis) return 'sin_bicis';
  if (sinAnclajes) return 'sin_anclajes';
  return 'normal';
}

export interface Presentacion {
  /** Texto de la spec (detalle, leyenda y aria-label). */
  etiqueta: string;
  /** Texto corto visible en el marcador (null = se muestra el nº de bicis). */
  corta: string | null;
  /** Icono decorativo; siempre acompañado de texto (FR-008: no solo color). */
  icono: string;
  /** Clases Tailwind del marcador / insignia. */
  clase: string;
}

export const PRESENTACION: Record<EstadoVisual, Presentacion> = {
  normal: {
    etiqueta: 'Con disponibilidad',
    corta: null,
    icono: '🚲',
    clase: 'bg-emerald-600 text-white border-emerald-800',
  },
  sin_bicis: {
    etiqueta: 'Sin bicicletas',
    corta: 'Sin bicis',
    icono: '∅',
    clase: 'bg-amber-100 text-amber-900 border-amber-600',
  },
  sin_anclajes: {
    etiqueta: 'Sin anclajes libres',
    corta: 'Llena',
    icono: '▣',
    clase: 'bg-sky-100 text-sky-900 border-sky-700',
  },
  sin_bicis_y_sin_anclajes: {
    etiqueta: 'Sin bicicletas ni anclajes libres',
    corta: 'Sin bicis · Llena',
    icono: '⊗',
    clase: 'bg-orange-100 text-orange-900 border-orange-700',
  },
  fuera_de_servicio: {
    etiqueta: 'Fuera de servicio',
    corta: 'Fuera de servicio',
    icono: '✕',
    clase: 'bg-slate-200 text-slate-700 border-slate-500 border-dashed',
  },
};

/** Descripción completa para lectores de pantalla. */
export function descripcionAccesible(e: EstacionEstado): string {
  const v = estadoVisual(e);
  const base = `${e.nombre}: ${PRESENTACION[v].etiqueta}`;
  if (v === 'fuera_de_servicio') return base;
  return `${base}. ${e.bicis_disponibles} bicicletas disponibles, ${e.anclajes_libres} anclajes libres`;
}
