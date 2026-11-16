// Layout común (T022): cabecera, navegación por áreas filtrada por rol y guard
// de sesión. Cada spec añade su entrada en AREAS.
import { For, Show, type ParentProps } from 'solid-js';
import { A, Navigate } from '@solidjs/router';
import AvisoViaje from './AvisoViaje';
import { cerrarSesion, sesion, type Rol } from '../stores/sesion';

const TODOS: Rol[] = ['usuario', 'operario_taller', 'operario_logistica', 'gestor'];

const AREAS: { ruta: string; etiqueta: string; roles: Rol[] }[] = [
  { ruta: '/', etiqueta: 'Mapa', roles: TODOS }, // spec-001
  { ruta: '/alquiler', etiqueta: 'Desbloquear', roles: ['usuario'] }, // spec-002
  { ruta: '/devolucion', etiqueta: 'Devolver', roles: ['usuario'] }, // spec-003
  { ruta: '/viajes', etiqueta: 'Mis viajes', roles: ['usuario'] }, // spec-004
  { ruta: '/averias', etiqueta: 'Reportar averia', roles: ['usuario'] }, // spec-005
  { ruta: '/incidencias', etiqueta: 'Incidencias', roles: ['operario_taller'] }, // spec-005
  { ruta: '/balanceo', etiqueta: 'Balanceo', roles: ['operario_logistica'] }, // spec-007
  { ruta: '/analitica', etiqueta: 'Analítica', roles: ['gestor'] }, // spec-008
];

export default function Layout(props: ParentProps) {
  const areas = () => {
    const rol = sesion.usuario()?.rol;
    return AREAS.filter(
      (a) => a.roles === TODOS || (rol !== undefined && a.roles.includes(rol)),
    );
  };
  return (
    <Show when={sesion.autenticado()} fallback={<Navigate href="/login" />}>
      <div class="flex h-dvh flex-col">
        <header class="flex items-center gap-4 bg-emerald-800 px-4 py-2 text-white">
          <span class="font-semibold">SI-16 · Bicicletas</span>
          <nav aria-label="Áreas" class="flex gap-3 text-sm">
            <For each={areas()}>
              {(a) => (
                <A href={a.ruta} end class="hover:underline" activeClass="underline">
                  {a.etiqueta}
                </A>
              )}
            </For>
          </nav>
          <span class="ml-auto text-sm">{sesion.usuario()?.nombre}</span>
          <button type="button" class="text-sm underline" onClick={() => cerrarSesion()}>
            Cerrar sesión
          </button>
        </header>
        <AvisoViaje />
        <main class="relative min-h-0 flex-1">{props.children}</main>
      </div>
    </Show>
  );
}
