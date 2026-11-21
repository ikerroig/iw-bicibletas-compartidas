# DemoSSD_GXX

Esta carpeta contiene la demostración audiovisual de SI-16 exigida en el enunciado.

- `demo-si16.mp4`: grabación funcional de las 8 historias de usuario, con casos correctos y de error.
- `index.html`: página de reproducción con accesos directos por historia y escenario.
- `record_demo.mjs`: script reproducible que recorre la interfaz mediante Chrome CDP.

Todas las vistas proceden de la aplicación en ejecución, conectada al backend
Go y a los datos seed de PostgreSQL y Redis. Los rótulos superpuestos identifican
la historia, el escenario y el resultado observado. El vídeo dura 1 min 15 s y
está codificado en H.264 a 1280 × 720.

La demostración incluye autenticación, mapa y listado, desbloqueo, devolución,
historial, reporte de avería, intervención de taller, balanceo, analítica,
exportación CSV, validaciones y controles de permisos. Los cambios de estado
que aparecen en pantalla son respuestas reales de la API.

Para abrir la página localmente se puede abrir `index.html` directamente en el navegador. Si el navegador bloquea el vídeo por restricciones locales, ejecutar desde esta carpeta:

```powershell
python -m http.server 8000
```

y visitar `http://localhost:8000`.
