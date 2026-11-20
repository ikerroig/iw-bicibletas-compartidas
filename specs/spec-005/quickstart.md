# Quickstart: HU-05

1. Arrancar PostgreSQL y Redis con `docker compose up -d`.
2. Aplicar las migraciones: `cd backend` y `go run ./cmd/migrate up`.
3. Cargar datos de demostración con `go run ./cmd/seed`.
4. Arrancar el backend con `go run ./cmd/server` y el frontend con
   `cd frontend; npm install; npm run dev`.
5. Entrar como `ana@si16.test` con contraseña `password`, abrir «Reportar
   averia», seleccionar una estación, introducir un código de bicicleta
   anclada y una descripción.
6. Comprobar que la respuesta confirma el reporte y que el mapa deja de contar
   esa bicicleta.
7. Entrar como `taller@si16.test` y abrir «Incidencias»; el reporte debe
   mostrar bicicleta, estación, tipo, descripción y fecha.

## Comprobaciones de error

- Descripción vacía: `422 ENTRADA_INVALIDA`.
- Usuario sin sesión: `401`.
- Rol distinto de usuario en el reporte: `403`.
- Reporte repetido: `200`, sin una segunda incidencia abierta.
- Bicicleta averiada: el desbloqueo responde `409 BICICLETA_NO_DISPONIBLE`.

## Validación realizada

- `go test ./tests -run 'TestAveria' -count=1 -v`: 3 pruebas superadas,
  incluyendo el rechazo de una estación distinta.
- `npm test -- --run --reporter=dot`: 10 ficheros y 61 pruebas superadas.
- `npm run build`: compilación TypeScript y Vite superada.
- Vite respondió `200` en `/`, `/averias` e `/incidencias` durante la
  comprobación temporal del frontend.

## Evidencia de refinamiento

- El frontend se volvió a verificar el 2026-11-20: `npm test -- --run
  --reporter=dot` terminó con 10 ficheros y 61 pruebas superadas; `npm run
  build` también terminó correctamente.
- La repetición backend del 2026-11-20 quedó pendiente porque PostgreSQL no
  estaba escuchando en `127.0.0.1:55432`. La evidencia backend anterior se
  conserva en la sección de validación superior y debe repetirse antes del
  cierre final con Docker activo.
- Los avisos `Not implemented: navigation` de `jsdom` pertenecen a las pruebas
  existentes de descarga CSV de HU-08; no provocaron fallos: las 61 pruebas
  terminaron en verde.

## Recorrido manual de cierre previo a revisión

- El 2026-11-20, con el backend en `8080` y Vite en `5173`, Ana reportó
  `BICI-0001` en la estación 1 con una avería de frenos.
- La API devolvió `201` y la estación pasó de 5 a 4 bicicletas disponibles.
- El panel autenticado como `taller@si16.test` mostró el código de bicicleta,
  estación, tipo, descripción, estado y fecha de la incidencia.
- Las rutas frontend `/averias` e `/incidencias` respondieron `200`. La página
  conserva el refresco automático de 10 segundos y la actualización manual.
- En esta máquina Windows el puerto `55432` está reservado; para esta
  comprobación se publicó PostgreSQL temporalmente en `5433`, sin modificar
  `.env` ni los datos persistentes.
