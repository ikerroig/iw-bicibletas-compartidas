/**
 * Tests de componente de la HU-03 (T023).
 *
 * NB: el registro npm disponible solo ofrece msw@3.0.3, cuyos subpaths de
 * Node (`"browser": null`) son irresolubles para el pipeline jsdom de
 * Vitest 5 + Vite 6 (el resolutor aplica la condición `browser` y aborta
 * antes de poder externalizar). Se sustituye MSW por un stub global de
 * `fetch` con la misma cobertura: éxito 200 y conflictos 409 del contrato.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@solidjs/testing-library";
import { Devolucion } from "./index";
import { fijarViajeEnCurso, viajeEnCurso } from "../../stores/viaje";

const viajeGratuito = {
  id: 7,
  usuario_id: 1,
  bicicleta_id: 1,
  estacion_origen_id: 1,
  estacion_destino_id: 2,
  fecha_inicio: "2026-10-27T10:00:00Z",
  fecha_fin: "2026-10-27T10:25:00Z",
  duracion_minutos: 25,
  coste_total: 0,
  estado: "finalizado",
  etiqueta: "Tramo gratuito (0,00 €)",
};

function jsonResponse(body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

let ultimaPeticion: { url: string; body: string } | null = null;
let responder: () => Response = () => jsonResponse(viajeGratuito, 200);

beforeEach(() => {
  ultimaPeticion = null;
  responder = () => jsonResponse(viajeGratuito, 200);
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      ultimaPeticion = {
        url: String(input),
        body: typeof init?.body === "string" ? init.body : "",
      };
      return responder();
    }),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
  cleanup();
});

async function enviarDevolucion(codigo: string) {
  fireEvent.input(screen.getByTestId("input-anclaje"), {
    target: { value: codigo },
  });
  fireEvent.click(screen.getByTestId("boton-devolver"));
}

describe("HU-03 · página de devolución", () => {
  it("viaje ≤ 30 min: pinta la etiqueta de tramo gratuito y el coste 0,00 €", async () => {
    render(() => <Devolucion />);
    await enviarDevolucion("AN-00002");

    const exito = await screen.findByTestId("exito-devolucion");
    expect(exito).toBeTruthy();
    expect(screen.getByTestId("duracion").textContent).toContain("25 min");
    expect(screen.getByTestId("coste").textContent).toContain("0.00 €");
    expect(screen.getByTestId("etiqueta").textContent).toBe(
      "Tramo gratuito (0,00 €)",
    );
    expect(screen.queryByTestId("error-devolucion")).toBeNull();
    // Payload del contrato (D6).
    expect(ultimaPeticion).not.toBeNull();
    expect(ultimaPeticion!.url).toBe("/api/v1/devoluciones");
    expect(JSON.parse(ultimaPeticion!.body)).toEqual({
      codigo_anclaje: "AN-00002",
    });
  });

  it("409 SIN_VIAJE_ACTIVO: muestra el mensaje y no cambia a la vista de éxito", async () => {
    responder = () =>
      jsonResponse(
        {
          error: {
            codigo: "SIN_VIAJE_ACTIVO",
            mensaje: "No tienes ningún viaje en curso",
          },
        },
        409,
      );
    render(() => <Devolucion />);
    await enviarDevolucion("AN-00002");

    const error = await screen.findByTestId("error-devolucion");
    expect(error.textContent).toContain("No tienes ningún viaje en curso");
    expect(screen.queryByTestId("exito-devolucion")).toBeNull();
    // El formulario sigue disponible (misma vista).
    expect(screen.getByTestId("input-anclaje")).toBeTruthy();
  });

  it("409 ANCLAJE_OCUPADO: muestra el mensaje del contrato", async () => {
    responder = () =>
      jsonResponse(
        {
          error: {
            codigo: "ANCLAJE_OCUPADO",
            mensaje: "El anclaje está ocupado, elige otro",
          },
        },
        409,
      );
    render(() => <Devolucion />);
    await enviarDevolucion("AN-00003");

    const error = await screen.findByTestId("error-devolucion");
    expect(error.textContent).toContain("El anclaje está ocupado, elige otro");
    expect(screen.queryByTestId("exito-devolucion")).toBeNull();
  });
});

describe("HU-03 · integración con el aviso de viaje en curso (HU-02)", () => {
  it("al devolver, el aviso de viaje en curso desaparece", async () => {
    fijarViajeEnCurso({
      ...viajeGratuito,
      estacion_destino_id: null,
      fecha_fin: null,
      duracion_minutos: null,
      estado: "en_curso",
      bicicleta_codigo: "BICI-0001",
      estacion_origen_nombre: "Colón",
    });
    render(() => <Devolucion />);
    await enviarDevolucion("AN-00002");

    await screen.findByTestId("exito-devolucion");
    expect(viajeEnCurso()).toBeNull();
  });
});
