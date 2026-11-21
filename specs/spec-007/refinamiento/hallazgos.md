# Hallazgos — spec-007 / HU-07

**Fecha de registro**: 2026-11-20 · **Responsable por defecto**: Alberto Copa (spec-007)
**Origen**: `/speckit-analyze` inicial → [analyze-inicial.md](./analyze-inicial.md)
**Convención de estado**: `Abierto` (pendiente de corrección) · `Corregido` (verificado en repetición de analyze) · `Pendiente de acuerdo del equipo` (no se toca sin consenso, const. III)

| ID | Criterio | Gravedad | Evidencia | Decisión | Responsable | Estado |
|----|----------|----------|-----------|----------|-------------|--------|
| H-001 | FR-005 / SC-002 — desempate del orden del panel | HIGH | clarify Q3 (spec.md Clarifications Session 2026-11-20) fijaba «tie-break by station ID (ascending)»; data-model.md L83-86, handler.go L29-32, tasks.md T012 y quickstart.md L104 («empate de score → `fecha_generacion` ASC ✅») usan/validan `fecha_generacion` ASC | **(b) Aceptada por Alberto el 2026-11-20**: enmendar spec (FR-005 + bullet de Clarifications) y clarify.md a desempate por `fecha_generacion` ASC (alerta más antigua primero). Justificación: es el criterio ya implementado, testeado y validado en T7d/quickstart; determinista igualmente; cambiar código/data-model/quickstart reabriría trabajo ya validado y añadiría riesgo de regresión sin beneficio. No se tocó código | Alberto Copa | Corregido |
| H-002 | Trazabilidad tareas ↔ requisitos | MEDIUM | tasks.md L81 (T020): «barrido de FR-001…FR-008»; la spec post-clarify tiene FR-009 | Enumeración ampliada a **FR-001…FR-009** en T020 | Alberto Copa | Corregido |
| H-003 | plan.md refleja el clarify 2026-11-20 | MEDIUM | plan.md L33 «0 NEEDS CLARIFICATION» sin mención de Q1–Q4; Q3 además contradecía el plan (H-001) | Input de plan.md y sección Unknowns actualizados: sesión clarify 2026-11-20 (Q1/Q2/Q4 ya cubiertas por V1/V7/T011/T013/T008/T016; Q3 resuelta por H-001). Constitution Check re-verificado: PASS sin cambios (el desempate no afecta a I–V) | Alberto Copa | Corregido |
| H-004 | Terminología del rol del panel | MEDIUM | spec.md US1/Assumptions «operario de transporte y calle»/«operario de calle» ↔ data-model V1, T013 y seed: rol `operario_logistica` | Término canónico **«operario de logística»** establecido en US1 (primera frase) y Assumptions, con los sinónimos de la narrativa de la HU-07 registrados una sola vez y el rol `operario_logistica` mapeado | Alberto Copa | Corregido |
| H-005 | FR-009 forma de la denegación | LOW | spec.md FR-009 «vista de error»; contracts/balanceo.md §autorización define `403 ROL_NO_PERMITIDO` y `401` sin token (data-model V1) | FR-009 referencia ahora `contracts/balanceo.md` como forma de la denegación (incluye peticiones sin autenticar), sin volcar códigos HTTP en el spec | Alberto Copa | Corregido |
| H-006 | T021 — auditoría `traslados_balanceo` | INFO | tasks.md T021; data-model.md §"Migraciones aprobadas"; spec.md Key Entities | El acuerdo posterior del equipo del 2026-11-20 revocó la decisión inicial de descarte y aprobó la tabla de auditoría. Migración `0004_spec-007_traslados.sql`; el servicio registra origen/destino/unidades/operario dentro de la transacción. T7f verifica la fila y T7g que traslados rechazados no se auditan | Equipo (registro: Alberto Copa) | Aprobado, implementado y validado con la suite completa de Go |
| H-007 | T022 — score de urgencia del panel | INFO | tasks.md T022; data-model.md §"Migraciones aprobadas"; handler.go | El acuerdo posterior del equipo del 2026-11-20 revocó la decisión inicial de descarte y aprobó la solución más simple: un índice parcial de expresión, sin cambiar el SQL del handler. Migración `0005_spec-007_score.sql`; T7d verifica el índice y conserva el orden validado | Equipo (registro: Alberto Copa) | Aprobado, implementado y validado con la suite completa de Go |

## Resumen

- Abiertos: 0 · Cerrados: 7 (H-001…H-007) · Bloqueados: 0
- Actualización posterior: el equipo revocó el descarte inicial de T021/T022 y aprobó las migraciones 0004/0005; ver el addendum de `clarify.md`.
- Validación final: `go test -p 1 ./...`, `go vet ./...`, `npm run build` y las 97 pruebas frontend pasan; ver [converge-1.md](./converge-1.md).
- `/speckit-converge` completado: no encontró brechas ni añadió tareas. La HU-07 queda lista para revisión cruzada.

## Revisión cruzada de la PR

**Fecha**: 2026-11-21
**Alcance**: revisión local de `refinamiento/spec-007` en el commit `9180ff4`,
comparada con `origin/main`. La HU-07 contiene una única historia de usuario;
se seleccionan varios criterios representativos, incluyendo un criterio de
funcionalidad, uno de estado y uno de permisos.

| Historia/criterio revisado | Tarea y evidencia comprobadas | Resultado |
| --- | --- | --- |
| HU-07 / FR-005 y SC-002: panel ordenado por urgencia | T012, T018, T7d y `analyze-final.md`; el backend devuelve alertas abiertas con nombre de estación y orden por score y fecha | Cobertura correcta, sin incidencia |
| HU-07 / FR-006 y FR-007: traslado válido y estados recalculados | T011, T021, T7f; el traslado `34 → 6` de 1 unidad devolvió ambos estados recalculados | Cobertura correcta, sin incidencia |
| HU-07 / FR-009: autorización del operario de logística | T013, T7e; un usuario normal recibió `403` en el panel | Cobertura correcta, sin incidencia |
| HU-07 / SC-001 y SC-004: generación y resolución de alertas | T010, T018, T019, T7a/T7b/T7h; la suite cubre umbrales, unicidad, worker y resolución | Cobertura correcta, sin incidencia |
| HU-07 / caso límite de capacidad | T006, T011, T7g; un traslado de 999 unidades devolvió `409` sin aplicar el movimiento | Cobertura correcta, sin incidencia |

### Verificación ejecutada

- `go test -p 1 ./...`: correcto, incluyendo `backend/tests` (92,698 s),
  contratos e integración.
- `go vet ./...`: correcto.
- `npm test -- --run --reporter=dot`: 16 ficheros y 97 pruebas superadas.
- `npm run build`: correcto.
- Recorrido manual: login de `logistica@si16.test`, panel con 7 alertas,
  traslado válido `34 → 6`, respuesta `403` para `ana@si16.test`, respuesta
  `409` para capacidad insuficiente y ruta `/balanceo` con `200`.

**Resultado de la revisión**: no se han encontrado incidencias accionables ni
contradicciones entre spec, plan, tareas, código y pruebas. No se añaden tareas
correctoras. El merge queda condicionado únicamente a la revisión formal de la
PR por el equipo.
