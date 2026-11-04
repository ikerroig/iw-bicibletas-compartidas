// T028 — FR-009: la opción de desbloqueo solo aparece al rol `usuario`.
import { Route, Router } from '@solidjs/router';
import { render, screen } from '@solidjs/testing-library';
import { describe, expect, it, vi } from 'vitest';
import Layout from '../../src/components/Layout';

const estado = vi.hoisted(() => ({ rol: null as string | null }));

vi.mock('../../src/stores/sesion', () => ({
  sesion: {
    token: () => 't',
    autenticado: () => true,
    usuario: () => (estado.rol ? { nombre: 'Prueba', rol: estado.rol } : null),
  },
  cerrarSesion: () => {},
}));

function montar(rol: string | null) {
  estado.rol = rol;
  render(() => (
    <Router>
      <Route path="/" component={Layout} />
    </Router>
  ));
}

describe('menú de áreas', () => {
  it('el rol usuario ve Desbloquear', async () => {
    montar('usuario');
    expect(await screen.findByRole('link', { name: 'Desbloquear' })).toBeInTheDocument();
  });

  it.each(['gestor', 'operario_taller', 'operario_logistica'])('%s no ve Desbloquear', async (rol) => {
    montar(rol);
    expect(await screen.findByRole('link', { name: 'Mapa' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Desbloquear' })).not.toBeInTheDocument();
  });

  it('sin usuario guardado no se ve Desbloquear, pero sí el Mapa', async () => {
    montar(null);
    expect(await screen.findByRole('link', { name: 'Mapa' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Desbloquear' })).not.toBeInTheDocument();
  });
});
