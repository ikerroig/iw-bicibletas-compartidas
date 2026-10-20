# Feature Specification: Monitorización de desequilibrios y alertas de balanceo de flota

**Feature Branch**: `spec-007`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-07 — Monitorización de desequilibrios y alertas de balanceo de flota (historia.md, SI-16)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Panel de alertas de balanceo para el operario de calle (Priority: P1)

Un operario de transporte y calle necesita planificar la ruta de la furgoneta
para redistribuir bicicletas. El sistema vigila la ocupación de cada estación
y, cuando una cae por debajo del 10 % o supera el 90 % de su capacidad,
genera automáticamente una alerta de prioridad alta. El operario abre su panel
de logística y ve la lista de estaciones descompensadas ordenada por urgencia
de intervención. Cuando realiza la carga o descarga de un lote de bicicletas y
lo confirma en su dispositivo, la disponibilidad de la estación de origen y la
de destino se recalcula de forma inmediata.

**Why this priority**: Sin alertas ni panel, el operario no sabe a dónde ir y
las estaciones se quedan vacías o saturadas, rompiendo el servicio para todos
los usuarios.

**Independent Test**: Mover manualmente el estado de estaciones por debajo del
10 % y por encima del 90 %, verificar la generación de alertas altas y su
orden en el panel; confirmar un traslado y comprobar el recálculo inmediato de
ambas estaciones.

**Acceptance Scenarios**:

1. **Given** el estado de ocupación de la red, **When** una estación cae por
   debajo del 10 % o supera el 90 % de su capacidad, **Then** el sistema
   genera una alerta automática de prioridad alta.
2. **Given** un operario que accede a su panel de logística, **When** abre la
   lista de balanceo, **Then** visualiza las estaciones descompensadas
   ordenadas por urgencia de intervención.
3. **Given** que un operario realiza la carga o descarga de un lote de
   bicicletas, **When** confirma la acción en su dispositivo, **Then** la
   disponibilidad de ambas estaciones se recalcula de forma inmediata.

### Edge Cases

- Una estación queda exactamente en el 10 % o en el 90 % de ocupación: no
  cruza los umbrales (son estrictos: <10 % y >90 %) y no se genera alerta.
- Una estación en alerta vuelve a entrar en el rango normal: su alerta se
  resuelve automáticamente y sale de la lista de pendientes.
- Varias estaciones desbordan la capacidad o quedan sin bicicletas: cada una
  genera su alerta; si ya existe alerta abierta para esa estación y umbral, no
  se duplica.
- Una estación fuera de servicio o sin capacidad computable: se excluye del
  cálculo de ocupación y no genera alertas.
- El operario confirma un traslado con datos algo desactualizados: el recálculo
  posterior a la confirmación usa los valores confirmados de origen y destino,
  corrigiendo los contadores.
- Un lote confirmado con más unidades de las disponibles en origen: se valida
  y se rechaza la confirmación inconsistente sin alterar estaciones.
- Estaciones que alternan rápidamente entre umbrales: el panel siempre refleja
  el estado actual (alertas abiertas o resueltas) sin duplicados.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST calcular para cada estación su porcentaje de
  ocupación como bicicletas operativas presentes respecto a su capacidad
  total.
- **FR-002**: El sistema MUST generar automáticamente una alerta de prioridad
  alta cuando una estación queda por debajo del 10 % o por encima del 90 % de
  su capacidad.
- **FR-003**: El sistema MUST mantener una única alerta activa por estación y
  umbral cruzado, sin duplicados, hasta su resolución.
- **FR-004**: El sistema MUST resolver automáticamente la alerta cuando la
  ocupación de la estación vuelve a situarse dentro del rango normal.
- **FR-005**: El sistema MUST ofrecer al operario un panel de logística con la
  lista de estaciones descompensadas ordenada por urgencia de intervención
  (primero los extremos más críticos).
- **FR-006**: El sistema MUST permitir al operario confirmar la carga o
  descarga de un lote de bicicletas indicando estación de origen, estación de
  destino y número de unidades.
- **FR-007**: Tras la confirmación de un traslado, el sistema MUST recalcular
  y actualizar la disponibilidad de ambas estaciones de forma inmediata en la
  vista del operario y con un retardo máximo de 30 segundos en el estado
  global (consistente con HU-01).
- **FR-008**: El sistema MUST excluir del cómputo de ocupación las bicicletas
  averiadas (consistente con HU-05).

### Key Entities

- **Estación**: Punto de la red; capacidad total, bicicletas operativas
  presentes y su porcentaje de ocupación derivado.
- **Alerta de balanceo**: Aviso automático; estación, umbral cruzado (vacía
  <10 % o saturada >90 %), prioridad alta, fecha de apertura y estado
  (activa/resuelta).
- **Operación de traslado**: Confirmación del operario; estación de origen,
  estación de destino, número de unidades movidas y fecha.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las estaciones que cruzan los umbrales (<10 % o
  >90 %) generan una alerta de prioridad alta en menos de 1 minuto.
- **SC-002**: El panel del operario muestra siempre las estaciones
  descompensadas ordenadas por urgencia, con 0 duplicados de alerta por
  estación y umbral.
- **SC-003**: Tras confirmar un traslado, la vista del operario refleja los
  nuevos valores de inmediato y el estado global en un máximo de 30 segundos
  en el 100 % de los casos.
- **SC-004**: El 100 % de las alertas se resuelven automáticamente al volver
  la estación a un rango de ocupación normal, dejando 0 alertas obsoletas en
  el panel.

## Assumptions

- Ocupación = bicicletas operativas presentes / capacidad total de la
  estación; las averiadas no cuentan (HU-05).
- Umbrales estrictos: alerta solo con ocupación <10 % o >90 % exactamente, tal
  como indica la HU-07.
- El panel de logística es una vista exclusiva del operario de transporte y
  calle (rol identificado); la gestión de reparaciones pertenece a HU-06.
- La prioridad de las alertas es alta por definición de la HU; el orden por
  urgencia sitúa primero los extremos más alejados del rango normal.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-07 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
