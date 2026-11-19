// T106 — Sesión caducada con el mapa abierto (HU-01 FR-012, escenario 7;
// research R22). La parte del servidor (401 sin token o con token falso) la
// cubre TestEstacionesExigeSesion en backend/tests/integration. Aquí se prueba
// el cliente con el store de sesión REAL y el formulario de login real:
// 401 durante el polling → cerrar sesión → /login → entrar → vuelta al mapa.
import { Route, Router } from '@solidjs/router';
import { fireEvent, render, screen, waitFor } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import { afterEach, describe, expect, it } from 'vitest';
import Layout from '../../src/components/Layout';
import Login from '../../src/pages/Login';
import MapaRed from '../../src/pages/mapa/MapaRed';
import { cerrarSesion, sesion } from '../../src/stores/sesion';
import { ESTACIONES, snapshot } from '../handlers';
import { server } from '../server';

function marcadores() {
  return document.querySelectorAll('.leaflet-marker-icon');
}

function montarApp() {
  window.history.replaceState({}, '', '/');
  render(() => (
    <Router>
      <Route path="/login" component={Login} />
      <Route path="/" component={Layout}>
        <Route path="/" component={() => <MapaRed intervaloMs={50} />} />
      </Route>
    </Router>
  ));
}

async function entrar() {
  fireEvent.input(await screen.findByLabelText('Email'), { target: { value: 'ana@si16.test' } });
  fireEvent.input(screen.getByLabelText('Contraseña'), { target: { value: 'password' } });
  fireEvent.click(screen.getByRole('button', { name: 'Entrar' }));
}

afterEach(() => {
  cerrarSesion();
  window.history.replaceState({}, '', '/');
});

describe('sesión no válida durante la consulta (FR-012)', () => {
  it('sin sesión, la vista de disponibilidad redirige al login y, tras entrar, vuelve al mapa', async () => {
    cerrarSesion();
    montarApp();
    await waitFor(() => expect(window.location.pathname).toBe('/login'));
    expect(marcadores()).toHaveLength(0);

    await entrar();
    await waitFor(() => expect(window.location.pathname).toBe('/'));
    await waitFor(() => expect(marcadores()).toHaveLength(ESTACIONES.length));
  });

  it('si la sesión caduca con el mapa abierto: cierra sesión, va al login y al entrar vuelve al mapa', async () => {
    cerrarSesion();
    montarApp();
    await entrar();
    await waitFor(() => expect(marcadores()).toHaveLength(ESTACIONES.length));
    expect(sesion.autenticado()).toBe(true);

    // El servidor deja de aceptar el token (p. ej. ha caducado).
    server.use(
      http.get('*/api/v1/estaciones', () =>
        HttpResponse.json(
          { error: { codigo: 'UNAUTENTICADO', mensaje: 'Sesión no válida o caducada' } },
          { status: 401 },
        ),
      ),
    );
    await waitFor(() => expect(window.location.pathname).toBe('/login'));
    expect(sesion.autenticado()).toBe(false);
    expect(marcadores()).toHaveLength(0);

    // Con una sesión nueva, el servidor vuelve a responder.
    server.use(http.get('*/api/v1/estaciones', () => HttpResponse.json(snapshot())));
    await entrar();
    await waitFor(() => expect(window.location.pathname).toBe('/'));
    await waitFor(() => expect(marcadores()).toHaveLength(ESTACIONES.length));
  });
});
