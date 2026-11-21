# DemoSSD_GXX

Esta carpeta contiene la demostración audiovisual de SI-16 exigida en el enunciado.

- `demo-si16.mp4`: vídeo con portada, 24 escenarios de las 8 historias de usuario y cierre.
- `index.html`: página de reproducción con accesos directos por historia y escenario.
- `frames/`: imágenes intermedias usadas para generar el vídeo.
- `generate_demo.py`: script reproducible para regenerar las escenas.

Las vistas del vídeo son ilustraciones de los flujos documentados; no son capturas de la aplicación en ejecución. El vídeo dura 2 min 36 s y está codificado en H.264 a 1920 × 1080.

Para abrir la página localmente se puede abrir `index.html` directamente en el navegador. Si el navegador bloquea el vídeo por restricciones locales, ejecutar desde esta carpeta:

```powershell
python -m http.server 8000
```

y visitar `http://localhost:8000`.
