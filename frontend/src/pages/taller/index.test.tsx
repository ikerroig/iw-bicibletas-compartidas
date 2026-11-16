import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@solidjs/testing-library';
import { Taller } from './index';

const fichaAveriada = {
  bicicleta: { id: 10, codigo: 'T-HU06-1', estado: 'averiada' },
  incidencias: [
    {
      id: 1,
      bicicleta_id: 10,
      usuario_reporta_id: 2,
      operario_id: null,
      tipo_fallo: 'frenos',
      descripcion_reporte: 'La maneta no frena',
      piezas_sustituidas: null,
      notas_taller: null,
      fecha_reporte: '2026-11-16T10:00:00Z',
      fecha_reparacion: null,
      estado: 'pendiente',
    },
  ],
};

const fichaOperativa = {
  bicicleta: { id: 10, codigo: 'T-HU06-1', estado: 'disponible' },
  incidencias: [
    {
      ...fichaAveriada.incidencias[0],
      operario_id: 3,
      piezas_sustituidas: 'Maneta derecha',
      notas_taller: 'Frenos ajustados',
      fecha_reparacion: '2026-11-16T10:10:00Z',
      estado: 'resuelta',
    },
  ],
};

const respuesta = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });

afterEach(() => {
  vi.unstubAllGlobals();
  cleanup();
});

describe('HU-06 · página de taller', () => {
  it('consulta la ficha, guarda el parte y permite ubicar la bicicleta', async () => {
    let reparada = false;
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (init?.method === 'POST' && url.includes('/intervenciones')) {
        reparada = true;
        return respuesta(fichaOperativa.incidencias[0], 201);
      }
      if (init?.method === 'POST' && url.includes('/ubicacion')) {
        return respuesta({
          id: 1,
          nombre: 'Taller test',
          bicis_disponibles: 1,
          anclajes_libres: 0,
          actualizado_en: '2026-11-16T10:11:00Z',
        });
      }
      return respuesta(reparada ? fichaOperativa : fichaAveriada);
    });
    vi.stubGlobal('fetch', fetchMock);

    render(() => <Taller />);
    fireEvent.input(screen.getByTestId('input-bicicleta-id'), { target: { value: '10' } });
    fireEvent.click(screen.getByTestId('boton-buscar-ficha'));

    expect(await screen.findByText('T-HU06-1')).toBeInTheDocument();
    expect(screen.getByTestId('descripcion-incidencia')).toHaveTextContent('La maneta no frena');

    fireEvent.input(screen.getByTestId('input-piezas'), { target: { value: 'Maneta derecha' } });
    fireEvent.input(screen.getByTestId('input-notas'), { target: { value: 'Frenos ajustados' } });
    fireEvent.click(screen.getByTestId('boton-guardar-parte'));

    expect(await screen.findByTestId('mensaje-taller')).toHaveTextContent('Parte de reparación');
    expect(await screen.findByTestId('input-codigo-anclaje')).toBeInTheDocument();

    fireEvent.input(screen.getByTestId('input-codigo-anclaje'), { target: { value: 'T06-A02' } });
    fireEvent.click(screen.getByTestId('boton-ubicar'));
    expect(await screen.findByTestId('mensaje-taller')).toHaveTextContent('Taller test');

    const peticiones = fetchMock.mock.calls;
    expect(peticiones.some(([, init]) => init?.method === 'POST')).toBe(true);
  });
});
