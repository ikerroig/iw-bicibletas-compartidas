# Quickstart: HU-06

1. Arrancar PostgreSQL y Redis con docker-compose up -d.
2. Aplicar migraciones desde backend con go run ./cmd/migrate up.
3. Cargar datos demo con go run ./cmd/seed.
4. Arrancar backend con go run ./cmd/server y frontend con
   cd frontend; npm install; npm run dev.
5. Iniciar sesión como taller@si16.test con contraseña password.
6. Abrir http://localhost:5173/taller, introducir el ID de una bicicleta y
   consultar su ficha.
7. Completar piezas o notas, marcarla como operativa y guardar el parte.
8. Introducir el código de un anclaje libre y confirmar la ubicación.
9. Comprobar en el mapa que la bicicleta vuelve a contar como disponible.

## Comprobaciones de error

- Parte sin piezas ni notas: 422 ENTRADA_INVALIDA.
- Rol usuario en la ficha o en las operaciones: 403.
- Bicicleta en viaje: 409 BICICLETA_EN_USO.
- Anclaje inexistente: 404 ANCLAJE_NO_ENCONTRADO.
- Anclaje ocupado o estación fuera de servicio: 409.
- Bicicleta operativa sin ubicación: no aparece en los contadores hasta ubicarla.

## Validación realizada

- go test ./tests -run 'TestTaller' -count=1 -v: 3 pruebas superadas.
- npm test -- --run --reporter=dot: 11 ficheros y 62 pruebas superadas.
- npm run build: compilación TypeScript y Vite superada.

## Evidencia de refinamiento

- El 2026-11-20, `go test ./tests -run 'TestTaller' -count=1 -v` terminó con
  las tres pruebas de HU-06 superadas.
- `npm test -- --run --reporter=dot` terminó con 11 ficheros y 62 pruebas
  superadas; `npm run build` también terminó correctamente.
- La validación final del 2026-11-25 ejecutó `go test -count=1 ./...`,
  `go vet ./...`, los 16 ficheros y 98 pruebas actuales del frontend y el
  build de producción; todas las comprobaciones terminaron correctamente.

## Recorrido manual de cierre previo a revisión

- Con `taller@si16.test`, la ficha de `BICI-0001` mostró el código y el
  historial vacío inicial.
- Se registró un mantenimiento preventivo con notas. La API devolvió el parte
  resuelto con `operario_id` y `fecha_reparacion`, y la bicicleta siguió
  disponible sin anclaje.
- Se confirmó la ubicación en el anclaje libre `E01-A01`. La respuesta mostró
  la estación y sus contadores; la bicicleta quedó ubicada y volvió a contar
  como disponible.
- La ruta frontend `/taller` respondió `200`.
- En esta máquina Windows PostgreSQL se publicó temporalmente en `5433`, ya
  que el puerto configurado `55432` está bloqueado por el sistema; no se
  modificó `.env`.
