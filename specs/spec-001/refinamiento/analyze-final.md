# Informe final de analyze — spec-001 (HU-01)

**Responsable**: Iker Roig
**Fecha**: 2026-11-19
**Comando**: `/speckit-analyze` (segunda ejecución, enunciado §5.2 paso 5:
«Repetir analyze después de los ajustes»)
**Informe inicial**: [analyze.md](./analyze.md) · **Registro de hallazgos**:
[registro-hallazgos.md](./registro-hallazgos.md)
**Artefactos revisados**: `spec.md` (con FR-013, añadido después del informe
inicial), `plan.md`, `research.md` (R20–R23), `tasks.md` (T101–T110),
`quickstart.md` (S1f–S1i), `contracts/openapi.yaml`, la constitución v1.0.1 y
los ficheros de test existentes de la HU-01.

## Cambios desde el informe inicial

- **Spec**: FR-013 (pregunta 5 del clarify, a raíz del hallazgo H1 de las
  pruebas) y escenarios de aceptación 4–8 en formato Dado/Cuando/Entonces.
- **Plan**: nota de actualización del 2026-11-19, `ListaEstaciones` en el
  árbol de carpetas, re-check de la constitución y decisiones R20–R23 en
  `research.md`.
- **Tareas**: T101–T110 añadidas; objetivo de la fase 3 actualizado (I1); nota
  sobre el estado de las fases 4–10 (I4).
- **Quickstart**: escenarios S1f–S1i.

## Hallazgos de esta ejecución

| ID | Categoría | Severidad | Ubicación | Resumen | Resolución |
|----|-----------|-----------|-----------|---------|------------|
| F1 | Inconsistencia | MEDIA | tasks.md T030 | La T030 (terminada) decía «si el usuario concede permiso, centra el mapa en su posición», lo que contradice el FR-013 | **Corregido en esta misma revisión**: nota en la T030 que remite a la T108, sin reescribir la tarea cerrada |
| F2 | Cobertura | BAJA | spec.md FR-013 | El FR-013 no tiene un criterio de éxito SC propio | **Aceptado**: es un requisito binario (encuadre `red` o `usuario`), ya verificable con el escenario de aceptación 8, la T107 y el S1i. Un SC porcentual no aportaría nada |

No quedan hallazgos de los anteriores (G1–G4, I1–I4, A1) sin tarea ni
decisión; su detalle está en el registro.

## Cobertura de requisitos

| Requisito | ¿Tiene tarea? | Tareas | Prueba prevista |
|-----------|---------------|--------|-----------------|
| FR-001 | Sí | T027, T030, T031 | Vitest SC-001, S1 |
| FR-002 | Sí | T020, T027, T031 | Go + Vitest, S1 |
| FR-003 | Sí | T025, T026, T029, T099 | Go (barrido, Redis caído), S1d |
| FR-004 | Sí | T027, T031, T105 | Vitest (detalle en vivo + paso a fuera de servicio) |
| FR-005 | Sí | T011, T025 | Go (TTL, fallback) |
| FR-006 | Sí | T020, T025, T031 | Go (averiadas) |
| FR-007 | Sí | T029, T097, T098 | Vitest (30/31 s), S1a |
| FR-008 | Sí | T030, T031, T095, T096 | Vitest, S1b |
| FR-009 | Sí | T030, T031, T095 | Go + Vitest, S1c |
| FR-010 | Sí | T101, T102 | Vitest (actualizado), S1f |
| FR-011 | Sí | T103, T104 | Vitest, S1g |
| FR-012 | Sí | T014, T021, T106 | Go `TestEstacionesExigeSesion` + Vitest, S1h |
| FR-013 | Sí | T107, T108 | Vitest (4 casos), S1i |
| SC-001 | Sí | T032 | Vitest |
| SC-002 | Sí | T025, T099 | Go |
| SC-003 | Sí | T025 | Go |
| SC-004 | Sí | T100 | S1e (manual) |
| SC-005 | Sí | T098 | Vitest |
| SC-006 | Sí | T096, T032 | Vitest |
| SC-007 | Sí | T025, T032 | Go + Vitest |
| SC-008 | Sí | T103, T104 | Vitest, S1g |

## Constitución

**Sin conflictos.** R20–R23 son solo de frontend y no cambian la API, el
esquema ni Redis (principio III). El FR-012 se valida en el servidor
(principio I). El listado hereda la frescura ≤ 30 s del store (principio IV).

## Métricas: inicial → final

| Métrica | Informe inicial | Informe final |
|---------|-----------------|---------------|
| Requisitos | 20 (12 FR + 8 SC) | 21 (13 FR + 8 SC) |
| Cobertura (requisitos con al menos una tarea) | 18/20 = 90 % (3 parciales) | **21/21 = 100 %** |
| Requisitos con prueba prevista | (no se midió) | **21/21** |
| Tareas de la fase 3 (US1) | 15 | 25 (+ T101–T110) |
| Hallazgos críticos | 0 | 0 |
| Hallazgos altos | 1 (G1) | 0 |
| Hallazgos medios / bajos abiertos | 7 / — | 0 (F1 corregido, F2 aceptado) |
| Ambigüedades | 1 (A1) | 0 (A1 aceptado y resuelto en R21) |

## Próximos pasos (enunciado §5.3)

1. `/speckit-implement` sobre T101–T108 (los tests primero).
2. Ejecutar `go test ./...` y `npm test`.
3. `/speckit-converge` para contrastar el código con spec, plan y tareas.
4. T109 (antes/después en el navegador) y T110 (regresión).
5. Que un compañero revise las filas de HU-01 en `docs/matriz-trazabilidad.md`.
