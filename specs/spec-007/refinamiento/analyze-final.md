# Informe final de análisis — spec-007 / HU-07 (post-corrección)

**Fecha**: 2026-11-20 · **Rama**: `refinamiento/spec-007` · **Spec Kit**: 1.1.3
**Nota de vigencia**: este informe refleja el estado al finalizar el primer análisis, antes del acuerdo posterior que revocó el descarte de T021/T022. El addendum al final registra ese cambio y su implementación.
**Artefactos analizados**: [spec.md](../spec.md) (clarify Session 2026-11-20 + enmienda H-001), [plan.md](../plan.md) (Input/Unknowns actualizados), [tasks.md](../tasks.md) (T020 actualizado), [data-model.md](../data-model.md), [quickstart.md](../quickstart.md), [contracts/balanceo.md](../contracts/balanceo.md), `.specify/memory/constitution.md` (v1.0.1)
**Predecesor**: [analyze-inicial.md](./analyze-inicial.md) (pre-corrección) · **Registro de decisiones**: [hallazgos.md](./hallazgos.md)
**Correcciones aplicadas**: H-001…H-005 (opción (b) en H-001, decidida por Alberto). **T021/T022 no se tocaron** (pendientes de acuerdo del equipo, const. III).

## Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| H-001 | Inconsistency | ~~HIGH~~ → **RESUELTO** | spec.md FR-005 + bullet Clarifications; refine/clarify.md Q3 | Desempate enmendado a `fecha_generacion` ASC: spec, clarify, data-model, handler, quickstart y código quedan **coherentes** | Ninguna (verificado en este pase) |
| H-002 | Coverage | ~~MEDIUM~~ → **RESUELTO** | tasks.md T020 | Enumeración ampliada a FR-001…FR-009 | Ninguna |
| H-003 | Inconsistency | ~~MEDIUM~~ → **RESUELTO** | plan.md Input + Unknowns | Clarify 2026-11-20 referenciado; Q1/Q2/Q4 confirmadas ya cubiertas por el plan; Q3 resuelta por H-001; Constitution Check re-verificado PASS | Ninguna |
| H-004 | Terminology | ~~MEDIUM~~ → **RESUELTO** | spec.md US1 + Assumptions | Término canónico «operario de logística» con sinónimos de la HU-07 y rol `operario_logistica` mapeados | Ninguna |
| H-005 | Ambiguity | ~~LOW~~ → **RESUELTO** | spec.md FR-009 | Denegación referenciada a `contracts/balanceo.md` (incluye peticiones sin autenticar), sin volcar códigos HTTP en el spec | Ninguna |
| R-001 | Terminology | LOW | spec.md L22 (título US1), tasks.md L46, research.md L51 | Títulos/registros históricos conservan «operario de calle»/«operario de transporte y calle» | **Aceptado**: el título de la US se mantiene idéntico en tasks.md (referencias cruzadas); research.md es registro de decisiones pasadas. El cuerpo del spec ya usa el término canónico (H-004) |
| H-006 | Pending | INFO | tasks.md T021 | Auditoría `traslados_balanceo` — intacta | Mantener [BLOCKED] hasta acuerdo del equipo |
| H-007 | Pending | INFO | tasks.md T022 | Score materializado — intacto | Mantener [BLOCKED] hasta acuerdo del equipo |

**Sin hallazgos nuevos** en este pase. Duplicaciones: 0. Placeholders: 0. Contradicciones residuales del clarify: 0 (las únicas menciones de «station ID» que quedan son el registro histórico de la respuesta original en clarify.md y la nota de enmienda, ambas intencionadas).

## Coverage Summary Table

| Requirement Key | Has Task? | Task IDs | Notes |
|---|---|---|---|
| FR-001 ocupación % | ✔ | T005, T009 | |
| FR-002 alerta <10 %>90 % | ✔ | T005, T007 (T7a), T009 | |
| FR-003 única alerta abierta | ✔ | T007 (T7b), T009 | |
| FR-004 resolución automática | ✔ | T007 (T7b), T009 | |
| FR-005 panel ordenado + empate | ✔ | T007 (T7d), T012 | **Coherente**: empate `fecha_generacion` ASC = data-model/handler/quickstart/código |
| FR-006 traslado + validaciones | ✔ | T006, T011, T007 (T7g) | Incluye validación de destino (Q1) |
| FR-007 recálculo ≤30 s | ✔ | T011, T016, T019 | |
| FR-008 averiadas excluidas | ✔ | T005, T007 (T7c), T009 | |
| FR-009 acceso operario | ✔ | T013, T007 (T7e), T016 | `RequireRole("operario_logistica")`; forma de denegación en contracts (Q2) |
| SC-001 alerta <60 s | ✔ | T010, T007 (T7h), T019 | |
| SC-002 panel íntegro | ✔ | T007 (T7d/T7b), T008, T016 | Incluye estado vacío (Q4) |
| SC-003 vista inmediata + ≤30 s | ✔ | T011, T019 | |
| SC-004 0 alertas obsoletas | ✔ | T007 (T7b), T019 | |

## Constitution Alignment Issues

Ninguna. Re-verificación tras las correcciones: **PASS (I–V, 0 violaciones)**.

- Las enmiendas son de documentación/spec: no tocan stack (II), esquema (III — sin migraciones; T021/T022 siguen bloqueados), frescura (IV), reglas server-side (I) ni independencia (V).
- Gate de calidad: pruebas de reglas críticas intactas (T7a/T7b/T7g).

## Unmapped Tasks

Ninguna (igual que en el informe inicial): T001–T020 mapean a FRs/SCs o setup/polish; T021/T022 a las propuestas PENDIENTE (bloqueadas por diseño).

## Metrics

- **Total Requirements**: 13 (FR-001…FR-009 = 9; SC-001…SC-004 = 4)
- **Total Tasks**: 22 (T001–T020 [X]; T021–T022 [BLOCKED] — intactos)
- **Coverage %**: 100 % (13/13 con ≥1 tarea)
- **Ambiguity Count**: 0 (H-005 resuelto; R-001 aceptado como LOW)
- **Duplication Count**: 0
- **Critical Issues Count**: 0 · HIGH: 0 (H-001 resuelto)
- **Hallazgos**: 7 → 5 Corregidos, 2 Pendiente de acuerdo del equipo, 0 Abiertos

## Validación ejecutada

1. `git diff` revisado: solo docs (spec.md, plan.md, tasks.md, checklists/requirements.md sin cambios, archivos de refinamiento).
2. Búsqueda de residuos (`station ID`/`identificador de estación`/`FR-001…FR-008`): sin contradicciones; solo registros históricos intencionados (clarify respuesta original, analyze-inicial, nota de enmienda).
3. Checklist de calidad `checklists/requirements.md`: **16/16 → 16/16** (ningún marcador cambió; sin NEEDS CLARIFICATION; sin fugas de implementación: los nombres citados — `fecha_generacion`, `operario_logistica`, `contracts/balanceo.md` — son vocabulario de dominio/contrato, no stack).
4. Constitution Check re-verificado: PASS.
5. Cobertura FR/SC ↔ tareas: 100 %.

## Next Actions

- Informe final disponible: informe inicial + hallazgos + este informe completan el paso 3 de `revisión-final-bicicletas.txt`.
- **No se avanza a `/speckit.plan`** (decisión del usuario). En el momento de este informe, T021/T022 seguían pendientes de acuerdo; su aprobación e implementación posterior quedan registradas en el addendum.

## Addendum — acuerdo e implementación de T021/T022

Posteriormente, el equipo aprobó implementar las dos tareas, revocando la decisión inicial de descartarlas. Se añadieron `0004_spec-007_traslados.sql` (tabla de auditoría con inserción atómica desde el servicio) y `0005_spec-007_score.sql` (índice parcial de expresión para el orden del panel). La definición del esquema se actualizó también en `specs/spec-001/data-model.md`.

Las pruebas de integración T7d/T7f/T7g verifican el índice, el registro de traslados exitosos y la ausencia de registros para traslados rechazados. La suite completa `go test -p 1 ./...` pasó contra una PostgreSQL temporal aislada; también pasaron `go vet ./...`, `npm run build` y las 97 pruebas frontend. El informe original conserva el resultado del análisis en su momento; la validación de cierre está en [converge-1.md](./converge-1.md).
