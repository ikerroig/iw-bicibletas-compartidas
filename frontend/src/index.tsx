/* @refresh reload */
import { render } from "solid-js/web";
import { Router, Route, A } from "@solidjs/router";
import "./index.css";
import { routes } from "./routes";

const root = document.getElementById("root");
if (!root) throw new Error("No existe #root");

render(
  () => (
    <Router>
      <nav class="flex gap-4 p-4 border-b">
        <A href="/">Inicio</A>
        {/* spec-003 (T022): HU-03 */}
        <A href="/devolucion">Devolver</A>
      </nav>
      <main class="p-4">
        {routes.map((r) => (
          <Route path={r.path} component={r.component} />
        ))}
      </main>
    </Router>
  ),
  root,
);
