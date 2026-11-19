# Feature Specification: Desbloqueo de bicicleta mediante código de anclaje

**Feature Branch**: `spec-002`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-02 — Desbloqueo de bicicleta mediante código de anclaje (historia.md, SI-16)."

## Clarifications

### Session 2026-11-04

- Q: Si un abono vence el 4 de noviembre, ¿hasta qué momento exacto puede desbloquear ese usuario? → A: Hasta las 23:59:59 del día de vencimiento, hora de España (Europe/Madrid).
- Q: ¿Debe aceptarse el código del anclaje aunque el usuario lo escriba en minúsculas o con espacios alrededor? → A: Sí; se ignoran las mayúsculas y los espacios alrededor (`e01-a01` equivale a `E01-A01`).
- Q: ¿Qué datos debe ver el usuario de su viaje activo, tanto al desbloquear como al volver a abrir la aplicación? → A: Hora de inicio, código de la bicicleta y nombre de la estación de origen.
- Q: Cuando la spec dice que «la bicicleta se desacopla», ¿el sistema debe enviar una orden real al anclaje físico o basta con registrar el cambio en la aplicación? → A: Solo se registra en el sistema (anclaje libre, bicicleta en uso); la orden física al anclaje queda fuera de alcance.
- Q: ¿Quién puede desbloquear una bicicleta: solo los clientes o también el personal con abono vigente, como operarios y gestores? → A: Solo el rol `usuario`; los demás roles reciben un rechazo aunque tengan abono y no ven la opción en el menú.

### Session 2026-11-19

- Q: Cuando un usuario sin abono vigente intenta desbloquear, ¿qué debe ver y en qué momento? → A: Mensajes distintos según la causa («No tienes abono anual» / «Tu abono anual caducó el DD/MM/AAAA») y, además, aviso del motivo al abrir la pantalla de desbloqueo, con el formulario desactivado; el backend sigue rechazando igualmente.
- Q: Si el usuario pulsa «Desbloquear» y la respuesta del servidor no llega, ¿qué debe ver y cómo sabe si el viaje se ha creado? → A: Mensaje «No sabemos si se ha desbloqueado; comprobando…»; la aplicación consulta el viaje activo y muestra el viaje si existe o «No se ha desbloqueado, inténtalo de nuevo» si no, con un botón para volver a comprobar. No se reintenta el desbloqueo automáticamente.
- Q: Si un gestor u operario escribe directamente la dirección de la pantalla de desbloqueo, ¿qué debe pasar? → A: Se le redirige a su página de inicio (el mapa) sin mostrar el formulario; el backend sigue rechazando la operación.
- Q: ¿Debe limitarse el número de intentos de desbloqueo con códigos erróneos que puede hacer un usuario en poco tiempo? → A: No; queda fuera de alcance porque los códigos de anclaje son públicos y predecibles, y las reglas de abono vigente y viaje único ya limitan el daño.
- Q: Si una solicitud de desbloqueo incumple varias reglas a la vez, ¿qué causa se le muestra al usuario? → A: Una sola causa, en orden fijo: abono → viaje activo → anclaje (existe → estación en servicio → anclaje en servicio → tiene bicicleta → bicicleta disponible).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Inicio de viaje con código de anclaje (Priority: P1)

Un usuario con abono anual en regla introduce en la aplicación el código
identificador del anclaje donde está estacionada la bicicleta que quiere
usar. Si el anclaje tiene una bicicleta disponible y el usuario cumple los
requisitos, la bicicleta se desacopla, se registra la hora de inicio y queda
creado su viaje activo. Un usuario sin abono vigente o con otro viaje en curso
recibe un rechazo inmediato con el motivo.

**Why this priority**: Es la puerta de entrada al servicio de alquiler; sin
desbloqueo no hay viajes ni facturación. Además es el primer punto donde se
aplican las reglas de negocio críticas (abono activo y un solo viaje por
usuario).

**Independent Test**: Con un usuario de prueba con abono activo y sin viajes,
introducir un código de anclaje con bicicleta disponible y verificar que se
crea el viaje con hora de inicio; repetir con abono caducado y con un usuario
ya con viaje activo, verificando ambos rechazos.

**Acceptance Scenarios**:

1. **Given** un usuario con abono activo sin viajes en curso, **When**
   introduce el código de un anclaje con bicicleta disponible, **Then** la
   bicicleta se desacopla, se registra la hora de inicio y se crea un registro
   de viaje activo.
2. **Given** un usuario con abono inactivo o caducado, **When** solicita un
   desbloqueo, **Then** el sistema rechaza la operación y muestra un mensaje
   que distingue «No tienes abono anual» de «Tu abono anual caducó el
   DD/MM/AAAA».
4. **Given** un usuario sin abono vigente, **When** abre la pantalla de
   desbloqueo, **Then** ve el motivo antes de introducir ningún código y el
   formulario aparece desactivado.
5. **Given** un usuario con abono vigente que envía un código válido, **When**
   la respuesta del servidor no llega o es un error interno (error de red,
   error 5xx o más de 10 s), **Then**
   la aplicación muestra «No sabemos si se ha desbloqueado; comprobando…»,
   consulta su viaje activo y le enseña el viaje si se creó o «No se ha
   desbloqueado, inténtalo de nuevo» si no, sin reenviar el desbloqueo.
6. **Given** un gestor u operario con sesión iniciada, **When** escribe la
   dirección de la pantalla de desbloqueo, **Then** se le lleva al mapa sin
   mostrarle el formulario, y una solicitud directa al servidor se rechaza.
7. **Given** un usuario con el abono caducado, **When** introduce un código
   de anclaje que no existe, **Then** el sistema rechaza la operación con la
   causa del abono, porque es la primera regla que falla.
3. **Given** un usuario que ya cuenta con un trayecto iniciado, **When**
   intenta liberar otra bicicleta, **Then** el sistema deniega el acceso e
   impide realizar múltiples viajes simultáneos.

### Edge Cases

- El código introducido no existe o contiene errores: se rechaza sin crear
  viaje ni modificar el anclaje.
- El anclaje apunta a una estación fuera de servicio: se rechaza con causa.
- El anclaje está vacío o la bicicleta asociada está averiada o fuera de
  servicio: se rechaza y no se inicia viaje.
- Dos desbloqueos casi simultáneos del mismo usuario: como máximo uno crea un
  viaje activo.
- El abono caducar entre el inicio de sesión y el desbloqueo: se detecta en el
  momento de la solicitud y se rechaza.
- Un rechazo (abono, viaje previo, anclaje inválido) no altera el estado del
  anclaje, de la bicicleta ni de ningún viaje existente.
- Un usuario con viaje activo pierde la conexión al intentar desbloquear: no
  se crea un viaje parcial; el usuario conserva únicamente su viaje previo.
- La respuesta a un desbloqueo no llega o no permite saber el resultado
  (error de red, error interno del servidor o más de 10 s sin respuesta): la aplicación no da el desbloqueo por fallido ni por hecho;
  comprueba el viaje activo y muestra el resultado real (FR-011).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST aceptar el código identificador de un anclaje
  como solicitud de desbloqueo de la bicicleta asociada, sin distinguir
  mayúsculas e ignorando los espacios alrededor.
- **FR-002**: El sistema MUST verificar que el usuario tiene el abono anual
  activo; si está inactivo o caducado MUST rechazar la operación indicando la
  causa. El abono es vigente hasta las 23:59:59 (hora de España,
  Europe/Madrid) de su día de vencimiento, incluido. La causa MUST distinguir
  «No tienes abono anual» (nunca activado o desactivado) de «Tu abono anual
  caducó el DD/MM/AAAA» (fecha de vencimiento pasada).
- **FR-003**: El sistema MUST verificar que el usuario no tiene ningún viaje
  activo; si lo tiene MUST denegar el desbloqueo, impidiendo viajes
  simultáneos.
- **FR-004**: El sistema MUST verificar que el anclaje existe, está en servicio
  y contiene una bicicleta disponible; en caso contrario MUST
  rechazar con causa.
- **FR-005**: Si todas las condiciones se cumplen, el sistema MUST desacoplar
  la bicicleta, registrar la hora de inicio y crear un registro de viaje
  activo asociado al usuario.
- **FR-006**: El viaje activo MUST permanecer asociado al usuario hasta su
  devolución (HU-03), durante la cual el usuario no puede iniciar otro viaje.
- **FR-007**: Todas las verificaciones y el alta del viaje MUST realizarse de
  forma atómica: nunca puede existir más de un viaje activo por usuario ni un
  viaje activo sin hora de inicio.
- **FR-008**: Tras el desbloqueo, y cada vez que el usuario vuelva a consultar
  la aplicación durante el viaje, el sistema MUST mostrarle la hora de inicio,
  el código de la bicicleta y el nombre de la estación de origen.
- **FR-009**: Solo el rol `usuario` MUST poder desbloquear. Los demás roles
  (operarios y gestores) MUST recibir un rechazo aunque tengan abono vigente, y
  la opción de desbloqueo MUST NOT aparecerles en el menú. Si acceden a la
  pantalla de desbloqueo escribiendo su dirección, MUST ser redirigidos al
  mapa sin ver el formulario.
- **FR-010**: Al abrir la pantalla de desbloqueo, un usuario sin abono vigente
  MUST ver el motivo (con el mismo texto que FR-002) y el formulario MUST
  aparecer desactivado. Este aviso es informativo: el sistema MUST seguir
  rechazando en el momento de la solicitud (FR-002), aunque la pantalla se
  haya abierto con el abono todavía vigente.
- **FR-011**: Si la respuesta a un desbloqueo no llega o no permite saber el
  resultado (error de red, error interno del servidor o más de 10 s sin
  respuesta), la aplicación MUST mostrar «No sabemos si se ha
  desbloqueado; comprobando…» y consultar el viaje activo del usuario: si
  existe, MUST mostrarlo como en FR-008; si no existe, MUST mostrar «No se ha
  desbloqueado, inténtalo de nuevo». Si la consulta tampoco responde, MUST
  ofrecer un botón para volver a comprobar. La aplicación MUST NOT reenviar el
  desbloqueo automáticamente.
- **FR-012**: Si una solicitud incumple varias reglas a la vez, el sistema MUST
  devolver una sola causa, la primera que falle en este orden: abono vigente
  (FR-002) → sin viaje activo (FR-003) → anclaje existe → estación en
  servicio → anclaje en servicio → anclaje con bicicleta → bicicleta
  disponible (FR-004).

### Key Entities

- **Usuario**: Cuenta del servicio; incluye el atributo preexistente de abono
  anual (activo/inactivo) y sus viajes asociados.
- **Anclaje**: Hueco individual de una estación; código identificador único y
  estado (ocupado/libre).
- **Bicicleta**: Unidad de la flota; estado (disponible, en uso, averiada, en taller) y ubicación
  por anclaje cuando está estacionada.
- **Viaje**: Registro de alquiler en curso; usuario, bicicleta, estación de
  origen, hora de inicio y estado (en curso/finalizado); un viaje activo es el que está en curso.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las solicitudes de desbloqueo de usuarios sin abono
  activo son rechazadas con mensaje de causa, y el mensaje distingue «sin
  abono» de «abono caducado» (con la fecha) en el 100 % de los casos.
- **SC-002**: El 100 % de los usuarios con un viaje activo reciben denegación
  ante un segundo desbloqueo; el sistema mantiene 0 usuarios con más de un
  viaje activo.
- **SC-003**: Un desbloqueo válido se confirma en menos de 10 segundos desde
  la introducción del código, con la hora de inicio registrada.
- **SC-004**: El 100 % de los desbloqueos sobre anclajes sin bicicleta
  disponible se rechazan sin efectos secundarios.

## Assumptions

- El abono anual activo/inactivo es un atributo preexistente del usuario en la
  base de datos; su alta y pago quedan fuera del alcance de esta spec (HU no
  incluida entre las 8 validadas).
- Cada anclaje tiene un código identificador único conocido por el usuario
  (impreso en el anclaje).
- Solo las bicicletas en estado disponible son desbloqueables; las averiadas
  quedan inhabilitadas (HU-05).
- La validación de las reglas críticas se exige en el sistema completo; los
  rechazos devuelven mensajes claros de causa al usuario.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- «Desacoplar» es un cambio de estado en el sistema (anclaje libre, bicicleta
  en uso). La orden física al anclaje y sus fallos quedan fuera de alcance.
- No se limita el número de intentos con códigos erróneos: los códigos de
  anclaje no son secretos (van impresos y siguen un patrón predecible) y el
  daño posible ya lo acotan las reglas de abono vigente (FR-002) y de un solo
  viaje activo (FR-003). Queda fuera de alcance.
- El alcance es exactamente la HU-02 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
