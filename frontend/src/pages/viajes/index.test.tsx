/**
 * Tests de componente de la HU-04 (T006): página "Mis Viajes".
 * Stub global de `fetch` con `vi.stubGlobal` (patrón de pages/devolucion,
 * NO MSW): render con datos, filtro mes/año, estado vacío, error de red
 * y paginación.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@solidjs/testing-library";
import { MisViajes } from "./index";

const itemGratuito = {
  id: 42,
  usuario_id: 7,
  bicicleta_id: 15,
  estacion_origen_id: 3,
  estacion_destino_id: 21,
  fecha_inicio: "2026-10-27T10:00:00Z", // 12:00 Madrid (CEST)
  fecha_fin: "2026-10-27T10:25:00Z",
  duracion_minutos: 25,
  coste_total: 0,
  estado: "finalizado",
  etiqueta: "Tramo gratuito (0,00 €)",
  estacion_origen_nombre: "Estació Universitat",
  estacion_destino_nombre: "Estació Mercat",
};

const itemDePago = {
  id: 41,
  usuario_id: 7,
  bicicleta_id: 12,
  estacion_origen_id: 21,
  estacion_destino_id: 3,
  fecha_inicio: "2026-10-20T08:00:00Z", // 10:00 Madrid
  fecha_fin: "2026-10-20T08:40:00Z",
  duracion_minutos: 40,
  coste_total: 2.5,
  estado: "finalizado",
  estacion_origen_nombre: "Estació Mercat",
  estacion_destino_nombre: "Estació Universitat",
};

function respuesta(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

const historialConDatos = {
  total: 2,
  suma_coste: 2.5,
  items: [itemGratuito, itemDePago],
};

let peticiones: string[] = [];
let responder: () => Promise<Response>;

beforeEach(() => {
  peticiones = [];
  responder = async () => respuesta(historialConDatos);
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      peticiones.push(String(input));
      return responder();
    }),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
  cleanup();
});

async function elegirFiltro(mes: string, anio: string) {
  fireEvent.change(screen.getByTestId("select-mes"), {
    target: { value: mes },
  });
  fireEvent.change(screen.getByTestId("select-anio"), {
    target: { value: anio },
  });
}

describe("HU-04 · página Mis Viajes", () => {
  it("render con datos: filas, fecha dd/mm/aaaa HH:MM, nombres, duración, coste, etiqueta y sumatorio", async () => {
    render(() => <MisViajes />);

    const filas = await screen.findAllByTestId("fila-viaje");
    expect(filas).toHaveLength(2);

    // Viaje gratuito: fecha formateada en Europe/Madrid (27-oct-2026 ya en
    // CET: 10:00Z = 11:00), nombres, etiqueta tal cual.
    expect(screen.getAllByTestId("fila-fecha")[0].textContent).toBe(
      "27/10/2026 11:00",
    );
    expect(screen.getAllByTestId("fila-origen")[0].textContent).toContain(
      "Estació Universitat",
    );
    expect(screen.getAllByTestId("fila-destino")[0].textContent).toContain(
      "Estació Mercat",
    );
    expect(screen.getAllByTestId("fila-duracion")[0].textContent).toContain(
      "25 min",
    );
    expect(screen.getAllByTestId("fila-coste")[0].textContent).toContain("0,00");
    expect(screen.getAllByTestId("fila-etiqueta")[0].textContent).toBe(
      "Tramo gratuito (0,00 €)",
    );

    // Sumatorio del periodo formateado es-ES.
    expect(screen.getByTestId("sumatorio").textContent).toContain("2,50");

    // La petición inicial es sin filtro, primera página.
    expect(peticiones[0]).toBe("/api/v1/viajes?page=1&page_size=50");
  });

  it("cambio de filtro mes/año envía desde/hasta del mes y resetea a página 1", async () => {
    render(() => <MisViajes />);
    await screen.findAllByTestId("fila-viaje");

    await elegirFiltro("10", "2026");
    await screen.findAllByTestId("fila-viaje");

    const ultima = peticiones[peticiones.length - 1];
    expect(ultima).toContain("desde=2026-10-01");
    expect(ultima).toContain("hasta=2026-10-31");
    expect(ultima).toContain("page=1");
  });

  it("periodo sin viajes: estado vacío 'sin viajes en el periodo'", async () => {
    responder = async () => respuesta({ total: 0, suma_coste: 0, items: [] });
    render(() => <MisViajes />);

    // 2027 es la opción "año actual + 1" del selector (research D7) y no
    // tiene viajes: periodo vacío sin error.
    await elegirFiltro("1", "2027");
    // El estado vacío se re-renderiza al cambiar el filtro: se espera al texto.
    await screen.findByText(/Sin viajes en el periodo/);
    expect(screen.getByTestId("sumatorio").textContent).toContain("0,00");
  });

  it("historial sin viajes (sin filtro): 'aún no tienes viajes'", async () => {
    responder = async () => respuesta({ total: 0, suma_coste: 0, items: [] });
    render(() => <MisViajes />);

    const vacio = await screen.findByTestId("estado-vacio");
    expect(vacio.textContent).toContain("Aún no tienes viajes");
  });

  it("error de red: mensaje con role='alert'", async () => {
    responder = async () => {
      throw new TypeError("NetworkError");
    };
    render(() => <MisViajes />);

    const error = await screen.findByTestId("error-historial");
    expect(error.getAttribute("role")).toBe("alert");
    expect(screen.queryByTestId("fila-viaje")).toBeNull();
  });

  it("filtro a medias (mes sin año): no envía desde/hasta y avisa", async () => {
    render(() => <MisViajes />);
    await screen.findAllByTestId("fila-viaje");

    fireEvent.change(screen.getByTestId("select-mes"), {
      target: { value: "10" },
    });
    const aviso = await screen.findByTestId("aviso-filtro-incompleto");
    expect(aviso.textContent).toContain("año");

    // Sin ambos lados no hay rango: se consulta el historial completo.
    const ultima = peticiones[peticiones.length - 1];
    expect(ultima).not.toContain("desde=");
    expect(ultima).toContain("page=1");
  });

  it("siguiente página pide page=2 y la información de paginación refleja el total", async () => {
    responder = async () =>
      respuesta({
        total: 60,
        suma_coste: 12.5,
        items: [itemDePago],
      });
    render(() => <MisViajes />);

    await screen.findByTestId("fila-viaje");
    expect(screen.getByTestId("pagina-info").textContent).toContain("Página 1 de 2");

    fireEvent.click(screen.getByTestId("boton-siguiente"));
    await screen.findByText(/Página 2 de 2/);
    expect(peticiones[peticiones.length - 1]).toContain("page=2");
  });
});
