// Mapa genérico (T030): Leaflet + teselas de OpenStreetMap con marcadores
// L.divIcon. No conoce la lógica de estados: recibe el HTML y el aria-label de
// cada marcador ya calculados (pages/mapa/).
//
// Encuadre (caso límite de HU-01): si el usuario concede su ubicación, se
// centra en ella; si la deniega, no está disponible o tarda más de 5 s, se
// encuadra en toda la red. El atributo data-encuadre lo refleja para tests.
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

    const encuadrarRed = (lista: MarcadorMapa[]) => {
      if (lista.length === 0 || encuadre() === 'usuario') return;
      const limites = L.latLngBounds(lista.map((m) => [m.lat, m.lng] as L.LatLngTuple));
      // Sin tamaño (p. ej. jsdom o contenedor oculto) fitBounds no puede calcular el zoom.
      if (mapa.getSize().x > 0 && mapa.getSize().y > 0)
        mapa.fitBounds(limites, { padding: [24, 24], maxZoom: 16 });
      else mapa.setView(limites.getCenter(), 13);
      setEncuadre('red');
    };

    // Geolocalización con plazo propio: una respuesta tardía se ignora.
    if (props.centrarEnUsuario !== false && navigator.geolocation) {
      let vencido = false;
      const plazo = setTimeout(() => (vencido = true), props.timeoutGeolocalizacionMs ?? 5_000);
      onCleanup(() => clearTimeout(plazo));
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          clearTimeout(plazo);
          if (vencido) return;
          mapa.setView([pos.coords.latitude, pos.coords.longitude], ZOOM_USUARIO);
          setEncuadre('usuario');
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
        if (actual.html !== m.html) {
          actual.marcador.setIcon(icono(m.html));
          actual.html = m.html;
        }
        const el = actual.marcador.getElement();
        if (el && (actual.aria !== m.ariaLabel || el.getAttribute('aria-label') !== m.ariaLabel)) {
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
      if (encuadre() === 'pendiente') encuadrarRed(lista);
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

function icono(html: string): L.DivIcon {
  // iconSize 0: el contenido se posiciona con CSS (ancla abajo-centro).
  return L.divIcon({ className: 'marcador-estacion', html, iconSize: [0, 0] });
}
