// T050 — FR-009 (research D9): un rol distinto de `usuario` que escribe la
// dirección de la pantalla de desbloqueo vuelve al mapa sin ver el formulario.
// El backend sigue rechazando con 403 (TestDesbloqueoRequiereRolUsuario).
import { screen } from '@solidjs/testing-library';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { server } from '../server';
import { activo, montarDesbloqueo } from './montar';

const estado = vi.hoisted(() => ({ rol: null as string | null }));

vi.mock('../../src/stores/sesion', () => ({
  sesion: { usuario: () => (estado.rol ? { id: 1, nombre: 'Prueba', rol: estado.rol } : null) },
}));

describe('guardia de la pantalla de desbloqueo', () => {
  let consultas = 0;
  beforeEach(() => {
    // jsdom no implementa scrollTo, que el router llama al redirigir.
    vi.spyOn(window, 'scrollTo').mockImplementation(() => {});
    consultas = 0;
    server.use(
      http.get('*/api/v1/alquileres/activo', () => {
        consultas++;
        return HttpResponse.json({ viaje: null, abono: { vigente: true, codigo: null, mensaje: null } });
      }),
    );
  });

  it.each(['gestor', 'operario_taller', 'operario_logistica', null])(
    'con rol %s lleva al mapa sin mostrar el formulario',
    async (rol) => {
      estado.rol = rol;
      const history = montarDesbloqueo();
      expect(await screen.findByRole('heading', { name: 'Mapa' })).toBeInTheDocument();
      expect(history.get()).toBe('/');
      expect(screen.queryByLabelText('Código de anclaje')).not.toBeInTheDocument();
      expect(consultas).toBe(0);
    },
  );

  it('con rol usuario muestra el formulario', async () => {
    estado.rol = 'usuario';
    server.use(activo(null));
    montarDesbloqueo();
    expect(await screen.findByLabelText('Código de anclaje')).toBeInTheDocument();
  });
});
