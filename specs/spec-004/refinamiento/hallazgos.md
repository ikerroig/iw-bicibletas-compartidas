# Registro de hallazgos — spec-004 (HU-04)

**Fecha**: 2026-11-19 · **Responsable**: Alex Morell · **Rama**: `spec-004`
**Fuente**: [analyze-inicial.md](./analyze-inicial.md) · **Destino**: Phase 5 de `tasks.md`

| ID | Ubicación | Gravedad | Hallazgo | Decisión | Responsable | Estado |
|----|-----------|----------|----------|----------|-------------|--------|
| A1 | spec AC4 vs `tests/historial_test.go` | ALTA | «Cualquier rol autenticado consulta su propio historial» sin prueba (fixture solo rol `usuario`; T4d solo aislamiento + 401) | **Válido.** Test **T016** (T4k): tokens `operario_taller` y `gestor` → `200` con historial propio; además se ignora un parámetro de usuario ajeno | Alex Morell | Tareas añadidas |
| A2 | quickstart T4j vs `TestHistorialT4jLatencia` | MEDIA | Latencia prometida «con un año completo de viajes» medida con fixture de 5 viajes | **Válido.** Test **T017**: ampliar la fixture de T4j a ~365 viajes finalizados antes de cronometrar (< 3 s) | Alex Morell | Tareas añadidas |
| A3 | quickstart § T4 | MEDIA | Filas T4a–T4j escritas antes del clarify 2026-11-19; AC4 sin fila y AC5/AC6 sin mapear | **Válido.** Doc **T018**: fila T4k (AC4 roles) + mapeo AC5→T4a, AC6→T4a/T4e en la tabla SC | Alex Morell | Tareas añadidas |
| A4 | `checklists/requirements.md` | BAJA | Sin tabla *Acceptance Coverage* (spec-003 sí la tiene); AC4–AC6 sin verificación rastreable | **Válido.** Doc **T019**: crear la tabla FR/SC/AC → test | Alex Morell | Tareas añadidas |
| A5 | plan.md Unknowns (línea 33) | BAJA | Solo menciona las 5 preguntas de 2026-10-28; no la sesión 2026-11-19 | **Válido.** Doc **T020**: actualizar la nota y remitir a `refinamiento/clarify.md` | Alex Morell | Tareas añadidas |
| A6 | clarify.md vs research.md | BAJA | Colisión de numeración D6–D8 entre ambos documentos (decisión ≠ research) | **Válido, menor.** Doc **T021**: renombrar decisiones de `clarify.md` a C1–C13 | Alex Morell | Tareas añadidas |

**Resumen**: 6 hallazgos (1 ALTA, 2 MEDIA, 3 BAJA) · 0 descartados ·
todos con tarea asignada en la Phase 5 de `tasks.md`.
