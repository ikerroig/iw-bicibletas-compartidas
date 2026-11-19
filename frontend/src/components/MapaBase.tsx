// Mapa genérico (T030): Leaflet + teselas de OpenStreetMap con marcadores
// L.divIcon. No conoce la lógica de estados: recibe el HTML y el aria-label de
// cada marcador ya calculados (pages/mapa/).
//
// Encuadre (HU-01 FR-013, research R23): se encuadra toda la red, salvo que el
// usuario conceda su ubicación y esta caiga DENTRO del área de la red, en cuyo
// caso se centra en él. Si la posición llega antes que las estaciones, se
// guarda y se decide al llegar estas. Ubicación denegada, no disponible, tardía
// (> 5 s) o fuera del área → toda la red. data-encuadre lo refleja para tests.
import L from 'leaflet';
import { createEffect, createSignal, onCleanup, onMount } from 'solid-js';

export interface MarcadorMapa {
  id: number;
  lat: number;
  lng: number;
  /** HTML del contenido del marcador (texto + icono; no solo color). */
  html: string;
  /** Descripción para lectores de pantalla. */
  ariaLabel: string;
}

interface Props {
  marcadores: MarcadorMapa[];
  onSeleccionar?: (id: number) => void;
  centrarEnUsuario?: boolean;
  timeoutGeolocalizacionMs?: number;
  class?: string;
}

const CENTRO_POR_DEFECTO: L.LatLngExpression = [39.4699, -0.3763]; // València
const ZOOM_USUARIO = 15;

type Encuadre = 'pendiente' | 'red' | 'usuario';

export default function MapaBase(props: Props) {
  let contenedor!: HTMLDivElement;
  const [encuadre, setEncuadre] = createSignal<Encuadre>('pendiente');

  onMount(() => {
    const mapa = L.map(contenedor, { zoomControl: true }).setView(CENTRO_POR_DEFECTO, 13);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '&copy; colaboradores de OpenStreetMap',
    }).addTo(mapa);

    const capas = new Map<number, { marcador: L.Marker; html: string; aria: string }>();

    // Última lista de marcadores y posición del usuario (si ya llegó): con las
    // dos se decide el encuadre (FR-013, R23), sin importar cuál llega antes.
    let listaActual: MarcadorMapa[] = [];
    let posicionUsuario: L.LatLng | null = null;

    const encuadrarRed = (lista: MarcadorMapa[]) => {
      if (lista.length === 0 || encuadre() === 'usuario') return;
      const limites = limitesDe(lista);
      // Sin tamaño (p. ej. jsdom o contenedor oculto) fitBounds no puede calcular el zoom.
      if (mapa.getSize().x > 0 && mapa.getSize().y > 0)
        mapa.fitBounds(limites, { padding: [24, 24], maxZoom: 16 });
      else mapa.setView(limites.getCenter(), 13);
      setEncuadre('red');
    };

    /** Centra en el usuario solo si está dentro del área de la red. */
    const centrarSiDentro = (): boolean => {
      if (!posicionUsuario || listaActual.length === 0) return false;
      if (!areaDeLaRed(listaActual).contains(posicionUsuario)) return false;
      mapa.setView(posicionUsuario, ZOOM_USUARIO);
      setEncuadre('usuario');
      return true;
    };

    // Geolocalización con plazo propio: una respuesta tardía se ignora.
    if (props.centrarEnUsuario !== false && navigator.geolocation) {
      let vencido = false;
      const plazo = setTimeout(() => (vencido = true), props.timeoutGeolocalizacionMs ?? 5_000);
      // Al desmontar, una respuesta posterior se ignora: el mapa ya no existe.
      onCleanup(() => {
        vencido = true;
        clearTimeout(plazo);
      });
      navigator.geolocation.getCurrentPosition(
        // Callback del navegador (una sola vez): lee el encuadre en ese momento, no necesita seguimiento.
        // eslint-disable-next-line solid/reactivity
        (pos) => {
          clearTimeout(plazo);
          if (vencido) return;
          // Si aún no hay estaciones, se guarda y se decide cuando lleguen.
          // Fuera del área de la red se ignora: se queda el encuadre de la red.
          posicionUsuario = L.latLng(pos.coords.latitude, pos.coords.longitude);
          if (encuadre() !== 'usuario') centrarSiDentro();
        },
        () => {
          clearTimeout(plazo); // denegada o no disponible: se queda el encuadre de la red
        },
        { timeout: props.timeoutGeolocalizacionMs ?? 5_000, maximumAge: 60_000 },
      );
    }

    createEffect(() => {
      const lista = props.marcadores;
      const vistos = new Set<number>();
      for (const m of lista) {
        vistos.add(m.id);
        const actual = capas.get(m.id);
        if (!actual) {
          const marcador = L.marker([m.lat, m.lng], {
            icon: icono(m.html),
            keyboard: true,
            title: m.ariaLabel,
          })
            // Callback de Leaflet: lee la prop en el momento del clic (no necesita seguimiento).
            // eslint-disable-next-line solid/reactivity
            .on('click', () => props.onSeleccionar?.(m.id))
            .addTo(mapa);
          marcador.getElement()?.setAttribute('aria-label', m.ariaLabel);
          capas.set(m.id, { marcador, html: m.html, aria: m.ariaLabel });
          continue;
        }
        actual.marcador.setLatLng([m.lat, m.lng]);
        // setIcon reconstruye el icono con options.title: se actualiza antes para
        // que el tooltip nativo no vuelva a la etiqueta del primer pintado.
        actual.marcador.options.title = m.ariaLabel;
        if (actual.html !== m.html) {
          actual.marcador.setIcon(icono(m.html));
          actual.html = m.html;
        }
        const el = actual.marcador.getElement();
        if (
          el &&
          (actual.aria !== m.ariaLabel ||
            el.getAttribute('aria-label') !== m.ariaLabel ||
            el.getAttribute('title') !== m.ariaLabel)
        ) {
          el.setAttribute('aria-label', m.ariaLabel);
          el.setAttribute('title', m.ariaLabel);
          actual.aria = m.ariaLabel;
        }
      }
      for (const [id, c] of capas) {
        if (!vistos.has(id)) {
          c.marcador.remove();
          capas.delete(id);
        }
      }
      listaActual = lista;
      if (encuadre() === 'pendiente' && !centrarSiDentro()) encuadrarRed(lista);
    });

    const observador =
      typeof ResizeObserver !== 'undefined'
        ? new ResizeObserver(() => mapa.invalidateSize())
        : null;
    observador?.observe(contenedor);

    onCleanup(() => {
      observador?.disconnect();
      mapa.remove();
    });
  });

  return (
    <div
      ref={contenedor}
      class={props.class}
      data-encuadre={encuadre()}
      role="application"
      aria-label="Mapa de estaciones"
    />
  );
}

function limitesDe(lista: MarcadorMapa[]): L.LatLngBounds {
  return L.latLngBounds(lista.map((m) => [m.lat, m.lng] as L.LatLngTuple));
}

/** Margen mínimo del área de la red, en grados (≈ 2 km). */
const MARGEN_MINIMO_GRADOS = 0.02;

/**
 * Área de la red (FR-013, R23): rectángulo que abarca todas las estaciones,
 * ampliado un 20 % y, como mínimo, ~2 km por cada lado (para que una red muy
 * compacta no deje fuera a quien está en la calle de al lado).
 */
export function areaDeLaRed(lista: MarcadorMapa[]): L.LatLngBounds {
  const b = limitesDe(lista).pad(0.2);
  const so = b.getSouthWest();
  const ne = b.getNorthEast();
  return L.latLngBounds(
    [so.lat - MARGEN_MINIMO_GRADOS, so.lng - MARGEN_MINIMO_GRADOS],
    [ne.lat + MARGEN_MINIMO_GRADOS, ne.lng + MARGEN_MINIMO_GRADOS],
  );
}

function icono(html: string): L.DivIcon {
  // iconSize 0: el contenido se posiciona con CSS (ancla abajo-centro).
  return L.divIcon({ className: 'marcador-estacion', html, iconSize: [0, 0] });
}
