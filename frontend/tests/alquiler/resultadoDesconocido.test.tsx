// T049 — FR-011 (research D17): si la respuesta al desbloqueo no llega o es un
// 5xx, la página no da el desbloqueo por fallido ni por hecho: consulta el
// viaje activo y enseña el resultado real. Nunca reenvía el POST.
import { screen } from '@solidjs/testing-library';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { server } from '../server';
import { ABONO_VIGENTE, VIAJE, montarDesbloqueo } from './montar';

vi.mock('../../src/stores/sesion', () => ({
  sesion: { usuario: () => ({ id: 7, nombre: 'Eva', rol: 'usuario' }) },
}));

const COMPROBANDO = 'No sabemos si se ha desbloqueado; comprobando…';
const NO_DESBLOQUEADO = 'No se ha desbloqueado, inténtalo de nuevo';

/** Escenario: `post` responde al desbloqueo; `consultas` devuelve, en orden, lo que responde cada GET posterior a la carga. */
function escenario(post: () => Response | Promise<Response>, consultas: (() => Response)[]) {
  const cuenta = { posts: 0, gets: 0 };
  server.use(
    http.post('*/api/v1/alquileres/desbloqueo', () => {
      cuenta.posts++;
      return post();
    }),
    http.get('*/api/v1/alquileres/activo', () => {
      const n = cuenta.gets++;
      if (n === 0) return HttpResponse.json({ viaje: null, abono: ABONO_VIGENTE }); // carga de la página
      const r = consultas[Math.min(n - 1, consultas.length - 1)];
      return r();
    }),
  );
  return cuenta;
}

const conViaje = () => HttpResponse.json({ viaje: VIAJE, abono: ABONO_VIGENTE });
const sinViaje = () => HttpResponse.json({ viaje: null, abono: ABONO_VIGENTE });
const caido = () => HttpResponse.error();

async function enviar(u = userEvent.setup()) {
  await u.type(await screen.findByLabelText('Código de anclaje'), 'E01-A01');
  await u.click(screen.getByRole('button', { name: 'Desbloquear' }));
  return u;
}

afterEach(() => {
  vi.useRealTimers();
});

describe('resultado desconocido del desbloqueo (FR-011)', () => {
  it('red caída al enviar y el viaje sí se creó → avisa, comprueba y muestra el viaje', async () => {
    let soltarConsulta!: () => void;
    const consultaRetenida = new Promise<void>((r) => (soltarConsulta = r));
    const cuenta = escenario(caido, [conViaje]);
    server.use(
      http.get('*/api/v1/alquileres/activo', async () => {
        if (cuenta.gets++ === 0) return sinViaje();
        await consultaRetenida;
        return conViaje();
      }),
    );
    montarDesbloqueo();
    await enviar();
    expect(await screen.findByText(COMPROBANDO)).toBeInTheDocument();
    soltarConsulta();
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
    expect(cuenta.posts).toBe(1);
  });

  it('error 500 y el viaje no existe → «No se ha desbloqueado» con el formulario activo', async () => {
    const cuenta = escenario(
      () => HttpResponse.json({ error: { codigo: 'ERROR_INTERNO', mensaje: 'x' } }, { status: 500 }),
      [sinViaje],
    );
    montarDesbloqueo();
    await enviar();
    expect(await screen.findByText(NO_DESBLOQUEADO)).toBeInTheDocument();
    expect(screen.getByLabelText('Código de anclaje')).toBeEnabled();
    expect(screen.getByRole('button', { name: 'Desbloquear' })).toBeEnabled();
    expect(cuenta.posts).toBe(1);
    expect(cuenta.gets).toBe(2);
  });

  it('sin respuesta durante más de 10 s → cancela y comprueba', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const cuenta = escenario(() => new Promise<Response>(() => {}), [sinViaje]);
    montarDesbloqueo();
    await enviar(userEvent.setup({ advanceTimers: vi.advanceTimersByTime }));
    await vi.advanceTimersByTimeAsync(9_000);
    expect(screen.queryByText(NO_DESBLOQUEADO)).not.toBeInTheDocument();
    expect(cuenta.gets).toBe(1);
    await vi.advanceTimersByTimeAsync(1_500);
    expect(await screen.findByText(NO_DESBLOQUEADO)).toBeInTheDocument();
    expect(cuenta.posts).toBe(1);
  });

  it('si la comprobación también falla → botón «Volver a comprobar», que solo repite la consulta', async () => {
    const cuenta = escenario(caido, [caido, conViaje]);
    montarDesbloqueo();
    const u = await enviar();
    const volver = await screen.findByRole('button', { name: 'Volver a comprobar' });
    expect(screen.getByText(/No sabemos si se ha desbloqueado/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Desbloquear' })).toBeDisabled();
    await u.click(volver);
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
    expect(cuenta.posts).toBe(1);
    expect(cuenta.gets).toBe(3);
  });

  it('un 4xx es un rechazo con causa: lo muestra y no consulta el viaje', async () => {
    const cuenta = escenario(
      () =>
        HttpResponse.json(
          { error: { codigo: 'ANCLAJE_NO_EXISTE', mensaje: 'No existe ningún anclaje con ese código' } },
          { status: 404 },
        ),
      [conViaje],
    );
    montarDesbloqueo();
    await enviar();
    expect(await screen.findByRole('alert')).toHaveTextContent('No existe ningún anclaje con ese código');
    expect(screen.queryByText(COMPROBANDO)).not.toBeInTheDocument();
    expect(cuenta.gets).toBe(1);
    expect(cuenta.posts).toBe(1);
  });
});
