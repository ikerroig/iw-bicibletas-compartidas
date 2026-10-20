# Feature Specification: Notificación de averías e inhabilitación preventiva

**Feature Branch**: `spec-005`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-05 — Notificación de averías e inhabilitación preventiva (historia.md, SI-16)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reporte de avería desde la aplicación (Priority: P1)

Un usuario identifica una deficiencia en una bicicleta estacionada (frenos,
cadena, neumático) y la reporta directamente desde la aplicación en la estación
donde la encuentra, describiendo el problema. La bicicleta cambia
inmediatamente su estado a "Averiada": queda inhabilitada para el alquiler, de
ja de sumar entre las bicicletas operativas, no aparece como disponible en el
mapa ni en las listas de anclajes, y el reporte alimenta el panel de
incidencias pendientes del equipo técnico.

**Why this priority**: Es la compuerta de seguridad de la flota; una bicicleta
averiada en alquiler pone en riesgo a los usuarios y su bloqueo inmediato es
regla crítica del sistema.

**Independent Test**: Con una bicicleta operativa anclada, enviar un reporte
con descripción y verificar el cambio inmediato a "Averiada", su exclusión de
los contadores disponibles y la aparición de la incidencia en el panel de
operarios.

**Acceptance Scenarios**:

1. **Given** un usuario que reporta una avería en una estación, **When** envía
   el formulario con la descripción del problema, **Then** la bicicleta cambia
   su estado a "Averiada".
2. **Given** que una bicicleta se marca como "Averiada", **When** se actualiza
   el mapa global y la lista de anclajes, **Then** la bicicleta queda
   inhabilitada para el alquiler y deja de sumar al conteo de bicicletas
   operativas.
3. **Given** el reporte de una avería, **When** la petición es procesada en el
   sistema, **Then** se emite un evento que actualiza el panel de incidencias
   pendientes para los operarios.

### Edge Cases

- El usuario intenta reportar una bicicleta ya marcada como averiada: se
  acepta sin duplicar el bloqueo (la incidencia original se mantiene o se
  actualiza) y se confirma al usuario.
- El formulario se envía sin descripción: se exige una descripción antes de
  procesar el reporte.
- Dos reportes simultáneos sobre la misma bicicleta: la transición a "Averiada"
  se aplica una sola vez y no se generan estados inconsistentes.
- La bicicleta reportada está en viaje activo: se admite el reporte y la
  bicicleta queda averiada; su devolución y bloqueo se gestionan en los flujos
  de viaje (HU-02/HU-03), sin permitir nuevos alquileres.
- Un reporte falla a mitad de proceso: el estado de la bicicleta y el panel de
  incidencias nunca quedan a medias (o se aplica todo, o nada).
- El usuario pierde conexión tras enviar: al reintentar, si el reporte ya se
  registró no se duplica.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST permitir a un usuario enviar un reporte de
  avería sobre una bicicleta desde la aplicación, incluyendo una descripción
  del fallo técnico (obligatoria) y la estación donde se encuentra.
- **FR-002**: Al procesarse el reporte el sistema MUST cambiar el estado de la
  bicicleta a "Averiada" de forma inmediata.
- **FR-003**: Una bicicleta en estado "Averiada" MUST quedar inhabilitada para
  el alquiler: no puede ser desbloqueada (HU-02) ni contada como disponible.
- **FR-004**: El sistema MUST excluir las bicicletas averiadas del mapa, de la
  lista de anclajes y del conteo de bicicletas operativas de cada estación.
- **FR-005**: El sistema MUST registrar el reporte como incidencia pendiente y
  MUST hacerla visible en el panel de incidencias del equipo técnico, con
  descripción, estación y fecha.
- **FR-006**: El sistema MUST confirmar al usuario que su reporte fue
  registrado.
- **FR-007**: El sistema MUST evitar transiciones duplicadas: una bicicleta
  averiada reportada de nuevo no genera un segundo bloqueo inconsistente.

### Key Entities

- **Bicicleta**: Unidad de la flota; estado (operativa, averiada) que determina
  su disponibilidad y su cómputo en estaciones.
- **Reporte de avería** (Incidencia): Registro del fallo; bicicleta,
  descripción del problema, estación, usuario reportante, fecha y estado
  (pendiente/atendida).
- **Estación**: Punto que aloja la bicicleta; sus contadores de operativas se
  recalculan al producirse averías.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de los reportes con descripción cambian la bicicleta a
  "Averiada" en el momento de su procesamiento.
- **SC-002**: Tras un reporte, la incidencia aparece en el panel de operarios
  en menos de 30 segundos y la bicicleta no puede ser desbloqueada (0
  alquileres de unidades averiadas).
- **SC-003**: El 100 % de los mapas y listas excluyen las bicicletas averiadas
  de los contadores de disponibilidad de forma consistente con HU-01.
- **SC-004**: 0 bicicletas quedan en estado inconsistente (averiada pero
  disponible para alquiler) tras cualquier secuencia de reportes.

## Assumptions

- El usuario reportante está identificado con sesión iniciada; el reporte
  incluye la estación en la que se encuentra la bicicleta.
- El flujo principal reporta bicicletas ancladas; una avería detectada durante
  un viaje activo también puede reportarse y su efecto sobre el estado es el
  mismo.
- Un reporte repetido sobre una bicicleta ya averiada no duplica la
  inhabilitación ni el bloqueo.
- El panel de incidencias es la vista de trabajo de operarios y taller
  (HU-06), cuyo contenido de reparación queda fuera del alcance de esta spec.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-05 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
