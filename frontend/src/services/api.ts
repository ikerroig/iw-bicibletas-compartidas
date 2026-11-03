// Cliente HTTP común (T021): base /api/v1, cabecera Bearer y envelope de error
// {"error":{"codigo","mensaje"}} de contracts/README.md.

export const API_BASE = '/api/v1';

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

let obtenerToken: () => string | null = () => null;
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

export async function apiFetch<T>(ruta: string, init: RequestInit = {}): Promise<T> {
  const cabeceras = new Headers(init.headers);
  cabeceras.set('Accept', 'application/json');
  if (init.body !== undefined && !cabeceras.has('Content-Type')) {
    cabeceras.set('Content-Type', 'application/json');
  }
  const token = obtenerToken();
  if (token) cabeceras.set('Authorization', `Bearer ${token}`);

  let res: Response;
  try {
    res = await fetch(urlApi(ruta), { ...init, headers: cabeceras });
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw e;
    throw new ApiError(0, 'SIN_CONEXION', 'No se pudo conectar con el servidor');
  }

  if (!res.ok) {
    let codigo = 'ERROR';
    let mensaje = `Error ${res.status}`;
    try {
      const cuerpo = (await res.json()) as { error?: { codigo?: string; mensaje?: string } };
      codigo = cuerpo.error?.codigo ?? codigo;
      mensaje = cuerpo.error?.mensaje ?? mensaje;
    } catch {
      /* cuerpo no JSON */
    }
    if (res.status === 401 && token) alNoAutorizado();
    throw new ApiError(res.status, codigo, mensaje);
  }
  return (await res.json()) as T;
}
