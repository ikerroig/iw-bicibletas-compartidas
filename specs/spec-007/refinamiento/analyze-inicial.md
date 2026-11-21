# Informe inicial de análisis — spec-007 / HU-07 (pre-corrección)

**Fecha**: 2026-11-20 · **Rama**: `refinamiento/spec-007` · **Spec Kit**: 1.1.3
**Artefactos analizados**: [spec.md](../spec.md) (con clarify Session 2026-11-20), [plan.md](../plan.md), [tasks.md](../tasks.md), [data-model.md](../data-model.md), [quickstart.md](../quickstart.md), [contracts/balanceo.md](../contracts/balanceo.md), `.specify/memory/constitution.md` (v1.0.1)
**Nota**: informe guardado **antes de corregir nada**, según `revisión-final-bicicletas.txt` (paso 3). Este informe es de solo lectura; no se modificaron artefactos para generarlo.

## Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| H-001 | Inconsistency | HIGH | spec.md FR-005 (clarify Q3) ↔ data-model.md L83-86, handler.go L29-32, tasks.md T012, quickstart.md L104 | El clarify (2026-11-20) fijó el desempate del panel por **ID de estación ascendente**, pero plan/data-model/tasks/quickstart/código usan y validan **`fecha_generacion` ASC** como desempate («empate de score → fecha_generacion ASC ✅») | Decidir un solo criterio: (a) aplicar station ID en `ORDER BY` + data-model + quickstart + T7d, o (b) enmendar FR-005 y el registro de clarify a `fecha_generacion` ASC (criterio ya implementado y validado). No dejar coexistir ambos |
| H-002 | Coverage | MEDIUM | tasks.md L81 (T020), checklists/requirements.md | T020 enumera «barrido de FR-001…FR-008»: la spec ahora tiene **FR-009** (acceso operario) y el barrido de tareas quedó desactualizado tras el clarify | Ampliar la enumeración a FR-001…FR-009 en T020 y revisar el checklist contra la spec post-clarify |
| H-003 | Inconsistency | MEDIUM | plan.md L33 («Unknowns: 0 NEEDS CLARIFICATION», Input) | El plan no referencia la sesión de clarify 2026-11-20 ni sus 4 decisiones (Q1 destino, Q2 acceso, Q3 empate, Q4 estado vacío); Q1/Q2/Q4 ya coinciden con el plan (V1/V7/T013/T011/T008), pero Q3 no (H-001) | Añadir nota de clarify en plan.md y resolver Q3 según H-001; re-verificar Constitution Check post-cambio |
| H-004 | Terminology | MEDIUM | spec.md US1/Assumptions ↔ data-model.md V1, tasks.md T013, seed `logistica@si16.test` | La spec habla de «operario de transporte y calle»/«operario de calle»; el código y el modelo usan el rol **`operario_logistica`**. Sin mapa explícito, un tester no sabe qué rol probar | Normalizar en spec: término canónico «operario de logística (rol `operario_logistica`)`» con nota del sinónimo, en US1 y Assumptions |
| H-005 | Ambiguity | LOW | spec.md FR-009 | «Vista de error» no define forma: el contrato ya fija `403 ROL_NO_PERMITIDO` (sin token → `401`) en contracts/balanceo.md §autorización y data-model V1 | Referenciar el contrato en FR-009 («denegado según contracts/balanceo.md») sin meter detalle de implementación |
| H-006 | Pending | INFO | tasks.md L89 (T021), data-model.md §PENDIENTE 1, spec.md Key Entities | Tabla de auditoría `traslados_balanceo`: pendiente de acuerdo del equipo (const. III); MVP no depende de ella | **Mantener [BLOCKED]** — no tocar hasta acuerdo del grupo |
| H-007 | Pending | INFO | tasks.md L90 (T022), data-model.md §PENDIENTE 2 | Score de urgencia materializado (columna/índice): pendiente de acuerdo del equipo (const. III) | **Mantener [BLOCKED]** — no tocar hasta acuerdo del grupo |

Sin hallazgos de duplicación; sin placeholders TODO/TKTK en los tres artefactos.

## Coverage Summary Table

| Requirement Key | Has Task? | Task IDs | Notes |
|---|---|---|---|
| FR-001 ocupación % | ✔ | T005, T009 | R9 en `umbrales.go` |
| FR-002 alerta <10 %>90 % | ✔ | T005, T007 (T7a), T009 | Umbrales estrictos |
| FR-003 única alerta abierta | ✔ | T007 (T7b), T009 | `ux_alerta_abierta` |
| FR-004 resolución automática | ✔ | T007 (T7b), T009 | SC-004 |
| FR-005 panel ordenado + empate | ✔ (conflicto) | T007 (T7d), T012 | **H-001**: desempate spec ≠ código |
| FR-006 confirmar traslado + validaciones | ✔ | T006, T011, T007 (T7g) | Q1 (destino) ya cubierta: `≤ anclajes libres en destino` |
| FR-007 recálculo inmediato ≤30 s | ✔ | T011, T016, T019 | Redis post-commit |
| FR-008 averiadas excluidas | ✔ | T005, T007 (T7c), T009 | Consistente HU-05 |
| FR-009 acceso solo operario | ✔ | T013, T007 (T7e), T016 | Q2 ya alineada: `RequireRole("operario_logistica")` → 403; 401 sin token |
| SC-001 alerta <60 s | ✔ | T010, T007 (T7h), T019 | Worker 55 s + barrido <5 s |
| SC-002 panel íntegro | ✔ | T007 (T7d/T7b), T008, T016 | Incluye Q4 (estado vacío: T008/T016) |
| SC-003 vista inmediata + ≤30 s | ✔ | T011, T019 | |
| SC-004 0 alertas obsoletas | ✔ | T007 (T7b), T019 | |

**Acceptance scenario 4 (estado vacío, Q4)**: cubierto por T008/T016 («Sin estaciones descompensadas»).

## Constitution Alignment Issues

Ninguna. Constitution Check PASS (I–V) según plan.md; re-chequeo post-Phase 1 PASS.

- **Const. I** (reglas server-side): umbrales, unicidad y validaciones del traslado en `services/balanceo` + tests de rechazo (T005–T007, T018) — PASS.
- **Const. III** (esquema compartido): sin migraciones; T021/T022 correctamente bloqueados — PASS (y así deben permanecer).
- **Const. IV** (≤30 s): Redis post-commit — PASS.
- **Const. V** (independencia): paquetes propios + puntos comunes de una línea — PASS.
- **Gate de calidad**: pruebas automatizadas de reglas críticas presentes (T7a/T7b/T7g) — PASS.

## Unmapped Tasks

Ninguna: T001–T020 mapean a FRs/SCs o a setup/polish transversal; T021/T022 mapean a las propuestas «PENDIENTE DE ACUERDO DEL EQUIPO» de data-model.md y spec.md Key Entities (bloqueadas por diseño, const. III).

## Metrics

- **Total Requirements**: 13 (FR-001…FR-009 = 9; SC-001…SC-004 = 4)
- **Total Tasks**: 22 (T001–T020 marcadas [X]; T021–T022 [BLOCKED])
- **Coverage %**: 100 % (13/13 con ≥1 tarea)
- **Ambiguity Count**: 1 (H-005)
- **Duplication Count**: 0
- **Critical Issues Count**: 0 (sin violaciones de constitución; 1 HIGH: H-001)

## Next Actions

- Sin CRITICAL: no bloquea por constitución, pero **H-001 (HIGH) debe resolverse antes de repetir `/speckit-analyze`** (paso 3 de la revisión: corregir y generar `analyze-final.md`).
- Correcciones sugeridas (por orden): resolver H-001 → actualizar T020/plan (H-002, H-003) → normalizar terminología (H-004) → referenciar contrato en FR-009 (H-005).
- **No** avanzar a `/speckit.plan` (decisión del usuario). T021/T022 siguen pendientes de acuerdo del equipo.
