/**
 * Tests de componente de la HU-08 (T008): página de analítica.
 * Stub global de `fetch` con `vi.stubGlobal` (patrón de pages/viajes, NO
 * MSW). El wrapper del gráfico (GraficoBarras → chart.js, pendiente T020)
 * se MOCKEA para no requerir la dependencia en Vitest.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@solidjs/testing-library";
import { PanelAnalitica } from "./index";

vi.mock("./GraficoBarras", () => ({
  GraficoBarras: (props: {
    titulo: string;
    etiquetas: string[];
    datos: number[];
    testid: string;
  }) => {
    // Stub DOM plano: evita depender de chart.js (T020 bloqueada) y del
    // runtime de JSX del factory (vi.mock es hoisted).
    const div = document.createElement("div");
    div.setAttribute("data-testid", props.testid);
    div.textContent = `${props.titulo}:${props.etiquetas.length}:${props.datos.length}:${props.datos.join("|")}`;
    return div;
  },
}));

const analiticaConDatos = {
  desde: "2026-10-01",
  hasta: "2026-10-31",
  total_viajes: 3,
  total_coste: 6.5,
  por_hora: Array.from({ length: 24 }, (_, i) => (i === 9 ? 2 : i === 17 ? 1 : 0)),
  por_dia_semana: [1, 0, 0, 1, 0, 1, 0],
};

const analiticaVacia = {
  desde: "2099-01-01",
  hasta: "2099-01-31",
  total_viajes: 0,
  total_coste: 0,
  por_hora: Array(24).fill(0),
  por_dia_semana: Array(7).fill(0),
};

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function csv(): Response {
  return new Response("﻿id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur\r\n1,2026-10-05 08:15,A,B,12,0.00\r\n", {
    status: 200,
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition": "attachment; filename=viajes_20261001_20261031.csv",
    },
  });
}

let peticiones: { url: string; init: RequestInit | undefined }[] = [];
let responderViajes: () => Promise<Response>;
let responderExport: () => Promise<Response>;
let exportPendiente: Promise<Response> | null;

beforeEach(() => {
  peticiones = [];
  responderViajes = async () => json(analiticaConDatos);
  responderExport = async () => csv();
  exportPendiente = null;
  // Es la misma clave que usa el store de sesión real.
  localStorage.setItem("si16.token", "tok-test");
  // jsdom no implementa object URLs (descarga vía Blob + <a download>).
  URL.createObjectURL = () => "blob:mock";
  URL.revokeObjectURL = () => {};
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      peticiones.push({ url, init });
      if (url.includes("/analitica/export")) {
        if (exportPendiente) return exportPendiente;
        return responderExport();
      }
      return responderViajes();
    }),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
  localStorage.clear();
  cleanup();
});

describe("HU-08 · página de analítica", () => {
  it("render con datos: dos gráficos con 24/7 etiquetas, datos y totales", async () => {
    render(() => <PanelAnalitica />);

    const gHora = await screen.findByTestId("grafico-hora");
    expect(gHora.textContent).toContain("Viajes por hora del día");
    expect(gHora.textContent).toContain(":24:24:"); // 24 etiquetas, 24 datos
    expect(gHora.textContent).toContain("|2|"); // bucket hora 9 = 2

    const gDia = await screen.findByTestId("grafico-dia");
    expect(gDia.textContent).toContain("Viajes por día de la semana");
    expect(gDia.textContent).toContain(":7:7:");

    expect(screen.getByTestId("total-viajes").textContent).toContain("3");
    expect(screen.getByTestId("total-coste").textContent).toContain("6,50");
    expect(screen.queryByTestId("sin-datos")).toBeNull();
  });

  it("total_viajes = 0 → aviso 'sin datos en el periodo'", async () => {
    responderViajes = async () => json(analiticaVacia);
    render(() => <PanelAnalitica />);

    const aviso = await screen.findByTestId("sin-datos");
    expect(aviso.textContent).toContain("sin datos en el periodo");
  });

  it("click en Descargar CSV → fetch autenticado a /analitica/export y descarga Blob", async () => {
    render(() => <PanelAnalitica />);
    await screen.findByTestId("grafico-hora");

    const clicks = vi.fn();
    const originalCreate = document.createElement.bind(document);
    vi.spyOn(document, "createElement").mockImplementation((tag: string) => {
      const el = originalCreate(tag);
      if (tag === "a") el.addEventListener("click", clicks);
      return el;
    });

    fireEvent.click(screen.getByTestId("boton-descargar"));

    await vi.waitFor(() => {
      const exp = peticiones.find((p) => p.url.includes("/analitica/export"));
      expect(exp).toBeTruthy();
      const auth = (exp!.init?.headers as Record<string, string>)?.Authorization;
      expect(auth).toBe("Bearer tok-test");
      expect(exp!.url).toContain("desde=");
      expect(exp!.url).toContain("hasta=");
    });
    await vi.waitFor(() => expect(clicks).toHaveBeenCalledTimes(1));
    vi.restoreAllMocks();
  });

  it("doble clic en Descargar CSV → un único archivo consistente (botón deshabilitado)", async () => {
    let resolverExport: (r: Response) => void = () => {};
    exportPendiente = new Promise<Response>((res) => {
      resolverExport = res;
    });
    render(() => <PanelAnalitica />);
    await screen.findByTestId("grafico-hora");

    const boton = screen.getByTestId("boton-descargar") as HTMLButtonElement;
    fireEvent.click(boton);
    // Durante la descarga el botón queda deshabilitado.
    await vi.waitFor(() => expect(boton.disabled).toBe(true));
    fireEvent.click(boton);
    fireEvent.click(boton);
    resolverExport(csv());
    await vi.waitFor(() => expect(boton.disabled).toBe(false));

    const exportCalls = peticiones.filter((p) => p.url.includes("/analitica/export"));
    expect(exportCalls).toHaveLength(1);
  });

  it("respuesta 403 → mensaje de permiso con role='alert'", async () => {
    responderViajes = async () =>
      json({ error: { codigo: "ROL_NO_PERMITIDO", mensaje: "Tu rol no tiene permiso" } }, 403);
    render(() => <PanelAnalitica />);

    const err = await screen.findByTestId("error-analitica");
    expect(err.getAttribute("role")).toBe("alert");
    expect(err.textContent).toContain("permiso");
  });

  it("respuesta 422 → mensaje de rango inválido", async () => {
    responderViajes = async () =>
      json({ error: { codigo: "ENTRADA_INVALIDA", mensaje: "desde/hasta deben tener formato YYYY-MM-DD" } }, 422);
    render(() => <PanelAnalitica />);

    const err = await screen.findByTestId("error-analitica");
    expect(err.getAttribute("role")).toBe("alert");
    expect(err.textContent!.toLowerCase()).toContain("rango");
  });
});
