# Informe inicial de analyze — spec-004 (HU-04)

**Comando**: `/speckit.analyze` (§5.3 paso 2, primera pasada) · **Spec Kit**: 1.1.3
**Fecha**: 2026-11-19 · **Responsable**: Alex Morell
**Funcionalidad activa**: `specs/spec-004` · **Rama**: `spec-004`
**Estado de partida**: clarify 2026-11-19 hecho (`2f22e79`: D6–D8, AC4–AC6,
FR-001 orden, FR-005 JWT/401); implementación T001–T015 completa y en verde.

## Specification Analysis Report

| ID | Categoría | Gravedad | Ubicación | Hallazgo | Recomendación |
|----|-----------|----------|-----------|----------|---------------|
| A1 | Coverage gap | **ALTA** | spec AC4 vs `tests/historial_test.go` T4d | La parte «cualquier rol autenticado (usuario, operario, gestor) consulta su propio historial» no tiene prueba: `TestHistorialT4dAislamientoY401` solo verifica aislamiento entre usuarios y 401 sin token; la fixture (`prepararHistorial`) solo crea usuarios rol `usuario` | Test nuevo T4k: tokens `operario_taller` y `gestor` → `200` con su propio historial (y, de paso, parámetro ajeno ignorado) |
| A2 | Inconsistencia | MEDIA | quickstart T4j vs `TestHistorialT4jLatencia` | El quickstart promete latencia «con un año completo de viajes»; el test mide con la fixture estándar de 5 viajes | Ampliar la fixture del test de latencia a ~365 viajes finalizados antes de cronometrar |
| A3 | Coverage gap | MEDIA | quickstart § T4 (filas T4a–T4j) | Los escenarios se escribieron antes del clarify 2026-11-19: sin fila para AC4 (roles); AC5 y AC6 implícitas en T4a/T4e sin mapear | Añadir fila T4k (AC4) y mapear AC5→T4a, AC6→T4a+T4e en la tabla SC |
| A4 | Inconsistencia | BAJA | `checklists/requirements.md` | Sin tabla *Acceptance Coverage* (spec-003 sí la tiene desde su T027); AC4–AC6 sin verificación rastreable en el checklist | Crear la tabla con FR/SC/AC → test que lo demuestra |
| A5 | Inconsistencia | BAJA | plan.md (Unknowns, línea 33) | Solo menciona «las 5 preguntas de `/speckit.clarify`» (2026-10-28); no la sesión 2026-11-19 (D6–D8 del refine) | Añadir la nota de la segunda sesión y remitir a `refinamiento/clarify.md` |
| A6 | Terminología | BAJA | clarify.md vs research.md | Colisión de prefijo «D»: clarify **D6–D8** (permisos / viaje activo / vista inicial) ≠ research **D6–D8** (validación de entrada / offsets / …) usados por tasks T004 | Renombrar las decisiones de `clarify.md` a **C1–C13** (o explicitar el prefijo) |

**Hallazgos nuevos en esta pasada**: 6 (0 CRITICAL, 1 HIGH, 2 MEDIUM, 3 LOW).

## Coverage Summary Table

| Requisito | ¿Tarea? | Tareas | Nota |
|-----------|---------|--------|------|
| FR-001 (lista + orden DESC) | Sí | T005/T007 | T4a aserta orden estricto |
| FR-002 (filtro mes/año Europe/Madrid) | Sí | T005/T007 | T4b/T4g |
| FR-003 (sumatorio periodo/historial) | Sí | T005/T007 | T4b/T4e/T4f (incluye vista inicial sin filtro) |
| FR-004 (etiqueta tramo gratuito) | Sí | T004/T005 | T4c |
| FR-005 (solo propietario, JWT, 401, cualquier rol) | Sí | T005 | **Parcial**: roles sin probar (A1) |
| FR-006 (origen/destino, sin GPS) | Sí | T005 | T4a (nombres por JOIN) |
| FR-007 (coste/duración = HU-03) | Sí | T005 | campos del response |
| FR-008 (paginación 50/200, suma del periodo) | Sí | T004/T005 | T4e/T4f/T4i |
| SC-001…SC-003 | Sí | T005/T013 | quickstart § SC |
| SC-004 (< 3 s con un año) | Sí | T005 (T4j) | **Parcial**: fixture de 5 viajes (A2) |
| AC1–AC3, AC5, AC6 | Sí | T4a–T4e | mapeados en A3 |
| AC4 (roles + 401) | Parcial | T4d | **roles sin probar** (A1) |

**Constitution Alignment Issues**: ninguno — I (sumatorio/etiqueta/filtro en
backend), II (pila fijada), III (0 migraciones; `ix_viajes_usuario_fecha`
ya existe), IV (N/A: sin Redis), V (paquetes propios + 1 línea compartida).
Tasks T001–T015 todos trazan a un requisito; sin tareas huérfanas.

**Metrics**:

- Total requisitos: 12 (8 FR + 4 SC) + 6 AC · Total tareas: 15 (T001–T015)
- Cobertura: FR 8/8 (100 %), SC 4/4 (con A2), AC 5/6 plenos + 1 parcial
- Ambigüedades: 0 · Duplicaciones: 0 · Inconsistencias: 5 (A2–A6)
- Criterios críticos/altos sin resolver: 1 (A1)

## Next Actions

1. Registrar `hallazgos.md` con decisión/responsable/estado por hallazgo.
2. Correcciones: Phase 5 en `tasks.md` (T016 test AC4 roles, T017 año en
   T4j, T018 quickstart, T019 checklist, T020 plan, T021 numeración) +
   reaplicar `/speckit.analyze` → `analyze-final.md`.
3. `/speckit-implement` sobre las tuebas nuevas → `converge-1` → correcciones →
   `converge-2`.

*Primera pasada; este informe se conserva sin modificar. Sin hooks de
extensión (`.specify/extensions.yml` no existe).*
