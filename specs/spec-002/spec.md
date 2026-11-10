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
   informativo de la causa.
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

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST aceptar el código identificador de un anclaje
  como solicitud de desbloqueo de la bicicleta asociada, sin distinguir
  mayúsculas e ignorando los espacios alrededor.
- **FR-002**: El sistema MUST verificar que el usuario tiene el abono anual
  activo; si está inactivo o caducado MUST rechazar la operación indicando la
  causa. El abono es vigente hasta las 23:59:59 (hora de España,
  Europe/Madrid) de su día de vencimiento, incluido.
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
  la opción de desbloqueo MUST NOT aparecerles en el menú.

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
  activo son rechazadas con mensaje de causa.
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
- El alcance es exactamente la HU-02 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
