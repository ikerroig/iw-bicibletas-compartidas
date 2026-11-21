// Graba una demostracion reproducible de la interfaz real mediante Chrome CDP.
// Cada fotograma procede de http://127.0.0.1:5173; los rotulos solo explican
// la historia y el resultado que se esta mostrando.
import { spawn } from 'node:child_process';
import { mkdir, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';

const ROOT = path.dirname(new URL(import.meta.url).pathname.replace(/^\/(.:)/, '$1'));
const FRAMES = path.join(ROOT, 'recording-frames');
const CHROME = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE = 'http://127.0.0.1:5173';
const PORT = 9333;

await rm(FRAMES, { recursive: true, force: true });
await mkdir(FRAMES, { recursive: true });

const browser = spawn(CHROME, [
  '--headless=new',
  `--remote-debugging-port=${PORT}`,
  '--remote-allow-origins=*',
  '--disable-gpu',
  '--hide-scrollbars',
  '--window-size=1280,720',
  '--force-device-scale-factor=1',
  '--no-first-run',
  '--no-default-browser-check',
  `--user-data-dir=${path.join(ROOT, '.demo-browser')}`,
  `${BASE}/login`,
], { stdio: 'ignore' });

let frame = 0;
let nextId = 1;
let socket;
const pending = new Map();

const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function endpoint() {
  for (let attempt = 0; attempt < 60; attempt++) {
    try {
      const pages = await fetch(`http://127.0.0.1:${PORT}/json`);
      const list = await pages.json();
      const page = list.find((item) => item.type === 'page');
      if (page) return page.webSocketDebuggerUrl;
    } catch {
      // Chrome aun esta arrancando.
    }
    await wait(200);
  }
  throw new Error('No se pudo conectar con Chrome CDP');
}

function command(method, params = {}) {
  const id = nextId++;
  return new Promise((resolve, reject) => {
    pending.set(id, { resolve, reject });
    socket.send(JSON.stringify({ id, method, params }));
  });
}

async function js(expression) {
  const result = await command('Runtime.evaluate', {
    expression,
    awaitPromise: true,
    returnByValue: true,
  });
  if (result.exceptionDetails) {
    throw new Error(result.exceptionDetails.text || 'Error al evaluar JavaScript');
  }
  return result.result?.value;
}

async function capture() {
  const shot = await command('Page.captureScreenshot', {
    format: 'jpeg',
    quality: 82,
    fromSurface: true,
    captureBeyondViewport: false,
  });
  frame += 1;
  const name = `frame-${String(frame).padStart(5, '0')}.jpg`;
  await writeFile(path.join(FRAMES, name), Buffer.from(shot.data, 'base64'));
}

async function hold(seconds = 2.8) {
  const count = Math.ceil(seconds * 5);
  for (let i = 0; i < count; i++) {
    await capture();
    await wait(200);
  }
}

async function overlay(story, title, detail, tone = '#047857') {
  await js(`(() => {
    document.getElementById('demo-caption')?.remove();
    const box = document.createElement('aside');
    box.id = 'demo-caption';
    box.style.cssText = 'position:fixed;left:22px;right:22px;bottom:18px;z-index:2147483647;background:rgba(15,23,42,.94);color:white;border-left:8px solid ${tone};border-radius:12px;padding:12px 18px;box-shadow:0 12px 30px rgba(0,0,0,.35);font-family:Segoe UI,Arial,sans-serif;pointer-events:none';
    box.innerHTML = '<div style="font-size:14px;font-weight:800;letter-spacing:.08em;color:#a7f3d0">${story}</div><div style="font-size:23px;font-weight:750;margin-top:2px">${title}</div><div style="font-size:15px;color:#dbeafe;margin-top:3px">${detail}</div>';
    document.body.appendChild(box);
  })()`);
}

async function scene(story, title, detail, seconds = 2.8, tone = '#047857') {
  await overlay(story, title, detail, tone);
  await hold(seconds);
}

async function navigate(route) {
  await command('Page.navigate', { url: `${BASE}${route}` });
  await wait(1300);
}

async function setValue(selector, value) {
  const encodedSelector = JSON.stringify(selector);
  const encodedValue = JSON.stringify(value);
  const ok = await js(`(() => {
    const element = document.querySelector(${encodedSelector});
    if (!element) return false;
    const prototype = element instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype :
      element instanceof HTMLSelectElement ? HTMLSelectElement.prototype : HTMLInputElement.prototype;
    const setter = Object.getOwnPropertyDescriptor(prototype, 'value').set;
    setter.call(element, ${encodedValue});
    element.dispatchEvent(new Event('input', { bubbles: true }));
    element.dispatchEvent(new Event('change', { bubbles: true }));
    return true;
  })()`);
  if (!ok) throw new Error(`No existe el control ${selector}`);
}

async function click(selector) {
  const ok = await js(`(() => {
    const element = document.querySelector(${JSON.stringify(selector)});
    if (!element) return false;
    element.scrollIntoView({ block: 'center' });
    element.click();
    return true;
  })()`);
  if (!ok) throw new Error(`No existe el elemento ${selector}`);
  await wait(900);
}

async function clickText(text, tag = '*') {
  const ok = await js(`(() => {
    const target = [...document.querySelectorAll(${JSON.stringify(tag)})]
      .find((element) => element.textContent.trim().includes(${JSON.stringify(text)}));
    if (!target) return false;
    target.scrollIntoView({ block: 'center' });
    target.click();
    return true;
  })()`);
  if (!ok) throw new Error(`No existe el texto ${text}`);
  await wait(900);
}

async function login(email, password = 'password') {
  await navigate('/login');
  await setValue('input[type=email]', email);
  await setValue('input[type=password]', password);
  await capture();
  await click('button[type=submit]');
  await wait(1100);
}

async function logout() {
  await clickText('Cerrar sesi');
  await wait(500);
}

try {
  socket = new WebSocket(await endpoint());
  await new Promise((resolve, reject) => {
    socket.addEventListener('open', resolve, { once: true });
    socket.addEventListener('error', reject, { once: true });
  });
  socket.addEventListener('message', ({ data }) => {
    const message = JSON.parse(data);
    if (!message.id) return;
    const item = pending.get(message.id);
    if (!item) return;
    pending.delete(message.id);
    if (message.error) item.reject(new Error(message.error.message));
    else item.resolve(message.result);
  });
  await command('Page.enable');
  await command('Runtime.enable');
  await command('Emulation.setDeviceMetricsOverride', {
    width: 1280, height: 720, deviceScaleFactor: 1, mobile: false,
  });
  await wait(900);

  await scene('SI-16 · DEMOSTRACIÓN FUNCIONAL', 'Aplicación ejecutándose en local',
    'Frontend SolidJS conectado a la API Go, PostgreSQL y Redis.', 3.5);

  await setValue('input[type=email]', 'ana@si16.test');
  await setValue('input[type=password]', 'incorrecta');
  await click('button[type=submit]');
  await scene('AUTENTICACIÓN · CASO DE ERROR', 'Credenciales incorrectas',
    'La API rechaza el acceso y la interfaz explica el error sin iniciar sesión.', 2.5, '#dc2626');

  await login('ana@si16.test');
  await scene('HU-01 · DISPONIBILIDAD', 'Mapa de la red en tiempo real',
    'La aplicación carga las estaciones y sus contadores desde el backend.', 3.2);
  await clickText('Lista', 'button');
  await scene('HU-01 · ACCESIBILIDAD', 'Listado textual de estaciones',
    'La misma disponibilidad se puede consultar sin depender del mapa.', 2.6);
  await click('ul[aria-label="Listado de estaciones"] button');
  await scene('HU-01 · DETALLE', 'Bicicletas y anclajes de una estación',
    'El detalle muestra los dos contadores y el estado operativo.', 2.6);

  await navigate('/alquiler');
  await setValue('input[placeholder="E01-A01"]', 'E01-A02');
  await scene('HU-02 · DESBLOQUEO', 'Código de anclaje listo para validar',
    'El usuario con abono vigente solicita una bicicleta disponible.', 2.5);
  await click('button[type=submit]');
  await scene('HU-02 · DESBLOQUEO', 'Viaje iniciado correctamente',
    'La bicicleta queda en uso y la interfaz muestra el único viaje activo.', 3.0);

  await navigate('/devolucion');
  await setValue('[data-testid=input-anclaje]', 'CODIGO-INEXISTENTE');
  await click('[data-testid=boton-devolver]');
  await scene('HU-03 · DEVOLUCIÓN · ERROR', 'Anclaje inválido rechazado',
    'No se finaliza el viaje ni se alteran los datos con un código inexistente.', 2.7, '#dc2626');
  await setValue('[data-testid=input-anclaje]', 'E01-A08');
  await click('[data-testid=boton-devolver]');
  await scene('HU-03 · DEVOLUCIÓN', 'Viaje finalizado y tarifa calculada',
    'El anclaje válido cierra el viaje y muestra duración, coste y tramo gratuito.', 3.2);

  await navigate('/viajes');
  await scene('HU-04 · HISTORIAL', 'Viajes ordenados y coste acumulado',
    'El recorrido recién finalizado aparece junto al historial del usuario.', 3.2);
  await setValue('[data-testid=select-mes]', '');
  await scene('HU-04 · FILTRO · LÍMITE', 'Filtro incompleto avisado',
    'La interfaz no aplica un mes sin año y orienta al usuario.', 2.4, '#d97706');

  await navigate('/analitica');
  await scene('HU-08 · PERMISOS', 'Acceso de usuario normal denegado',
    'El backend protege la analítica incluso si se escribe la URL directamente.', 2.6, '#dc2626');

  await navigate('/averias');
  await setValue('input[placeholder="BICI-0001"]', 'BICI-0003');
  await setValue('select', '1');
  await setValue('textarea', 'Freno delantero con respuesta irregular durante la inspección de la demo.');
  await scene('HU-05 · AVERÍAS', 'Formulario con datos reales de la estación',
    'El usuario identifica la bicicleta, la estación y el fallo observado.', 2.6);
  await click('button[type=submit]');
  await scene('HU-05 · AVERÍAS', 'Reporte registrado e inhabilitación inmediata',
    'La API confirma la incidencia y bloquea la bicicleta para nuevos alquileres.', 3.0);

  await logout();
  await login('taller@si16.test');
  await navigate('/incidencias');
  await scene('HU-05 · AVISO AL TALLER', 'Incidencias pendientes actualizadas',
    'El reporte del usuario ya aparece para el operario de mantenimiento.', 3.2);

  await navigate('/taller');
  await setValue('[data-testid=input-bicicleta-id]', '3');
  await click('[data-testid=boton-buscar-ficha]');
  await scene('HU-06 · TALLER', 'Ficha e historial de la bicicleta',
    'El operario consulta el reporte original antes de intervenir.', 3.0);
  await setValue('[data-testid=input-piezas]', 'Zapatas de freno delanteras');
  await setValue('[data-testid=input-notas]', 'Ajuste y prueba de frenado completados durante la demo.');
  await click('[data-testid=boton-guardar-parte]');
  await scene('HU-06 · INTERVENCIÓN', 'Parte de reparación guardado',
    'Quedan registradas las piezas, notas, fecha y operario responsable.', 3.0);
  if (await js(`Boolean(document.querySelector('[data-testid=input-codigo-anclaje]'))`)) {
    await setValue('[data-testid=input-codigo-anclaje]', 'E01-A03');
    await click('[data-testid=boton-ubicar]');
    await scene('HU-06 · REACTIVACIÓN', 'Ubicación confirmada tras reparar',
      'La bicicleta vuelve a estar disponible únicamente después de confirmar el anclaje.', 3.0);
  }

  await logout();
  await login('logistica@si16.test');
  await navigate('/balanceo');
  await scene('HU-07 · BALANCEO', 'Alertas ordenadas por urgencia',
    'El panel muestra estaciones por debajo del 10 % o por encima del 90 %.', 3.2);
  await setValue('[data-testid=input-origen]', '10');
  await setValue('[data-testid=input-destino]', '10');
  await setValue('[data-testid=input-unidades]', '2');
  await click('[data-testid=boton-traslado]');
  await scene('HU-07 · VALIDACIÓN', 'Traslado inconsistente rechazado',
    'Origen y destino deben ser estaciones distintas y los datos deben ser válidos.', 2.7, '#dc2626');
  await setValue('[data-testid=input-destino]', '8');
  await click('[data-testid=boton-traslado]');
  await scene('HU-07 · TRASLADO', 'Disponibilidad recalculada',
    'La operación válida actualiza origen, destino y alertas de balanceo.', 3.2);

  await logout();
  await login('gestor@si16.test');
  await navigate('/analitica');
  await setValue('[data-testid=input-desde]', '2026-01-01');
  await setValue('[data-testid=input-hasta]', '2026-12-31');
  await wait(1400);
  await scene('HU-08 · ANALÍTICA', 'Gráficos y totales del periodo',
    'El gestor consulta viajes por hora y por día de la semana.', 3.4);
  await click('[data-testid=boton-descargar]');
  await scene('HU-08 · EXPORTACIÓN', 'CSV descargado con el mismo filtro',
    'La exportación usa los mismos viajes y totales que los gráficos.', 2.6);
  await setValue('[data-testid=input-desde]', '2026-12-31');
  await setValue('[data-testid=input-hasta]', '2026-01-01');
  await scene('HU-08 · RANGO INVÁLIDO', 'Fechas invertidas rechazadas',
    'No se generan gráficos ni CSV cuando la fecha inicial supera a la final.', 2.7, '#dc2626');

  await navigate('/');
  await scene('SI-16 · CIERRE', 'Ocho historias ejecutadas sobre la aplicación real',
    'Se han mostrado usuarios, errores, permisos, cambios de estado y los cuatro roles.', 4.2);

  console.log(`Grabados ${frame} fotogramas reales en ${FRAMES}`);
} finally {
  if (socket?.readyState === WebSocket.OPEN) socket.close();
  browser.kill();
}
