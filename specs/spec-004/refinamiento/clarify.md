# Registro de decisiones — clarify (spec-004 · HU-04)

**Rama**: `spec-004` · **Commit de partida**: `0d6f478d511427172aa8975fb7f9c4d3b3db1eb5`
(main tras fast-forward desde `10ff0bf`) · **Spec Kit**: `1.1.3` (`.specify/init-options.json`)
**Fecha de la sesión de refinamiento**: 2026-11-19 · **Funcionalidad activa**: `specs/spec-004`
**Comando**: `/speckit.clarify` (integración opencode, `.opencode/commands/speckit.clarify.md`)
**Spec resultante**: `specs/spec-004/spec.md` § Clarifications (Session 2026-10-28 + Session 2026-11-19)

Esta spec ya contaba con una sesión previa (2026-10-28, 5 preguntas). La sesión
de refinamiento de la fase 4 añade la sesión 2026-11-19 para completar los tres
casos exigidos por el enunciado. Ambas se registran aquí con respuesta,
justificación y HU afectada.

## Casos probados por historia (3/3 exigidos)

| # | Categoría | Caso probado | Sesión | Resultado |
|---|-----------|--------------|--------|-----------|
| 1 | Recorrido normal | Lista cronológica, campos, presentación; y qué ve el usuario en la primera carga sin filtro | 2026-10-28 (Q3, Q4) + 2026-11-19 (Q3) | Formato `dd/mm/aaaa HH:MM` Europe/Madrid; primera carga = todo el historial sin filtro, más recientes primero, sumatorio total |
| 2 | Entrada inválida / caso límite | Paginación con muchos viajes, página fuera de rango, periodo vacío | 2026-10-28 (Q1, Q5) | Página fuera de rango → `items` vacío **sin error** con el sumatorio del periodo; el 0,00 € se reserva a periodos sin viajes |
| 3 | Permisos / estado | ¿Se puede ver historial ajeno? ¿Aparece el viaje en curso? | 2026-11-19 (Q1, Q2) | Solo historial propio (propietario desde el JWT, nunca por parámetro; sin sesión → 401); viaje activo no se lista hasta que HU-03 lo cierra |

## Decisiones (pregunta → respuesta → justificación → HU)

### Sesión 2026-11-19 (refinamiento fase 4)

#### C1 — Permisos del historial · permisos · HU-04

- **Pregunta**: ¿Puede algún usuario consultar el historial de viajes de
  otra persona, y qué ve un operario o un gestor cuando entra en «Mis
  Viajes»?
- **Respuesta**: **Solo historial propio**. El propietario se resuelve
  siempre desde la sesión (JWT), nunca por parámetro de entrada; cualquier
  rol autenticado (usuario, operario, gestor) consulta su propio
  historial; sin sesión → 401.
- **Justificación**: FR-005 ya prohibía ver viajes ajenos pero no fijaba
  *cómo* se garantiza ni qué reciben los otros roles. Con JWT como única
  fuente no existe forma de pedir historial ajeno (mínimo privilegio);
  operarios y gestores también pueden ser usuarios del servicio y deben
  ver *su propio* viaje — la HU-04 es una consulta sin efectos, a
  diferencia de la devolución (HU-03), que sí se restringe a rol
  `usuario`. Un endpoint de lectura de historiales ajenos añadiría alcance
  no pedido por el enunciado.
- **HU afectada**: HU-04 (FR-005, AC4); refuerzo del aislamiento
  compartido con HU-03.

#### C2 — Viaje activo en el historial · estado · HU-04

- **Pregunta**: ¿Debe aparecer el viaje activo (en curso) en el historial
  «Mis Viajes» mientras el usuario sigue andando?
- **Respuesta**: **No; solo viajes finalizados**. El viaje en curso no se
  lista ni computa en el sumatorio; entra en el historial cuando HU-03 lo
  cierra.
- **Justificación**: un viaje en curso no tiene todavía
  `estacion_destino_id`, `fecha_fin` ni `coste_total` (los CHECKs del
  modelo compartido lo exigen así), por lo que no puede mostrarse con los
  campos que la HU-04 exige ni sumarse a un gasto aún no devengado. El
  criterio ya existía como caso límite del spec; se confirma y se eleva a
  criterio de aceptación explícito para poder probarlo.
- **HU afectada**: HU-04 (AC5, edge cases); dependencia de HU-03 (FR-005
  de spec-003 persiste el cierre).

#### C3 — Vista inicial sin filtro · recorrido normal · HU-04

- **Pregunta**: Cuando el usuario entra en «Mis Viajes» por primera vez
  sin tocar ningún filtro, ¿qué debe ver y con qué sumatorio?
- **Respuesta**: **Todos sus viajes finalizados, más recientes primero, con
  el sumatorio de todo el historial**; el filtro mes/año es opcional y solo
  estrecha la lista.
- **Justificación**: el criterio de HU-04 solo describía la vista *con*
  filtro; sin fijar la vista inicial, la primera carga podía interpretarse
  como «mes en curso» o «año completo». FR-003 ya exigía el sumatorio «del
  historial completo cuando no hay filtro», así que la opción elegida es
  la única coherente con el propio spec, con el caso límite «usuario sin
   viajes» y con el estado inicial `sin filtro` del plan (research D7).
- **HU afectada**: HU-04 (FR-001 orden, FR-003, AC6).

### Sesión 2026-10-28 (previa, registrada aquí por trazabilidad)

#### C4 — Presentación de la lista · recorrido normal · HU-04

- **Pregunta**: ¿Debe mostrarse en cada fila del historial la fecha y hora
  de inicio de cada viaje?
- **Respuesta**: Sí; cada fila muestra fecha y hora de inicio en formato
  `dd/mm/aaaa HH:MM`.
- **Justificación**: los criterios de HU-04 nombran origen, destino,
  duración y coste pero no la fecha visible; sin ella la lista no permite
  localizar un cargo concreto (fin de «no entiendo un cargo» de la
  narrativa). Se formatea en Europe/Madrid para que la fila y el filtro
  hablen del mismo calendario.
- **HU afectada**: HU-04 (FR-001).

#### C5 — Paginación server-side · caso límite · HU-04

- **Pregunta**: ¿Cómo se entrega la lista cuando el periodo contiene
  muchos viajes?
- **Respuesta**: Paginación server-side con `page`/`page_size` del
  contrato global (50 por defecto, máx. 200); el sumatorio se calcula
  sobre el periodo completo sea cual sea la página visible.
- **Justificación**: la primera respuesta había fijado `limit` 20/100
  divergiendo del OpenAPI global; se adopta el contrato compartido sin
  modificar el fichero (const. III) y el sumatorio del periodo evita que
  el gasto mostrado dependa de la página.
- **HU afectada**: HU-04 (FR-008, SC-002).

#### C6 — Zona horaria del filtro · caso límite · HU-04

- **Pregunta**: ¿En qué zona horaria se interpretan el mes y el año de
  cada viaje al aplicar el filtro?
- **Respuesta**: Europe/Madrid (hora local española).
- **Justificación**: un viaje iniciado el día 31 a las 23:30 hora
  española debe caer en el mes que ve el usuario; con UTC ese viaje
  saltaría de mes según el huso del servidor. La conversión se hace en
  PostgreSQL (const. I: la regla vive en el backend).
- **HU afectada**: HU-04 (FR-002, SC-002).

#### C7 — Contrato `GET /viajes` · caso límite · HU-04

- **Pregunta**: El OpenAPI global define `desde`/`hasta` + `page`/
  `page_size`, mientras que la primera respuesta fijó filtro mes/año +
  `limit` 20/100. ¿Qué contrato adoptamos?
- **Respuesta**: El contrato global, sin modificar `openapi.yaml`: el
  filtro mes/año se expresa como rango `desde`/`hasta` (lo construye la
  UI) y la paginación usa `page`/`page_size` (50/200).
- **Justificación**: el OpenAPI es un acuerdo compartido por las 8 specs
  (const. III); adaptar la UI al rango de fechas es un detalle de
  contrato y no cambia el criterio «filtrar por mes y año» de la HU.
- **HU afectada**: HU-04 (FR-002, FR-008).

#### C8 — Sumatorio con página fuera de rango · caso límite · HU-04

- **Pregunta**: Si el periodo tiene viajes pero la página pedida está fuera
  de rango, ¿qué sumatorio devuelve la API?
- **Respuesta**: `items` vacío pero con el sumatorio real del periodo
  completo; el 0,00 € se reserva a los periodos sin viajes.
- **Justificación**: el sumatorio es propiedad del periodo, no de la
  página visible; devolver 0,00 € por pedir una página inexistente
  habría mentido sobre el gasto del usuario (y SC-002 lo prohíbe).
- **HU afectada**: HU-04 (FR-008, SC-002).

## Criterios modificados (formato «Dado… cuando… entonces…»)

Añadidos/actualizados en `spec.md` (US1 Acceptance Scenarios 4–6 y FR):

1. **Dado** un usuario con sesión iniciada de cualquier rol, **cuando**
   consulta su historial, **entonces** ve únicamente sus propios viajes
   finalizados, resueltos desde la sesión y nunca desde un parámetro; sin
   sesión la API responde 401 y no se muestra ningún dato. *(AC4, nuevo —
   FR-005 reforzado)*
2. **Dado** un usuario con un viaje en curso, **cuando** abre «Mis
   Viajes», **entonces** el viaje activo no aparece en la lista ni computa
   en el sumatorio hasta que HU-03 lo cierra como finalizado. *(AC5,
   nuevo — formaliza un caso límite existente)*
3. **Dado** un usuario que entra en «Mis Viajes» sin aplicar ningún
   filtro, **cuando** la vista se renderiza por primera vez, **entonces**
   ve todos sus viajes finalizados ordenados por fecha de inicio
   descendente con el sumatorio de todo el historial. *(AC6, nuevo —
   C3; refuerza FR-001 con el orden y añade edge case de sesión 401)*

## Diferencia entre la especificación anterior y la revisada

`git diff 0d6f478 -- specs/spec-004/spec.md` → **+44 / −6 líneas**.
Resumen de secciones tocadas:

| Sección | Antes | Después |
|---|---|---|
| `## Clarifications` | Session 2026-10-28 (5 Q→A) | + Session 2026-11-19 (3 Q→A: C1–C3) |
| US1 Acceptance Scenarios | 3 escenarios | 6 escenarios (AC4–AC6 nuevos, en GWT) |
| Edge Cases | 6 casos | 7 casos (+ sesión ausente/expirada → 401) |
| FR-001 | «lista cronológica…» | + orden por fecha de inicio descendente |
| FR-005 | «nunca los viajes de otros usuarios» | + propietario desde JWT, nunca por parámetro; 401 sin sesión; cualquier rol autenticado |
| Assumptions (orden) | «más recientes primero cuando la HU no fija dirección» | remite a FR-001 y a la clarificación 2026-11-19 |

El diff completo se conserva en el commit de esta rama
(`docs(spec-004): registrar decisiones de clarify y actualizar spec (fase 4)`).

## Re-validación del checklist de calidad

`specs/spec-004/checklists/requirements.md`: sin regresiones; los ítems
siguen en verde (marcadores, testeabilidad, medición y casos límite
actualizados con los nuevos AC/FR).

## Notas de proceso

- **Numeración**: este documento numera sus decisiones **C1–C8**; el
  `research.md` conserva su numeración propia **D1–D8** (decisiones
  técnicas). No colisionan (hallazgo A6 del analyze 2026-11-19).
- C4–C8 (sesión 2026-10-28) ya estaban codificadas como Q→A en `spec.md`;
  aquí se les añade la justificación que el enunciado de clarify exige,
  extraída de `research.md` (D1–D8).
- C1–C3 (sesión 2026-11-19) son la aportación de la fase 4: cubren los
  casos de permisos y estado que faltaban y concretan la vista inicial.
- Toda modificación queda sujeta a PR revisado por otro integrante del
  grupo (constitución, Acuerdos de Colaboración).
