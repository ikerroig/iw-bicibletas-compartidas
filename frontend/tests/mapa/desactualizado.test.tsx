// T098 — Aviso de «datos desactualizados» (FR-007, SC-005; research R17).
import { render, screen } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import AvisoDesactualizado from '../../src/pages/mapa/AvisoDesactualizado';
import MapaRed from '../../src/pages/mapa/MapaRed';
import { crearRedEnVivo } from '../../src/stores/red';
import { snapshot } from '../handlers';
import { server } from '../server';

const URL = '*/api/v1/estaciones';
const ok = () => http.get(URL, () => HttpResponse.json(snapshot()));
const caida = () => http.get(URL, () => HttpResponse.error());

function Arnes() {
  const red = crearRedEnVivo();
  return (
    <>
      <AvisoDesactualizado
        ultimaActualizacionOk={red.ultimaActualizacionOk()}
        desactualizado={red.desactualizado()}
      />
      <span data-testid="contadores">{red.estaciones().length}</span>
    </>
  );
}

/** Avanza el reloj falso dejando resolver las promesas de fetch/MSW. */
async function avanzar(ms: number) {
  await vi.advanceTimersByTimeAsync(ms);
}

describe('aviso de datos desactualizados', () => {
  beforeEach(() => {
    vi.useFakeTimers({
      toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'Date'],
      now: new Date('2026-11-03T10:00:00Z'),
    });
  });
  afterEach(() => vi.useRealTimers());

  it('a los 30 s exactos no avisa; a los 31 s sí, con los contadores visibles; una respuesta correcta lo quita', async () => {
    // Primera petición correcta (t = 0) y después la conexión cae.
    // En server.use el primer handler tiene prioridad: una respuesta correcta y luego caída.
    server.use(
      http.get(URL, () => HttpResponse.json(snapshot()), { once: true }),
      caida(),
    );
    render(() => <Arnes />);
    await avanzar(0);
    expect(screen.getByTestId('contadores')).toHaveTextContent('5');
    expect(screen.getByText(/Actualizado a las \d{2}:\d{2}:\d{2}/)).toBeInTheDocument();
    const horaInicial = screen.getByText(/Actualizado a las/).textContent;

    await avanzar(30_000); // polls fallidos a los 15 y 30 s
    expect(screen.queryByText('Datos desactualizados')).not.toBeInTheDocument();

    await avanzar(1_000); // 31 s > 30 s
    expect(screen.getByText('Datos desactualizados')).toBeInTheDocument();
    expect(screen.getByTestId('contadores')).toHaveTextContent('5'); // FR-007: siguen visibles
    expect(screen.getByText(/Actualizado a las/).textContent).toBe(horaInicial);

    server.use(ok()); // vuelve la conexión
    await avanzar(14_000); // siguiente poll a los 45 s
    expect(screen.queryByText('Datos desactualizados')).not.toBeInTheDocument();
    expect(screen.getByText(/Actualizado a las/).textContent).not.toBe(horaInicial);
  });

  it('el aviso es una región accesible con texto, no solo color', async () => {
    // En server.use el primer handler tiene prioridad: una respuesta correcta y luego caída.
    server.use(
      http.get(URL, () => HttpResponse.json(snapshot()), { once: true }),
      caida(),
    );
    render(() => <Arnes />);
    await avanzar(0);
    await avanzar(31_000);
    const region = screen.getByRole('status');
    expect(region).toHaveTextContent('Datos desactualizados');
  });

  // T112 — converge F2. Antes de FR-010 este test exigía el aviso «Datos
  // desactualizados» también sin datos, pero ese aviso dice «Mostramos los
  // últimos datos conocidos» y no hay ninguno: sin datos manda FR-010.
  it('sin ninguna carga correcta NO muestra «Datos desactualizados» (no hay datos que mostrar)', async () => {
    server.use(caida());
    render(() => <Arnes />);
    await avanzar(31_000);
    expect(screen.queryByText('Datos desactualizados')).not.toBeInTheDocument();
    expect(screen.queryByText(/últimos datos conocidos/)).not.toBeInTheDocument();
    expect(screen.getByText(/Sin datos todavía/)).toBeInTheDocument();
  });

  it('sin datos y pasados 30 s, el mapa muestra solo el mensaje de FR-010', async () => {
    server.use(caida());
    render(() => <MapaRed />);
    await avanzar(31_000);
    expect(screen.getByRole('alert')).toHaveTextContent('No se pudo cargar la disponibilidad');
    expect(screen.queryByText('Datos desactualizados')).not.toBeInTheDocument();
  });

  it('al volver a una pestaña oculta más de 30 s no parpadea el aviso si la respuesta llega bien', async () => {
    let visible: DocumentVisibilityState = 'visible';
    vi.spyOn(document, 'visibilityState', 'get').mockImplementation(() => visible);
    render(() => <Arnes />);
    await avanzar(0);

    visible = 'hidden';
    document.dispatchEvent(new Event('visibilitychange'));
    await avanzar(60_000); // oculta 1 min: sin polling

    // La respuesta al volver tarda 1 s: durante la gracia de 2 s no hay aviso.
    server.use(
      http.get(URL, async () => {
        await new Promise((r) => setTimeout(r, 1_000));
        return HttpResponse.json(snapshot());
      }),
    );
    visible = 'visible';
    document.dispatchEvent(new Event('visibilitychange'));
    await avanzar(0);
    expect(screen.queryByText('Datos desactualizados')).not.toBeInTheDocument();
    await avanzar(1_000);
    expect(screen.queryByText('Datos desactualizados')).not.toBeInTheDocument();
  });
});
