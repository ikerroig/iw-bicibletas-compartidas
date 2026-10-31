/* @refresh reload */
import { render } from "solid-js/web";
import { Router, Route, A } from "@solidjs/router";
import type { RouteSectionProps } from "@solidjs/router";
import "./index.css";
import { routes } from "./routes";

const root = document.getElementById("root");
if (!root) throw new Error("No existe #root");

// Layout raíz (prop `root`): en @solidjs/router 1.0 los <A> solo funcionan
// dentro del contexto de rutas; la ruta activa llega como props.children
// (no existe <Outlet /> en esta versión).
function Layout(props: RouteSectionProps) {
  return (
    <>
      <nav class="flex gap-4 p-4 border-b">
        <A href="/">Inicio</A>
        {/* spec-003 (T022): HU-03 */}
        <A href="/devolucion">Devolver</A>
        {/* spec-004 (T012): HU-04 */}
        <A href="/viajes">Mis viajes</A>
        {/* spec-007 (T017): HU-07 */}
        <A href="/balanceo">Balanceo</A>
      </nav>
      <main class="p-4">{props.children}</main>
    </>
  );
}

render(
  () => (
    <Router root={Layout}>
      {routes.map((r) => (
        <Route path={r.path} component={r.component} />
      ))}
    </Router>
  ),
  root,
);
