// Utilidades de los tests de HU-02: la página de desbloqueo necesita un router
// porque redirige al mapa a los roles distintos de `usuario` (FR-009, D9).
// Cada fichero de test simula su propia sesión con vi.mock('…/stores/sesion').
import { MemoryRouter, Route, createMemoryHistory } from '@solidjs/router';
import { render } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import Desbloqueo from '../../src/pages/alquiler/Desbloqueo';
import type { EstadoAbono, ViajeActivo } from '../../src/services/api/alquiler';

export const VIAJE: ViajeActivo = {
  id: 1,
  usuario_id: 7,
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

export const ABONO_VIGENTE: EstadoAbono = { vigente: true, codigo: null, mensaje: null };

/** Respuesta de GET /alquileres/activo (D16: incluye siempre el abono). */
export const activo = (viaje: ViajeActivo | null, abono: EstadoAbono = ABONO_VIGENTE) =>
  http.get('*/api/v1/alquileres/activo', () => HttpResponse.json({ viaje, abono }));

export const rechazo = (status: number, codigo: string, mensaje: string) =>
  http.post('*/api/v1/alquileres/desbloqueo', () =>
    HttpResponse.json({ error: { codigo, mensaje } }, { status }),
  );

/** Monta la aplicación mínima en `/alquiler`, con el mapa como destino de la redirección. */
export function montarDesbloqueo() {
  const history = createMemoryHistory();
  history.set({ value: '/alquiler' });
  render(() => (
    <MemoryRouter history={history}>
      <Route path="/" component={() => <h1>Mapa</h1>} />
      <Route path="/alquiler" component={Desbloqueo} />
    </MemoryRouter>
  ));
  return history;
}
