# DemoSSD_G13

Esta carpeta contiene la demostración audiovisual de SI-16 exigida en el
enunciado. Sigue el formato de la demo de referencia: cada historia dispone de
su propio vídeo y todos los segmentos se concatenan en una reproducción
completa.

## Contenido

- `demo-si16.mp4`: vídeo completo de las ocho historias (**42 min 52 s**).
- `video/hu-01-*.mp4` … `video/hu-08-*.mp4`: un vídeo independiente por
  historia, todos iniciados desde la pantalla de login.
- `video/*.jpg`: miniaturas de los ocho recorridos.
- `index.html` y `datos.js`: reproductor con modo completo/por historia,
  navegación por escenarios, velocidad y descargas.
- `record_demo.mjs`: automatización reproducible de Edge mediante CDP.
- `build_videos.mjs`: codificación H.264 y concatenación mediante FFmpeg.
- `recording-manifest.json`: duración y capítulos obtenidos de la grabación.

Los **101 capítulos** muestran la interfaz real conectada a la API Go,
PostgreSQL y Redis: login, permisos, mapa/lista/detalle, desbloqueo, devolución
y tarifa, historial y filtros, averías, taller, balanceo, analítica y CSV. Se
incluyen recorridos satisfactorios y rechazos importantes de cada historia.

## Reproducción

Se puede abrir `index.html` directamente. Si el navegador restringe la carga
de vídeos desde `file://`, se debe servir esta carpeta con cualquier servidor
HTTP estático y abrir la URL local. Por ejemplo, con Node.js y el paquete
`http-server`:

```powershell
npx http-server . -p 8000
```

Después se visita `http://localhost:8000`.

Todos los vídeos son MP4 H.264, 1280 × 720, 4 fotogramas por segundo. Esta
cadencia es deliberada: la demo se compone principalmente de formularios y
estados de interfaz, por lo que mantiene los textos legibles y reduce mucho el
tamaño del entregable.
