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
