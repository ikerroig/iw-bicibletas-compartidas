# Feature Specification: Registro de intervenciones técnicas y reparación en taller

**Feature Branch**: `spec-006`

**Created**: 2026-10-20

**Status**: Implemented locally; refinement in progress

## Clarifications

Las decisiones de refinamiento de HU-06 se detallan en [research.md](./research.md)
con los identificadores C-001 a C-005:

- reparar y ubicar una bicicleta son operaciones independientes;
- una bicicleta operativa sin anclaje no aparece como disponible en el mapa;
- la avería detectada se conserva en el historial y el parte añade piezas y
  notas firmadas por el operario;
- el mantenimiento preventivo se registra como intervención resuelta;
- la ubicación exige un anclaje libre y una estación operativa.

**Input**: User description: "HU-06 — Registro de intervenciones técnicas y reparación en taller (historia.md, SI-16)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cierre de reparación desde el taller (Priority: P1)

Un operario de mantenimiento consulta la ficha de una bicicleta averiada y
visualiza el reporte original introducido por el usuario. Al concluir la
reparación cumplimenta el parte técnico —avería detectada, piezas sustituidas,
notas y su identificación como técnico— y marca la unidad como "Operativa"; la
intervención firmada por el operario queda guardada en el historial de
mantenimiento de la bicicleta. Cuando la bicicleta reparada se devuelve a un
anclaje y se confirma su ubicación, vuelve a figurar como disponible para los
usuarios en el mapa.

**Why this priority**: Es la única vía para devolver bicicletas a servicio con
garantía; sin registro de reparación la flota se encoge y las averías
(HU-05) no se resuelven.

**Independent Test**: Con una bicicleta averiada con reporte de usuario,
abrir su ficha como operario, registrar el parte de reparación marcándola como
"Operativa", verificar la intervención guardada con autoría y confirmar que la
bicicleta reaparece disponible en el mapa tras ubicarla en un anclaje.

**Acceptance Scenarios**:

1. **Given** un operario en el taller, **When** consulta la ficha de una
   bicicleta averiada, **Then** visualiza el reporte original introducido por
   el usuario.
2. **Given** que un operario concluye una reparación, **When** cumplimenta el
   formulario técnico (piezas cambiadas, notas) y marca la unidad como
   "Operativa", **Then** se guarda la intervención firmada por el operario.
3. **Given** que una bicicleta es reparada y devuelta a un anclaje, **When**
   se confirma su ubicación, **Then** vuelve a figurar como disponible para
   los usuarios en el mapa.

### Edge Cases

- El operario intenta marcar "Operativa" sin completar los campos
  obligatorios del parte (piezas o avería detectada): se bloquea el guardado y
  se solicita completarlos.
- Varias incidencias abiertas existen para la misma bicicleta: se muestran en
  la ficha y la reparación las resuelve (la incidencia original queda atendida).
- Se registra una intervención sobre una bicicleta no averiada: se permite como
  mantenimiento preventivo, dejando la intervención en el historial sin
  alterar su disponibilidad si ya estaba operativa.
- El operario cancela el parte sin guardar: no se modifica estado ni historial.
- Se confirma la ubicación en un anclaje ocupado o inexistente: se rechaza y
  la bicicleta no pasa a disponible en una ubicación inválida.
- Dos operarios editan la misma ficha a la vez: solo el último guardado
  completo prevalece sin perder intervenciones ya confirmadas.
- La bicicleta se marca "Operativa" pero no se ha confirmado ubicación: no
  aparece todavía como disponible en el mapa hasta ubicarla en un anclaje.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST mostrar al operario la ficha de la bicicleta con
  su historial de averías y, cuando exista, el reporte original introducido
  por el usuario (HU-05).
- **FR-002**: El sistema MUST permitir registrar un parte de reparación con
  avería detectada, piezas sustituidas, notas del técnico e identificación del
  operario, con campos obligatorios validados.
- **FR-003**: Al guardar el parte con la unidad marcada como "Operativa" el
  sistema MUST cambiar el estado de la bicicleta a operativa y MUST persistir
  la intervención firmada por el operario con su fecha.
- **FR-004**: El sistema MUST conservar el historial de mantenimiento completo
  de cada bicicleta (intervenciones y averías asociadas).
- **FR-005**: Al confirmar la ubicación de la bicicleta reparada en un anclaje
  libre, el sistema MUST registrar su ubicación y MUST mostrarla como
  disponible para los usuarios en el mapa y en los contadores de la estación.
- **FR-006**: El sistema MUST resolver la incidencia de avería atendida al
  completarse la reparación, de modo que no siga en el panel de pendientes.
- **FR-007**: Una bicicleta "Operativa" sin ubicación confirmada en anclaje
  MUST NOT figurar como disponible en el mapa hasta que se ubique.

### Key Entities

- **Bicicleta**: Unidad de la flota; estado (operativa, averiada) y ubicación
  actual (anclaje) que determinan su disponibilidad.
- **Reporte de avería** (Incidencia): Fallo reportado por usuarios; se muestra
  en la ficha y se resuelve al reparar.
- **Intervención de taller**: Parte de reparación; bicicleta, avería detectada,
  piezas sustituidas, notas, operario autor, fecha y estado resultante.
- **Anclaje**: Hueco de una estación donde se confirma la ubicación de la
  bicicleta reparada.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las bicicletas averiadas con reporte de usuario
  muestran ese reporte original en su ficha de taller.
- **SC-002**: El 100 % de las reparaciones concluidas generan una intervención
  persistida con autoría del operario y fecha; 0 reparaciones quedan sin
  registro.
- **SC-003**: Tras confirmar la ubicación, la bicicleta aparece como disponible
  en el mapa y en los contadores de la estación en un máximo de 30 segundos
  (consistente con HU-01).
- **SC-004**: 0 bicicletas figuran simultáneamente como "Operativa" con
  incidencia pendiente tras su reparación; el panel de taller refleja el cierre
  de la incidencia.

## Assumptions

- Solo personal identificado como operario de taller accede a la ficha y al
  registro de intervenciones; los usuarios solo envían reportes (HU-05).
- El parte de reparación requiere como mínimo avería detectada y piezas o
  notas; la identificación del operario se registra automáticamente con su
  sesión.
- La disponibilidad final en el mapa exige dos condiciones: estado "Operativa"
  y ubicación confirmada en un anclaje.
- El flujo principal parte de bicicletas con incidencia abierta; el registro
  de mantenimiento preventivo sobre unidades operativas se admite.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-06 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.

## Decisiones aclaradas durante la implementación

- La ficha se consulta por el identificador interno de la bicicleta y devuelve
  también el código visible y todo su historial de incidencias.
- Una reparación completada deja la bicicleta `disponible` sin anclaje. La
  ubicación se confirma en una operación posterior y solo entonces vuelve a
  contar en la estación.
- El mantenimiento preventivo se registra como una incidencia resuelta de tipo
  `otro`; como no procede de un reporte de usuario, puede quedar sin estación.
