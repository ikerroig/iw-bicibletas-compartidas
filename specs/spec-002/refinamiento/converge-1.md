# Converge 1 — spec-002 (HU-02)

**Responsable**: Iker Roig · **Revisor**: Alex Morell · **Fecha**: 2026-11-19
**Comando**: `/speckit-converge` · **Spec Kit**: 1.1.3
**Versión analizada**: commit `b69dc49` (Phase 6 implementada: T043–T060)
**Enunciado**: §5.3. El converge va **después** de implement, para comprobar
que el código cubre la spec, el plan y las tareas; no para planificar desde cero.

## Resultado: `tasks_appended` (3 tareas, Phase 7: Convergence)

| ID | Tipo | Gravedad | Origen | Evidencia | Tarea |
|----|------|----------|--------|-----------|-------|
| F1 | partial | ALTA | FR-011, D17 | `cargarViajeEnCurso()` (`frontend/src/stores/viaje.ts`) no tiene plazo. Si la consulta de comprobación se queda colgada (no falla, simplemente no responde), la página se queda en «comprobando…» sin botón. FR-011 exige el botón «si la consulta tampoco responde». Los tests de T049 solo cubren una consulta que **falla**, no una que **no responde**; las pruebas manuales tampoco (V12a usa `TypeError`) | T061 |
| F2 | unrequested | BAJA | plan D16 | `viajeActivo()` en `frontend/src/services/api/alquiler.ts` ya no tiene usos (`grep` en `src/` y `tests/`) | T062 |
| F3 | unrequested | BAJA | research D17 | `resultadoDesconocido()` devuelve `true` para cualquier error que no sea `ApiError` (por ejemplo, un cuerpo 201 que no se puede leer). Es seguro porque solo provoca una consulta, pero D17 solo enumera tres casos | T063 |

## Qué se comprobó

| Ámbito | Elementos | Resultado |
|--------|-----------|-----------|
| Requisitos funcionales | FR-001…FR-012 | FR-011 parcial (F1); el resto se cumple |
| Criterios de éxito | SC-001…SC-004 | Se cumplen (tests y `verificacion-final.md`) |
| Escenarios de aceptación | US1/AC1…AC7 | Se cumplen |
| Casos límite | 9 | Se cumplen; el de «respuesta que no llega» depende de F1 en el caso de que la comprobación tampoco responda |
| Decisiones del plan | D1–D17 | D16 deja código sin usar (F2); D17 más amplio en el código que en el documento (F3) |
| Constitución | I–V | Sin incumplimientos: las reglas y los textos se deciden en el servidor; sin cambios de esquema ni de Redis; guardia en la página propia |

Ningún hallazgo es crítico ni contradice la spec. F1 es el único que afecta a
un requisito: lo encontró el converge, no las pruebas, porque los tests de
T049 simulaban una consulta que falla y no una que se queda sin respuesta.
