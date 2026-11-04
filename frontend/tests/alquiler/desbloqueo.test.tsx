// T039 — Página de desbloqueo (HU-02): estados de viaje activo y causa de rechazo.
import { render, screen } from '@solidjs/testing-library';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import Desbloqueo from '../../src/pages/alquiler/Desbloqueo';
import type { Viaje } from '../../src/services/api/alquiler';
import { server } from '../server';

const VIAJE: Viaje = {
  id: 1,
  usuario_id: 7,
  bicicleta_id: 42,
  estacion_origen_id: 1,
  estacion_destino_id: null,
  fecha_inicio: '2026-11-04T09:15:00Z',
  fecha_fin: null,
  duracion_minutos: null,
  coste_total: 0,
  estado: 'en_curso',
};

const activo = (viaje: Viaje | null) =>
  http.get('*/api/v1/alquileres/activo', () => HttpResponse.json({ viaje }));

const rechazo = (status: number, codigo: string, mensaje: string) =>
  http.post('*/api/v1/alquileres/desbloqueo', () =>
    HttpResponse.json({ error: { codigo, mensaje } }, { status }),
  );

async function teclearYEnviar(codigo: string) {
  const u = userEvent.setup();
  await u.type(await screen.findByLabelText('Código de anclaje'), codigo);
  await u.click(screen.getByRole('button', { name: 'Desbloquear' }));
}

describe('desbloqueo', () => {
  it('muestra el viaje activo en lugar del formulario', async () => {
    server.use(activo(VIAJE));
    render(() => <Desbloqueo />);
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
    expect(screen.queryByLabelText('Código de anclaje')).not.toBeInTheDocument();
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
    render(() => <Desbloqueo />);
    await teclearYEnviar('  E01-A01 ');
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
    expect(enviado).toEqual({ codigo_anclaje: 'E01-A01' });
  });

  it('muestra la causa que devuelve el backend y conserva el formulario', async () => {
    server.use(activo(null), rechazo(403, 'ABONO_VENCIDO', 'Tu abono anual no está vigente'));
    render(() => <Desbloqueo />);
    await teclearYEnviar('E01-A01');
    expect(await screen.findByRole('alert')).toHaveTextContent('Tu abono anual no está vigente');
    expect(screen.getByLabelText('Código de anclaje')).toHaveValue('E01-A01');
  });

  it('con VIAJE_ACTIVO recupera el viaje iniciado en otro dispositivo', async () => {
    let consultas = 0;
    server.use(
      http.get('*/api/v1/alquileres/activo', () =>
        HttpResponse.json({ viaje: consultas++ === 0 ? null : VIAJE }),
      ),
      rechazo(409, 'VIAJE_ACTIVO', 'Ya tienes un viaje en curso'),
    );
    render(() => <Desbloqueo />);
    await teclearYEnviar('E01-A01');
    expect(await screen.findByRole('heading', { name: 'Viaje en curso' })).toBeInTheDocument();
  });
});
