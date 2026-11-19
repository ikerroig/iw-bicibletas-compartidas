// T039 — Página de desbloqueo (HU-02): estados de viaje activo y causa de rechazo.
// T048 — FR-010 (D16): aviso del abono al abrir la pantalla, con el formulario
// desactivado; FR-002 (D15): mensajes distintos para cada causa del abono.
import { screen } from '@solidjs/testing-library';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { server } from '../server';
import { VIAJE, activo, montarDesbloqueo, rechazo } from './montar';

vi.mock('../../src/stores/sesion', () => ({
  sesion: { usuario: () => ({ id: 7, nombre: 'Eva', rol: 'usuario' }) },
}));

async function teclearYEnviar(codigo: string) {
  const u = userEvent.setup();
  await u.type(await screen.findByLabelText('Código de anclaje'), codigo);
  await u.click(screen.getByRole('button', { name: 'Desbloquear' }));
}

describe('desbloqueo', () => {
  it('muestra el viaje activo en lugar del formulario', async () => {
    server.use(activo(VIAJE));
    montarDesbloqueo();
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
    expect(screen.queryByLabelText('Código de anclaje')).not.toBeInTheDocument();
  });

  it('muestra inicio, código de bicicleta y estación sin pedir el detalle de la estación', async () => {
    let peticionesDetalle = 0;
    server.use(
      activo(VIAJE),
      http.get('*/api/v1/estaciones/:id', () => {
        peticionesDetalle++;
        return HttpResponse.json({});
      }),
    );
    montarDesbloqueo();
    expect(await screen.findByText('BICI-0007')).toBeInTheDocument();
    expect(screen.getByText('Colón')).toBeInTheDocument();
    expect(screen.getByTestId('viaje-inicio')).not.toBeEmptyDOMElement();
    expect(screen.queryByText(/n.º 42/)).not.toBeInTheDocument();
    expect(peticionesDetalle).toBe(0);
  });

  it('envía el código sin espacios y pasa al viaje en curso', async () => {
    let enviado: unknown;
    server.use(
      activo(null),
      http.post('*/api/v1/alquileres/desbloqueo', async ({ request }) => {
        enviado = await request.json();
        return HttpResponse.json(VIAJE, { status: 201 });
      }),
    );
    montarDesbloqueo();
    await teclearYEnviar('  E01-A01 ');
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
    expect(enviado).toEqual({ codigo_anclaje: 'E01-A01' });
  });

  it.each([
    ['SIN_ABONO', 'No tienes abono anual'],
    ['ABONO_VENCIDO', 'Tu abono anual caducó el 09/11/2026'],
    ['ANCLAJE_NO_EXISTE', 'No existe ningún anclaje con ese código'],
  ])('muestra la causa %s que devuelve el backend y conserva el formulario', async (codigo, mensaje) => {
    server.use(activo(null), rechazo(codigo === 'ANCLAJE_NO_EXISTE' ? 404 : 403, codigo, mensaje));
    montarDesbloqueo();
    await teclearYEnviar('E01-A01');
    expect(await screen.findByRole('alert')).toHaveTextContent(mensaje);
    expect(screen.getByLabelText('Código de anclaje')).toHaveValue('E01-A01');
  });

  it('con VIAJE_ACTIVO recupera el viaje iniciado en otro dispositivo', async () => {
    let consultas = 0;
    server.use(
      http.get('*/api/v1/alquileres/activo', () =>
        HttpResponse.json({
          viaje: consultas++ === 0 ? null : VIAJE,
          abono: { vigente: true, codigo: null, mensaje: null },
        }),
      ),
      rechazo(409, 'VIAJE_ACTIVO', 'Ya tienes un viaje en curso'),
    );
    montarDesbloqueo();
    await teclearYEnviar('E01-A01');
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
  });
});

describe('aviso del abono al abrir la pantalla (FR-010)', () => {
  it.each([
    ['SIN_ABONO', 'No tienes abono anual'],
    ['ABONO_VENCIDO', 'Tu abono anual caducó el 09/11/2026'],
  ] as const)('con %s avisa antes de escribir y desactiva el formulario', async (codigo, mensaje) => {
    let posts = 0;
    server.use(
      activo(null, { vigente: false, codigo, mensaje }),
      http.post('*/api/v1/alquileres/desbloqueo', () => {
        posts++;
        return HttpResponse.json({}, { status: 500 });
      }),
    );
    montarDesbloqueo();
    expect(await screen.findByRole('alert')).toHaveTextContent(mensaje);
    expect(screen.getByLabelText('Código de anclaje')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Desbloquear' })).toBeDisabled();
    await userEvent.setup().click(screen.getByRole('button', { name: 'Desbloquear' }));
    expect(posts).toBe(0);
  });

  it('con el abono vigente el formulario está activo y sin aviso', async () => {
    server.use(activo(null));
    montarDesbloqueo();
    expect(await screen.findByLabelText('Código de anclaje')).toBeEnabled();
    expect(screen.getByRole('button', { name: 'Desbloquear' })).toBeEnabled();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
