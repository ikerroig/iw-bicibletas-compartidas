// Cliente de la API de HU-06 para el operario de taller.
import { apiFetch } from '../api';

export interface IncidenciaTaller {
  id: number;
  bicicleta_id: number;
  usuario_reporta_id: number | null;
  operario_id: number | null;
  tipo_fallo: string;
  descripcion_reporte: string;
  piezas_sustituidas: string | null;
  notas_taller: string | null;
  fecha_reporte: string;
  fecha_reparacion: string | null;
  estado: string;
}

export interface FichaBicicleta {
  bicicleta: {
    id: number;
    codigo: string;
    estado: string;
  };
  incidencias: IncidenciaTaller[];
}

export interface IntervencionRequest {
  piezas_sustituidas: string | null;
  notas_taller: string | null;
  marcar_operativa: boolean;
}

export interface EstadoEstacion {
  id: number;
  nombre: string;
  bicis_disponibles: number;
  anclajes_libres: number;
  actualizado_en: string;
}

export function obtenerFicha(bicicletaID: number): Promise<FichaBicicleta> {
  return apiFetch<FichaBicicleta>('/taller/bicicletas/' + bicicletaID);
}

export function registrarIntervencion(
  bicicletaID: number,
  request: IntervencionRequest,
): Promise<IncidenciaTaller> {
  return apiFetch<IncidenciaTaller>('/taller/bicicletas/' + bicicletaID + '/intervenciones', {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

export function confirmarUbicacion(
  bicicletaID: number,
  codigoAnclaje: string,
): Promise<EstadoEstacion> {
  return apiFetch<EstadoEstacion>('/taller/bicicletas/' + bicicletaID + '/ubicacion', {
    method: 'POST',
    body: JSON.stringify({ codigo_anclaje: codigoAnclaje }),
  });
}
