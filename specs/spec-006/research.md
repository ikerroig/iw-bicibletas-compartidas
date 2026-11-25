# Research: spec-006

## Decisiones

1. Se reutiliza incidencias_mantenimientos en lugar de crear otra tabla:
   contiene el reporte de usuario y los campos del parte técnico.
2. La identidad del operario se toma del JWT y nunca del body de la petición.
3. La bicicleta se bloquea con FOR UPDATE antes de modificar incidencias o
   estado, evitando dos partes simultáneos inconsistentes.
4. Al marcar una reparación como operativa, la bicicleta queda disponible pero
   sin anclaje. La ubicación posterior es la que la hace aparecer en el mapa.
5. El mantenimiento preventivo se representa con una incidencia resuelta de
   tipo otro; no necesita estación ni usuario reportante.
6. La operación de ubicación bloquea bicicleta y anclaje dentro de la misma
   transacción y recalcula la estación tras el commit.

## Decisiones descartadas

- No se añade un sistema de autenticación nuevo: se reutilizan los roles y el
  middleware común.
- No se modifica 0001_esquema_base.sql; la compatibilidad del mantenimiento
  preventivo se resuelve en 0003_spec006_mantenimiento_preventivo.sql.
- No se muestran bicicletas reparadas en el mapa antes de confirmar un anclaje.

## Registro de `clarify` y decisiones funcionales

Esta sección documenta las decisiones tomadas al revisar la primera versión
funcional de HU-06. No se amplía el alcance de la historia.

### C-001. Separación entre reparación y ubicación

- **Pregunta**: ¿una reparación marcada como operativa debe colocar también la
  bicicleta en un anclaje?
- **Respuesta**: no. La reparación cierra la incidencia y deja la bicicleta en
  estado `disponible` sin anclaje; la ubicación se confirma en una operación
  posterior.
- **Justificación**: FR-007 exige que una bicicleta sin anclaje no aparezca en
  el mapa ni en los contadores. Separar ambas acciones evita publicar una
  ubicación que el operario todavía no ha confirmado.
- **Criterio verificable**: Dado un parte completado, cuando se marca la
  bicicleta como operativa, entonces se guarda el parte pero no se incrementan
  los contadores hasta confirmar un anclaje libre.

### C-002. Campos del parte técnico

- **Pregunta**: ¿qué campos debe completar el operario para guardar un parte?
- **Respuesta**: se exige al menos piezas sustituidas o notas del taller. El
  reporte original de la incidencia conserva la avería detectada y se muestra
  junto al parte en el historial.
- **Justificación**: el modelo compartido no necesita duplicar el diagnóstico
  original; el parte añade la información técnica y queda firmado por la
  sesión del operario.
- **Criterio verificable**: Dado un formulario sin piezas ni notas, cuando se
  intenta guardar, entonces responde `422 ENTRADA_INVALIDA` y no cambia ni el
  estado ni el historial.

### C-003. Mantenimiento preventivo

- **Pregunta**: ¿cómo se conserva una intervención sobre una bicicleta que no
  tiene una avería abierta?
- **Respuesta**: se crea una incidencia resuelta de tipo `otro`, con la
  descripción `Mantenimiento preventivo`, fecha, operario y notas del parte.
- **Justificación**: se reutiliza el historial común sin crear una tabla nueva;
  una intervención preventiva no necesita estación ni usuario reportante.
- **Criterio verificable**: Dada una bicicleta operativa, cuando el taller
  guarda un parte preventivo válido, entonces aparece una entrada resuelta y
  la disponibilidad existente no se altera.

### C-004. Reglas para ubicar una bicicleta

- **Pregunta**: ¿qué condiciones debe cumplir el anclaje indicado?
- **Respuesta**: debe existir, estar libre, pertenecer a una estación operativa
  y la bicicleta debe estar disponible sin ubicación previa.
- **Justificación**: la ubicación es una transición física que debe ser
  atómica; si falla, la bicicleta y el anclaje conservan sus estados.
- **Criterio verificable**: Dado un anclaje ocupado, inexistente o fuera de
  servicio, cuando se intenta ubicar la bicicleta, entonces se rechaza la
  operación y no se cambia la ubicación.

### C-005. Autorización y concurrencia

- **Pregunta**: ¿quién puede consultar o modificar una ficha y qué ocurre con
  dos partes simultáneos?
- **Respuesta**: solo `operario_taller` puede acceder; la bicicleta se bloquea
  con `FOR UPDATE`, por lo que las operaciones se serializan y no se pierden
  intervenciones ya confirmadas.
- **Justificación**: la identidad del técnico debe proceder del JWT y las
  transacciones deben preservar la consistencia del historial.
- **Criterio verificable**: Dado un usuario normal, cuando accede a taller,
  entonces recibe `403`; dado dos partes concurrentes, entonces cada operación
  se confirma de forma serializada.

## Informe de `analyze`

El análisis compara `spec.md`, `plan.md`, `tasks.md` y la implementación actual.
Los hallazgos identifican evidencias que deben conservarse, sin cambiar el
alcance de HU-06.

| ID | Requisito o criterio | Evidencia | Decisión | Estado |
| --- | --- | --- | --- | --- |
| A-001 | FR-001, FR-004, SC-001 | `GET /taller/bicicletas/{id}` devuelve código, reporte original e historial | Mantener | Cubierto por prueba y recorrido manual |
| A-002 | FR-002, FR-003, FR-006, SC-002, SC-004 | Parte transaccional, operario del JWT, fecha y cierre de incidencias | Mantener | Cubierto por `TestTallerFichaReparacionYUbicacion` |
| A-003 | FR-005, FR-007, SC-003 | Ubicación transaccional y recálculo de estación después del commit | Mantener | Cubierto por prueba y recorrido manual |
| A-004 | mantenimiento preventivo | Incidencia resuelta de tipo `otro` sin estación | Mantener | Cubierto por `TestTallerMantenimientoPreventivo` |
| A-005 | permisos y entradas inválidas | Middleware de rol y validación de parte vacío | Mantener | Cubierto por `TestTallerValidaRolYCampos` |
| A-006 | anclaje ocupado, inexistente o fuera de servicio | Errores diferenciados en servicio y handler | Mantener; conservar evidencia de error | Cubierto por `TestTallerUbicacionErroresAnclaje` (T020, 2026-11-25) |

Las filas A-001–A-005 están respaldadas por pruebas automatizadas o por el
recorrido manual. A-006 quedó señalado en el primer análisis porque ninguna
prueba automática comprobaba los errores de anclaje; se cerró con T020 (véase
el informe de `converge`).

## Informe de `converge`

El estado de referencia para esta revisión es el commit `ea9aad5` de la rama
`spec-006`. La comparación código ↔ especificación confirma la ficha, el parte
firmado, el mantenimiento preventivo, el cierre de incidencias, la ubicación
transaccional y la separación entre estado disponible y anclaje.

La verificación previa no detectó brechas funcionales nuevas. La revisión
independiente quedó registrada en la PR #16: `acf91-ua` aprobó el commit
`1fd6ebc` el 2026-11-20. La matriz global conserva el criterio revisado, su
tarea, prueba, responsable y revisor.

La convergencia final se repitió el 2026-11-25 sobre la versión de entrega. La
suite completa de Go, las 98 pruebas de frontend, `go vet` y el build de Vite
terminaron correctamente. Al revisar la entrega se vio que A-006 seguía sin
prueba automática: la revisión de la PR #16 comprobó FR-005/FR-007, no los
escenarios de error. Se añadió la tarea correctora T020, que incorpora
`TestTallerUbicacionErroresAnclaje`: anclaje ocupado (409
`ANCLAJE_OCUPADO`), fuera de servicio (409 `ANCLAJE_FUERA_DE_SERVICIO`) e
inexistente (404 `ANCLAJE_NO_ENCONTRADO`), sin que la bicicleta quede
anclada. El código ya cumplía y no fue necesario cambiarlo. Con T020 no quedan
brechas accionables de HU-06.
