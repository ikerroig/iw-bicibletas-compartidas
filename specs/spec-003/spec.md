# Feature Specification: Devolución de bicicleta y cálculo de tarificación

**Feature Branch**: `spec-003`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-03 — Devolución de bicicleta y cálculo de tarificación (historia.md, SI-16)."

## Clarifications

### Session 2026-11-19

- Q: Para el tiempo que excede de los 30 minutos, ¿se aplica una única
  tarifa del servicio o cada usuario tiene su propia tarifa? — A: Tarifa
  única del servicio (`TARIFA_MINUTO`, global en configuración); no existe
  tarifa por usuario. La «tarifa del usuario» de la narrativa se interpreta
  como la tarifa aplicable al viaje de ese usuario.
- Q: Cuando un viaje dura 30 min 45 s, ¿cómo se cobra el tiempo que excede
  de los 30 minutos iniciales? — A: Minuto empezado → se cobra entero
  (redondeo hacia arriba): 30:45 → 1 minuto excedente.
- Q: ¿Quién puede ejecutar la devolución (cierre) de una bicicleta? — A:
  Solo el usuario autenticado con rol `usuario` propietario del viaje
  activo; sin sesión → 401; operarios y gestores → 403, sin efectos sobre
  viaje ni anclaje.
- Q: Si el abono del usuario caduca mientras su viaje sigue en curso,
  ¿puede devolver la bicicleta? — A: Sí; la devolución siempre está
  permitida para el propietario del viaje activo. El abono solo condiciona
  el desbloqueo (HU-02), no el cierre.
- Q: ¿Se mantiene el criterio de que, si la estación destino no tiene
  anclajes libres, la devolución no se completa y el viaje sigue activo?
  — A: Sí, se mantiene (criterio existente confirmado).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fin de viaje y tarifa (Priority: P1)

Un usuario que está realizando un trayecto lleva la bicicleta a cualquier
estación de la red y la fija en un anclaje vacante. El sistema registra la
hora de llegada, da por finalizado el viaje, libera al usuario para que pueda
iniciar otro y bloquea la bicicleta en el anclaje. Si el viaje duró 30 minutos
o menos, el coste es de 0,00 € (incluido en el abono); si superó los 30
minutos, se cobra únicamente el tramo de tiempo que excedió de los 30
minutos iniciales.

**Why this priority**: Cierra el ciclo de alquiler y genera el ingreso del
servicio; sin devolución el usuario queda con viaje activo y la flota pierde
bicicletas.

**Independent Test**: Con un viaje activo de duración controlada, fijar la
bicicleta en un anclaje libre y verificar hora de llegada, liberación del
usuario, bloqueo de la bicicleta y coste correcto (0,00 € si ≤ 30 min; solo
excedente si > 30 min).

**Acceptance Scenarios**:

1. **Given** un viaje activo, **When** la bicicleta es introducida en un
   anclaje libre, **Then** el sistema registra la hora de llegada, libera el
   estado del usuario y bloquea la bicicleta en el anclaje.
2. **Given** un viaje cuya duración total sea menor o igual a 30 minutos,
   **When** el sistema calcula el importe del trayecto, **Then** asigna un
   coste de 0,00 € (incluido en el abono).
3. **Given** un viaje cuya duración supere los 30 minutos, **When** se procesa
   la devolución, **Then** se tarifica y cobra el tramo de tiempo que exceda
   de los 30 minutos iniciales, aplicando la tarifa única del servicio y
   cobrando por completo cada minuto empezado del excedente (redondeo hacia
   arriba: 30:45 → 1 minuto excedente).
4. **Given** un usuario autenticado con rol distinto de `usuario`
   (operario o gestor) o sin sesión, **When** intenta ejecutar una
   devolución, **Then** el sistema responde 403 (401 sin sesión) y no
   modifica el viaje ni el anclaje.
5. **Given** un viaje activo cuyo propietario tiene el abono caducado,
   **When** fija la bicicleta en un anclaje libre, **Then** el viaje se
   cierra con normalidad (hora de llegada, liberación del usuario, anclaje
   ocupado) y el importe queda registrado; el abono solo condiciona el
   desbloqueo (HU-02).
6. **Given** un viaje de 30 min 45 s con `TARIFA_MINUTO = 0,20 €`,
   **When** el sistema calcula el importe al devolver, **Then** el coste es
   0,20 € (1 minuto excedente × tarifa), nunca 0,00 € ni 0,15 €.

### Edge Cases

- Todos los anclajes de la estación de destino están ocupados: la devolución
  no se completa; el usuario debe probar otra estación y su viaje sigue
  activo.
- Se intenta devolver sin ningún viaje activo: se rechaza con causa y no se
  modifica el anclaje.
- Duración exacta de 30 minutos (30:00): aplica el tramo gratuito, coste
  0,00 €.
- La bicicleta se reporta averiada durante el viaje (HU-05): la devolución
  sigue siendo posible para estacionarla; la avería no altera la tarificación
  del tramo ya recorrido.
- Un cálculo de importe falla o queda sin confirmar: el viaje no se cierra de
  forma incompleta — o se cierra con el último estado válido y se registra el
  importe pendiente, o se reintenta; nunca se libera al usuario sin registro
  de cierre.
- Dos confirmaciones de devolución casi simultáneas para el mismo viaje: solo
  la primera tiene efecto.
- Duración negativa o reloj inconsistente: se detecta y se usa la hora de
  llegada validada; el coste nunca resulta negativo.
- Abono caducado durante el viaje: la devolución del propietario se permite
  siempre; rechazarla dejaría la bicicleta y el viaje bloqueados sin
  solución de continuidad.
- Intento de devolución por un rol distinto de `usuario` o sin sesión:
  rechazo 403/401 sin efectos sobre el viaje ni el anclaje (solo el
  propietario con viaje activo puede cerrarlo).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST permitir finalizar un viaje fijando la bicicleta
  en un anclaje libre de cualquier estación de la red.
- **FR-002**: Al completarse la devolución el sistema MUST registrar la hora
  de llegada, MUST liberar al usuario (puede iniciar un nuevo viaje) y MUST
  bloquear la bicicleta en el anclaje.
- **FR-003**: Para viajes con duración ≤ 30 minutos el sistema MUST asignar
  un coste de 0,00 €.
- **FR-004**: Para viajes con duración > 30 minutos el sistema MUST
  tarificar y cobrar únicamente el tramo de tiempo que exceda de los 30
  minutos iniciales, aplicando la tarifa única del servicio
  (`TARIFA_MINUTO`, sin tarifas por usuario) y cobrando por completo cada
  minuto empezado del excedente (redondeo hacia arriba).
- **FR-005**: El sistema MUST persistir en el viaje la hora de inicio, la hora
  de llegada, la duración total y el coste final, para su consulta en el
  historial (HU-04).
- **FR-006**: El sistema MUST garantizar que un viaje solo puede cerrarse una
  vez y que una bicicleta solo puede estar en un anclaje a la vez.
- **FR-007**: El sistema MUST rechazar la devolución cuando el usuario no
  tenga viaje activo, sin alterar el estado del anclaje.
- **FR-008**: El sistema MUST permitir la devolución únicamente al usuario
  autenticado con rol `usuario` propietario del viaje activo: sin sesión
  MUST responder 401 y con cualquier otro rol (operario, gestor) MUST
  responder 403, en ambos casos sin modificar viaje ni anclaje.
- **FR-009**: El sistema MUST permitir la devolución aunque el abono del
  propietario esté caducado; la vigencia del abono solo condiciona el
  desbloqueo (HU-02), nunca el cierre del viaje.

### Key Entities

- **Viaje**: Registro de alquiler; usuario, bicicleta, estación de origen y
  destino, hora de inicio, hora de llegada, duración, coste y estado
  (activo/finalizado).
- **Anclaje**: Hueco individual de una estación; estado (ocupado/libre) y la
  bicicleta que aloja.
- **Bicicleta**: Unidad de la flota; estado y ubicación por anclaje.
- **Tarifa**: Regla de precio del servicio, única para todo el servicio (no
  por usuario); incluye la gratuidad de los primeros 30 minutos y el precio
  aplicable al tiempo excedente con redondeo al minuto empezado.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de los viajes con duración ≤ 30 minutos registran un
  coste final de 0,00 €.
- **SC-002**: El 100 % de los viajes con duración > 30 minutos registran un
  coste igual a `TARIFA_MINUTO` × minutos excedentes redondeados hacia
  arriba (minuto empezado → entero), sin cobrar los 30 minutos iniciales.
- **SC-003**: Una devolución válida se completa en menos de 10 segundos desde
  que la bicicleta se fija en el anclaje, dejando el anclaje ocupado y el
  usuario libre.
- **SC-004**: Tras cerrar un viaje, el 100 % de los usuarios pueden iniciar un
  nuevo desbloqueo de inmediato; existen 0 viajes activos con doble cierre.

## Assumptions

- La tarifa del tiempo excedente es una **tarifa única del servicio**
  (`TARIFA_MINUTO`, configuración operativa); no hay tarifa por usuario
  (clarificación 2026-11-19: la narrativa «la tarifa del usuario» se
  interpreta como la tarifa aplicable al viaje de ese usuario; el modelo
  compartido no prevé tarifas individuales — constitución III).
- La duración se calcula como diferencia entre hora de llegada y hora de
  inicio; la hora de cierre es la registrada en el momento de fijar la
  bicicleta en el anclaje.
- Devolver significa fijar la bicicleta en cualquier anclaje libre de la red,
  no necesariamente en la estación de origen.
- El usuario se identifica con sesión iniciada y su viaje activo se determina
  por el estado de sus registros de viaje.
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-03 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
