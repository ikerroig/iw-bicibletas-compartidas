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
| Decisiones del clarify | _(p. ej., AC4 / AC5 / AC6: abono caducado no bloquea devolución, tarifa y redondeo por minuto empezado)_ | _(lectura de `clarify.md` y contraste con la narrativa)_ | | |
| Cobertura de un criterio en la matriz | _(p. ej., FR-003 / SC-001 tramo gratuito ≤ 30 min → T014, T016 o FR-008 permisos → T028)_ | _(tarea, test previsto y ejecución de pruebas)_ | | |
| Registro de hallazgos | _(p. ej., A1 permisos o A2 abono vencido)_ | _(contraste con el informe analyze y los tests)_ | | |

**Firma del revisor**: _(Iker Roig — fecha)_
