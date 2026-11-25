# Tasks: spec-006

- [x] T001 Revisar HU-06, el modelo común y los contratos globales.
- [x] T002 Añadir la migración para permitir mantenimiento preventivo sin
      estación.
- [x] T003 Implementar la ficha de bicicleta con historial completo.
- [x] T004 Implementar el registro transaccional de partes e intervenciones.
- [x] T005 Resolver incidencias abiertas y firmarlas con operario y fecha.
- [x] T006 Implementar la confirmación de ubicación en anclaje libre.
- [x] T007 Registrar endpoints y autorización exclusiva de taller.
- [x] T008 Implementar cliente API y pantalla frontend de taller.
- [x] T009 Añadir pruebas backend de reparación, preventivo, ubicación y roles.
- [x] T010 Añadir prueba frontend de ficha, parte y ubicación.

## Verificación y cierre

- [x] T011 Ejecutar las pruebas de integración de HU-06 con PostgreSQL y Redis.
- [x] T012 Ejecutar las pruebas frontend y la compilación de producción.
- [x] T013 Comprobar que la aplicación expone la ruta /taller.

## Refinamiento SDD y revisión cruzada

- [x] T014 Registrar en `research.md` las decisiones de `clarify` sobre
      separación reparación/ubicación, parte técnico, preventivo, anclajes y
      concurrencia, con criterios Dado/Cuando/Entonces.
- [x] T015 Elaborar el informe inicial de `analyze` y la matriz de trazabilidad
      de FR-001…FR-007 y SC-001…SC-004 en `research.md` y `plan.md`.
- [x] T016 Repetir las pruebas backend de HU-06 con PostgreSQL y Redis
      disponibles: reparación, preventivo, ubicación, rol y validación.
- [x] T017 Ejecutar el recorrido manual del taller: consultar ficha, guardar
      parte firmado, comprobar bicicleta disponible sin anclaje, ubicarla y
      comprobar mapa/contadores y la ruta `/taller`.
- [x] T018 Registrar la revisión cruzada de HU-06: la PR #16 fue aprobada por
      `acf91-ua` el 2026-11-20 sobre `1fd6ebc`; la matriz global conserva el
      criterio FR-005/FR-007, sus tareas y la prueba verificada.
- [x] T019 Ejecutar `converge` final tras T018. Repetido el 2026-11-25: suite
      Go completa, `go vet`, 98 pruebas frontend y build correctos, sin
      brechas funcionales.
- [x] T020 Tarea correctora de `converge`: cubrir A-006 con
      `TestTallerUbicacionErroresAnclaje` en `backend/tests/taller_test.go`
      (anclaje ocupado → 409, fuera de servicio → 409, inexistente → 404, la
      bicicleta sigue sin anclaje). Pasa el 2026-11-25 sin cambios en el código.

## Estado de convergencia

Las tareas T014–T020 documentan el ciclo completo: decisiones, análisis,
pruebas, recorrido manual, revisión independiente, convergencia final y la
prueba correctora de A-006. No
queda trabajo accionable dentro del alcance de HU-06.
