# Feature Specification: Consulta de disponibilidad de estaciones en tiempo real

**Feature Branch**: `spec-001`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-01 — Consulta de disponibilidad de estaciones en tiempo real (historia.md, SI-16)."

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

### Edge Cases

- Una estación muestra 0 bicicletas y 0 anclajes libres (llena u operativamente
  inactiva): se representa correctamente y entra en los umbrales de alerta de
  ocupación (fuera del alcance de esta spec, ver HU-07).
- La estación está fuera de servicio: se muestra su estado para que el usuario
  no planifique sobre ella.
- La ubicación del usuario no está disponible o fue denegada: el mapa sigue
  siendo consultable sin centrado automático en la posición del usuario.
- Un evento de actualización en vivo no llega o la conexión falla: la vista
  conserva el último estado conocido, indica la hora de la última
  actualización y reintenta automáticamente.
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
  30 segundos.
- **FR-004**: El sistema MUST permitir seleccionar una estación del mapa para
  ver su detalle (contadores y estado de servicio).
- **FR-005**: Los contadores mostrados MUST provenir del estado en vivo
  gestionado por el sistema; queda prohibido presentarlos como datos manuales o
  históricos estáticos sin expiración.
- **FR-006**: El sistema MUST excluir del cómputo de disponibilidad las
  bicicletas marcadas como averiadas (consistente con HU-05).

### Key Entities

- **Estación**: Punto físico de la red; ubicación geográfica, capacidad total,
  estado de servicio y sus anclajes asociados.
- **Anclaje**: Hueco individual de una estación; código identificador y estado
  (ocupado por una bicicleta o libre).
- **Bicicleta**: Unidad de la flota; estado (operativa, averiada) y anclaje en
  el que se encuentra cuando está estacionada.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las estaciones de la red aparecen geolocalizadas en
  el mapa al cargar la pantalla principal.
- **SC-002**: El 100 % de los cambios de disponibilidad registrados en el
  servidor se reflejan en la vista del usuario en un máximo de 30 segundos.
- **SC-003**: En el 100 % de las consultas de detalle, los contadores de
  bicicletas disponibles y anclajes libres coinciden con el estado real del
  sistema.
- **SC-004**: El mapa completo de la red se muestra en menos de 5 segundos en
  una conexión estándar, y seleccionar una estación resuelve en menos de
  2 segundos.

## Assumptions

- La consulta del mapa está disponible para usuarios con sesión iniciada; el
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
