/**
 * Cliente API tipado del SI-16 (spec-003: API base compartida + su endpoint).
 * Maneja el envelope de error del contrato { error: { codigo, mensaje } }.
 */

export interface ApiError {
  codigo: string;
  mensaje: string;
}

export class FetchError extends Error {
  readonly codigo: string;
  readonly status: number;

  constructor(codigo: string, mensaje: string, status: number) {
    super(mensaje);
    this.name = "FetchError";
    this.codigo = codigo;
    this.status = status;
  }
}

function token(): string | null {
  return localStorage.getItem("token");
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string> | undefined),
  };
  const t = token();
  if (t) headers.Authorization = `Bearer ${t}`;

  const res = await fetch(`/api/v1${path}`, { ...init, headers });

  if (!res.ok) {
    let codigo = "ERROR_DESCONOCIDO";
    let mensaje = `Error ${res.status}`;
    try {
      const body = (await res.json()) as { error?: ApiError };
      if (body.error) {
        codigo = body.error.codigo;
        mensaje = body.error.mensaje;
      }
    } catch {
      /* respuesta sin JSON: conserva el mensaje genérico */
    }
    throw new FetchError(codigo, mensaje, res.status);
  }
  return (await res.json()) as T;
}

export const api = { request };
