// Convierte los fotogramas de record_demo.mjs en ocho MP4 y concatena la demo.
// Usa FFMPEG_PATH o, si no se indica, el paquete temporal ffmpeg-static.
import { copyFile, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const ROOT = path.dirname(fileURLToPath(import.meta.url));
const FRAMES = path.join(ROOT, 'recording-frames');
const VIDEO = path.join(ROOT, 'video');
const COMPLETE = path.join(ROOT, 'demo-si16.mp4');
const MANIFEST_PATH = path.join(ROOT, 'recording-manifest.json');
const manifest = JSON.parse(await readFile(MANIFEST_PATH, 'utf8'));

const candidates = [
  process.env.FFMPEG_PATH,
  path.join(process.env.TEMP ?? '', 'codex-demo-tools', 'node_modules', 'ffmpeg-static', 'ffmpeg.exe'),
  path.join(process.env.TEMP ?? '', 'codex-demo-tools', 'node_modules', 'ffmpeg-static', 'ffmpeg'),
].filter(Boolean);
const ffmpeg = candidates.find((candidate) => existsSync(candidate));
if (!ffmpeg) {
  throw new Error('No se encontró ffmpeg. Define FFMPEG_PATH o instala ffmpeg-static en %TEMP%\\codex-demo-tools.');
}

function run(args) {
  const result = spawnSync(ffmpeg, args, { cwd: ROOT, stdio: 'inherit', windowsHide: true });
  if (result.status !== 0) throw new Error(`ffmpeg terminó con código ${result.status}`);
}

await rm(VIDEO, { recursive: true, force: true });
await mkdir(VIDEO, { recursive: true });

for (const story of manifest.stories) {
  const input = path.join(FRAMES, story.id, 'frame-%05d.jpg');
  const output = path.join(VIDEO, `${story.id}.mp4`);
  console.log(`Codificando ${story.spec}: ${story.title}`);
  run([
    '-hide_banner', '-loglevel', 'warning', '-y',
    '-framerate', String(manifest.fps), '-i', input,
    '-vf', `scale=${manifest.width}:${manifest.height}:flags=lanczos`,
    '-c:v', 'libx264', '-preset', 'medium', '-crf', '24',
    '-pix_fmt', 'yuv420p', '-movflags', '+faststart',
    '-r', String(manifest.fps), output,
  ]);
  await copyFile(path.join(FRAMES, story.id, 'poster.jpg'), path.join(VIDEO, `${story.id}.jpg`));
}

const concatPath = path.join(VIDEO, 'lista.txt');
const concatLines = manifest.stories.map((story) => {
  const absolute = path.join(VIDEO, `${story.id}.mp4`).replaceAll('\\', '/').replaceAll("'", "'\\''");
  return `file '${absolute}'`;
});
await writeFile(concatPath, concatLines.join('\n') + '\n', 'utf8');
run([
  '-hide_banner', '-loglevel', 'warning', '-y',
  '-f', 'concat', '-safe', '0', '-i', concatPath,
  '-c', 'copy', '-movflags', '+faststart', COMPLETE,
]);
await rm(concatPath, { force: true });

let start = 0;
const stories = manifest.stories.map((story) => {
  const item = {
    ...story,
    start: Number(start.toFixed(2)),
    video: `video/${story.id}.mp4`,
    poster: `video/${story.id}.jpg`,
  };
  start += story.duration;
  return item;
});
const data = {
  complete: 'demo-si16.mp4',
  duration: Number(start.toFixed(2)),
  generated: manifest.generated,
  stories,
};
await writeFile(path.join(ROOT, 'datos.js'), `window.DEMO = ${JSON.stringify(data, null, 2)};\n`, 'utf8');

// Decodificación completa: detecta un MP4 truncado o una concatenación dañada.
run(['-hide_banner', '-v', 'error', '-i', COMPLETE, '-f', 'null', '-']);
console.log(`Demo completa: ${COMPLETE} (${start.toFixed(1)} s)`);
