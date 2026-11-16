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
