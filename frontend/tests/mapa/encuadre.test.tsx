// T107 — Encuadre inicial con geolocalización (HU-01 FR-013, escenario 8;
// research R23; hallazgo H1 de specs/spec-001/refinamiento/pruebas-previas.md).
// El mapa solo se centra en el usuario si está dentro del área de la red; si
// está fuera, se encuadra toda la red, llegue la posición antes o después que
// los datos. data-encuadre ('red' | 'usuario') refleja la decisión.
import { render, waitFor } from '@solidjs/testing-library';
import { afterEach, describe, expect, it, vi } from 'vitest';
import MapaRed from '../../src/pages/mapa/MapaRed';
import { ESTACIONES } from '../handlers';

// Las estaciones de prueba están en València (39,47 / −0,37).
const DENTRO = { lat: 39.472, lng: -0.371 };
const FUERA = { lat: 38.385, lng: -0.513 }; // Alicante, a ~120 km

type Resultado = 'denegada' | { lat: number; lng: number };

/** Geolocalización simulada. Con `diferida`, la posición llega cuando se llama a `responder()`. */
function simularGeolocalizacion(resultado: Resultado, diferida = false) {
  let pendiente: (() => void) | null = null;
  const getCurrentPosition = vi.fn((ok: PositionCallback, error?: PositionErrorCallback | null) => {
    const responder = () => {
      if (resultado === 'denegada')
        error?.({ code: 1, message: 'denegada', PERMISSION_DENIED: 1 } as GeolocationPositionError);
      else
        ok({
          coords: { latitude: resultado.lat, longitude: resultado.lng },
        } as GeolocationPosition);
    };
    if (diferida) pendiente = responder;
    else responder();
  });
  Object.defineProperty(navigator, 'geolocation', {
    configurable: true,
    value: { getCurrentPosition },
  });
  return { responder: () => pendiente?.() };
}

const encuadre = (c: HTMLElement) =>
  c.querySelector('[data-encuadre]')?.getAttribute('data-encuadre');
const marcadores = (c: HTMLElement) => c.querySelectorAll('.leaflet-marker-icon');

afterEach(() => {
  Object.defineProperty(navigator, 'geolocation', { configurable: true, value: undefined });
});

describe('encuadre inicial del mapa (FR-013)', () => {
  it('(a) usuario FUERA del área y su posición llega ANTES que los datos → se encuadra la red', async () => {
    simularGeolocalizacion(FUERA); // responde en el acto, antes del fetch
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(marcadores(container)).toHaveLength(ESTACIONES.length));
    await waitFor(() => expect(encuadre(container)).toBe('red'));
  });

  it('(b) usuario FUERA del área y su posición llega DESPUÉS que los datos → se mantiene la red', async () => {
    const geo = simularGeolocalizacion(FUERA, true);
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(encuadre(container)).toBe('red'));
    geo.responder();
    await new Promise((r) => setTimeout(r, 20));
    expect(encuadre(container)).toBe('red');
  });

  it('(c) usuario DENTRO del área → se centra en él (antes o después de los datos)', async () => {
    simularGeolocalizacion(DENTRO);
    const antes = render(() => <MapaRed />);
    await waitFor(() => expect(encuadre(antes.container)).toBe('usuario'));
    antes.unmount();

    const geo = simularGeolocalizacion(DENTRO, true);
    const despues = render(() => <MapaRed />);
    await waitFor(() => expect(encuadre(despues.container)).toBe('red'));
    geo.responder();
    await waitFor(() => expect(encuadre(despues.container)).toBe('usuario'));
  });

  it('(d) permiso denegado → se encuadra la red', async () => {
    simularGeolocalizacion('denegada');
    const { container } = render(() => <MapaRed />);
    await waitFor(() => expect(encuadre(container)).toBe('red'));
  });
});
