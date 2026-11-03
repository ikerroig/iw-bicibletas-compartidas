// Layout común (T022): cabecera, navegación por áreas filtrada por rol y guard
// de sesión. Cada spec añade su entrada en AREAS.
import { For, Show, type ParentProps } from 'solid-js';
import { A, Navigate } from '@solidjs/router';
import { cerrarSesion, sesion, type Rol } from '../stores/sesion';

const TODOS: Rol[] = ['usuario', 'operario_taller', 'operario_logistica', 'gestor'];

const AREAS: { ruta: string; etiqueta: string; roles: Rol[] }[] = [
  { ruta: '/', etiqueta: 'Mapa', roles: TODOS }, // spec-001
];

export default function Layout(props: ParentProps) {
  const areas = () => AREAS.filter((a) => a.roles.includes(sesion.usuario()?.rol ?? 'usuario'));
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
        <main class="relative min-h-0 flex-1">{props.children}</main>
      </div>
    </Show>
  );
}
