# Informe de analyze final — spec-002 (HU-02)

**Responsable**: Iker Roig · **Revisor**: Alex Morell
**Comando**: `/speckit-analyze` (segunda pasada) · **Spec Kit**: 1.1.3 · **Fecha**: 2026-11-19
**Versión analizada**: commit `50f0c71` (correcciones del [registro de hallazgos](./registro-hallazgos.md)), más la corrección F1 de este informe
**Enunciado**: §5.2, paso 5: «Repetir analyze después de los ajustes».

## Estado de los hallazgos del informe inicial

| ID | Gravedad inicial | Comprobación en la segunda pasada | Estado |
|----|------------------|-----------------------------------|--------|
| I2 | ALTA (MEDIA tras contraste) | `quickstart.md` explica cómo crear `eva` y `fran` y los usa en V2–V5 y V9; Ana solo aparece como usuaria con viaje activo | Cerrado |
| I1 | MEDIA | Escenarios de aceptación 1–7 en orden | Cerrado |
| U1 | MEDIA | Entidad Usuario con «activo/inactivo y fecha de vencimiento» | Cerrado |
| U2 | MEDIA | D2 (research), plan (deltas), contrato (422) y tareas T059/T060 coinciden en los dos mensajes | Cerrado en documentos; código pendiente (T059, T060) |
| I3 | MEDIA | No queda «bicicleta fuera de servicio» en la spec | Cerrado |
| I4 | MEDIA | Entidad Anclaje con «ocupado, libre o fuera de servicio» | Cerrado |
| I5 | MEDIA | Nota de la ampliación a 5xx en la spec (Clarifications) y en `clarify.md` | Cerrado |
| A1 | BAJA | «error 5xx del servidor» en FR-011, escenario 5 y caso límite; ningún «error interno» en la spec | Cerrado |
| I6 | BAJA | Aceptado con justificación en el registro (D9 aplica el criterio del menú) | Aceptado |
| D1 | BAJA | El caso límite antiguo remite a FR-011 | Cerrado |
| C1 | BAJA | «Incremental Delivery» con el paso 6; dependencias y checkpoint con T059/T060 | Cerrado |
| C2 | BAJA | Árbol del plan con los tests nuevos | Cerrado |
| K1 | BAJA | T043 pide avisar al equipo antes de empezar | Cerrado en documentos; el aviso se hará al empezar T043 |

## Hallazgos nuevos de la segunda pasada

| ID | Categoría | Gravedad | Ubicación | Resumen | Decisión |
|----|-----------|----------|-----------|---------|----------|
| F1 | Cobertura | BAJA | `quickstart.md` V7; `tasks.md` T056 | Introducido al corregir U2: V7 seguía esperando solo «422, nunca 500», sin los dos mensajes nuevos, y T056 no validaba V7 | **Corregido en esta pasada**: V7 incluye el código vacío y los dos mensajes; T056 valida V7 y toma P2.3 como evidencia «antes» |
| F2 | Inconsistencia | — | `research.md` D11 | D11 dice que un plazo agotado «responde error interno». Se comprobó que no contradice FR-011: es justo uno de los casos de resultado desconocido que el cliente resuelve consultando el viaje activo (D17) | **Sin cambios** (verificado, no es un hallazgo) |

El checkpoint de la fase 3 de `tasks.md` sigue diciendo «FR-001…FR-009». Es el
checkpoint histórico de esa fase, no el actual (el de la Phase 6 cubre
FR-001…FR-012), así que se deja como está.

## Cobertura

| Requisito | Tareas | Prueba prevista |
|-----------|--------|-----------------|
| FR-001 | T011, T017, T059, T060 | Go + V2, V7 |
| FR-002, SC-001 | T001, T010, T016, T044, T045, T051 | Go (unitario e integración) + V1 |
| FR-003, SC-002 | T002 | Go + V4 |
| FR-004, SC-004 | T002, T012, T018 | Go + V5 |
| FR-005 | T004 | Go + V2, V8 |
| FR-006 | T004, T007, T041 | Vitest + V3 |
| FR-007 | T002, T036, T037, T039, T040 | Go |
| FR-008 | T014, T015, T019, T021, T022, T038, T041, T042 | Go + Vitest + V2, V9 |
| FR-009 | T028, T029, T050, T055 | Go + Vitest + V6, V13 |
| FR-010 | T047, T048, T052, T053 | Go (contrato) + Vitest + V10, V11 |
| FR-011 | T049, T054 | Vitest + V12 |
| FR-012 | T046 | Go + V14 |
| SC-003 | T013, T030, T035 | Go |

Detalle por criterio, con nombres de test: `docs/matriz-trazabilidad.md`,
sección HU-02.

**Problemas con la constitución**: ninguno. Principio I: las reglas y los
textos de causa se deciden en el servidor (D15, D16), el aviso previo es
informativo y la guardia de ruta (D9) no sustituye al 403. Principio III: sin
migraciones ni claves Redis; los cambios compartidos (OpenAPI, `httperr`,
`frontend/tests/handlers.ts`) son aditivos y se comunican antes (T043).
Principio V: la guardia vive en la página propia, no en `routes.tsx`.

## Métricas

| Métrica | Inicial | Final |
|---------|---------|-------|
| Requisitos (FR + SC) | 16 | 16 |
| Tareas | 58 | 60 (+ T059, T060) |
| Cobertura (requisitos con ≥ 1 tarea y prueba prevista) | 100 % | 100 % |
| Hallazgos CRÍTICOS | 0 | 0 |
| Hallazgos ALTOS | 1 | 0 |
| Hallazgos abiertos | 13 | 0 (12 cerrados, 1 aceptado; F1 corregido en esta pasada) |

## Conclusión

Se cumplen las condiciones de finalización del §5.2: no quedan contradicciones
críticas ni de gravedad alta, y todos los criterios del alcance tienen tareas y
una prueba prevista. Lo único pendiente es código (Phase 6: T043–T060), que se
aborda en el §5.3 con `/speckit-implement` y `/speckit-converge`, y la revisión
de Alex Morell en [registro-hallazgos.md](./registro-hallazgos.md).
