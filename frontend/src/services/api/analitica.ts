/**
 * Cliente del endpoint HU-08: GET /api/v1/analitica/viajes y
 * GET /api/v1/analitica/export.
 * Contrato: specs/spec-008/contracts/analitica.md — schema global
 * AnaliticaViajes; CSV con BOM UTF-8 descargado vía Blob autenticado
 * (el JWT NO viaja en la URL — research D6).
 */
import { api, FetchError } from "../api";

export interface AnaliticaViajes {
  desde: string;
  hasta: string;
  total_viajes: number;
  total_coste: number;
  por_hora: number[]; // 24 posiciones (00–23)
  por_dia_semana: number[]; // 7 posiciones (lunes=0 … domingo=6)
}

/** Agregados del panel para el rango inclusivo dado (YYYY-MM-DD). */
export function consultarAnalitica(
  desde: string,
  hasta: string,
): Promise<AnaliticaViajes> {
  const q = new URLSearchParams({ desde, hasta });
  return api.request<AnaliticaViajes>(`/analitica/viajes?${q.toString()}`);
}

/** Descarga el CSV del rango: fetch autenticado + Blob + <a download>. */
export async function descargarCSV(
  desde: string,
  hasta: string,
): Promise<void> {
  const q = new URLSearchParams({ desde, hasta });
  const headers: Record<string, string> = {};
  const t = localStorage.getItem("token");
  if (t) headers.Authorization = `Bearer ${t}`;

  const res = await fetch(`/api/v1/analitica/export?${q.toString()}`, {
    headers,
  });
  if (!res.ok) {
    let codigo = "ERROR_DESCONOCIDO";
    let mensaje = `Error ${res.status}`;
    try {
      const body = (await res.json()) as { error?: { codigo: string; mensaje: string } };
      if (body.error) {
        codigo = body.error.codigo;
        mensaje = body.error.mensaje;
      }
    } catch {
      /* sin JSON: conserva el genérico */
    }
    throw new FetchError(codigo, mensaje, res.status);
  }

  const blob = await res.blob();
  const cd = res.headers.get("Content-Disposition") ?? "";
  const m = cd.match(/filename=([^;]+)/);
  const nombre =
    m?.[1]?.trim() ?? `viajes_${desde.replaceAll("-", "")}_${hasta.replaceAll("-", "")}.csv`;

  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = nombre;
  a.click();
  URL.revokeObjectURL(url);
}
