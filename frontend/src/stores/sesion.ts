// Sesión del usuario (T021): token JWT y datos del usuario, persistidos en
// localStorage como comodidad (si el almacenamiento falla, la sesión vive solo
// en memoria).
import { createSignal } from 'solid-js';
import { apiFetch, configurarApi } from '../services/api';

export type Rol = 'usuario' | 'operario_taller' | 'operario_logistica' | 'gestor';

export interface Usuario {
  id: number;
  nombre: string;
  email: string;
  rol: Rol;
  abono_activo: boolean;
  fecha_vencimiento_abono: string | null;
}

const CLAVE_TOKEN = 'si16.token';
const CLAVE_USUARIO = 'si16.usuario';

function leer(clave: string): string | null {
  try {
    return localStorage.getItem(clave);
  } catch {
    return null;
  }
}

function escribir(clave: string, valor: string | null): void {
  try {
    if (valor === null) localStorage.removeItem(clave);
    else localStorage.setItem(clave, valor);
  } catch {
    /* almacenamiento no disponible */
  }
}

function leerUsuario(): Usuario | null {
  const crudo = leer(CLAVE_USUARIO);
  if (!crudo) return null;
  try {
    return JSON.parse(crudo) as Usuario;
  } catch {
    return null;
  }
}

const [token, setToken] = createSignal<string | null>(leer(CLAVE_TOKEN));
const [usuario, setUsuario] = createSignal<Usuario | null>(leerUsuario());

export const sesion = {
  token,
  usuario,
  autenticado: () => token() !== null,
};

export async function iniciarSesion(email: string, password: string): Promise<Usuario> {
  const r = await apiFetch<{ token: string; usuario: Usuario }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  setToken(r.token);
  setUsuario(r.usuario);
  escribir(CLAVE_TOKEN, r.token);
  escribir(CLAVE_USUARIO, JSON.stringify(r.usuario));
  return r.usuario;
}

export function cerrarSesion(): void {
  setToken(null);
  setUsuario(null);
  escribir(CLAVE_TOKEN, null);
  escribir(CLAVE_USUARIO, null);
}

configurarApi({ token, alNoAutorizado: cerrarSesion });
