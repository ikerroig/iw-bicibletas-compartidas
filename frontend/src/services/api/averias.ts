import { apiFetch } from '../api';

export type TipoFallo = 'frenos' | 'cadena' | 'neumatico' | 'otro';

export interface Incidencia {
  id: number;
  bicicleta_id: number;
  bicicleta_codigo: string;
  estacion_id: number;
  estacion_nombre: string;
  usuario_reporta_id: number | null;
  tipo_fallo: TipoFallo;
  descripcion_reporte: string;
  fecha_reporte: string;
  estado: 'pendiente' | 'en_reparacion' | 'resuelta';
}

export interface ReporteAveria {
  bicicleta_codigo: string;
  estacion_id: number;
  tipo_fallo: TipoFallo;
  descripcion: string;
}

export interface RespuestaReporte {
  incidencia: Incidencia;
  mensaje: string;
  repetido: boolean;
}

export function reportarAveria(reporte: ReporteAveria): Promise<RespuestaReporte> {
  return apiFetch<RespuestaReporte>('/averias', {
    method: 'POST',
    body: JSON.stringify(reporte),
  });
}

export async function listarIncidencias(): Promise<Incidencia[]> {
  const respuesta = await apiFetch<{ items: Incidencia[] }>('/incidencias');
  return respuesta.items;
}
