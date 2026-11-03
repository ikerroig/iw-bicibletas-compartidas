// Handlers MSW derivados de specs/spec-001/contracts/openapi.yaml (T023).
// Los tests de cada spec pueden sustituirlos con server.use(...).
import { http, HttpResponse } from 'msw';
import type { EstacionEstado, SnapshotEstaciones } from '../src/services/api/estaciones';

const AHORA = '2026-11-03T10:00:00Z';

function estacion(
  p: Partial<EstacionEstado> & Pick<EstacionEstado, 'id' | 'nombre'>,
): EstacionEstado {
  return {
    latitud: 39.47 + p.id / 1000,
    longitud: -0.37,
    capacidad_total: 10,
    estado: 'operativa',
    bicis_disponibles: 5,
    anclajes_libres: 5,
    ocupacion_pct: 50,
    actualizado_en: AHORA,
    ...p,
  };
}

/** Una estación por cada estado visual de R18. */
export const ESTACIONES: EstacionEstado[] = [
  estacion({ id: 1, nombre: 'Normal', bicis_disponibles: 7, anclajes_libres: 3 }),
  estacion({ id: 2, nombre: 'Sin bicis', bicis_disponibles: 0, anclajes_libres: 9 }),
  estacion({ id: 3, nombre: 'Llena', bicis_disponibles: 9, anclajes_libres: 0 }),
  estacion({ id: 4, nombre: 'Averiadas', bicis_disponibles: 0, anclajes_libres: 0 }),
  estacion({
    id: 5,
    nombre: 'Cerrada',
    estado: 'fuera_de_servicio',
    bicis_disponibles: 0,
    anclajes_libres: 4,
    ocupacion_pct: 0,
  }),
];

export function snapshot(estaciones: EstacionEstado[] = ESTACIONES): SnapshotEstaciones {
  return { actualizado_en: AHORA, estaciones };
}

export const handlers = [
  http.post('*/api/v1/auth/login', () =>
    HttpResponse.json({
      token: 'token-de-prueba',
      usuario: {
        id: 1,
        nombre: 'Ana García',
        email: 'ana@si16.test',
        rol: 'usuario',
        abono_activo: true,
        fecha_vencimiento_abono: '2027-06-01',
      },
    }),
  ),
  http.get('*/api/v1/estaciones', () => HttpResponse.json(snapshot())),
  http.get('*/api/v1/estaciones/:id', ({ params }) => {
    const e = ESTACIONES.find((x) => x.id === Number(params.id));
    return e
      ? HttpResponse.json({ ...e, anclajes: [] })
      : HttpResponse.json(
          { error: { codigo: 'NO_ENCONTRADO', mensaje: 'Estación no encontrada' } },
          { status: 404 },
        );
  }),
];
