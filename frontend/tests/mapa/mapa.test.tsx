// T032 — Mapa y detalle de HU-01 (FR-001, FR-004, FR-006, FR-008, FR-009,
// SC-001, caso límite de geolocalización).
import { fireEvent, render, screen, waitFor, within } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import { afterEach, describe, expect, it, vi } from 'vitest';
import DetalleEstacion from '../../src/pages/mapa/DetalleEstacion';
import MapaRed from '../../src/pages/mapa/MapaRed';
import { ESTACIONES, snapshot } from '../handlers';
import { server } from '../server';

const porNombre = (n: string) => ESTACIONES.find((e) => e.nombre === n)!;

function marcadores(container: HTMLElement) {
  return Array.from(container.querySelectorAll<HTMLElement>('.leaflet-marker-icon'));
}

function simularGeolocalizacion(resultado: 'denegada' | { lat: number; lng: number }) {
  const getCurrentPosition = vi.fn((ok: PositionCallback, error?: PositionErrorCallback | null) => {
    if (resultado === 'denegada') {
      error?.({
        code: 1,
        message: 'User denied Geolocation',
        PERMISSION_DENIED: 1,
      } as GeolocationPositionError);
    } else {
      ok({ coords: { latitude: resultado.lat, longitude: resultado.lng } } as GeolocationPosition);
    }
  });
  Object.defineProperty(navigator, 'geolocation', {
    configurable: true,
    value: { getCurrentPosition },
  });
  return getCurrentPosition;
}

afterEach(() => {
  Object.defineProperty(navigator, 'geolocation', { configurable: true, value: undefined });
});

describe('DetalleEstacion', () => {
  it('operativa con disponibilidad: muestra los dos contadores', () => {
    render(() => <DetalleEstacion estacion={porNombre('Normal')} />);
    expect(screen.getByTestId('contador-bicis')).toHaveTextContent('7');
    expect(screen.getByTestId('contador-anclajes')).toHaveTextContent('3');
    expect(screen.queryByText(/Sin bicicletas/)).not.toBeInTheDocument();
  });

  it('0 bicicletas → «0 · Sin bicicletas» (FR-008)', () => {
    render(() => <DetalleEstacion estacion={porNombre('Sin bicis')} />);
    expect(screen.getByTestId('contador-bicis')).toHaveTextContent('0 · Sin bicicletas');
    expect(screen.getByTestId('contador-anclajes')).toHaveTextContent('9');
  });

  it('0 anclajes libres → «0 · Sin anclajes libres» (FR-008)', () => {
    render(() => <DetalleEstacion estacion={porNombre('Llena')} />);
    expect(screen.getByTestId('contador-anclajes')).toHaveTextContent('0 · Sin anclajes libres');
  });

  it('ambos a 0 → señala los dos estados, sin desglose de averiadas (FR-006)', () => {
    const { container } = render(() => <DetalleEstacion estacion={porNombre('Averiadas')} />);
    expect(screen.getByTestId('contador-bicis')).toHaveTextContent('0 · Sin bicicletas');
    expect(screen.getByTestId('contador-anclajes')).toHaveTextContent('0 · Sin anclajes libres');
    // Solo los dos contadores: ningún tercer valor con el desglose de averías.
    expect(container.querySelectorAll('dd')).toHaveLength(2);
  });

  it('fuera de servicio → «Fuera de servicio» y «—» sin contadores ni acciones (FR-009)', () => {
    render(() => <DetalleEstacion estacion={porNombre('Cerrada')} />);
    expect(screen.getAllByText('Fuera de servicio').length).toBeGreaterThan(0);
    expect(screen.getByTestId('contador-bicis')).toHaveTextContent('—');
    expect(screen.getByTestId('contador-anclajes')).toHaveTextContent('—');
    // anclajes_libres = 4 en los datos, pero no se muestra.
    expect(screen.getByTestId('contador-anclajes')).not.toHaveTextContent('4');
    expect(
      screen.queryByRole('button', { name: /recoger|desbloquear|devolver/i }),
    ).not.toBeInTheDocument();
  });
});

describe('MapaRed', () => {
  it('pinta TODAS las estaciones de la API (SC-001) con texto/aria-label distinto por estado', async () => {
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    const etiquetas = marcadores(container).map((m) => m.getAttribute('aria-label') ?? '');
    expect(new Set(etiquetas).size).toBe(ESTACIONES.length);
    expect(etiquetas.find((l) => l.startsWith('Cerrada'))).toMatch(/Fuera de servicio/);
    expect(etiquetas.find((l) => l.startsWith('Sin bicis'))).toMatch(/Sin bicicletas/);
    expect(etiquetas.find((l) => l.startsWith('Llena'))).toMatch(/Sin anclajes libres/);
    // Texto visible en el propio marcador, no solo una clase de color (FR-008).
    const cerrada = marcadores(container).find((m) =>
      m.getAttribute('aria-label')?.startsWith('Cerrada'),
    )!;
    expect(cerrada).toHaveTextContent(/Fuera de servicio/);
  });

  it('un marcador fuera de servicio es seleccionable y abre su detalle (FR-009)', async () => {
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    const cerrada = marcadores(container).find((m) =>
      m.getAttribute('aria-label')?.startsWith('Cerrada'),
    )!;
    fireEvent.click(cerrada);
    const detalle = await screen.findByRole('region', { name: /Cerrada/ });
    expect(within(detalle).getByTestId('contador-bicis')).toHaveTextContent('—');
  });

  it('el detalle se actualiza con los datos en vivo de la estación seleccionada', async () => {
    const { container } = render(() => <MapaRed intervaloMs={50} />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    fireEvent.click(
      marcadores(container).find((m) => m.getAttribute('aria-label')?.startsWith('Normal'))!,
    );
    const detalle = await screen.findByRole('region', { name: /Normal/ });
    expect(within(detalle).getByTestId('contador-bicis')).toHaveTextContent('7');

    server.use(
      http.get('*/api/v1/estaciones', () =>
        HttpResponse.json(
          snapshot(ESTACIONES.map((e) => (e.id === 1 ? { ...e, bicis_disponibles: 0 } : e))),
        ),
      ),
    );
    await waitFor(() =>
      expect(within(detalle).getByTestId('contador-bicis')).toHaveTextContent('0 · Sin bicicletas'),
    );

    // T105 — FR-004 ampliado (escenario 6): paso a fuera de servicio con el
    // detalle abierto; no se cierra ni pierde la selección.
    server.use(
      http.get('*/api/v1/estaciones', () =>
        HttpResponse.json(
          snapshot(
            ESTACIONES.map((e) =>
              e.id === 1 ? { ...e, estado: 'fuera_de_servicio' as const } : e,
            ),
          ),
        ),
      ),
    );
    await waitFor(() => expect(detalle).toHaveTextContent('Fuera de servicio'));
    expect(within(detalle).getByTestId('contador-bicis')).toHaveTextContent('—');
    expect(within(detalle).getByTestId('contador-anclajes')).toHaveTextContent('—');
    expect(screen.getByRole('region', { name: /Normal/ })).toBe(detalle);
  });

  it('con la geolocalización denegada se encuadra en toda la red sin error', async () => {
    const geo = simularGeolocalizacion('denegada');
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    expect(geo).toHaveBeenCalled();
    expect(container.querySelector('[data-encuadre]')).toHaveAttribute('data-encuadre', 'red');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('con permiso de ubicación centra el mapa en el usuario', async () => {
    simularGeolocalizacion({ lat: 39.47, lng: -0.376 });
    const { container } = render(() => <MapaRed />);
    await waitFor(() =>
      expect(container.querySelector('[data-encuadre]')).toHaveAttribute(
        'data-encuadre',
        'usuario',
      ),
    );
  });

  it('sin estaciones muestra un estado vacío', async () => {
    server.use(http.get('*/api/v1/estaciones', () => HttpResponse.json(snapshot([]))));
    render(() => <MapaRed />);
    expect(await screen.findByText(/No hay estaciones para mostrar/)).toBeInTheDocument();
  });

  it('si la conexión falla conserva el último estado conocido', async () => {
    // El primer handler tiene prioridad: una respuesta correcta y después caída.
    server.use(
      http.get('*/api/v1/estaciones', () => HttpResponse.json(snapshot()), { once: true }),
      http.get('*/api/v1/estaciones', () => HttpResponse.error()),
    );
    const { container } = render(() => <MapaRed intervaloMs={50} />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    expect(await screen.findByText(/Reintentando/)).toBeInTheDocument();
    expect(marcadores(container)).toHaveLength(ESTACIONES.length);
  });

  // T101 — FR-010 (escenario 4). Con el intervalo de producción (15 s), la
  // petición inmediata tras pulsar «Reintentar» solo puede venir del botón.
  it('si la primera carga falla: mensaje, botón «Reintentar» y ningún marcador (FR-010)', async () => {
    let peticiones = 0;
    server.use(
      http.get('*/api/v1/estaciones', () => {
        peticiones++;
        return HttpResponse.error();
      }),
    );
    const { container } = render(() => <MapaRed />);

    const alerta = await screen.findByRole('alert');
    expect(alerta).toHaveTextContent('No se pudo cargar la disponibilidad');
    expect(marcadores(container)).toHaveLength(0);
    expect(screen.queryByTestId('contador-bicis')).not.toBeInTheDocument();

    server.use(
      http.get('*/api/v1/estaciones', () => {
        peticiones++;
        return HttpResponse.json(snapshot());
      }),
    );
    const antes = peticiones;
    fireEvent.click(within(alerta).getByRole('button', { name: 'Reintentar' }));

    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    expect(peticiones).toBe(antes + 1);
    expect(screen.queryByText('No se pudo cargar la disponibilidad')).not.toBeInTheDocument();
  });

  it('si la primera carga falla, reintenta solo y el mensaje desaparece al recibir datos (FR-010)', async () => {
    server.use(
      http.get('*/api/v1/estaciones', () => HttpResponse.error(), { once: true }),
      http.get('*/api/v1/estaciones', () => HttpResponse.json(snapshot())),
    );
    const { container } = render(() => <MapaRed intervaloMs={50} />);
    expect(await screen.findByText('No se pudo cargar la disponibilidad')).toBeInTheDocument();
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    expect(screen.queryByText('No se pudo cargar la disponibilidad')).not.toBeInTheDocument();
  });

  it('con datos previos, un fallo NO muestra el mensaje de primera carga (se mantiene R17)', async () => {
    server.use(
      http.get('*/api/v1/estaciones', () => HttpResponse.json(snapshot()), { once: true }),
      http.get('*/api/v1/estaciones', () => HttpResponse.error()),
    );
    const { container } = render(() => <MapaRed intervaloMs={50} />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    expect(await screen.findByText(/Reintentando/)).toBeInTheDocument();
    expect(screen.queryByText('No se pudo cargar la disponibilidad')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Reintentar' })).not.toBeInTheDocument();
  });
});
