# Informe de clarify — spec-002 (HU-02: Desbloqueo de bicicleta mediante código de anclaje)

**Responsable**: Iker Roig · **Revisor**: Alex Morell
**Comando**: `/speckit-clarify` · **Spec Kit**: 1.1.3
**Spec**: [`specs/spec-002/spec.md`](../spec.md)
**Commit de referencia** (antes del refinamiento): `0d6f478`
**Diferencia entre la spec anterior y la revisada**: `git diff 0d6f478 -- specs/spec-002/spec.md`

La spec-002 ha pasado por dos sesiones de clarificación. La primera se hizo
antes del `/speckit-plan`. La segunda se hizo dentro de la fase de
refinamiento (enunciado §5.1), con la funcionalidad ya implementada y
**después** de las pruebas previas, en el orden que pide el enunciado.

**Fuente de las justificaciones**: la descripción narrativa del sistema
(`SI-16-bicicletas-compartidas-descripcion-narrativa.pdf`, en adelante
«narrativa») y la constitución del proyecto. Sobre la HU-02, la narrativa
dice que el usuario «introduce en la aplicación el código del anclaje y la
bicicleta se desbloquea», que el sistema «solo lo permite si el abono está
activo y si esa persona no tiene ya otro viaje en curso», y que el sistema
actual «genera muchas reclamaciones». Cuando la narrativa no determina una
única respuesta, se indica como **decisión del grupo** y se explica el motivo.
Ninguna opción se ha elegido porque la aplicación ya funcionase así: en la
sesión del 2026-11-19, cuatro de las cinco respuestas obligan a cambiar el
código.

## Sesión 2026-11-04 (antes del plan)

| # | Pregunta | Respuesta | Justificación | Criterios |
|---|----------|-----------|---------------|-----------|
| 1 | ¿Hasta qué momento exacto vale un abono el día que vence? | Hasta las 23:59:59 de ese día, hora de España | La narrativa habla de un abono «anual» sin precisar la hora. **Decisión del grupo**: el usuario entiende «vence el 4» como «vale todo el día 4», y el servicio es de una ciudad española, así que se usa su zona horaria y no la del servidor | FR-002 |
| 2 | ¿Se acepta el código en minúsculas o con espacios? | Sí, se normaliza | El usuario teclea el código que ve impreso en el anclaje, desde el móvil y en la calle. Rechazar `e01-a01` por las minúsculas sería un error evitable que acabaría en reclamación | FR-001 |
| 3 | ¿Qué datos ve el usuario de su viaje activo? | Hora de inicio, código de bici y estación de origen | La narrativa dice que el sistema «registra la hora de inicio» y que el importe depende de la duración. Mostrarla permite al usuario controlar los 30 minutos gratuitos | FR-008 |
| 4 | ¿«Desacoplar» envía una orden física al anclaje? | No, solo cambia el estado en el sistema | La narrativa no describe ninguna integración con el hardware de las estaciones; añadirla ampliaría el alcance de la HU | FR-005, Supuestos |
| 5 | ¿Quién puede desbloquear? | Solo el rol `usuario` | La narrativa distingue usuarios, operarios y gestores; el desbloqueo es la acción del usuario del servicio, y el personal trabaja con sus propias pantallas | FR-009 |

## Sesión 2026-11-19 (refinamiento, §5.1)

**Pruebas previas**: [pruebas-previas.md](./pruebas-previas.md), con 26
comprobaciones en las tres situaciones (habitual, límite o inválida, y
permisos o estado). Ninguna falla contra la spec anterior, pero dejaron siete
huecos en la propia spec. Cada pregunta sale de uno o dos de ellos.

| # | Pregunta | Respuesta | Justificación | Criterios | ¿Cambia el código? |
|---|----------|-----------|---------------|-----------|--------------------|
| 1 | ¿Qué ve un usuario **sin abono vigente** y cuándo? *(huecos 1 y 6)* | Mensaje distinto para «No tienes abono anual» y «Tu abono anual caducó el DD/MM/AAAA»; además, aviso al abrir la pantalla con el formulario desactivado | El sistema solo lo permite «si el abono está activo», y «muchas reclamaciones vienen de usuarios que no entienden» lo que les pasa. Un mensaje único no dice si hay que contratar o renovar. Avisar al abrir la pantalla evita que el usuario vaya hasta el anclaje para descubrirlo allí (decisión del grupo). El rechazo del servidor se mantiene (constitución I). Se descartó la opción A (como estaba) porque P3.3 mostró que no indica la causa real | FR-002, FR-010, SC-001, escenarios 2 y 4 | Sí (P3.3) |
| 2 | ¿Qué ve el usuario si la **respuesta no llega**? *(hueco 7)* | «No sabemos si se ha desbloqueado; comprobando…», consulta del viaje activo y resultado real; botón para volver a comprobar; nunca se reintenta solo | El desbloqueo ocurre en la calle, con cobertura variable. Si el viaje se creó y la app dice «error», el usuario deja la bici y luego recibe un cargo que no entiende: justo la reclamación que la narrativa quiere evitar. Se descartó reintentar automáticamente (C) porque, si el primero entró, el reintento devuelve «ya tienes un viaje en curso» y confunde más | FR-011, caso límite nuevo | Sí (hoy muestra un error genérico) |
| 3 | ¿Qué pasa si un **gestor u operario** escribe la dirección de la pantalla? *(hueco 3)* | Se le redirige al mapa sin ver el formulario | Coherente con haber ocultado la opción del menú (sesión 2026-11-04, pregunta 5). La seguridad no depende de esto: el servidor ya rechaza a esos roles (P3.7), pero mostrarles un formulario inútil es incoherente (decisión del grupo) | FR-009 | Sí (P3.9) |
| 4 | ¿Se **limitan los intentos** con códigos erróneos? *(hueco 5)* | No; queda fuera de alcance, con el motivo escrito en la spec | Los códigos de anclaje no son secretos: van impresos y siguen un patrón (`E03-A01`). Limitar intentos no protege nada que no sea ya público. El daño real (usar bicis sin abono o varias a la vez) ya lo impiden las reglas de la narrativa: abono activo y un solo viaje | Supuestos | No |
| 5 | Si fallan **varias reglas a la vez**, ¿qué causa se muestra? *(hueco 2)* | Una sola, en orden fijo: abono → viaje activo → anclaje (existe → estación en servicio → anclaje en servicio → tiene bici → bici disponible) | Primero lo que depende de la persona, porque no se arregla cambiando de anclaje: así nadie prueba varios anclajes para acabar descubriendo que su abono ha caducado. Además, no se informa del estado de los anclajes a quien no puede alquilar. Fijar el orden hace que las pruebas de aceptación tengan un único resultado correcto | FR-012 | Sí, en parte: el orden ya coincide (P3.5), pero falta probarlo |

**Hueco no preguntado**: el 4 (mensaje para un código demasiado largo, P2.3).
Es de bajo impacto: el rechazo ya funciona y solo cambia el texto. Se deja
para el analyze, por si conviene unificar los mensajes de validación.
*Resultado*: el analyze lo recogió como hallazgo U2 y el grupo decidió
tratarlo como validación técnica del plan, con un mensaje propio («El código
no es válido»), sin tocar la spec (research D2, opción A del registro de
hallazgos).

**Ampliación posterior de la pregunta 2** (hallazgo I5 del analyze): al
actualizar el plan, FR-011 se amplió para tratar también un **error 5xx del
servidor** como resultado desconocido, no solo la falta de respuesta. No es
una decisión funcional nueva, sino la misma respuesta aplicada a un caso que
la pregunta no nombraba: un 5xx tampoco garantiza que el viaje no se haya
creado, porque el commit puede haberse confirmado (research D12). Consultar el
viaje activo es inofensivo si no se creó. Queda anotado en la spec, junto a la
respuesta.

**Historia afectada**: todas las preguntas afectan solo a la HU-02. Para
cumplir la pregunta 1 el frontend necesita conocer la fecha de vencimiento,
que hoy el modelo compartido `Usuario` no expone (`json:"-"`). Para no tocar
un contrato que usan otras historias (constitución, principio III), en el plan
se resolverá con datos que devuelvan los propios endpoints de la HU-02
(`/alquileres/...`). Si al final hiciera falta cambiar el modelo compartido,
se acordará con el equipo antes.

## Decisión que explica el alumno (§5.1, finalización)

**Iker Roig — pregunta 2 de la sesión del 2026-11-19: qué ve el usuario cuando
la respuesta al desbloqueo no llega (FR-011).**

**Qué duda había.** La spec garantizaba que nunca se crea un viaje a medias,
pero no decía qué ve el usuario cuando la petición se envía y la respuesta no
vuelve. En ese caso hay dos posibilidades que la app no puede distinguir: que
la petición no llegase al servidor (no hay viaje) o que llegase y se perdiera
la respuesta (sí hay viaje y la bici está desbloqueada).

**Qué opciones había.** (A) Mostrar un error genérico, como hasta ahora.
(B) Decir que no se sabe el resultado y comprobarlo consultando el viaje
activo. (C) Reintentar el desbloqueo automáticamente.

**Por qué elegí B.** Con A, si el viaje sí se creó, el usuario cree que ha
fallado, deja la bici y el viaje sigue contando: luego le llega un cargo que no
entiende, que es justo el tipo de reclamación que la narrativa dice que hay que
evitar. Con C, si el primer intento entró, el reintento responde «ya tienes un
viaje en curso», que confunde todavía más. B no adivina: pregunta al servidor,
que es quien tiene el dato real (constitución I), y enseña lo que hay.

**Cómo se comprueba.** El caso aplica si la red falla, si pasan 10 s sin
respuesta o si el servidor devuelve un error 5xx, porque en ninguno de los
tres se sabe si el viaje se creó. La app muestra «No sabemos si se ha
desbloqueado; comprobando…» y consulta el viaje activo una sola vez: si existe
lo enseña, y si no, muestra «No se ha desbloqueado, inténtalo de nuevo». Si la
consulta tampoco responde, aparece el botón «Volver a comprobar», que repite
solo la consulta. Nunca se envía un segundo desbloqueo. Lo cubren los tests de
`resultadoDesconocido.test.tsx` y las pruebas V12a y V12b de
[verificacion-final.md](./verificacion-final.md); en V12b el servidor creó el
viaje, la respuesta se perdió y la app lo mostró tras comprobarlo, con un solo
`POST`.
