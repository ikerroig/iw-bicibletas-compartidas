# Registro de hallazgos del analyze — spec-003 (HU-03)

**Enunciado**: §5.2, paso 3. Contraste de cada hallazgo del
[informe inicial](./analyze-inicial.md) con el material del proyecto.
**Responsable**: Alex Morell · **Revisor**: Iker Roig (pendiente) ·
**Fecha**: 2026-11-19

**Estados**: *Corregido* = el documento ya está corregido · *Tareas añadidas*
= la corrección documental está hecha y las pruebas quedan para
`/speckit-implement` (§5.3) · *Aceptado* = no requiere cambio, con
justificación.

| ID | Criterio | Gravedad | Evidencia | Decisión | Responsable | Estado |
|----|----------|----------|-----------|----------|-------------|--------|
| A1 | HU-03 · FR-008 / AC4 (401 sin sesión, 403 rol ≠ usuario) | ALTA | `RequireAuth`+`RequireRole("usuario")` en `routes/spec_003_devolucion.go:14-17` (cumple), pero `backend/tests/devolucion_test.go` no tiene ningún aserto 401/403 para `POST /devoluciones` y el quickstart acaba en T3j | **Válido.** El código cumple pero el criterio está sin verificar. Se añade la tarea **T028** (test de integración 401/403 + fila en el checklist) | Alex Morell | Tareas añadidas |
| A2 | HU-03 · FR-009 / AC5 (devolución con abono caducado) | ALTA | Cero tareas mencionan FR-009; `services/devolucion/service.go` no consulta `abono_activo` (cumple por omisión); el seed ya crea un usuario con abono vencido (T008) pero nadie prueba el cierre con él | **Válido.** Riesgo real de regresión si alguien añade un chequeo de abono al cierre. Se añade la tarea **T029** (test de integración: usuario con abono vencido cierra viaje → 200) | Alex Morell | Tareas añadidas |
| A3 | HU-03 · FR-004, SC-002, AC6 («TARIFA_MINUTO» en la spec) | MEDIA | `checklists/requirements.md` exige «No implementation details»; el nombre de variable aparece en requisitos normativos tras el clarify | **Válido.** Se reformula en la spec como «tarifa unitaria del servicio» (conservando el ejemplo 0,20 € para que SC-002 sea testeable); `TARIFA_MINUTO` sigue en Clarifications como registro de la decisión y en plan/tasks | Alex Morell | Corregido |
| A4 | HU-03 · checklist (tabla Acceptance Coverage) | MEDIA | La tabla llega a FR-007/SC-004; sin filas de FR-008, FR-009 ni AC4–AC6 | **Válido.** Se añade la tarea **T031** para ampliar la tabla apuntando a las pruebas de T028/T029/T030 (se completa al ejecutarlas) | Alex Morell | Tareas añadidas |
| A5 | HU-03 · `plan.md` (Unknowns) | BAJA | «0 NEEDS CLARIFICATION… se resuelven en research.md», sin mencionar la sesión 2026-11-19 | **Válido.** Se actualiza la nota y el Summary con la sesión de clarify y sus FR | Alex Morell | Corregido |
| A6 | HU-03 · `tasks.md` (cabecera Tests) | BAJA | Dice «T3a–T3i»; el quickstart define T3j y T015 ya la usa | **Válido.** Referencia corregida a T3a–T3j | Alex Morell | Corregido |
| A7 | HU-03 · AC6 vs T3d (fracción > 1 min) | BAJA | T3d prueba 30:01 (ceil → 1 min); ningún caso con excedente fraccionario mayor (p. ej. 46:30 → 17 min) | **Válido, menor.** Se añade el aserto en el test unitario de tarifa como tarea **T030** | Alex Morell | Tareas añadidas |

## Resumen

| | Total | Válidos | Falsos positivos | Aceptados sin cambio |
|---|---|---|---|---|
| Hallazgos del analyze | 7 | 7 | 0 | 0 |

Ningún hallazgo crítico. Los dos altos (A1, A2) son huecos de **verificación**
sobre comportamiento que el código ya cumple: se cierran añadiendo pruebas,
no tocando el comportamiento. Ninguna decisión funcional pendiente → no hace
falta volver a clarify.

## Dependencias con historias de otros integrantes

- **spec-002 (HU-02, Iker)**: `abono_activo` solo se valida en el
  desbloqueo; FR-009 **no** pide cambios en spec-002 (el cierre no consulta
  el abono). Sin acuerdo de equipo necesario (const. III).
- **spec-004 (HU-04, propio)**: consume `fecha_fin`/`coste_total` que
  FR-005 persiste; el redondeo de AC6 es visible en el historial — ya
  reflejado en el clarify de spec-004 (D2).
- **spec-005 (HU-05, Josep)**: T3h (bici averiada durante el viaje) sigue
  intacto tras el merge de main.

## Revisión entre compañeros (§5.1 y §5.2) — a rellenar por Iker Roig

> El enunciado pide que el revisor contraste las decisiones del clarify, elija
> **al menos un criterio** de la historia y compruebe que la tarea propuesta
> permitiría verificarlo. Si no hay incidencias reales, se documenta la
> cobertura comprobada, sin inventarse errores.

| Qué se revisó | Criterio / decisión | Cómo se comprobó | Resultado | Fecha |
|---------------|---------------------|------------------|-----------|-------|
| Decisiones del clarify | D1–D5, en especial **D4** (AC5: el abono caducado no impide devolver) y **D2** (AC6: minuto empezado) | Lectura de `clarify.md` y contraste con la narrativa y con HU-02 (spec-002, FR-002/FR-009: el abono solo condiciona el desbloqueo) | **De acuerdo con las 5 decisiones.** D4 evita un estado sin salida (bici sin anclar y usuario bloqueado) y es coherente con HU-02: el abono se comprueba al desbloquear, no al cerrar. D2 no se apoya solo en «ya funciona así»: también aporta que nunca deja tiempo consumido sin cobrar. D3 coincide con HU-02 FR-009 (solo el rol `usuario` alquila) | 2026-11-19 |
| Cobertura de un criterio en la matriz | **FR-008** (AC4, permisos → T028), **FR-004/SC-002** (AC6, redondeo → T030) y el contrato de `codigo_anclaje` ≤ 64 caracteres (F1 → T032) | Tarea → test → ejecución: `TestDevolucion_Permisos_SinSesionYRolNoUsuario` (401 sin sesión; 403 `ROL_NO_PERMITIDO` con `operario_taller` y `gestor`; viaje `en_curso` y anclaje intacto), `TestCalcularCoste_FraccionesDeExcedente` (30:45 → 0,20 €; 46:30 → 3,40 €) y `TestDevolucion_CodigoLargo_ValidacionContrato`. Reproducción propia del antes de T032 sobre `9dbed49~1`: 65 caracteres → `409 ANCLAJE_NO_ENCONTRADO`. Fusión simulada `main` (con #9) + #11: sin conflictos, `go vet` y `go test ./... -count=1` en verde, `npm test` 97/97 | **Cobertura verificada.** Las tareas permiten verificar los criterios y los tests fallan si se rompen (comprobado con T032). Las filas HU-03 de la matriz solo citan tests que existen en la rama | 2026-11-19 |
| Registro de hallazgos y converge | A1, A2 y A7 (huecos de verificación) · F1 (contrato ≤ 64) | Contraste con `analyze-inicial.md`, el código (`routes/spec_003_devolucion.go`, `services/devolucion/service.go`, `handler.go`) y los tests | **Las decisiones del registro se sostienen** (A1/A2 eran huecos de prueba, no de comportamiento). La revisión encontró **4 incidencias reales**, todas corregidas en `c10ad5a` y `09a76a8`: (1) faltaba el antes/después de T032 (§5.3.4); (2) T028–T033 sin responsable ni revisor (§5.3.3); (3) `len()` contaba bytes y no caracteres (33 «Ñ» → 422 indebido; ahora `utf8.RuneCountInString` + test); (4) la matriz incluía filas de HU-04 con tests de otra rama. También se corrigieron los opcionales: rol `gestor` en AC4, comentario de código de error y evidencia de la prueba previa con un usuario real del seed (`ana@si16.test`) | 2026-11-19 |

**Resultado general de la revisión**: sin incidencias abiertas. Las 4
incidencias de la revisión (comentarios de la PR #11) quedan resueltas y
verificadas; la HU-03 queda trazada, cubierta y con pruebas en verde.

**Firma del revisor**: Iker Roig — 2026-11-19
