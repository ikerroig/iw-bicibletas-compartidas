import type { Component } from "solid-js";
import { Devolucion } from "./pages/devolucion";
import { MisViajes } from "./pages/viajes";

/**
 * Registro único de rutas del SPA (fichero compartido entre specs).
 * Cada spec añade SU entrada (una línea/objeto) — conflictos de merge triviales.
 */
export interface AppRoute {
  path: string;
  component: Component;
}

const Home: Component = () => <p>SI-16 — Bicicletas compartidas</p>;

export const routes: AppRoute[] = [
  { path: "/", component: Home },
  // spec-003 (HU-03): devolución y tarificación
  { path: "/devolucion", component: Devolucion },
  // spec-004 (HU-04): historial de viajes y desglose mensual
  { path: "/viajes", component: MisViajes },
];
