# Feature Specification: Consulta de disponibilidad de estaciones en tiempo real

**Feature Branch**: `spec-001`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-01 — Consulta de disponibilidad de estaciones en tiempo real (historia.md, SI-16)."

## Clarifications

### Session 2026-11-03

- Q: ¿Desde qué momento y hasta qué momento se mide el límite de 30 segundos para que un cambio se vea en la pantalla del usuario? → A: Desde que el cambio queda confirmado (commit) en el servidor hasta que el contador nuevo aparece pintado en la pantalla de un usuario con la aplicación abierta y conexión activa (extremo a extremo).
- Q: ¿Cuándo debe avisar la pantalla al usuario de que los datos que ve pueden estar desactualizados? → A: Cuando han pasado más de 30 s desde la última actualización correcta; se muestra un aviso visible de «datos desactualizados» con la hora de la última actualización y los contadores siguen visibles.
- Q: ¿Cómo debe distinguir el mapa una estación sin bicicletas disponibles y una estación sin anclajes libres de una estación normal? → A: En el mapa y en el detalle: el marcador tiene un estado visual distinto para «sin bicicletas» (0 disponibles) y para «sin anclajes libres» (0 libres), no basado solo en el color (también icono o texto); el detalle muestra el contador en 0 con el texto correspondiente.
- Q: ¿Qué debe ver el usuario de una estación fuera de servicio, en el mapa y en su detalle? → A: Aparece en el mapa con un estado visual propio «Fuera de servicio» (no solo color), con prioridad sobre «sin bicicletas»/«sin anclajes libres»; es seleccionable; su detalle muestra «Fuera de servicio» y no muestra contadores («—»), sin presentarla como opción de recogida ni de devolución.
- Q: ¿Cómo debe contarse un anclaje ocupado por una bicicleta averiada: como libre, como disponible o como ninguno de los dos? ¿Y se le dice algo al usuario? → A: No cuenta ni como bicicleta disponible ni como anclaje libre; el detalle muestra solo esos dos contadores, sin desglose de averías, y se acepta que su suma sea menor que la capacidad total.

### Session 2026-11-19

- Q: ¿Qué debe ver el usuario si la primera carga del mapa falla y todavía no hay ningún dato anterior que enseñar? → A: Un mensaje de error claro («No se pudo cargar la disponibilidad») con botón «Reintentar» y reintento automático; no se pintan marcadores ni contadores.
- Q: Además del mapa, ¿debe haber un listado de texto con las estaciones y sus contadores, para quien no pueda usar el mapa o cuando el mapa de fondo no cargue? → A: Sí: un listado de texto siempre disponible (p. ej., pestaña «Lista») con nombre, estado y contadores de cada estación, que sigue funcionando aunque el mapa de fondo no cargue.
- Q: Si el usuario tiene abierto el detalle de una estación y sus datos cambian en el servidor, ¿el detalle se actualiza solo? → A: Sí: el detalle abierto se actualiza solo dentro de los mismos 30 s (contadores y estado, incluido el paso a «Fuera de servicio»), sin cerrarse ni perder la selección.
- Q: ¿Qué pasa si alguien sin sesión iniciada intenta ver el mapa, o si la sesión caduca mientras lo está mirando? → A: El backend rechaza las peticiones de disponibilidad sin sesión válida; el frontend redirige al inicio de sesión y, tras autenticarse, devuelve al usuario al mapa. Si la sesión caduca con el mapa abierto, ocurre lo mismo.
- Q: Si el usuario ha dado permiso de ubicación pero está fuera del área que cubre la red de estaciones, ¿dónde debe centrarse el mapa al abrirse? → A: Solo se centra en el usuario si está dentro del área de la red (rectángulo que abarca todas las estaciones más un margen); si está fuera, se ignora su posición y se encuadra toda la red, también cuando la posición llega antes que los datos.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Mapa en vivo de la red de estaciones (Priority: P1)

Un usuario registrado con abono anual activo abre la aplicación y ve un mapa
interactivo con la ubicación geolocalizada de todas las estaciones de la red
(~60). Al seleccionar una estación consulta en detalle cuántas bicicletas hay
disponibles y cuántos anclajes están libres, para planificar dónde recoger una
bicicleta y dónde devolverla. Los reflejos de alzas y bajas de bicicletas en
cualquier estación aparecen en su pantalla con un retardo máximo de 30 segundos.

**Why this priority**: Es la pantalla principal y el punto de entrada del
servicio; sin disponibilidad fiable el usuario no puede planificar ni iniciar
ningún viaje, por lo que es el MVP del sistema.

**Independent Test**: Dar de alta estaciones con estados variados, abrir la
aplicación, verificar que todas aparecen en el mapa, seleccionar una estación y
comprobar que sus contadores coinciden con el estado real; provocar un cambio
en una estación y medir que se refleja en ≤ 30 s.

**Acceptance Scenarios**:

1. **Given** un usuario accede a la aplicación móvil o web, **When** se carga
   la pantalla principal, **Then** se muestra la ubicación geolocalizada de las
   ~60 estaciones de la red.
2. **Given** un usuario selecciona una estación concreta, **When** consulta sus
   detalles, **Then** se indica con precisión el número de bicicletas
   disponibles y el número de anclajes libres.
3. **Given** que ocurren alzas o bajas de bicicletas en cualquier estación,
   **When** se actualizan los datos en el servidor, **Then** la vista del
   usuario refleja el cambio con un retardo máximo de 30 segundos.
4. **Given** un usuario con sesión iniciada abre la aplicación y el servidor
   no responde, **When** falla la primera carga de disponibilidad sin ningún
   dato previo, **Then** la vista muestra «No se pudo cargar la
   disponibilidad» con un botón «Reintentar», no pinta marcadores ni
   contadores y reintenta automáticamente; **and when** el servidor vuelve a
   responder, el mapa se carga y el mensaje desaparece (FR-010).
5. **Given** un usuario con sesión iniciada y el mapa de fondo sin cargar (o
   un usuario que no puede usar el mapa), **When** abre el listado de
   estaciones, **Then** ve todas las estaciones con su nombre, su estado
   («Sin bicicletas», «Sin anclajes libres», «Fuera de servicio») y sus
   contadores, iguales a los del mapa y con la misma frescura (FR-011,
   SC-008).
6. **Given** un usuario tiene abierto el detalle de una estación operativa,
   **When** cambian sus contadores en el servidor o la estación pasa a fuera
   de servicio, **Then** en un máximo de 30 s el detalle muestra los valores
   nuevos (o «Fuera de servicio» con «—») sin cerrarse ni perder la selección
   (FR-004).
7. **Given** una persona sin sesión válida (no ha iniciado sesión o su sesión
   ha caducado), **When** intenta consultar la disponibilidad, **Then** el
   servidor rechaza la consulta sin devolver datos, la aplicación la lleva al
   inicio de sesión **and** tras autenticarse vuelve a la vista de
   disponibilidad (FR-012).
8. **Given** un usuario que ha concedido su ubicación y se encuentra fuera del
   área de la red, **When** abre el mapa (llegue su posición antes o después
   que los datos), **Then** el mapa se encuadra en toda la red y las
   estaciones quedan visibles; **and given** un usuario dentro del área,
   **then** el mapa puede centrarse en su posición (FR-013).

### Edge Cases

- Una estación operativa tiene 0 bicicletas disponibles: su marcador muestra
  el estado «sin bicicletas» y el detalle indica «Sin bicicletas» con el
  contador en 0 (ver FR-008).
- Una estación operativa tiene 0 anclajes libres (llena): su marcador muestra
  el estado «sin anclajes libres» y el detalle indica «Sin anclajes libres» con
  el contador en 0 (ver FR-008).
- Una estación operativa tiene a la vez 0 bicicletas disponibles y 0 anclajes
  libres (p. ej., todos sus anclajes ocupados por bicicletas averiadas): se
  señalan ambos estados. Las alertas de ocupación asociadas quedan fuera del
  alcance de esta spec (ver HU-07).
- La estación está fuera de servicio: sigue apareciendo en el mapa con el
  estado visual «Fuera de servicio», que prevalece sobre «sin bicicletas» y
  «sin anclajes libres»; al seleccionarla, el detalle indica «Fuera de
  servicio» y muestra «—» en lugar de contadores, para que el usuario no
  planifique sobre ella (ver FR-009).
- La ubicación del usuario no está disponible o fue denegada: el mapa sigue
  siendo consultable sin centrado automático en la posición del usuario y se
  encuadra en toda la red (ver FR-013).
- La ubicación del usuario está disponible pero queda fuera del área de la
  red: se ignora para el encuadre y el mapa muestra toda la red, aunque la
  posición llegue antes que los datos de las estaciones (ver FR-013).
- Un evento de actualización en vivo no llega o la conexión falla: la vista
  conserva el último estado conocido (contadores visibles), indica la hora de
  la última actualización y reintenta automáticamente; si han pasado más de
  30 s desde la última actualización correcta, muestra además un aviso visible
  de «datos desactualizados» (ver FR-007).
- El mapa de fondo (proveedor externo de teselas) no carga o el usuario no
  puede usar un mapa (p. ej., lector de pantalla): la disponibilidad sigue
  consultable desde el listado de estaciones (ver FR-011).
- La primera carga de disponibilidad falla y no existe ningún estado previo
  conocido: la vista muestra el mensaje «No se pudo cargar la disponibilidad»
  con un botón «Reintentar» y reintenta automáticamente; no se pintan
  marcadores ni contadores hasta recibir datos correctos (ver FR-010).
- El usuario tiene abierto el detalle de una estación cuando esta cambia: el
  detalle se actualiza en ≤ 30 s sin cerrarse; si la estación pasa a fuera de
  servicio, sus contadores se sustituyen por «—» (ver FR-004).
- Un usuario sin sesión intenta abrir el mapa, o su sesión caduca con el mapa
  abierto: el servidor rechaza la consulta y el usuario es llevado al inicio
  de sesión; tras entrar vuelve al mapa (ver FR-012).
- Dos cambios sucesivos en la misma estación dentro de la ventana de 30 s: la
  vista termina reflejando el estado final más reciente.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST mostrar a los usuarios registrados un mapa
  interactivo con la ubicación geolocalizada de la red completa de estaciones
  (~60 estaciones).
- **FR-002**: Para cada estación el sistema MUST indicar el número actual de
  bicicletas disponibles y el número de anclajes libres.
- **FR-003**: El sistema MUST reflejar en la vista del usuario cualquier alza o
  baja de bicicletas producida en el servidor con un retardo máximo de
  30 segundos, medido extremo a extremo: desde que el cambio queda confirmado
  (commit) en el servidor hasta que el contador nuevo aparece pintado en la
  pantalla de un usuario con la aplicación abierta y conexión activa.
- **FR-004**: El sistema MUST permitir seleccionar una estación del mapa para
  ver su detalle (estado de servicio y, si está operativa, sus contadores;
  ver FR-009). Mientras el detalle está abierto, MUST actualizarse
  automáticamente con el mismo límite de FR-003 (contadores y estado, incluido
  el paso a «Fuera de servicio» o la vuelta a operativa), sin cerrarse ni
  perder la selección.
- **FR-005**: Los contadores mostrados MUST provenir del estado en vivo
  gestionado por el sistema; queda prohibido presentarlos como datos manuales o
  históricos estáticos sin expiración.
- **FR-006**: El sistema MUST excluir del cómputo de disponibilidad las
  bicicletas marcadas como averiadas (consistente con HU-05). Un anclaje
  ocupado por una bicicleta averiada MUST NOT contarse como bicicleta
  disponible ni como anclaje libre; el detalle muestra solo esos dos
  contadores (sin desglose de averías), por lo que su suma puede ser menor que
  la capacidad total de la estación.
- **FR-007**: La vista MUST mostrar siempre la hora de la última actualización
  correcta de los datos y, cuando hayan transcurrido más de 30 segundos desde
  ella, MUST mostrar un aviso visible de «datos desactualizados» manteniendo
  visibles los últimos contadores conocidos; el aviso MUST desaparecer en
  cuanto llegue una actualización correcta.
- **FR-008**: El mapa MUST representar con un estado visual propio, tanto en el
  marcador como en el detalle, las estaciones sin bicicletas disponibles
  (0 disponibles) y las estaciones sin anclajes libres (0 libres), distinguibles
  entre sí y de una estación con disponibilidad. La distinción MUST NOT
  depender solo del color (debe incluir icono o texto). El detalle MUST mostrar
  el contador en 0 acompañado del texto «Sin bicicletas» o «Sin anclajes
  libres» según corresponda.
- **FR-009**: Las estaciones fuera de servicio MUST seguir apareciendo en el
  mapa y ser seleccionables, con un estado visual propio «Fuera de servicio»
  que no dependa solo del color y que prevalezca sobre los estados de FR-008.
  Su detalle MUST mostrar «Fuera de servicio» y MUST NOT mostrar contadores de
  bicicletas disponibles ni de anclajes libres (se muestra «—»), de modo que no
  se presenten como opción de recogida ni de devolución.
- **FR-010**: Si la carga inicial de disponibilidad falla sin que exista un
  estado previo conocido, la vista MUST mostrar el mensaje «No se pudo cargar
  la disponibilidad» con un botón «Reintentar» y MUST reintentar
  automáticamente; MUST NOT pintar marcadores ni contadores (ni valores por
  defecto) hasta recibir una respuesta correcta, momento en que el mensaje
  desaparece.
- **FR-011**: Además del mapa, el sistema MUST ofrecer un listado textual de
  todas las estaciones, accesible en todo momento (p. ej., pestaña «Lista»),
  que muestre para cada una su nombre, su estado (con los mismos textos de
  FR-008 y FR-009) y sus contadores, con la misma frescura que el mapa
  (FR-003). El listado MUST seguir funcionando aunque el mapa de fondo no
  cargue.
- **FR-012**: El servidor MUST rechazar las consultas de disponibilidad que no
  vengan de una sesión válida. Ante un acceso sin sesión, o si la sesión
  caduca con la vista abierta, la aplicación MUST redirigir al inicio de sesión
  y, tras autenticarse, MUST devolver al usuario a la vista de disponibilidad.
- **FR-013**: Al abrir el mapa, el encuadre inicial MUST mostrar toda la red
  de estaciones, salvo que el usuario haya concedido su ubicación y esta
  quede **dentro del área de la red** (el rectángulo que abarca todas las
  estaciones más un margen), en cuyo caso el mapa MAY centrarse en él. Una
  posición fuera de esa área MUST ignorarse para el encuadre, con
  independencia de si llega antes o después que los datos de las estaciones.

### Key Entities

- **Estación**: Punto físico de la red; ubicación geográfica, capacidad total,
  estado de servicio y sus anclajes asociados.
- **Anclaje**: Hueco individual de una estación; código identificador y estado
  (libre, ocupado por una bicicleta o fuera de servicio; solo los libres
  cuentan como anclajes libres).
- **Bicicleta**: Unidad de la flota; estado (operativa —disponible para
  alquilar—, en uso, averiada o en taller) y anclaje en el que se encuentra
  cuando está estacionada. Solo las operativas ancladas cuentan como
  bicicletas disponibles.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las estaciones de la red aparecen geolocalizadas en
  el mapa al cargar la pantalla principal.
- **SC-002**: El 100 % de los cambios de disponibilidad registrados en el
  servidor se reflejan en la vista del usuario en un máximo de 30 segundos,
  medidos desde el commit del cambio en el servidor hasta que el contador
  nuevo se pinta en pantalla (cliente abierto y con conexión activa).
- **SC-003**: En el 100 % de las consultas de detalle, los contadores de
  bicicletas disponibles y anclajes libres coinciden con el estado real del
  sistema, donde bicicletas disponibles = anclajes con bicicleta operativa y
  anclajes libres = anclajes en servicio sin bicicleta (los anclajes con
  bicicleta averiada no suman en ninguno de los dos, y los anclajes fuera de
  servicio no cuentan como libres).
- **SC-004**: El mapa completo de la red se muestra en menos de 5 segundos en
  una conexión estándar (referencia: perfil de red «Fast 4G» de Chrome
  DevTools, con caché vacía), y seleccionar una estación resuelve en menos de
  2 segundos en las mismas condiciones.
- **SC-005**: Con la conexión cortada, el aviso de «datos desactualizados»
  aparece en el 100 % de los casos en cuanto pasan más de 30 s desde la última
  actualización correcta y desaparece al restablecerse la conexión y
  recibirse datos nuevos.
- **SC-006**: El 100 % de las estaciones con 0 bicicletas disponibles o con
  0 anclajes libres muestran en el mapa el estado visual correspondiente, y
  este cambia al estado normal en un máximo de 30 s cuando vuelve a haber
  disponibilidad.
- **SC-007**: El 100 % de las estaciones fuera de servicio aparecen en el mapa
  con el estado «Fuera de servicio» y su detalle no muestra contadores; el
  paso de una estación a fuera de servicio (o su vuelta a operativa) se refleja
  en la vista en un máximo de 30 s.
- **SC-008**: El 100 % de las estaciones del mapa aparecen también en el
  listado textual con el mismo estado y los mismos contadores, y el listado
  sigue siendo consultable con el mapa de fondo bloqueado.

## Assumptions

- La consulta del mapa está disponible solo para usuarios con sesión iniciada
  (ver FR-012); el
  estado del abono no restringe la lectura de disponibilidad (solo las
  operaciones que lo exigen, como el desbloqueo de HU-02).
- El abono anual activo/inactivo es un atributo preexistente de la cuenta de
  usuario, gestionado fuera del alcance de esta especificación.
- La red consta de ~60 estaciones; el mapa tolera variaciones de inventario
  sin cambiar la especificación.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-01 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
