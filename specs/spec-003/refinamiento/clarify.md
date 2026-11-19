# Registro de decisiones — clarify (spec-003 · HU-03)

**Rama**: `spec-003` · **Commit de partida**: `0d6f478d511427172aa8975fb7f9c4d3b3db1eb5`
(main tras fast-forward desde `10ff0bf`) · **Spec Kit**: `1.1.3` (`.specify/init-options.json`)
**Fecha**: 2026-11-19 · **Funcionalidad activa**: `specs/spec-003`
**Comando**: `/speckit.clarify` (integración opencode, `.opencode/commands/speckit.clarify.md`)
**Spec resultante**: `specs/spec-003/spec.md` § Clarifications (Session 2026-11-19)

## Casos probados por historia (3/3 exigidos)

| # | Categoría | Caso probado | Resultado |
|---|-----------|--------------|-----------|
| 1 | Recorrido normal | Cierre de un viaje de > 30 min: ¿qué tarifa se aplica al excedente? | **D1**: tarifa única del servicio (`TARIFA_MINUTO`), sin tarifa por usuario |
| 2 | Entrada inválida / caso límite | (a) viaje de 30:45 → ¿cuántos minutos se cobran? (b) estación destino sin anclajes libres → ¿se mantiene el rechazo? | **D2**: minuto empezado → entero (ceil). **D5**: criterio existente **mantenido** |
| 3 | Permisos / estado | (a) ¿quién puede ejecutar la devolución? (b) abono caducado durante el viaje → ¿se puede devolver? | **D3**: solo el propietario con rol `usuario` (401/403). **D4**: sí, la devolución siempre está permitida |

## Decisiones (pregunta → respuesta → justificación → HU)

### D1 — Tarifa aplicable al excedente · recorrido normal · HU-03

- **Pregunta**: ¿Se aplica una única tarifa del servicio o cada usuario tiene
  su propia tarifa para el tiempo que excede de los 30 minutos?
- **Respuesta**: **Tarifa única del servicio** (`TARIFA_MINUTO`, global en
  configuración). No existe tarifa por usuario.
- **Justificación**: el enunciado solo dice «se calcula el importe según la
  tarifa del usuario» y los criterios de HU-03 en `historia.md` ya no
  mencionan tipo de tarifa, así que la narrativa **no determina una única
  respuesta** y el grupo debe decidirlo. Se elige la tarifa única porque:
  (a) «la tarifa del usuario» se lee naturalmente como *la tarifa aplicable
  al viaje de ese usuario*; (b) el modelo de datos compartido no prevé
  ningún campo de tarifa por usuario y añadirlo sería un cambio de esquema
  que exige acuerdo del equipo (constitución III); (c) la regla de los 30
  minutos gratuitos ya es una regla de servicio, coherente con una tarifa
  única; (d) ninguna de las 8 HU menciona tarifas individuales.
- **HU afectada**: HU-03 (FR-004, SC-002, Assumptions, Key Entities).

### D2 — Redondeo del excedente · caso límite · HU-03

- **Pregunta**: Cuando un viaje dura 30 min 45 s, ¿cómo se cobra el tiempo
  que excede de los 30 minutos iniciales?
- **Respuesta**: **Minuto empezado → se cobra entero** (redondeo hacia
  arriba): 30:45 → 1 minuto excedente → 0,20 € con `TARIFA_MINUTO = 0,20`.
- **Justificación**: SC-002 decía «tarifa × minutos excedentes» sin definir
  qué es un «minuto excedente», lo que hacía no testeable el aserto T3d. El
  redondeo hacia arriba es el único que (a) nunca deja de cobrar tiempo
  realmente consumido, (b) coincide con lo ya implementado en
  `services/devolucion/tarifa.go` (`ceil((segundos−1800)/60)`) y con los
  tests T3b/T3d, y (c) evita importes fraccionarios confusos en el
  historial (HU-04).
- **HU afectada**: HU-03 (FR-004, SC-002); efecto visible en HU-04.

### D3 — Permisos de la devolución · permisos · HU-03

- **Pregunta**: ¿Quién puede ejecutar la devolución (cierre) de una
  bicicleta?
- **Respuesta**: **Solo el usuario autenticado con rol `usuario`
  propietario del viaje activo**. Sin sesión → 401; operarios y gestores →
  403, sin efectos sobre viaje ni anclaje.
- **Justificación**: la devolución pertenece al ciclo de alquiler del
  usuario (desbloqueo HU-02 → cierre HU-03); los operarios actúan sobre
  averías/taller (HU-05/06) y el gestor solo lee datos (HU-08). Además,
  cerrar el viaje de otro usuario alteraría su historial y su facturación
  (HU-04). El propietario se determina por el JWT, nunca por parámetro de
  entrada (constitución I: autoridad del backend). Coincide con el plan
  (research D7) y con el middleware ya implementado.
- **HU afectada**: HU-03 (FR-008, edge cases); refuerza HU-04 (FR-005).

### D4 — Abono caducado durante el viaje · estado · HU-03

- **Pregunta**: Si el abono del usuario caduca mientras su viaje sigue en
  curso, ¿puede devolver la bicicleta?
- **Respuesta**: **Sí, siempre puede devolver** el propietario del viaje
  activo; el importe queda registrado. El abono solo condiciona el
  desbloqueo (HU-02), nunca el cierre.
- **Justificación**: si el cierre exigiera abono vigente, un abono que
  caduca a mitad de trayecto dejaría la bicicleta anclada y el viaje
  `en_curso` de forma permanente (el usuario no podría cerrar ni volver a
  desbloquear): bicicleta varada y usuario bloqueado. La deuda queda
  representada por el `coste_total` persistido en el viaje, consultable en
  el historial (HU-04). El enunciado solo condiciona el *uso* del servicio
  (abono para desbloquear), no el cierre.
- **HU afectada**: HU-03 (FR-009, edge cases); relación con HU-02.

### D5 — Estación destino sin anclajes libres · caso límite · HU-03

- **Pregunta**: ¿Se mantiene el criterio de que, si la estación destino no
  tiene ningún anclaje libre, la devolución no se completa y el viaje sigue
  activo?
- **Respuesta**: **Sí, se mantiene** (criterio existente confirmado, no
  cambiado).
- **Justificación**: es la única alternativa coherente con el dominio: no
  hay dónde fijar la bicicleta, por lo que el viaje no puede cerrarse. El
  usuario prueba otra estación (la HU-01 le muestra anclajes libres) y su
  viaje permanece `en_curso`, que es el estado correcto. Se rechaza la cola
  de espera y el retorno forzoso a la estación de origen: ambas añadirían
  alcance no pedido por el enunciado. Se documenta explícitamente para
  poder expresarlo como «Dado… cuando… entonces…».
- **HU afectada**: HU-03 (edge cases, US1/AC — mantenimiento de criterio).

## Criterios modificados (formato «Dado… cuando… entonces…»)

Añadidos/actualizados en `spec.md` (US1 Acceptance Scenarios 3–6):

1. **Dado** un viaje > 30 minutos, **cuando** se procesa la devolución,
   **entonces** se cobra el excedente con la tarifa única del servicio y
   redondeo al minuto empezado (30:45 → 1 minuto excedente). *(AC3,
   modificado — antes no fijaba tarifa ni redondeo)*
2. **Dado** un operario/gestor autenticado o sin sesión, **cuando** intenta
   ejecutar una devolución, **entonces** responde 403/401 y no modifica
   viaje ni anclaje. *(AC4, nuevo — FR-008)*
3. **Dado** un viaje activo cuyo propietario tiene el abono caducado,
   **cuando** fija la bicicleta en un anclaje libre, **entonces** el viaje
   se cierra con normalidad y el importe queda registrado. *(AC5, nuevo —
   FR-009)*
4. **Dado** un viaje de 30 min 45 s con `TARIFA_MINUTO = 0,20 €`,
   **cuando** el sistema calcula el importe, **entonces** el coste es
   0,20 € (nunca 0,00 € ni 0,15 €). *(AC6, nuevo — operacionaliza D2)*

## Diferencia entre la especificación anterior y la revisada

`git diff 0d6f478 -- specs/spec-003/spec.md` → **+67 / −11 líneas**.
Resumen de secciones tocadas:

| Sección | Antes | Después |
|---|---|---|
| `## Clarifications` | (no existía) | Session 2026-11-19 con 5 Q→A (D1–D5) |
| US1 Acceptance Scenarios | 3 escenarios | 6 escenarios (AC3 afinado; AC4–AC6 nuevos, en GWT) |
| Edge Cases | 7 casos | 9 casos (+ abono caducado, + permisos 403/401) |
| FR-004 | «según la tarifa vigente del servicio» | tarifa única + redondeo al minuto empezado |
| FR-008 / FR-009 | (no existían) | permisos de cierre; devolución con abono caducado |
| SC-002 | «la tarifa × minutos excedentes» | `TARIFA_MINUTO` × minutos redondeados hacia arriba |
| Assumptions (tarifa) | «tarifa por unidad de tiempo… fuera de alcance» | tarifa única del servicio + referencia a la clarificación |
| Key Entities (Tarifa) | «regla de precio del servicio» | «única para todo el servicio… redondeo al minuto empezado» |

El diff completo se conserva en el commit de esta rama
(`docs(spec-003): registrar decisiones de clarify y actualizar spec`).

## Re-validación del checklist de calidad

`specs/spec-003/checklists/requirements.md`: **16/16 → 16/16** (sin
regresiones ni nuevos checked). La actualización no deja marcadores
`[NEEDS CLARIFICATION]`, los criterios siguen siendo testeables y medibles,
y los nuevos FR-008/FR-009 quedan expresados en términos observables
(códigos HTTP y efectos sobre el estado).

## Notas de proceso

- Las 5 preguntas se respondieron por el responsable de la spec (Alex
  Morell) a partir de la narrativa del enunciado; D1 es una decisión del
  grupo justificada porque el enunciado no determina una única respuesta.
- La sesión no detectó ambigüedades adicionales: no se inventan problemas
  para completar la actividad.
- Toda modificación de esta spec queda sujeta a PR revisado por otro
  integrante del grupo (constitución, Acuerdos de Colaboración).
