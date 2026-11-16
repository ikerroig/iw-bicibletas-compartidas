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
