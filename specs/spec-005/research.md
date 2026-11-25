# Research: spec-005

## Decisiones

1. Se reutiliza `incidencias_mantenimientos`, ya presente en el modelo común.
   La estación del reporte se añade con la migración `0002_spec005_estacion_incidencia.sql`;
   no se modifica la migración inicial compartida.
2. El reporte usa una transacción PostgreSQL y bloquea la fila de la bicicleta
   con `FOR UPDATE`. Así el INSERT de la incidencia y el estado `averiada` se
   confirman juntos o se deshacen juntos.
3. Solo existe una incidencia abierta por bicicleta mediante una consulta bajo
   bloqueo y un índice único parcial. Un reintento devuelve la incidencia
   existente y no repite la transición.
4. El usuario reporta con el código de bicicleta y la estación seleccionada.
   El servicio de taller consulta las incidencias abiertas con bicicleta,
   estación, descripción, tipo y fecha.
5. Después del commit se llama a `RecalcularEstacion`. Si Redis no responde,
   el cambio de PostgreSQL se conserva y los mecanismos de refresco/fallback
   existentes recuperan el estado.

## Decisiones descartadas

- No se edita `0001_esquema_base.sql`, porque es la base acordada por el equipo.
- No se permite reportar una bicicleta `en_taller`: el operario ya controla
  esa unidad y el reporte de usuario no debe alterar el flujo de reparación.

## Decisiones añadidas durante la implementación

6. La estación enviada por el usuario se contrasta con el anclaje actual de la
   bicicleta dentro de la transacción. Esto evita asociar una avería a una
   estación diferente por un código mal introducido o por datos obsoletos.
7. El panel del taller se refresca cada 10 segundos, por debajo del límite de
   30 segundos indicado para la HU-05.

## Registro de `clarify` y decisiones funcionales

Esta sección conserva las decisiones tomadas al revisar la primera versión
funcional de HU-05. No se añade alcance nuevo; se concretan reglas que la
narrativa dejaba abiertas.

### C-001. Estación válida para una bicicleta anclada

- **Pregunta**: ¿qué ocurre si el usuario selecciona una estación distinta de
  aquella en la que está anclada la bicicleta?
- **Respuesta**: el reporte se rechaza con `409 BICICLETA_EN_OTRA_ESTACION` y
  la bicicleta conserva su estado anterior.
- **Justificación**: el reporte debe describir dónde se encuentra la bicicleta;
  aceptar otra estación produciría una incidencia incoherente para el taller.
- **Criterio verificable**: Dado que la bicicleta está anclada en A, cuando se
  reporta indicando B, entonces se rechaza la petición y no cambia ni el estado
  de la bicicleta ni el número de incidencias abiertas.

### C-002. Reporte durante un viaje

- **Pregunta**: ¿debe exigirse una coincidencia con una estación si la
  bicicleta está en un viaje activo?
- **Respuesta**: no. Una bicicleta en viaje no tiene anclaje actual; se admite
  la estación declarada y se bloquea la bicicleta para nuevos alquileres.
- **Justificación**: la HU-05 permite reportar una avería detectada durante el
  viaje y la devolución se resuelve en HU-02/HU-03.
- **Criterio verificable**: Dado un viaje activo, cuando el usuario envía un
  reporte válido, entonces se registra una única incidencia y la bicicleta
  pasa a `averiada`.

### C-003. Reintentos y reportes duplicados

- **Pregunta**: ¿qué respuesta recibe el usuario si reenvía el mismo reporte
  tras perder la conexión?
- **Respuesta**: la API devuelve `200`, indica `repetido: true` y devuelve la
  incidencia abierta existente; no crea otra incidencia.
- **Justificación**: el reintento debe ser seguro y no puede duplicar el
  bloqueo de la bicicleta.
- **Criterio verificable**: Dado un reporte ya registrado, cuando se repite la
  misma petición, entonces se conserva una sola incidencia abierta.

### C-004. Actualización del panel de taller

- **Pregunta**: ¿cómo se cumple el refresco del panel en menos de 30 segundos?
- **Respuesta**: la interfaz consulta de nuevo el listado cada 10 segundos y
  mantiene además el botón de actualización manual.
- **Justificación**: 10 segundos queda por debajo del límite de 30 segundos sin
  introducir WebSockets ni cambiar el contrato existente.
- **Criterio verificable**: Dado que el panel está abierto, cuando pasan 10
  segundos, entonces se realiza una nueva consulta; el botón manual también
  permite actualizarlo inmediatamente.

## Informe de `analyze`

El primer análisis compara `spec.md`, `plan.md`, `tasks.md` y la implementación
actual. Los hallazgos no cambian el alcance de HU-05.

| ID | Requisito o criterio | Evidencia | Decisión | Estado |
| --- | --- | --- | --- | --- |
| A-001 | FR-001, FR-006 | Formulario, handler `POST /averias` y respuesta 201/200 | Mantener | Cubierto por prueba y recorrido manual |
| A-002 | FR-002, FR-003, FR-007, SC-001, SC-004 | Transacción con `FOR UPDATE`, incidencia única y `TestAveriaBloqueaYEsIdempotente` | Mantener | Cubierto por código y prueba |
| A-003 | FR-004, SC-003 | Recálculo de estación, mapa y listas compartidos | Mantener | Cubierto: la estación pasó de 5 a 4 bicicletas disponibles |
| A-004 | FR-005, SC-002 | `GET /incidencias`, panel de taller y control de rol | Mantener | Cubierto: el panel mostró bicicleta, estación, tipo, descripción y fecha |
| A-005 | FR-008 | Intervalo de 10 segundos y botón manual en `pages/incidencias` | Mantener | Cubierto por inspección de `setInterval`, ruta y botón manual |
| A-006 | permisos y estados incompatibles | Pruebas de rol de taller y rechazo de bicicleta en otra estación | Mantener | Cubierto por pruebas backend |

La matriz anterior se conserva como punto de partida del análisis. Las tareas
T015–T019 de `tasks.md` cierran las evidencias que faltan y permiten que otro
integrante revise al menos un criterio de esta historia.

## Informe de `converge`

El estado de referencia para esta revisión es el commit `4343015` de la rama
`spec-005`. La comparación código ↔ especificación confirma que las reglas
principales están implementadas: reporte transaccional, bloqueo de la bicicleta,
idempotencia, control de estación, autorización del panel y refresco cada 10
segundos.

La verificación previa no detectó brechas funcionales nuevas. La revisión
independiente quedó registrada en la PR #15: `acf91-ua` aprobó el commit
`be76c22` el 2026-11-20. La matriz global conserva el criterio revisado, su
tarea, prueba, responsable y revisor.

La convergencia final se repitió el 2026-11-25 sobre la versión de entrega. La
suite completa de Go, las 98 pruebas de frontend, `go vet` y el build de Vite
terminaron correctamente. No quedan brechas accionables de HU-05; no fue
necesario añadir tareas correctoras.
