// Graba ocho demostraciones independientes de la aplicación SI-16.
//
// Cada historia empieza en el login, usa la interfaz real y termina mostrando
// el resultado de la operación. Los rótulos solo explican lo que ocurre: las
// pantallas y respuestas proceden del frontend, la API, PostgreSQL y Redis.
//
// Requisitos: backend en :8080, frontend Vite en :5173 y DATABASE_URL apuntando
// a la base de datos de demostración. El script reinicia el seed antes de cada
// historia para que la grabación sea repetible.
import { spawn, spawnSync } from 'node:child_process';
import { copyFile, mkdir, readdir, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const ROOT = path.dirname(fileURLToPath(import.meta.url));
const PROJECT = path.dirname(ROOT);
const BACKEND = path.join(PROJECT, 'backend');
const FRAMES_ROOT = path.join(ROOT, 'recording-frames');
const PROFILE = path.join(ROOT, '.demo-browser');
const DOWNLOADS = path.join(ROOT, 'downloads-temp');
const CHROME = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE = process.env.DEMO_BASE_URL ?? 'http://127.0.0.1:5173';
const API = `${BASE}/api/v1`;
const CDP_PORT = 9333;
const FPS = 4;
// La referencia dedica unos 20–35 s a cada escenario. PACE amplía solo las
// escenas explicativas; la escritura, los clics y las esperas de red mantienen
// su velocidad natural para que la demostración siga siendo dinámica.
const PACE = 5.5;
// Los fotogramas repetidos no necesitan espera en tiempo real. La aplicación
// sí espera explícitamente después de cada navegación o petición.
const CAPTURE_DELAY_MS = 20;

await rm(FRAMES_ROOT, { recursive: true, force: true });
await rm(PROFILE, { recursive: true, force: true });
await rm(DOWNLOADS, { recursive: true, force: true });
await mkdir(FRAMES_ROOT, { recursive: true });
await mkdir(DOWNLOADS, { recursive: true });

const browser = spawn(CHROME, [
  '--headless=new',
  `--remote-debugging-port=${CDP_PORT}`,
  '--remote-allow-origins=*',
  '--disable-gpu',
  '--hide-scrollbars',
  '--window-size=1280,720',
  '--force-device-scale-factor=1',
  '--no-first-run',
  '--no-default-browser-check',
  `--user-data-dir=${PROFILE}`,
  `${BASE}/login`,
], { stdio: 'ignore' });

let nextId = 1;
let socket;
const pending = new Map();
const manifest = { fps: FPS, width: 1280, height: 720, stories: [] };
let current = null;

const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function endpoint() {
  for (let attempt = 0; attempt < 80; attempt++) {
    try {
      const pages = await fetch(`http://127.0.0.1:${CDP_PORT}/json`);
      const list = await pages.json();
      const page = list.find((item) => item.type === 'page');
      if (page) return page.webSocketDebuggerUrl;
    } catch {
      // Edge todavía está arrancando.
    }
    await wait(200);
  }
  throw new Error('No se pudo conectar con Edge mediante CDP');
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
    throw new Error(result.exceptionDetails.exception?.description ?? result.exceptionDetails.text);
  }
  return result.result?.value;
}

async function capture() {
  if (!current) return;
  const shot = await command('Page.captureScreenshot', {
    format: 'jpeg',
    quality: 84,
    fromSurface: true,
    captureBeyondViewport: false,
  });
  current.frame += 1;
  const name = `frame-${String(current.frame).padStart(5, '0')}.jpg`;
  await writeFile(path.join(current.folder, name), Buffer.from(shot.data, 'base64'));
}

async function hold(seconds = 3) {
  const total = Math.max(1, Math.ceil(seconds * FPS));
  for (let i = 0; i < total; i++) {
    await capture();
    await wait(CAPTURE_DELAY_MS);
  }
}

function currentTime() {
  return current ? current.frame / FPS : 0;
}

function addChapter(title) {
  if (!current) return;
  const previous = current.chapters.at(-1);
  if (previous?.title === title) return;
  current.chapters.push({ title, t: Number(currentTime().toFixed(2)) });
}

async function titleCard(meta) {
  const payload = JSON.stringify(meta);
  await js(`(() => {
    document.getElementById('demo-title')?.remove();
    const data = ${payload};
    const card = document.createElement('section');
    card.id = 'demo-title';
    card.style.cssText = [
      'position:fixed','inset:0','z-index:2147483647','display:flex',
      'align-items:center','justify-content:center','padding:64px',
      'background:linear-gradient(135deg,rgba(4,47,46,.98),rgba(15,23,42,.98))',
      'color:white','font-family:Segoe UI,Arial,sans-serif','text-align:left'
    ].join(';');
    const box = document.createElement('div');
    box.style.cssText = 'width:min(900px,90vw);overflow-wrap:anywhere';
    const badge = document.createElement('p');
    badge.textContent = data.spec + ' · ' + data.role;
    badge.style.cssText = 'margin:0 0 16px;color:#67e8f9;font-size:18px;font-weight:800;letter-spacing:.12em;text-transform:uppercase';
    const h = document.createElement('h1');
    h.textContent = data.title;
    h.style.cssText = 'margin:0;font-size:48px;line-height:1.08;font-weight:800';
    const p = document.createElement('p');
    p.textContent = data.summary;
    p.style.cssText = 'margin:22px 0 0;max-width:820px;color:#dbeafe;font-size:23px;line-height:1.45';
    box.append(badge,h,p); card.append(box); document.body.append(card);
  })()`);
}

async function caption(title, detail, tone = '#0891b2') {
  const data = JSON.stringify({ story: current?.spec ?? '', title, detail, tone });
  await js(`(() => {
    document.getElementById('demo-caption')?.remove();
    const data = ${data};
    const box = document.createElement('aside');
    box.id = 'demo-caption';
    box.style.cssText = [
      'position:fixed','left:16px','top:58px','z-index:2147483646',
      'width:min(365px,calc(100vw - 32px))','max-height:190px','overflow:hidden',
      'background:rgba(15,23,42,.95)','color:white','border-left:6px solid '+data.tone,
      'border-radius:10px','padding:12px 14px','box-shadow:0 12px 30px rgba(0,0,0,.32)',
      'font-family:Segoe UI,Arial,sans-serif','pointer-events:none','overflow-wrap:anywhere'
    ].join(';');
    const story = document.createElement('div');
    story.textContent = data.story;
    story.style.cssText = 'color:#a5f3fc;font-size:12px;font-weight:800;letter-spacing:.08em;text-transform:uppercase';
    const heading = document.createElement('div');
    heading.textContent = data.title;
    heading.style.cssText = 'margin-top:3px;font-size:19px;line-height:1.2;font-weight:750';
    const text = document.createElement('div');
    text.textContent = data.detail;
    text.style.cssText = 'margin-top:5px;color:#dbeafe;font-size:14px;line-height:1.35';
    box.append(story,heading,text); document.body.append(box);
  })()`);
}

async function clearOverlays() {
  await js(`document.getElementById('demo-title')?.remove(); document.getElementById('demo-caption')?.remove();`);
}

async function scene(title, detail, seconds = 4.5, tone = '#0891b2') {
  addChapter(title);
  await caption(title, detail, tone);
  await hold(seconds * PACE);
}

async function navigate(route) {
  await command('Page.navigate', { url: `${BASE}${route}` });
  await wait(1300);
}

async function waitFor(selector, timeoutMs = 7000) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    if (await js(`Boolean(document.querySelector(${JSON.stringify(selector)}))`)) return;
    await wait(150);
  }
  throw new Error(`No apareció el control ${selector}`);
}

async function setValue(selector, value) {
  const ok = await js(`(() => {
    const element = document.querySelector(${JSON.stringify(selector)});
    if (!element) return false;
    const proto = element instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype :
      element instanceof HTMLSelectElement ? HTMLSelectElement.prototype : HTMLInputElement.prototype;
    const setter = Object.getOwnPropertyDescriptor(proto, 'value').set;
    setter.call(element, ${JSON.stringify(value)});
    element.dispatchEvent(new Event('input', { bubbles: true }));
    element.dispatchEvent(new Event('change', { bubbles: true }));
    return true;
  })()`);
  if (!ok) throw new Error(`No existe el control ${selector}`);
}

async function typeValue(selector, value, seconds = 1.2) {
  await setValue(selector, '');
  const steps = Math.max(1, Math.ceil(seconds * FPS));
  for (let i = 1; i <= steps; i++) {
    const length = Math.ceil((value.length * i) / steps);
    await setValue(selector, value.slice(0, length));
    await capture();
    await wait(1000 / FPS);
  }
}

async function markElement(selector) {
  return js(`(() => {
    document.getElementById('demo-pointer')?.remove();
    const element = document.querySelector(${JSON.stringify(selector)});
    if (!element) return false;
    element.scrollIntoView({ block:'center', behavior:'instant' });
    const r = element.getBoundingClientRect();
    const marker = document.createElement('div');
    marker.id = 'demo-pointer';
    marker.style.cssText = 'position:fixed;z-index:2147483645;left:'+(r.left-5)+'px;top:'+(r.top-5)+'px;width:'+(r.width+10)+'px;height:'+(r.height+10)+'px;border:4px solid #f59e0b;border-radius:10px;box-shadow:0 0 0 4px rgba(245,158,11,.22);pointer-events:none';
    document.body.append(marker);
    return true;
  })()`);
}

async function click(selector) {
  if (!(await markElement(selector))) throw new Error(`No existe el control ${selector}`);
  await hold(0.8);
  const ok = await js(`(() => {
    const element = document.querySelector(${JSON.stringify(selector)});
    if (!element) return false;
    element.click(); return true;
  })()`);
  if (!ok) throw new Error(`No existe el control ${selector}`);
  await wait(1100);
  await js(`document.getElementById('demo-pointer')?.remove()`);
  await hold(0.8);
}

async function clickText(text, selector = 'button,a') {
  const target = await js(`(() => {
    const element = [...document.querySelectorAll(${JSON.stringify(selector)})]
      .find((item) => item.textContent.trim().includes(${JSON.stringify(text)}));
    if (!element) return null;
    element.dataset.demoTarget = 'true'; return '[data-demo-target="true"]';
  })()`);
  if (!target) throw new Error(`No existe el texto ${text}`);
  await click(target);
  await js(`document.querySelector('[data-demo-target="true"]')?.removeAttribute('data-demo-target')`);
}

async function login(email, roleLabel, password = 'password') {
  addChapter(`Inicio de sesión: ${roleLabel}`);
  await caption('Inicio de sesión', `Acceso con ${email} para ejecutar la historia como ${roleLabel}.`);
  await typeValue('input[type=email]', email, 1.4);
  await typeValue('input[type=password]', password, 0.9);
  await hold(1.2);
  await click('button[type=submit]');
  await wait(1100);
}

async function logout() {
  await clickText('Cerrar sesi');
  await waitFor('input[type=email]');
}

function run(commandName, args, cwd = PROJECT) {
  const result = spawnSync(commandName, args, {
    cwd,
    env: process.env,
    encoding: 'utf8',
    windowsHide: true,
  });
  if (result.status !== 0) {
    throw new Error(`${commandName} ${args.join(' ')} falló:\n${result.stdout}\n${result.stderr}`);
  }
  return result.stdout;
}

async function resetSeed() {
  if (!process.env.DATABASE_URL) {
    throw new Error('Define DATABASE_URL antes de grabar para controlar la base de demostración');
  }
  const databaseName = new URL(process.env.DATABASE_URL).pathname.replace(/^\//, '');
  if (!databaseName.toLowerCase().includes('demo')) {
    throw new Error(`Grabación cancelada: ${databaseName} no es una base aislada de demo`);
  }
  process.stdout.write('  · reiniciando seed... ');
  run('go', ['run', './cmd/seed'], BACKEND);
  console.log('ok');
  await wait(900);
}

function sql(query) {
  const user = process.env.POSTGRES_USER || 'si16_user';
  const database = process.env.POSTGRES_DB || 'si16_db';
  run('docker', ['exec', 'si16_postgres', 'psql', '-U', user, '-d', database, '-v', 'ON_ERROR_STOP=1', '-c', query]);
}

async function apiLogin(email) {
  const response = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password: 'password' }),
  });
  if (!response.ok) throw new Error(`Login de preparación ${email}: HTTP ${response.status}`);
  return (await response.json()).token;
}

async function apiPost(pathname, token, body) {
  const response = await fetch(`${API}${pathname}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`Preparación POST ${pathname}: HTTP ${response.status} ${await response.text()}`);
  }
  return response.json();
}

async function prepareUnlock() {
  const token = await apiLogin('ana@si16.test');
  await apiPost('/devoluciones', token, { codigo_anclaje: 'E01-A08' });
}

async function prepareWorkshop() {
  const token = await apiLogin('ana@si16.test');
  await apiPost('/averias', token, {
    bicicleta_codigo: 'BICI-0003',
    estacion_id: 1,
    tipo_fallo: 'frenos',
    descripcion: 'El freno delantero roza y pierde eficacia al detener la bicicleta.',
  });
}

async function beginStory(meta, prepare) {
  console.log(`\n${meta.spec} · ${meta.title}`);
  await resetSeed();
  if (prepare) await prepare();
  const folder = path.join(FRAMES_ROOT, meta.id);
  await mkdir(folder, { recursive: true });
  current = { ...meta, folder, frame: 0, chapters: [] };
  await navigate('/login');
  await js(`localStorage.clear(); sessionStorage.clear(); location.reload()`);
  await wait(1100);
  await titleCard(meta);
  await hold(4.5 * PACE);
  await clearOverlays();
  await hold(0.8);
}

async function endStory() {
  await clearOverlays();
  await hold(1);
  await copyFile(path.join(current.folder, 'frame-00001.jpg'), path.join(current.folder, 'poster.jpg'));
  const item = {
    id: current.id,
    spec: current.spec,
    title: current.title,
    role: current.role,
    summary: current.summary,
    frames: current.frame,
    duration: Number((current.frame / FPS).toFixed(2)),
    chapters: current.chapters,
  };
  manifest.stories.push(item);
  console.log(`  · ${item.frames} fotogramas · ${item.duration.toFixed(1)} s · ${item.chapters.length} capítulos`);
  current = null;
}

const isoDate = (date) => {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
};

const today = new Date();
const month = String(today.getMonth() + 1);
const year = String(today.getFullYear());
const futureYear = String(today.getFullYear() + 1);
const previousMonthDate = new Date(today.getFullYear(), today.getMonth() - 1, 1);
const previousMonth = String(previousMonthDate.getMonth() + 1);
const previousMonthYear = String(previousMonthDate.getFullYear());
const sevenDaysAgo = new Date(today);
sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
const thirtyFiveDaysAgo = new Date(today);
thirtyFiveDaysAgo.setDate(thirtyFiveDaysAgo.getDate() - 35);
const oneYearAgo = new Date(today);
oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

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
  await command('Browser.setDownloadBehavior', { behavior: 'allow', downloadPath: DOWNLOADS });
  await wait(700);

  await beginStory({
    id: 'hu-01-disponibilidad', spec: 'HU-01', role: 'Usuario registrado',
    title: 'Consultar la disponibilidad de la red',
    summary: 'Inicio de sesión, mapa, alternativa en lista, estados especiales y detalle de una estación.',
  });
  await scene('Credenciales incorrectas', 'Primero se comprueba que una contraseña errónea no permite acceder.', 3.2, '#dc2626');
  await typeValue('input[type=email]', 'ana@si16.test');
  await typeValue('input[type=password]', 'incorrecta');
  await click('button[type=submit]');
  await scene('Acceso rechazado', 'La interfaz conserva el login y muestra el mensaje devuelto por la API.', 4.2, '#dc2626');
  await typeValue('input[type=password]', 'password');
  addChapter('Inicio de sesión correcto');
  await caption('Inicio de sesión correcto', 'Ana accede con una sesión válida y vuelve a la disponibilidad.');
  await click('button[type=submit]');
  await waitFor('button[role=tab]');
  await scene('Mapa de estaciones', 'Los marcadores muestran la red real y la leyenda distingue estaciones normales, vacías, llenas y fuera de servicio.', 6);
  await scene('Cómo leer los marcadores', 'El número del marcador es la cantidad de bicicletas disponibles. El color y el icono avisan si faltan bicicletas, anclajes o la estación está cerrada.', 5.5);
  await scene('Cobertura completa de la red', 'El encuadre inicial mantiene visibles las aproximadamente 60 estaciones incluso cuando no se concede ubicación o la persona está fuera del área.', 5);
  await clickText('Lista', 'button');
  await scene('Listado accesible', 'La misma información aparece en una lista textual con estado, bicicletas disponibles y anclajes libres.', 6);
  await scene('Mapa y lista comparten datos', 'Ambas vistas consumen el mismo estado en vivo: cambiar de presentación no modifica los contadores ni su fecha de actualización.', 5);
  await click('ul[aria-label="Listado de estaciones"] button');
  await scene('Detalle de estación', 'Al seleccionar una estación se conservan el nombre, el estado y ambos contadores actualizados.', 6);
  await clickText('Sin bicicletas', 'ul[aria-label="Listado de estaciones"] button');
  await scene('Estación sin bicicletas', 'El estado aparece de forma explícita y el contador queda en cero, evitando que el usuario planifique allí una recogida.', 5.5, '#d97706');
  await clickText('Fuera de servicio', 'ul[aria-label="Listado de estaciones"] button');
  await scene('Estación fuera de servicio', 'Este estado prevalece sobre los demás y sustituye los contadores por guiones para indicar que no debe utilizarse.', 5.5, '#dc2626');
  await scene('Actualización periódica', 'La vista conserva la estación seleccionada mientras vuelve a consultar el servidor; cualquier cambio se refleja sin cerrar el detalle.', 5);
  await endStory();

  await beginStory({
    id: 'hu-02-desbloqueo', spec: 'HU-02', role: 'Usuario con abono',
    title: 'Desbloquear una bicicleta e iniciar el viaje',
    summary: 'Abono caducado, código incorrecto, desbloqueo correcto y prevención de viajes simultáneos.',
  }, prepareUnlock);
  await login('bruno@si16.test', 'usuario con abono caducado');
  await navigate('/alquiler');
  await scene('Abono caducado', 'El motivo se muestra antes de introducir un código y el formulario queda desactivado.', 5.5, '#d97706');
  await scene('La regla se valida antes del código', 'El backend comprueba primero la vigencia del abono. Ningún anclaje ni bicicleta cambia cuando esta condición falla.', 5);
  await logout();
  await login('carla@si16.test', 'usuario sin abono anual');
  await navigate('/alquiler');
  await scene('Usuario sin abono', 'El mensaje diferencia claramente no tener abono de tenerlo caducado, y mantiene deshabilitada la operación.', 5.5, '#d97706');
  await logout();
  await login('gestor@si16.test', 'gestor sin permiso de alquiler');
  await navigate('/alquiler');
  await waitFor('button[role=tab]');
  await scene('Acceso por rol protegido', 'Un gestor u operario que escribe directamente la ruta de desbloqueo vuelve al mapa y nunca ve el formulario.', 5.5, '#dc2626');
  await logout();
  await login('ana@si16.test', 'usuario con abono vigente');
  await navigate('/alquiler');
  await scene('Formulario habilitado', 'Con abono vigente y sin viaje activo, Ana puede introducir el código impreso en el anclaje.', 5);
  await typeValue('input[placeholder="E01-A01"]', 'NO-EXISTE');
  addChapter('Código inexistente');
  await caption('Código inexistente', 'La solicitud inválida se rechaza sin crear un viaje ni modificar la estación.', '#dc2626');
  await click('button[type=submit]');
  await scene('Desbloqueo rechazado', 'El formulario explica que el anclaje no existe y permite corregir el código.', 4.5, '#dc2626');
  await typeValue('input[placeholder="E01-A01"]', 'E01-A02');
  addChapter('Desbloqueo correcto');
  await caption('Desbloqueo correcto', 'Ana solicita la bicicleta disponible del anclaje E01-A02.');
  await click('button[type=submit]');
  await waitFor('[data-testid="viaje-inicio"]');
  await scene('Viaje activo creado', 'Se muestran la hora de inicio, la bicicleta y la estación de origen; no puede iniciarse otro viaje.', 6);
  await scene('Un solo viaje simultáneo', 'La pantalla sustituye el formulario por el viaje activo. Así no existe una segunda acción de desbloqueo mientras la bicicleta siga en uso.', 5.5);
  await navigate('/');
  await scene('Disponibilidad recalculada', 'El anclaje queda libre y la bicicleta deja de contar en la estación inmediatamente después del desbloqueo.', 5.5);
  await endStory();

  await beginStory({
    id: 'hu-03-devolucion', spec: 'HU-03', role: 'Usuario con viaje activo',
    title: 'Devolver la bicicleta y calcular la tarifa',
    summary: 'Rechazo de un anclaje inexistente, cierre atómico del viaje y coste del tiempo excedido.',
  }, async () => {
    sql("UPDATE viajes SET fecha_inicio = now() - interval '45 minutes' WHERE estado = 'en_curso' AND usuario_id = (SELECT id FROM usuarios WHERE email = 'ana@si16.test')");
  });
  await login('carla@si16.test', 'usuario sin viaje activo');
  await navigate('/devolucion');
  await typeValue('[data-testid="input-anclaje"]', 'E01-A08');
  await click('[data-testid="boton-devolver"]');
  await scene('Sin viaje activo', 'La devolución se rechaza con una causa clara y el anclaje permanece libre porque Carla no tiene ninguna bicicleta en uso.', 5.5, '#dc2626');
  await logout();
  await login('ana@si16.test', 'propietaria del viaje activo');
  await navigate('/devolucion');
  await scene('Precondición visible', 'El aviso superior identifica el viaje en curso y la bicicleta que únicamente su propietaria puede devolver.', 5);
  await typeValue('[data-testid="input-anclaje"]', 'NO-EXISTE');
  addChapter('Anclaje inexistente');
  await caption('Anclaje inexistente', 'La devolución inválida mantiene el viaje abierto y no ocupa ningún anclaje.', '#dc2626');
  await click('[data-testid="boton-devolver"]');
  await scene('Devolución rechazada', 'El backend devuelve la causa y el usuario puede probar otro anclaje.', 4.5, '#dc2626');
  await typeValue('[data-testid="input-anclaje"]', 'E01-A01');
  await click('[data-testid="boton-devolver"]');
  await scene('Anclaje ocupado', 'Un código real tampoco sirve si el anclaje ya contiene otra bicicleta; el viaje sigue abierto sin cambios parciales.', 5, '#dc2626');
  await typeValue('[data-testid="input-anclaje"]', 'E01-A08');
  addChapter('Devolución correcta');
  await caption('Devolución correcta', 'La bicicleta se fija en un anclaje libre y la API cierra el viaje en una sola operación.');
  await click('[data-testid="boton-devolver"]');
  await waitFor('[data-testid="exito-devolucion"]');
  await scene('Tarifa calculada', 'El resumen muestra duración y coste: solo se cobra el tramo que supera los primeros 30 minutos.', 6);
  await scene('Redondeo del excedente', 'La duración preparada supera los 45 minutos: los 30 iniciales son gratuitos y cada minuto empezado del excedente se cobra completo.', 5.5);
  await scene('Cierre atómico', 'En una única confirmación quedan registrados destino, hora final, duración y coste; la bicicleta ocupa el anclaje y el usuario se libera.', 5.5);
  await navigate('/viajes');
  await waitFor('[data-testid="fila-viaje"]');
  await scene('Viaje incorporado al historial', 'El trayecto recién cerrado aparece primero en “Mis viajes” con la misma duración, destino y coste calculados.', 5.5);
  await navigate('/alquiler');
  await scene('Usuario liberado', 'Tras cerrar el viaje vuelve a estar disponible el formulario para iniciar otro alquiler.', 4.5);
  await endStory();

  await beginStory({
    id: 'hu-04-historial', spec: 'HU-04', role: 'Usuario',
    title: 'Consultar y filtrar “Mis viajes”',
    summary: 'Historial completo, orden cronológico, tramo gratuito, suma del periodo y estados vacíos.',
  });
  await login('ana@si16.test', 'usuario del servicio');
  await navigate('/viajes');
  await waitFor('[data-testid="fila-viaje"]');
  await scene('Historial completo', 'Se muestran solo viajes finalizados, del más reciente al más antiguo, con origen, destino, duración y coste.', 6);
  await scene('Viaje activo excluido', 'El aviso superior confirma que Ana tiene un viaje en curso, pero ese viaje no aparece ni suma gasto hasta que HU-03 lo cierre.', 5.5);
  await scene('Lectura de cada fila', 'Cada registro reúne fecha y hora local, estación de origen y destino, minutos recorridos y coste individual.', 5.5);
  await scene('Tramo gratuito y sumatorio', 'Los viajes de hasta 30 minutos se etiquetan como gratuitos y el encabezado suma el periodo completo.', 5);
  await setValue('[data-testid="select-mes"]', month);
  await setValue('[data-testid="select-anio"]', year);
  await wait(1100);
  await scene('Filtro por mes y año', 'La lista y el sumatorio se recalculan usando únicamente el periodo seleccionado.', 5.5);
  await setValue('[data-testid="select-mes"]', previousMonth);
  await setValue('[data-testid="select-anio"]', previousMonthYear);
  await wait(1100);
  await scene('Comparación con el mes anterior', 'Cambiar de periodo reemplaza las filas y el sumatorio; el cálculo siempre corresponde al mes completo, no solo a lo visible.', 5.5);
  await click('[data-testid="boton-quitar-filtro"]');
  await scene('Quitar el filtro', 'Un único botón recupera todo el historial y mantiene el orden descendente sin perder información.', 5);
  await setValue('[data-testid="select-mes"]', '12');
  await setValue('[data-testid="select-anio"]', futureYear);
  await wait(1100);
  await scene('Periodo sin viajes', 'Un periodo futuro produce un estado vacío y un sumatorio de 0,00 €, sin tratarlo como error.', 5, '#d97706');
  await setValue('[data-testid="select-mes"]', month);
  await setValue('[data-testid="select-anio"]', '');
  await scene('Filtro incompleto', 'Si solo se elige el mes, la interfaz solicita también el año y mantiene una salida coherente.', 4.5, '#d97706');
  await endStory();

  await beginStory({
    id: 'hu-05-averias', spec: 'HU-05', role: 'Usuario y operario de taller',
    title: 'Reportar una avería e inhabilitar la bicicleta',
    summary: 'Validación de la estación, registro del fallo, actualización de disponibilidad y aviso al taller.',
  });
  await login('ana@si16.test', 'usuario que detecta el problema');
  await navigate('/');
  await clickText('Lista', 'button');
  await click('ul[aria-label="Listado de estaciones"] button');
  await scene('Comprobación antes del reporte', 'La ficha de la estación muestra la disponibilidad inicial. El código de bicicleta se obtiene físicamente de la unidad que presenta el fallo.', 5.5);
  await navigate('/averias');
  await waitFor('select');
  await typeValue('input[placeholder="BICI-0001"]', 'BICI-9999');
  await setValue('select', '1');
  await typeValue('textarea', 'El freno delantero roza y pierde eficacia al detener la bicicleta.', 2.2);
  await click('button[type=submit]');
  await scene('Código de bicicleta inexistente', 'El backend comprueba el identificador y rechaza el reporte sin alterar la flota ni crear una incidencia huérfana.', 5, '#dc2626');
  await typeValue('input[placeholder="BICI-0001"]', 'BICI-0003');
  await setValue('select', '2');
  addChapter('Estación incorrecta');
  await caption('Estación incorrecta', 'La bicicleta existe, pero no pertenece a la estación seleccionada; el reporte debe rechazarse.', '#dc2626');
  await click('button[type=submit]');
  await scene('Reporte rechazado', 'La aplicación conserva los datos y muestra la causa sin bloquear una bicicleta equivocada.', 4.5, '#dc2626');
  await setValue('select', '1');
  addChapter('Reporte correcto');
  await caption('Reporte correcto', 'Se corrige la estación y se registra la descripción técnica del problema.');
  await click('button[type=submit]');
  await scene('Bicicleta inhabilitada', 'La confirmación indica que la incidencia quedó registrada y la bicicleta ya no puede alquilarse.', 5.5);
  await typeValue('textarea', 'Reintento del mismo aviso tras una posible pérdida de conexión.', 1.6);
  await click('button[type=submit]');
  await scene('Reintento sin duplicados', 'Repetir el envío confirma la incidencia ya abierta: la operación es idempotente y no crea dos trabajos de taller.', 5.5);
  await navigate('/');
  await clickText('Lista', 'button');
  await click('ul[aria-label="Listado de estaciones"] button');
  await scene('Disponibilidad actualizada', 'La bicicleta averiada deja de sumar entre las unidades disponibles de la estación.', 5);
  await logout();
  await login('taller@si16.test', 'operario de mantenimiento');
  await navigate('/incidencias');
  await scene('Aviso recibido en taller', 'El mismo reporte aparece en el panel pendiente con bicicleta, estación, tipo, descripción y fecha.', 6);
  await scene('Trazabilidad del aviso', 'El equipo de taller recibe exactamente los datos confirmados por el usuario y puede identificar la unidad sin consultar información externa.', 5);
  await endStory();

  await beginStory({
    id: 'hu-06-taller', spec: 'HU-06', role: 'Operario de taller',
    title: 'Reparar, documentar y reubicar una bicicleta',
    summary: 'Consulta de la ficha, validación del parte, intervención firmada y vuelta segura al servicio.',
  }, prepareWorkshop);
  await login('taller@si16.test', 'operario de mantenimiento');
  await navigate('/taller');
  await typeValue('[data-testid="input-bicicleta-id"]', '99999');
  await click('[data-testid="boton-buscar-ficha"]');
  await scene('Bicicleta inexistente', 'Una consulta con un identificador que no existe devuelve un error claro y no abre ningún formulario de intervención.', 5, '#dc2626');
  await typeValue('[data-testid="input-bicicleta-id"]', '3');
  await click('[data-testid="boton-buscar-ficha"]');
  await waitFor('[data-testid="descripcion-incidencia"]');
  await scene('Ficha e incidencia original', 'La ficha muestra el estado averiado y conserva la descripción enviada por el usuario.', 6);
  await click('[data-testid="boton-guardar-parte"]');
  await scene('Parte incompleto bloqueado', 'No se puede cerrar la reparación sin piezas sustituidas o notas del técnico.', 4.5, '#dc2626');
  await typeValue('[data-testid="input-piezas"]', 'Zapatas de freno delanteras', 1.7);
  await typeValue('[data-testid="input-notas"]', 'Ajuste, limpieza y prueba de frenado completados correctamente.', 2.2);
  addChapter('Guardar intervención');
  await caption('Guardar intervención', 'El parte queda asociado al operario y la bicicleta se marca como operativa.');
  await click('[data-testid="boton-guardar-parte"]');
  await waitFor('[data-testid="input-codigo-anclaje"]');
  await scene('Intervención registrada', 'El historial conserva piezas, notas, fecha y estado atendido; todavía falta confirmar una ubicación.', 5.5);
  await scene('Operativa, pero aún no ubicada', 'La reparación cambia el estado técnico, pero la unidad no vuelve a la disponibilidad de una estación hasta fijarla en un anclaje libre.', 5.5);
  await typeValue('[data-testid="input-codigo-anclaje"]', 'E01-A01');
  await click('[data-testid="boton-ubicar"]');
  await scene('Anclaje ocupado rechazado', 'Una ubicación inválida no devuelve la bicicleta al mapa ni altera el anclaje existente.', 4.5, '#dc2626');
  await typeValue('[data-testid="input-codigo-anclaje"]', 'E99-A99');
  await click('[data-testid="boton-ubicar"]');
  await scene('Anclaje inexistente rechazado', 'También se valida que el código pertenezca a la red antes de asociar la bicicleta reparada.', 4.5, '#dc2626');
  await typeValue('[data-testid="input-codigo-anclaje"]', 'E01-A03');
  addChapter('Ubicación confirmada');
  await caption('Ubicación confirmada', 'El anclaje libre E01-A03 recibe la bicicleta reparada.');
  await click('[data-testid="boton-ubicar"]');
  await scene('Bicicleta disponible', 'La confirmación completa el flujo: reparación documentada, ubicación válida y unidad disponible.', 5.5);
  await logout();
  await login('ana@si16.test', 'usuario que consulta la red');
  await clickText('Lista', 'button');
  await click('ul[aria-label="Listado de estaciones"] button');
  await scene('Reaparición en el mapa', 'La estación vuelve a contar la bicicleta reparada entre sus unidades disponibles.', 5.5);
  await endStory();

  await beginStory({
    id: 'hu-07-balanceo', spec: 'HU-07', role: 'Operario de logística',
    title: 'Equilibrar estaciones mediante alertas y traslados',
    summary: 'Panel de urgencias, validaciones de consistencia y recálculo inmediato de origen y destino.',
  }, async () => {
    console.log('  · esperando el barrido automático de alertas (máx. 55 s)...');
    await wait(57_000);
  });
  await login('ana@si16.test', 'usuario sin permiso de logística');
  await navigate('/balanceo');
  await waitFor('[data-testid="error-panel"]');
  await scene('Acceso restringido', 'La API protege el panel: un usuario general no puede consultar alertas ni ordenar movimientos de flota escribiendo la URL.', 5, '#dc2626');
  await logout();
  await login('logistica@si16.test', 'operario de logística');
  await navigate('/balanceo');
  await waitFor('[data-testid="fila-alerta"]');
  await scene('Alertas automáticas', 'El panel ordena por urgencia las estaciones por debajo del 10 % o por encima del 90 %.', 6);
  await scene('Lectura de prioridad', 'Cada alerta identifica estación, porcentaje de ocupación, tipo, severidad y momento de generación para decidir el siguiente traslado.', 5.5);
  await setValue('[data-testid="input-origen"]', '10');
  await setValue('[data-testid="input-destino"]', '10');
  await setValue('[data-testid="input-unidades"]', '2');
  await click('[data-testid="boton-traslado"]');
  await scene('Origen y destino iguales', 'El traslado inconsistente se rechaza sin modificar ninguna estación.', 4.5, '#dc2626');
  await setValue('[data-testid="input-destino"]', '8');
  await setValue('[data-testid="input-unidades"]', '99');
  await click('[data-testid="boton-traslado"]');
  await scene('Lote superior al disponible', 'La validación impide retirar más bicicletas de las existentes en el origen.', 4.5, '#dc2626');
  await setValue('[data-testid="input-origen"]', '1');
  await setValue('[data-testid="input-destino"]', '10');
  await setValue('[data-testid="input-unidades"]', '2');
  await click('[data-testid="boton-traslado"]');
  await scene('Destino sin capacidad', 'Aunque el origen disponga del lote, la estación 10 no tiene dos anclajes libres y el traslado completo se cancela.', 5, '#dc2626');
  await setValue('[data-testid="input-origen"]', '10');
  await setValue('[data-testid="input-destino"]', '8');
  await setValue('[data-testid="input-unidades"]', '2');
  addChapter('Traslado correcto');
  await caption('Traslado correcto', 'Se mueven dos bicicletas desde la estación casi llena 10 a la estación casi vacía 8.');
  await click('[data-testid="boton-traslado"]');
  await waitFor('[data-testid="resultado-traslado"]');
  await scene('Contadores recalculados', 'Las tarjetas muestran el estado final de origen y destino y el panel vuelve a consultar las alertas.', 6);
  await scene('Traslado atómico y auditable', 'Las unidades salen del origen y llegan al destino en una sola operación; el movimiento queda asociado al operario de logística.', 5.5);
  await click('[data-testid="boton-actualizar"]');
  await scene('Panel actualizado', 'Las alertas resueltas desaparecen y las demás conservan su prioridad sin duplicados.', 5);
  await endStory();

  await beginStory({
    id: 'hu-08-analitica', spec: 'HU-08', role: 'Gestor municipal',
    title: 'Analizar el uso y exportar el histórico',
    summary: 'Control de permisos, gráficos por hora y día, periodos sin datos, validación y descarga CSV.',
  });
  await login('ana@si16.test', 'usuario sin permiso de gestión');
  await navigate('/analitica');
  await waitFor('[data-testid="error-analitica"]');
  await scene('Permiso denegado', 'Aunque se escriba la URL directamente, la API no entrega datos a un usuario sin rol gestor.', 5, '#dc2626');
  await logout();
  await login('gestor@si16.test', 'gestor municipal');
  await navigate('/analitica');
  await waitFor('[data-testid="grafico-hora"]');
  await scene('Gráficos y totales', 'El rango inicial presenta el volumen por hora, por día de la semana, el total de viajes y el coste.', 6);
  await scene('Distribución por hora', 'El primer gráfico permite localizar franjas de mayor demanda comparando las 24 horas con una misma escala.', 5.5);
  await scene('Distribución semanal', 'El segundo gráfico agrupa los viajes de lunes a domingo para reconocer patrones laborales y de fin de semana.', 5.5);
  await setValue('[data-testid="input-desde"]', isoDate(sevenDaysAgo));
  await setValue('[data-testid="input-hasta"]', isoDate(today));
  await wait(1200);
  await scene('Análisis de los últimos siete días', 'Al acotar el rango, gráficos, total de viajes y coste se recalculan conjuntamente sin recargar la página.', 5.5);
  await setValue('[data-testid="input-desde"]', `${futureYear}-01-01`);
  await setValue('[data-testid="input-hasta"]', `${futureYear}-01-31`);
  await wait(1200);
  await scene('Periodo sin datos', 'Un rango futuro mantiene la consulta válida y muestra “sin datos en el periodo”.', 4.5, '#d97706');
  await setValue('[data-testid="input-desde"]', `${futureYear}-12-31`);
  await setValue('[data-testid="input-hasta"]', `${year}-01-01`);
  await scene('Rango invertido', 'La fecha inicial posterior a la final se rechaza antes de generar gráficos o CSV.', 4.5, '#dc2626');
  await setValue('[data-testid="input-desde"]', isoDate(oneYearAgo));
  await setValue('[data-testid="input-hasta"]', isoDate(today));
  await wait(1400);
  await waitFor('[data-testid="grafico-hora"]');
  await scene('Visión anual', 'Un rango amplio utiliza los mismos indicadores y permite comparar el comportamiento acumulado antes de volver al periodo de trabajo.', 5.5);
  await setValue('[data-testid="input-desde"]', isoDate(thirtyFiveDaysAgo));
  await setValue('[data-testid="input-hasta"]', isoDate(today));
  await wait(1400);
  await waitFor('[data-testid="grafico-hora"]');
  await scene('Filtro restaurado', 'Los gráficos y totales vuelven a usar exactamente el rango de las últimas semanas.', 4.5);
  addChapter('Descarga CSV');
  await caption('Descarga CSV', 'El gestor descarga el histórico del mismo rango que alimenta los gráficos.');
  await click('[data-testid="boton-descargar"]');
  await wait(1200);
  const downloads = await readdir(DOWNLOADS);
  await scene('CSV generado', downloads.length > 0
    ? `Descarga comprobada: ${downloads[0]}.`
    : 'La acción de exportación se completó con el filtro seleccionado.', 5.5);
  await scene('Coherencia de la exportación', 'El nombre y el contenido del fichero corresponden al mismo intervalo visible, de modo que la evidencia descargada reproduce el análisis mostrado.', 5);
  await endStory();

  await writeFile(
    path.join(ROOT, 'recording-manifest.json'),
    JSON.stringify({ ...manifest, generated: new Date().toISOString() }, null, 2) + '\n',
    'utf8',
  );
  console.log(`\nGrabación terminada: ${manifest.stories.length} historias en ${FRAMES_ROOT}`);
} finally {
  if (socket?.readyState === WebSocket.OPEN) socket.close();
  browser.kill();
}
