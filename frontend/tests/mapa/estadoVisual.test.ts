// T096 — estadoVisual: todas las combinaciones (research R18, R19; FR-008, FR-009).
import { describe, expect, it } from 'vitest';
import { estadoVisual, PRESENTACION, type EstadoVisual } from '../../src/pages/mapa/estadoVisual';

type Caso = [
  estado: 'operativa' | 'fuera_de_servicio',
  bicis: number,
  libres: number,
  esperado: EstadoVisual,
];

const casos: Caso[] = [
  ['operativa', 5, 5, 'normal'],
  ['operativa', 0, 5, 'sin_bicis'],
  ['operativa', 5, 0, 'sin_anclajes'],
  ['operativa', 0, 0, 'sin_bicis_y_sin_anclajes'],
  // Fuera de servicio prevalece sobre cualquier combinación de contadores.
  ['fuera_de_servicio', 5, 5, 'fuera_de_servicio'],
  ['fuera_de_servicio', 0, 5, 'fuera_de_servicio'],
  ['fuera_de_servicio', 5, 0, 'fuera_de_servicio'],
  ['fuera_de_servicio', 0, 0, 'fuera_de_servicio'],
];

describe('estadoVisual', () => {
  it.each(casos)('%s con %i bicis y %i libres → %s', (estado, bicis, libres, esperado) => {
    expect(estadoVisual({ estado, bicis_disponibles: bicis, anclajes_libres: libres })).toBe(
      esperado,
    );
  });

  it('acepta que bicis + libres < capacidad (anclajes con averiadas, R19)', () => {
    // capacidad 10: 2 disponibles, 1 libre, 7 ocupados por averiadas.
    expect(estadoVisual({ estado: 'operativa', bicis_disponibles: 2, anclajes_libres: 1 })).toBe(
      'normal',
    );
  });

  it('cada estado tiene icono y etiqueta propios, distintos entre sí (no solo color)', () => {
    const estados = Object.keys(PRESENTACION) as EstadoVisual[];
    expect(estados).toHaveLength(5);
    const etiquetas = new Set(estados.map((e) => PRESENTACION[e].etiqueta));
    const iconos = new Set(estados.map((e) => PRESENTACION[e].icono));
    expect(etiquetas.size).toBe(5);
    expect(iconos.size).toBe(5);
    for (const e of estados) {
      expect(PRESENTACION[e].etiqueta.length).toBeGreaterThan(0);
      expect(PRESENTACION[e].clase).toMatch(/\S/);
    }
  });

  it('usa los textos de la spec', () => {
    expect(PRESENTACION.fuera_de_servicio.etiqueta).toBe('Fuera de servicio');
    expect(PRESENTACION.sin_bicis.etiqueta).toBe('Sin bicicletas');
    expect(PRESENTACION.sin_anclajes.etiqueta).toBe('Sin anclajes libres');
    expect(PRESENTACION.sin_bicis_y_sin_anclajes.etiqueta).toBe(
      'Sin bicicletas ni anclajes libres',
    );
  });
});
