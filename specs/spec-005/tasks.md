# Tasks: spec-005

- [x] T001 Revisar el modelo compartido y la especificación de HU-05.
- [x] T002 Añadir `estacion_id` e índice único de incidencias abiertas en la
      migración `0002_spec005_estacion_incidencia.sql`.
- [x] T003 Ampliar `models.Incidencia` con estación y datos de lectura.
- [x] T004 Implementar el servicio transaccional de reporte y el listado del
      panel de taller.
- [x] T005 Implementar handlers, validación y autorización por rol.
- [x] T006 Registrar rutas principales y alias compatibles.
- [x] T007 Implementar el cliente API y el formulario de usuario.
- [x] T008 Implementar el panel de incidencias para operarios de taller.
- [x] T009 Ejecutar las pruebas Go de integración con PostgreSQL/Redis
      disponibles en Docker.
- [x] T010 Ejecutar `npm test`, `npm run build` y la comprobación manual del
      recorrido de HU-05 cuando estén instaladas las dependencias locales.
- [x] T011 Validar que una bicicleta anclada no se pueda reportar desde otra
      estación y que la respuesta conserve el estado original de la bicicleta.
- [x] T012 Actualizar automáticamente el panel de taller cada 10 segundos y
      conservar la actualización manual.

## Verificación y cierre

- [x] T013 Ejecutar las pruebas de integración de averías y comprobar los
      casos de alta, reintento, autorización y estación incorrecta.
- [x] T014 Ejecutar las pruebas del frontend, la compilación de producción y
      las rutas `/`, `/averias` y `/incidencias` con la aplicación levantada.

## Refinamiento SDD y revisión cruzada

- [x] T015 Registrar en `research.md` las decisiones de `clarify` sobre
      estación real, viajes activos, reintentos idempotentes y refresco del
      panel, expresadas con criterios Dado/Cuando/Entonces.
- [x] T016 Elaborar el informe inicial de `analyze` y la matriz de trazabilidad
      de FR-001…FR-008 y SC-001…SC-004 en `research.md` y `plan.md`.
- [x] T017 Repetir las pruebas backend de HU-05 con PostgreSQL y Redis
      disponibles: las tres pruebas `TestAveria*` pasaron el 2026-11-20,
      cubriendo alta, reintento, rol del panel y estación incorrecta.
- [x] T018 Ejecutar el recorrido manual de mapa y panel el 2026-11-20:
      `BICI-0001` pasó de 5 a 4 bicicletas disponibles en la estación, el
      panel mostró todos los datos de la incidencia y `/averias` e
      `/incidencias` respondieron 200; el refresco de 10 segundos y el botón
      manual se comprobaron en `pages/incidencias`.
- [ ] T019 Realizar la revisión de compañero: seleccionar al menos un criterio
      de HU-05, comprobar su tarea, evidencia y prueba, y registrar el
      resultado, responsable y revisor.
- [ ] T020 Ejecutar `converge` final tras T017–T019, añadir aquí cualquier
      tarea correctora nueva y repetir implementación y pruebas si aparece una
      brecha real.

## Estado de convergencia

Las tareas T015–T018 documentan el refinamiento y la verificación previa. T019
y T020 permanecen abiertas hasta disponer de la revisión de otro integrante y
de la convergencia final; no se marcan como completadas solo porque el código
ya funcione.
