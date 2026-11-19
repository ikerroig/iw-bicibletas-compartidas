# Informe de analyze inicial — spec-002 (HU-02)

**Responsable**: Iker Roig · **Revisor**: Alex Morell
**Comando**: `/speckit-analyze` · **Spec Kit**: 1.1.3 · **Fecha**: 2026-11-19
**Versión analizada**: commit `1890a4a` (spec, plan y tareas ya actualizados tras el clarify, enunciado §5.2 pasos 1 y 2)
**Instrucción**: «Revisa la coherencia entre spec.md, plan.md y tasks.md tras el clarify de refinamiento del 2026-11-19. Presta atención a criterios sin tareas, tareas sin requisito que las justifique, reglas contradictorias y dependencias entre historias. Informa de la gravedad y de los documentos afectados.»

> Informe de solo lectura, conservado **sin corregir** (§5.2, paso 2). Cada
> hallazgo se contrasta con el proyecto y se decide en
> [registro-hallazgos.md](./registro-hallazgos.md).

## Hallazgos

| ID | Categoría | Gravedad | Ubicación | Resumen | Recomendación |
|----|-----------|----------|-----------|---------|---------------|
| I1 | Inconsistencia | MEDIA | spec.md L52-77 | Los escenarios de aceptación están desordenados (1, 2, 4, 5, 6, 7, 3): el 3 quedó al final al insertar los nuevos | Reordenar 1–7 sin cambiar el contenido, para que la matriz y los tests citen números estables |
| I2 | Inconsistencia | ALTA | quickstart.md V2–V4, V9; seed | V2 pide que Ana desbloquee, pero en el seed Ana ya tiene un viaje activo (P3.4 de las pruebas previas: 409). V2–V9 no se pueden ejecutar tal como están escritos, y T057 manda repetirlos | Indicar en el quickstart un usuario con abono vigente y sin viaje, o cómo crearlo |
| U1 | Infraespecificación | MEDIA | spec.md L154-155 (Key Entities) | La entidad Usuario solo menciona el abono «activo/inactivo», pero FR-002 y SC-001 usan ahora la **fecha de vencimiento** para el mensaje | Añadir la fecha de vencimiento a la entidad Usuario |
| U2 | Infraespecificación | MEDIA | spec.md FR-001; data-model.md; contracts | El límite de 1–64 caracteres, la prohibición de caracteres de control y el 422 solo están en el plan y el contrato; la spec no dice qué es un código mal formado ni qué mensaje recibe (hueco 4 de las pruebas previas, P2.3) | Decidir si es una validación técnica (se queda en el plan) o un criterio de la spec |
| I3 | Inconsistencia | MEDIA | spec.md L84-85 (Edge Cases); Key Entities L158 | El caso límite habla de bicicleta «fuera de servicio», un estado que no existe: los estados son disponible, en uso, averiada y en taller | Sustituir por «averiada o en taller» |
| I4 | Inconsistencia | MEDIA | spec.md L156-157 (Key Entities); FR-012 | La entidad Anclaje solo tiene «ocupado/libre», pero FR-004 y FR-012 usan «anclaje en servicio» y el esquema tiene `fuera_de_servicio` | Añadir el estado «fuera de servicio» a la entidad Anclaje |
| I5 | Inconsistencia | MEDIA | spec.md Clarifications 2026-11-19 (Q2) vs. FR-011 y escenario 5 | La respuesta del clarify habla de «la respuesta no llega», pero FR-011 se amplió después a los errores 5xx al actualizar el plan, sin dejar constancia en el clarify | Registrar la ampliación y su motivo (D12) en el clarify, o volver a preguntarla |
| A1 | Ambigüedad | BAJA | spec.md FR-011, L95, escenario 5 | El mismo caso se llama «error interno del servidor» (FR-011, caso límite) y «error 5xx» (escenario 5); D17 usa `status >= 500`, que incluye 502–504 de un proxy | Usar un solo término: «error 5xx del servidor» |
| I6 | Inconsistencia | BAJA | spec.md FR-009 vs. research D9 | D9 también redirige cuando hay token pero no usuario guardado; FR-009 solo habla de «los demás roles» | Aceptable (mismo criterio que el menú, T029), pero conviene citarlo en FR-009 o en D9 como decisión técnica |
| D1 | Duplicación | BAJA | spec.md L92-96 (Edge Cases) | Dos casos límite sobre la pérdida de conexión: el antiguo (usuario con viaje activo) y el nuevo (FR-011) se solapan en parte | Mantener ambos, pero que el antiguo remita a FR-011 |
| C1 | Cobertura | BAJA | tasks.md Phase 3 (Checkpoint), Implementation Strategy | El checkpoint de la fase 3 y la estrategia de entrega no mencionan la Phase 6 | Añadir la Phase 6 a «Incremental Delivery» |
| C2 | Cobertura | BAJA | plan.md Project Structure | El árbol del plan no recoge los ficheros de test nuevos de T044, T046, T049 y T050 (`causa_test.go`, `spec002_orden_test.go`, `resultadoDesconocido.test.tsx`, `guardia.test.tsx`) | Actualizar el árbol |
| K1 | Constitución | BAJA | tasks.md T058; constitución III | Los cambios compartidos (constante `SinAbono`, OpenAPI y `frontend/tests/handlers.ts`) se comunican al final, en la PR. No son cambios de esquema ni de Redis, así que el principio III no exige acuerdo previo, pero el acuerdo de colaboración pide comunicarlos | Avisar al equipo al empezar T043, no solo en la PR |

## Cobertura

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|-----------|---------------|--------|-------|
| FR-001 | Sí | T011, T017 | |
| FR-002 (ampliado) | Sí | T001, T010, T016, T044, T045, T051 | |
| FR-003 | Sí | T002 | |
| FR-004 | Sí | T002, T012, T018 | |
| FR-005 | Sí | T004 | |
| FR-006 | Sí | T004, T007 | |
| FR-007 | Sí | T002, T036, T037, T039, T040 | |
| FR-008 | Sí | T014, T015, T019, T021, T022, T038, T041, T042 | |
| FR-009 (ampliado) | Sí | T028, T029, T050, T055 | |
| FR-010 | Sí | T047, T048, T052, T053 | Validación manual V10, V11 (T056) |
| FR-011 | Sí | T049, T054 | Validación manual V12 (T056) |
| FR-012 | Sí | T046 | Test de caracterización; validación V14 |
| SC-001 (ampliado) | Sí | T044, T045 | |
| SC-002 | Sí | T002, T025 | |
| SC-003 | Sí | T013, T030, T035 | |
| SC-004 | Sí | T002 | |

**Tareas sin requisito**: ninguna. T043 y T058 son de contrato y comunicación; T056 y T057, de validación y regresión.

**Problemas con la constitución**: ninguno CRÍTICO. Principio I: las causas del abono y el aviso previo se calculan en el servidor (D15, D16) y el `POST` sigue rechazando. Principio III: no hay migraciones ni claves Redis nuevas. Ver K1.

**Dependencias entre historias**: spec-003 usa `fijarViajeEnCurso` y `viajeEnCurso`; T053 mantiene sus firmas. Ninguna otra spec consume `GET /alquileres/activo` ni `ABONO_VENCIDO`.

## Métricas

| Métrica | Valor |
|---------|-------|
| Requisitos (FR + SC) | 16 (12 FR + 4 SC) |
| Tareas | 58 (T001–T058) |
| Cobertura (requisitos con ≥ 1 tarea) | 100 % |
| Ambigüedades | 1 (A1) |
| Duplicaciones | 1 (D1) |
| Hallazgos CRÍTICOS | 0 |
| Hallazgos ALTOS | 1 (I2) |
| Total de hallazgos | 13 |

## Siguientes pasos

No hay hallazgos críticos. El ALTO (I2) afecta a la validación manual, no al
código, pero hay que resolverlo antes de T056 y T057. Los demás se pueden
corregir en la spec, el plan y las tareas antes de `/speckit-implement`.
