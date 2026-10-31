/**
 * Tests de componente de la HU-07 (T008): página "Balanceo".
 * Stub global de `fetch` con `vi.stubGlobal` (patrón de pages/viajes, NO MSW):
 * orden del backend, etiquetas de umbral, porcentaje es-ES, fecha Europe/Madrid,
 * estado vacío, error de red y envío del traslado (POST + tarjetas + recarga).
 * TDD: debe FALLAR hasta implementar pages/balanceo/index.tsx (T016).
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@solidjs/testing-library";
import { Balanceo } from "./index";

const alertaVacía = {
  id: 3,
  estacion_id: 3,
  estacion_nombre: "Estació Nord",
  tipo_alerta: "vacia_menos_10",
  porcentaje_ocupacion: 5,
  severidad: "alta",
  fecha_generacion: "2026-10-30T09:12:44Z", // 10:12 Madrid (CET)
  resuelta: false,
};

const alertaLlena = {
  id: 1,
  estacion_id: 4,
  estacion_nombre: "Estació Port",
  tipo_alerta: "llena_mas_90",
  porcentaje_ocupacion: 95,
  severidad: "alta",
  fecha_generacion: "2026-10-30T08:58:01Z", // 09:58 Madrid
  resuelta: false,
};

const estadoOrigen = {
  id: 4,
  nombre: "Estació Port",
  latitud: 39.4812,
  longitud: -0.3711,
  capacidad_total: 20,
  estado: "operativa",
  bicis_disponibles: 16,
  anclajes_libres: 4,
  ocupacion_pct: 80,
  actualizado_en: "2026-10-30T10:15:32Z",
};

const estadoDestino = {
  id: 3,
  nombre: "Estació Nord",
  latitud: 39.477,
  longitud: -0.3801,
  capacidad_total: 20,
  estado: "operativa",
  bicis_disponibles: 5,
  anclajes_libres: 15,
  ocupacion_pct: 25,
  actualizado_en: "2026-10-30T10:15:32Z",
};

function respuesta(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

interface Peticion {
  url: string;
  metodo: string;
  cuerpo: string;
}

let peticiones: Peticion[] = [];
let responder: (p: Peticion) => Promise<Response>;

beforeEach(() => {
  peticiones = [];
  responder = async (p) =>
    p.url.includes("/balanceo/alertas")
      ? respuesta({ items: [alertaVacía, alertaLlena] })
      : respuesta({ origen: estadoOrigen, destino: estadoDestino });
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const p: Peticion = {
        url: String(input),
        metodo: init?.method ?? "GET",
        cuerpo: typeof init?.body === "string" ? init.body : "",
      };
      peticiones.push(p);
      return responder(p);
    }),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
  cleanup();
});

describe("HU-07 · página Balanceo", () => {
  it("render con datos: orden del backend, etiquetas de umbral, porcentaje es-ES y fecha Madrid", async () => {
    render(() => <Balanceo />);

    const filas = await screen.findAllByTestId("fila-alerta");
    expect(filas).toHaveLength(2);

    // El orden es el del backend (const. I): la vacía llega primero.
    expect(screen.getAllByTestId("fila-nombre")[0].textContent).toContain("Estació Nord");
    expect(screen.getAllByTestId("fila-nombre")[1].textContent).toContain("Estació Port");

    expect(screen.getAllByTestId("fila-tipo")[0].textContent).toBe("Vacía (< 10 %)");
    expect(screen.getAllByTestId("fila-tipo")[1].textContent).toBe("Llena (> 90 %)");

    expect(screen.getAllByTestId("fila-porcentaje")[0].textContent).toContain("5,00");
    expect(screen.getAllByTestId("fila-porcentaje")[1].textContent).toContain("95,00");

    expect(screen.getAllByTestId("fila-fecha")[0].textContent).toBe("30/10/2026 10:12");
    expect(screen.getAllByTestId("fila-fecha")[1].textContent).toBe("30/10/2026 09:58");
  });

  it("red equilibrada: estado vacío 'Sin estaciones descompensadas'", async () => {
    responder = async () => respuesta({ items: [] });
    render(() => <Balanceo />);

    const vacio = await screen.findByTestId("estado-vacio");
    expect(vacio.textContent).toContain("Sin estaciones descompensadas");
    expect(screen.queryByTestId("fila-alerta")).toBeNull();
  });

  it("error de red: mensaje con role='alert'", async () => {
    responder = async () => {
      throw new TypeError("NetworkError");
    };
    render(() => <Balanceo />);

    const error = await screen.findByTestId("error-panel");
    expect(error.getAttribute("role")).toBe("alert");
    expect(screen.queryByTestId("fila-alerta")).toBeNull();
  });

  it("traslado: POST con los tres campos, tarjetas de origen/destino y recarga del panel", async () => {
    let postHecho = false;
    responder = async (p) => {
      if (p.metodo === "POST" && p.url.includes("/balanceo/traslados")) {
        postHecho = true;
        return respuesta({ origen: estadoOrigen, destino: estadoDestino });
      }
      // GET del panel: tras el 200 la recarga ya no debe traer la alerta resuelta.
      return postHecho
        ? respuesta({ items: [alertaLlena] })
        : respuesta({ items: [alertaVacía, alertaLlena] });
    };
    render(() => <Balanceo />);
    await screen.findAllByTestId("fila-alerta");

    fireEvent.change(screen.getByTestId("input-origen"), { target: { value: "4" } });
    fireEvent.change(screen.getByTestId("input-destino"), { target: { value: "3" } });
    fireEvent.change(screen.getByTestId("input-unidades"), { target: { value: "3" } });
    fireEvent.click(screen.getByTestId("boton-traslado"));

    const resultado = await screen.findByTestId("resultado-traslado");
    expect(resultado.textContent).toContain("Estació Port");
    expect(resultado.textContent).toContain("Estació Nord");

    const post = peticiones.find((p) => p.metodo === "POST");
    expect(post).toBeDefined();
    expect(post!.url).toContain("/api/v1/balanceo/traslados");
    expect(JSON.parse(post!.cuerpo)).toEqual({
      estacion_origen_id: 4,
      estacion_destino_id: 3,
      unidades: 3,
    });

    // Tarjetas EstacionEstado de la respuesta.
    expect(screen.getByTestId("tarjeta-origen").textContent).toContain("16");
    expect(screen.getByTestId("tarjeta-destino").textContent).toContain("5");

    // El panel se recarga tras el 200 (2ª petición GET al panel).
    const getsPanel = peticiones.filter((p) => p.metodo === "GET" && p.url.includes("/balanceo/alertas"));
    expect(getsPanel.length).toBeGreaterThanOrEqual(2);
    const filas = await screen.findAllByTestId("fila-alerta");
    expect(filas).toHaveLength(1); // solo queda la alerta no resuelta tras la recarga
  });
});
