# Feature Specification: Consulta de historial de viajes y desglose mensual

**Feature Branch**: `spec-004`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-04 — Consulta de historial de viajes y desglose mensual (historia.md, SI-16)."

## Clarifications

### Session 2026-10-28

- Q: Cuando el periodo consultado contiene muchos viajes, ¿cómo debe
  entregarse y presentarse la lista del historial? — A: Paginación
  server-side con los parámetros `page` y `page_size` del contrato global
  (50 viajes por defecto, máximo 200); el sumatorio se calcula sobre el
  periodo completo sea cual sea la página visible; una página fuera de rango
  devuelve lista vacía sin error conservando el sumatorio del periodo.
- Q: ¿En qué zona horaria se deben interpretar el mes y el año de cada
  viaje al aplicar el filtro? — A: Europe/Madrid (hora local española):
  mes y año se interpretan en ese huso para que el desglose mensual
  coincida con el calendario del usuario.
- Q: ¿Debe mostrarse en cada fila del historial la fecha y hora de inicio de
  cada viaje? — A: Sí; cada fila muestra fecha y hora de inicio en formato
  dd/mm/aaaa HH:MM.
- Q: El OpenAPI global define `GET /viajes` con `desde`/`hasta` y `page`/
  `page_size` (50 por defecto, máx. 200), mientras que la primera respuesta
  fijó filtro mes/año y `limit` 20/100. ¿Qué contrato adoptamos? — A: Se
  adopta el contrato global sin modificar el fichero compartido
  (`specs/spec-001/contracts/openapi.yaml`): el filtro de mes/año se
  expresa como rango de fechas `desde`/`hasta`, que la UI construye a partir
  del mes seleccionado, y la paginación usa `page`/`page_size` (50/200).
- Q: Si el periodo tiene viajes pero la página pedida está fuera de rango,
  ¿qué sumatorio devuelve la API? — A: `items` vacío pero con el sumatorio
  real del periodo completo: el sumatorio es propiedad del periodo, no de la
  página visible; el 0,00 € se reserva a los periodos sin viajes.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - "Mis Viajes" con filtros de periodo (Priority: P1)

Un usuario del servicio entra en la sección "Mis Viajes" y ve una lista
cronológica de todos sus trayectos: fecha y hora de inicio, estación de
origen, estación de destino, duración y coste individual de cada uno. Puede
filtrar por mes y año para
ver únicamente los viajes de ese periodo junto al sumatorio del gasto. Los
viajes de 30 minutos o menos aparecen identificados como "Tramo gratuito
(0,00 €)". "Ver mis rutas pasadas" significa consultar la estación de origen
y destino de cada viaje, sin seguimiento de trayectoria GPS.

**Why this priority**: Da transparencia al usuario sobre sus cobros y gastos;
es la función de confianza del servicio y responde a la verificación de
importes.

**Independent Test**: Con un usuario que tiene viajes repartidos en varios
meses, abrir "Mis Viajes", verificar campos y orden, aplicar un filtro
mes/año y comprobar que solo aparecen esos viajes con el sumatorio correcto y
la etiqueta en los viajes gratuitos.

**Acceptance Scenarios**:

1. **Given** que un usuario entra en la sección "Mis Viajes", **When** la vista
   se renderiza, **Then** se muestra una lista cronológica con fecha y hora de
   inicio, estación origen, estación destino, duración y coste individual.
2. **Given** el historial de trayectos, **When** el usuario aplica un filtro de
   mes y año específicos, **Then** la lista exhibe únicamente los viajes
   comprendidos en ese rango junto al sumatorio del gasto.
3. **Given** un viaje que no generó cargos por durar menos de 30 minutos,
   **When** se muestra en el historial, **Then** figura claramente identificado
   con el concepto "Tramo gratuito (0,00 €)".

### Edge Cases

- El periodo filtrado no contiene viajes: se muestra una lista vacía con
  indicación de "sin viajes en el periodo" y sumatorio de 0,00 €.
- El usuario aún tiene un viaje activo: no se lista como viaje finalizado
  hasta que se cierre (HU-03).
- El filtro apunta a un mes futuro o al mes en curso sin actividad: lista
  vacía sin error.
- Un periodo con muchos viajes: la lista se pagina en el servidor (50 viajes
  por página, ajustable hasta 200) y se mantiene ordenada; una página fuera
  de rango devuelve lista vacía sin error; el sumatorio corresponde siempre
  al periodo completo, no a la página visible.
- Un viaje del periodo tiene coste 0,00 € por duración ≤ 30 min: se muestra la
  etiqueta "Tramo gratuito (0,00 €)" y su importe computa 0 en el sumatorio.
- Usuario sin viajes registrados: se muestra el estado vacío del historial
  desde la primera carga.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST mostrar al usuario una lista cronológica de sus
  trayectos finalizados con fecha y hora de inicio (dd/mm/aaaa HH:MM),
  estación de origen, estación de destino, duración y coste individual de
  cada viaje.
- **FR-002**: El sistema MUST permitir filtrar el historial por mes y año,
  mostrando solo los viajes comprendidos en ese rango; el mes y el año de
  cada viaje MUST interpretarse en la zona horaria Europe/Madrid (hora
  local española).
- **FR-003**: El sistema MUST mostrar el sumatorio del gasto de los viajes
  listados (del periodo filtrado o del historial completo cuando no hay
  filtro).
- **FR-004**: El sistema MUST identificar cada viaje con coste 0,00 € con el
  concepto "Tramo gratuito (0,00 €)".
- **FR-005**: El sistema MUST mostrar únicamente los viajes finalizados del
  usuario con sesión iniciada; nunca los viajes de otros usuarios.
- **FR-006**: El sistema MUST ofrecer para cada viaje su estación de origen y
  destino como consulta de la "ruta" del viaje, sin seguimiento de trayectoria
  GPS.
- **FR-007**: El coste y la duración mostrados MUST coincidir con los
  registrados al cierre del viaje (HU-03).
- **FR-008**: El sistema MUST entregar el listado paginado por el servidor
  (50 viajes por página por defecto, máximo 200) manteniendo el orden
  cronológico, con el sumatorio calculado sobre el periodo completo
  consultado y no sobre la página visible; una página fuera de rango MUST
  devolver lista vacía sin error conservando el sumatorio del periodo
  (0,00 € únicamente cuando el periodo no contiene viajes).

### Key Entities

- **Viaje**: Registro finalizado; usuario, estación de origen, estación de
  destino, fecha/hora de inicio y llegada, duración y coste.
- **Usuario**: Cuenta con sesión iniciada; propietario de los viajes del
  historial.
- **Estación**: Punto de origen y destino referenciado en cada viaje.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de los viajes finalizados del usuario aparecen en su
  historial con origen, destino, duración y coste correctos.
- **SC-002**: El 100 % de los filtros por mes/año devuelven exactamente los
  viajes de ese rango; el sumatorio mostrado coincide con la suma de los
  costes de todos los viajes del periodo, sea cual sea la página visible,
  en el 100 % de los casos.
- **SC-003**: El 100 % de los viajes sin coste se muestran con la etiqueta
  "Tramo gratuito (0,00 €)".
- **SC-004**: El historial y sus filtros se muestran en menos de 3 segundos
  para un año completo de uso habitual.

## Assumptions

- "Ver mis rutas pasadas" se refiere a consultar la estación de origen y
  destino de cada viaje; no hay tracking GPS ni trayectorias intermedias
  (aclarecimiento de alcance del equipo, HU validada sin cambios).
- La lista se ordena cronológicamente con los viajes más recientes primero
  cuando la HU no fija dirección.
- Solo se listan viajes finalizados; el viaje activo se consulta en el flujo
  en curso, no en el historial.
- La identificación del usuario se realiza mediante sesión iniciada.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-04 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
