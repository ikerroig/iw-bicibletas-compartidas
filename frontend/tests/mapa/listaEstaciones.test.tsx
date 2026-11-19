// T103 — Listado textual de estaciones (HU-01 FR-011, SC-008, escenario 5;
// research R21). El listado usa el mismo store y la misma función estadoVisual
// que el mapa, así que debe mostrar exactamente el mismo estado y contadores.
import { fireEvent, render, screen, waitFor, within } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import MapaRed from '../../src/pages/mapa/MapaRed';
import { ESTACIONES, snapshot } from '../handlers';
import { server } from '../server';

function marcadores(container: HTMLElement) {
  return Array.from(container.querySelectorAll<HTMLElement>('.leaflet-marker-icon'));
}

async function montarYAbrirLista() {
  const r = render(() => <MapaRed intervaloMs={50} />);
  await waitFor(() => expect(marcadores(r.container)).toHaveLength(ESTACIONES.length));
  fireEvent.click(screen.getByRole('tab', { name: 'Lista' }));
  const lista = await screen.findByRole('list', { name: 'Listado de estaciones' });
  return { ...r, lista };
}

const filas = (lista: HTMLElement) => within(lista).getAllByRole('button');

describe('listado de estaciones (FR-011)', () => {
  it('el conmutador «Mapa / Lista» es accesible y se maneja con el teclado', async () => {
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));

    const pestanas = screen.getByRole('tablist', { name: 'Vista de la disponibilidad' });
    const mapa = within(pestanas).getByRole('tab', { name: 'Mapa' });
    const lista = within(pestanas).getByRole('tab', { name: 'Lista' });
    expect(mapa).toHaveAttribute('aria-selected', 'true');
    expect(lista).toHaveAttribute('aria-selected', 'false');

    mapa.focus();
    fireEvent.keyDown(mapa, { key: 'ArrowRight' });
    expect(lista).toHaveAttribute('aria-selected', 'true');
    expect(lista).toHaveFocus();
    expect(screen.getByRole('list', { name: 'Listado de estaciones' })).toBeInTheDocument();

    fireEvent.keyDown(lista, { key: 'ArrowLeft' });
    expect(mapa).toHaveAttribute('aria-selected', 'true');
    expect(screen.queryByRole('list', { name: 'Listado de estaciones' })).not.toBeInTheDocument();
  });

  it('contiene TODAS las estaciones con el mismo estado y contadores que el mapa (SC-008)', async () => {
    const { container, lista } = await montarYAbrirLista();
    const delMapa = marcadores(container)
      .map((m) => m.getAttribute('aria-label'))
      .sort();
    const delListado = filas(lista)
      .map((b) => b.getAttribute('aria-label'))
      .sort();
    expect(delListado).toHaveLength(ESTACIONES.length);
    expect(delListado).toEqual(delMapa);
  });

  it('las filas van ordenadas por nombre y muestran texto de estado, no solo color', async () => {
    const { lista } = await montarYAbrirLista();
    const nombres = filas(lista).map((b) => b.querySelector('[data-campo="nombre"]')?.textContent);
    expect(nombres).toEqual(
      [...ESTACIONES.map((e) => e.nombre)].sort((a, b) => a.localeCompare(b, 'es')),
    );

    const fila = (n: string) =>
      filas(lista).find((b) => b.getAttribute('aria-label')?.startsWith(n))!;
    expect(fila('Sin bicis')).toHaveTextContent('Sin bicicletas');
    expect(fila('Llena')).toHaveTextContent('Sin anclajes libres');
    expect(fila('Averiadas')).toHaveTextContent('Sin bicicletas ni anclajes libres');
    expect(fila('Normal')).toHaveTextContent('7');
    expect(fila('Normal')).toHaveTextContent('3');
  });

  it('una estación fuera de servicio aparece con «Fuera de servicio» y «—», sin contadores', async () => {
    const { lista } = await montarYAbrirLista();
    const cerrada = filas(lista).find((b) => b.getAttribute('aria-label')?.startsWith('Cerrada'))!;
    expect(cerrada).toHaveTextContent('Fuera de servicio');
    expect(cerrada).toHaveTextContent('—');
    // Cerrada tiene 4 anclajes libres en los datos: no deben mostrarse.
    expect(within(cerrada).queryByText('4')).not.toBeInTheDocument();
  });

  it('se actualiza con el siguiente refresco, igual que el mapa (FR-003)', async () => {
    const { lista } = await montarYAbrirLista();
    server.use(
      http.get('*/api/v1/estaciones', () =>
        HttpResponse.json(
          snapshot(ESTACIONES.map((e) => (e.id === 1 ? { ...e, bicis_disponibles: 0 } : e))),
        ),
      ),
    );
    await waitFor(() =>
      expect(
        filas(lista).find((b) => b.getAttribute('aria-label')?.startsWith('Normal')),
      ).toHaveTextContent('Sin bicicletas'),
    );
  });

  it('pulsar una fila abre el detalle de esa estación', async () => {
    const { lista } = await montarYAbrirLista();
    fireEvent.click(filas(lista).find((b) => b.getAttribute('aria-label')?.startsWith('Llena'))!);
    const detalle = await screen.findByRole('region', { name: /Llena/ });
    expect(within(detalle).getByTestId('contador-anclajes')).toHaveTextContent(
      '0 · Sin anclajes libres',
    );
  });

  it('sigue funcionando aunque las teselas del mapa de fondo no carguen', async () => {
    const { container } = render(() => <MapaRed intervaloMs={50} />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    // Simula la caída del proveedor de teselas: cada imagen de tesela falla.
    container.querySelectorAll('img.leaflet-tile').forEach((img) => fireEvent.error(img));
    container.querySelector('.leaflet-tile-pane')?.remove();

    fireEvent.click(screen.getByRole('tab', { name: 'Lista' }));
    const lista = await screen.findByRole('list', { name: 'Listado de estaciones' });
    expect(filas(lista)).toHaveLength(ESTACIONES.length);
  });
});
