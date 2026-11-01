import { createEffect, onCleanup, type Component } from "solid-js";
import {
  Chart,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  type ChartConfiguration,
} from "chart.js";

/**
 * Wrapper fino sobre Chart.js (R14 del plan global) para los gráficos de
 * barras de la HU-08: recibe etiquetas + datos ya calculados por el backend
 * (const. I — el frontend solo pinta los arrays). Interactividad = tooltip
 * al explorar (Assumption del spec).
 *
 * Nota: chart.js se instala al aprobarse T020 (frontend/package.json es
 * compartido); los tests de la página mockean este wrapper.
 */
Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

export interface GraficoBarrasProps {
  titulo: string;
  etiquetas: string[];
  datos: number[];
  testid: string;
}

export const GraficoBarras: Component<GraficoBarrasProps> = (props) => {
  let canvas: HTMLCanvasElement | undefined;
  let chart: Chart | undefined;

  createEffect(() => {
    // Suscripción reactiva a los datos del rango consultado.
    const etiquetas = props.etiquetas;
    const datos = props.datos;
    if (!canvas) return;
    chart?.destroy();
    const config: ChartConfiguration<"bar"> = {
      type: "bar",
      data: {
        labels: etiquetas,
        datasets: [
          {
            label: props.titulo,
            data: datos,
            backgroundColor: "rgb(37, 99, 235)",
          },
        ],
      },
      options: {
        responsive: true,
        plugins: { legend: { display: false } },
      },
    };
    chart = new Chart(canvas, config);
  });

  onCleanup(() => chart?.destroy());

  return (
    <div
      data-testid={props.testid}
      class="rounded-xl border border-gray-200 bg-white p-3"
    >
      <h2 class="mb-2 text-sm font-semibold text-gray-700">{props.titulo}</h2>
      <canvas ref={canvas} />
    </div>
  );
};

export default GraficoBarras;
