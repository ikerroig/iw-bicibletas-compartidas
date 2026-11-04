// T038 — FR-008: el aviso de viaje en curso aparece en toda la aplicación.
import { MemoryRouter, Route, createMemoryHistory } from '@solidjs/router';
import { render, screen } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import AvisoViaje from '../../src/components/AvisoViaje';
import type { ViajeActivo } from '../../src/services/api/alquiler';
import { fijarViajeEnCurso } from '../../src/stores/viaje';
import { server } from '../server';

const estado = vi.hoisted(() => ({ rol: 'usuario' }));

vi.mock('../../src/stores/sesion', () => ({
  sesion: { usuario: () => ({ id: 1, rol: estado.rol }) },
}));

const VIAJE: ViajeActivo = {
  id: 1,
  usuario_id: 1,
  bicicleta_id: 42,
  bicicleta_codigo: 'BICI-0007',
  estacion_origen_id: 1,
  estacion_origen_nombre: 'Colón',
  estacion_destino_id: null,
  fecha_inicio: '2026-11-04T09:15:00Z',
  fecha_fin: null,
  duracion_minutos: null,
  coste_total: 0,
  estado: 'en_curso',
};

function montar(ruta: string, viaje: ViajeActivo | null) {
  server.use(http.get('*/api/v1/alquileres/activo', () => HttpResponse.json({ viaje })));
  const history = createMemoryHistory();
  history.set({ value: ruta });
  render(() => (
    <MemoryRouter history={history}>
      <Route path="*" component={AvisoViaje} />
    </MemoryRouter>
  ));
}

describe('aviso de viaje en curso', () => {
  beforeEach(() => {
    estado.rol = 'usuario';
    fijarViajeEnCurso(null);
  });

  it('con viaje muestra hora, código y enlace al viaje', async () => {
    montar('/', VIAJE);
    expect(await screen.findByRole('link', { name: 'Ver' })).toHaveAttribute('href', '/alquiler');
    expect(screen.getByRole('status')).toHaveTextContent('BICI-0007');
    expect(screen.getByRole('status')).toHaveTextContent(/\d{2}:\d{2}/);
  });

  it('sin viaje no muestra nada', async () => {
    montar('/', null);
    await new Promise((r) => setTimeout(r, 50));
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('no aparece en /alquiler, que ya muestra el viaje', async () => {
    montar('/alquiler', VIAJE);
    await new Promise((r) => setTimeout(r, 50));
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('no consulta el viaje para otros roles', async () => {
    estado.rol = 'gestor';
    let consultas = 0;
    server.use(
      http.get('*/api/v1/alquileres/activo', () => {
        consultas++;
        return HttpResponse.json({ viaje: VIAJE });
      }),
    );
    const history = createMemoryHistory();
    render(() => (
      <MemoryRouter history={history}>
        <Route path="*" component={AvisoViaje} />
      </MemoryRouter>
    ));
    await new Promise((r) => setTimeout(r, 50));
    expect(consultas).toBe(0);
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});
