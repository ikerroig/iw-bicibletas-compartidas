// Cliente HTTP común (T021): base /api/v1, cabecera Bearer y envelope de
// error {"error":{"codigo","mensaje"}} de contracts/README.md.
//
// Unión de las dos mitades del equipo:
//  * apiFetch / ApiError / configurarApi — specs-001/002 (URL absoluta).
//  * api.request / FetchError — specs-003/004/007/008 (ruta relativa).

export const API_BASE = '/api/v1';

/** Lee el token: la sesión (spec-001/002) usa 'si16.token'; specs-00X 'token'. */
function tokenLocal(): string | null {
  return localStorage.getItem('si16.token') ?? localStorage.getItem('token');
}

let obtenerToken: () => string | null = tokenLocal;
let alNoAutorizado: () => void = () => {};

/** Lo llama stores/sesion.ts para inyectar el token sin dependencia circular. */
export function configurarApi(opciones: {
  token: () => string | null;
  alNoAutorizado: () => void;
}): void {
  obtenerToken = opciones.token;
  alNoAutorizado = opciones.alNoAutorizado;
}

/** URL absoluta en el origen actual (en dev, Vite reenvía /api al backend). */
export function urlApi(ruta: string): string {
  return new URL(API_BASE + ruta, window.location.origin).toString();
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly codigo: string,
    mensaje: string,
  ) {
    super(mensaje);
    this.name = 'ApiError';
  }
}

export class FetchError extends Error {
  readonly codigo: string;
  readonly status: number;

  constructor(codigo: string, mensaje: string, status: number) {
    super(mensaje);
    this.name = 'FetchError';
    this.codigo = codigo;
    this.status = status;
  }
}

/** Extrae {codigo, mensaje} del envelope de error; respuestas no JSON → genérico. */
async function leerError(res: Response): Promise<{ codigo: string; mensaje: string }> {
  let codigo = 'ERROR';
  let mensaje = `Error ${res.status}`;
  try {
    const cuerpo = (await res.json()) as { error?: { codigo?: string; mensaje?: string } };
    codigo = cuerpo.error?.codigo ?? codigo;
    mensaje = cuerpo.error?.mensaje ?? mensaje;
  } catch {
    /* cuerpo no JSON */
  }
  return { codigo, mensaje };
}

export async function apiFetch<T>(ruta: string, init: RequestInit = {}): Promise<T> {
  const cabeceras = new Headers(init.headers);
  cabeceras.set('Accept', 'application/json');
  if (init.body !== undefined && !cabeceras.has('Content-Type')) {
    cabeceras.set('Content-Type', 'application/json');
  }
  const t = obtenerToken();
  if (t) cabeceras.set('Authorization', `Bearer ${t}`);

  let res: Response;
  try {
    res = await fetch(urlApi(ruta), { ...init, headers: cabeceras });
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw e;
    throw new ApiError(0, 'SIN_CONEXION', 'No se pudo conectar con el servidor');
  }

  if (!res.ok) {
    const { codigo, mensaje } = await leerError(res);
    if (res.status === 401 && t) alNoAutorizado();
    throw new ApiError(res.status, codigo, mensaje);
  }
  return (await res.json()) as T;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string> | undefined),
  };
  const t = obtenerToken();
  if (t) headers.Authorization = `Bearer ${t}`;

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  if (!res.ok) {
    const { codigo, mensaje } = await leerError(res);
    throw new FetchError(codigo, mensaje, res.status);
  }
  return (await res.json()) as T;
}

export const api = { request };
