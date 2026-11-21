// Medición S1e (SC-004 de HU-01) con el perfil de red «Fast 4G» de Chrome DevTools,
// aplicado por CDP (Network.emulateNetworkConditions, el mismo comando que usa DevTools).
// Uso: node medir-s1e.mjs <urlBase> <etiqueta> <repeticiones>
import { spawn } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const [, , BASE = 'http://localhost:5173', ETIQUETA = 'dev', N = '3'] = process.argv;
const CHROME = 'C:/Program Files/Google/Chrome/Application/chrome.exe';
const PUERTO = 9333;
const ESTACIONES = 60;
// Valores del preset «Fast 4G» de DevTools (front_end/core/sdk/NetworkManager.ts):
// 9 Mbit/s bajada, 1,5 Mbit/s subida (×0,9) y 60 ms × 2,75 de latencia.
const FAST_4G = {
  offline: false,
  latency: 60 * 2.75,
  downloadThroughput: ((9 * 1000 * 1000) / 8) * 0.9,
  uploadThroughput: ((1.5 * 1000 * 1000) / 8) * 0.9,
};

const esperar = (ms) => new Promise((r) => setTimeout(r, ms));

async function login() {
  const r = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: 'ana@si16.test', password: 'password' }),
  });
  if (!r.ok) throw new Error(`login ${r.status}`);
  return r.json();
}

function conectar(wsUrl) {
  const ws = new WebSocket(wsUrl);
  let id = 0;
  const pendientes = new Map();
  const oyentes = [];
  ws.onmessage = (ev) => {
    const m = JSON.parse(ev.data);
    if (m.id && pendientes.has(m.id)) {
      const { ok, ko } = pendientes.get(m.id);
      pendientes.delete(m.id);
      m.error ? ko(new Error(JSON.stringify(m.error))) : ok(m.result);
    } else if (m.method) oyentes.forEach((f) => f(m));
  };
  const enviar = (method, params = {}) =>
    new Promise((ok, ko) => {
      const n = ++id;
      pendientes.set(n, { ok, ko });
      ws.send(JSON.stringify({ id: n, method, params }));
    });
  return new Promise((ok) => (ws.onopen = () => ok({ enviar, alEvento: (f) => oyentes.push(f), ws })));
}

async function evaluar(c, expr) {
  const r = await c.enviar('Runtime.evaluate', { expression: expr, awaitPromise: true, returnByValue: true });
  if (r.exceptionDetails) throw new Error(r.exceptionDetails.exception?.description ?? 'error JS');
  return r.result.value;
}

// Se inyecta antes que cualquier script de la página: marca el instante (desde el
// inicio de la navegación) en que hay ESTACIONES marcadores en el DOM y se ha
// pintado el siguiente frame.
const SONDA = `
  window.__s1e = {};
  new MutationObserver((_, obs) => {
    if (document.querySelectorAll('.marcador-estacion').length >= ${ESTACIONES} && !window.__s1e.marcadores) {
      window.__s1e.marcadores = -1;
      requestAnimationFrame(() => requestAnimationFrame(() => { window.__s1e.marcadores = performance.now(); }));
      obs.disconnect();
    }
  }).observe(document, { childList: true, subtree: true });
`;

async function pasada(c, sesion, ubicacion) {
  const origen = new URL(BASE).origin;
  // Sesión iniciada (como tras el login) y permiso de ubicación del escenario.
  await c.enviar('Network.emulateNetworkConditions', { offline: false, latency: 0, downloadThroughput: -1, uploadThroughput: -1 });
  await c.enviar('Page.navigate', { url: origen + '/' });
  await esperar(1500);
  await evaluar(c, `localStorage.setItem('si16.token', ${JSON.stringify(sesion.token)});
                    localStorage.setItem('si16.usuario', ${JSON.stringify(JSON.stringify(sesion.usuario))}); true`);
  await c.enviar('Browser.setPermission', { origin: origen, permission: { name: 'geolocation' }, setting: ubicacion === 'concedida' ? 'granted' : 'denied' });
  if (ubicacion === 'concedida') await c.enviar('Emulation.setGeolocationOverride', { latitude: 39.4699, longitude: -0.3763, accuracy: 30 });
  else await c.enviar('Emulation.clearGeolocationOverride');

  // Caché vacía + Fast 4G, y medición.
  await c.enviar('Network.clearBrowserCache');
  await c.enviar('Network.setCacheDisabled', { cacheDisabled: true });
  await c.enviar('Network.emulateNetworkConditions', FAST_4G);
  let bytes = 0;
  let peticiones = 0;
  const errores = [];
  c.alEvento((m) => {
    if (m.method === 'Network.loadingFinished') { bytes += m.params.encodedDataLength; peticiones++; }
    if (m.method === 'Runtime.exceptionThrown') errores.push(m.params.exceptionDetails.text);
    if (m.method === 'Runtime.consoleAPICalled' && m.params.type === 'error') errores.push(m.params.args.map((a) => a.value ?? a.description).join(' '));
  });
  await c.enviar('Page.navigate', { url: BASE + '/' });

  const limite = Date.now() + 30000;
  let tMarcadores = null;
  while (Date.now() < limite) {
    const v = await evaluar(c, `window.__s1e && window.__s1e.marcadores > 0 ? window.__s1e.marcadores : null`).catch(() => null);
    if (v) { tMarcadores = v; break; }
    await esperar(50);
  }
  const bytesHastaMarcadores = bytes;
  const peticionesHastaMarcadores = peticiones;
  if (!tMarcadores) return { ubicacion, error: 'los 60 marcadores no aparecieron en 30 s' };

  // Clic en un marcador y espera al detalle con datos.
  const tDetalle = await evaluar(c, `new Promise((ok, ko) => {
      const m = document.querySelectorAll('.marcador-estacion')[7];
      const t0 = performance.now();
      const listo = () => { const b = document.querySelector('[data-testid="contador-bicis"]'); return b && b.textContent.trim() !== ''; };
      const obs = new MutationObserver(() => { if (listo()) { obs.disconnect(); requestAnimationFrame(() => ok(performance.now() - t0)); } });
      obs.observe(document.body, { childList: true, subtree: true, characterData: true });
      m.dispatchEvent(new MouseEvent('click', { bubbles: true }));
      setTimeout(() => ko(new Error('el detalle no apareció en 10 s')), 10000);
    })`);
  const encuadre = await evaluar(c, `(() => { const z = document.querySelector('.leaflet-container'); return z ? 'mapa visible' : 'sin mapa'; })()`);
  return {
    ubicacion,
    marcadores_ms: Math.round(tMarcadores),
    detalle_ms: Math.round(tDetalle),
    kB_hasta_marcadores: Math.round(bytesHastaMarcadores / 1024),
    peticiones_hasta_marcadores: peticionesHastaMarcadores,
    mapa: encuadre,
    marcadores_dentro_de_la_vista: await evaluar(c, `(() => {
      const z = document.querySelector('.leaflet-container').getBoundingClientRect();
      return [...document.querySelectorAll('.marcador-estacion')].filter((m) => {
        const r = m.getBoundingClientRect();
        return r.left >= z.left && r.right <= z.right && r.top >= z.top && r.bottom <= z.bottom;
      }).length + '/' + document.querySelectorAll('.marcador-estacion').length;
    })()`),
    errores_consola: errores.filter((e) => !/tile|openstreetmap/i.test(e)),
  };
}

async function main() {
  const sesion = await login();
  const resultados = [];
  for (const ubicacion of ['concedida', 'denegada']) {
    for (let i = 1; i <= Number(N); i++) {
      const perfil = mkdtempSync(join(tmpdir(), 's1e-'));
      const chrome = spawn(CHROME, [
        '--headless=new', `--remote-debugging-port=${PUERTO}`, `--user-data-dir=${perfil}`,
        '--window-size=1366,768', '--no-first-run', '--no-default-browser-check', 'about:blank',
      ], { stdio: 'ignore' });
      try {
        let pagina;
        for (let k = 0; k < 50 && !pagina; k++) {
          await esperar(200);
          pagina = await fetch(`http://localhost:${PUERTO}/json/list`).then((r) => r.json()).then((l) => l.find((t) => t.type === 'page')).catch(() => null);
        }
        const c = await conectar(pagina.webSocketDebuggerUrl);
        await c.enviar('Page.enable');
        await c.enviar('Runtime.enable');
        await c.enviar('Network.enable');
        await c.enviar('Page.addScriptToEvaluateOnNewDocument', { source: SONDA });
        const r = await pasada(c, sesion, ubicacion);
        resultados.push({ servidor: ETIQUETA, pasada: i, ...r });
        console.log(JSON.stringify(resultados.at(-1)));
        c.ws.close();
      } finally {
        chrome.kill();
        await esperar(800);
        try { rmSync(perfil, { recursive: true, force: true }); } catch {}
      }
    }
  }
}

main().catch((e) => { console.error(e); process.exit(1); });
