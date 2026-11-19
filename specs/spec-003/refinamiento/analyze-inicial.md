# Informe inicial de analyze — spec-003 (HU-03)

**Comando**: `/speckit.analyze` (lectura; §5.2, paso 2) · **Spec Kit**: 1.1.3
**Fecha**: 2026-11-19 · **Responsable**: Alex Morell
**Funcionalidad activa**: `specs/spec-003` · **Rama**: `spec-003`
**Estado de partida**: clarify 2026-11-19 ya aplicado al `spec.md`
(commit `8c39452`); plan/tasks **todavía sin actualizar** — este informe se
conserva **antes de corregir nada** (§5.2, paso 2).
**Documentos analizados**: `spec.md`, `plan.md`, `tasks.md` + constitución;
referencia cruzada con el código y las pruebas existentes para contrastar
la cobertura (el analyze no lee los ficheros de test por sí mismo).

## Specification Analysis Report

| ID | Categoría | Gravedad | Ubicación | Resumen | Recomendación |
|----|-----------|----------|-----------|---------|---------------|
| A1 | Coverage gap | **ALTA** | `spec.md` FR-008 (AC4) · `tasks.md` · `quickstart.md` T3a–T3j | FR-008 exige 401 sin sesión y 403 para rol distinto de `usuario` en `POST /devoluciones`. La implementación cumple (`RequireAuth` + `RequireRole("usuario")` en `routes/spec_003_devolucion.go`) pero **ninguna tarea ni prueba lo verifica**: no hay test de 401/403 en `backend/tests/devolucion_test.go`, ni T3k en el quickstart | Añadir test de integración 401/403 + tarea en `tasks.md` + fila en el checklist + escenario T3k |
| A2 | Coverage gap | **ALTA** | `spec.md` FR-009 (AC5) · `tasks.md` | FR-009 (devolución permitida con abono caducado) **no tiene ninguna tarea que la mencione**. El código cumple por omisión (`services/devolucion/service.go` no consulta `abono_activo`), pero nada lo protege de una regresión futura (p. ej., si alguien añade un chequeo de abono al cierre) | Añadir test de integración con usuario de abono vencido que cierre viaje + tarea + fila checklist + escenario T3k/T3l |
| A3 | Inconsistencia / detalle de implementación | MEDIA | `spec.md` FR-004, SC-002, AC6 | El clarify introdujo el nombre de variable `TARIFA_MINUTO` en requisitos normativos de la spec; el checklist de calidad exige «No implementation details». El valor 0,20 € sí conviene dejarlo como ejemplo concreto para que SC-002 sea testeable | Reformular FR-004/SC-002/AC6 como «tarifa unitaria del servicio»; conservar `TARIFA_MINUTO` en Clarifications (registro de la decisión) y en plan/tasks |
| A4 | Documento desfasado | MEDIA | `checklists/requirements.md` (tabla Acceptance Coverage) | La tabla de cobertura llega hasta FR-007 y SC-004; no incluye las filas de FR-008, FR-009 ni los AC4–AC6 nuevos del clarify | Ampliar la tabla tras crear las pruebas de A1/A2 (dependencia explícita) |
| A5 | Documento desfasado | BAJA | `plan.md` § Technical Context (Unknowns) | Dice «0 NEEDS CLARIFICATION… se resuelven en `research.md`» y no menciona la sesión de clarify del 2026-11-19 (FR-008/FR-009/redondeo) | Actualizar la nota y el Summary para referenciar la sesión de clarify y sus FR |
| A6 | Inconsistencia menor | BAJA | `tasks.md` cabecera «Tests» | Referencia «T3a–T3i»; el quickstart define hasta **T3j** y T015 ya la usa | Corregir a T3a–T3j |
| A7 | Cobertura parcial | BAJA | `spec.md` AC6 vs `quickstart.md` T3d | AC6 fija 30:45 → 1 minuto; T3d prueba 30:01 → 0,20 €. Ambos ejercitan el mismo techo de `ceil`, pero ningún caso prueba un excedente de varios minutos con fracción (p. ej., 46:30 → 17 min) | Opcional: añadir un caso unitario con fracción > 1 min en T014 al ampliar A1/A2 |

**Coverage Summary Table** (requisitos con ≥1 tarea que los cubra):

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001 | Sí | T016–T020 | |
| FR-002 | Sí | T017 | |
| FR-003 | Sí | T014, T016 | T3a/T3c |
| FR-004 | Sí | T014, T016 | Redondeo cubierto por T3d (1801 s) |
| FR-005 | Sí | T017 | |
| FR-006 | Sí | T017 | T3f |
| FR-007 | Sí | T017 | T3e/T3g |
| FR-008 | **Parcial** | T020 (implementación) | **Sin tarea de prueba** → A1 |
| FR-009 | **No** | — | **Cero tareas**; el código cumple por omisión → A2 |
| SC-001…SC-004 | Sí | T014, T015, T024, T026 | SC-004/SC-003 con T3j/S3 |

**Constitution Alignment Issues**: ninguna violación. Principio I (reglas en
el backend + pruebas): el cálculo de tarifa y el cierre ya tienen pruebas
(T3a–T3i); A1/A2 son huecos de *verificación* sobre reglas ya implementadas
y se cierran con las tareas recomendadas. Principios II–V: sin impacto.

**Unmapped Tasks**: ninguna — T001–T027 trazan todas a US1/FR-001…FR-007.

**Dependencias entre historias**: FR-009 toca el campo `abono_activo` que
valida la HU-02 (spec-002), pero no exige cambios allí (el chequeo de abono
sigue siendo exclusivo del desbloqueo). FR-005 enlaza con HU-04
(spec-004, solo lectura del viaje cerrado). Sin acuerdos de equipo
pendientes (const. III intacta).

**Metrics**:

- Total requisitos evaluados: 13 (9 FR + 4 SC)
- Total tareas: 27 (T001–T027)
- Cobertura: 11/13 con tarea completa (84,6 %); 1 parcial (FR-008); 1 sin tarea (FR-009)
- Ambigüedades: 0 (el clarify cerró las 5 preguntas)
- Duplicaciones: 0
- Criterios críticos sin resolver: 0 · Altos: 2 (A1, A2)

## Next Actions

1. Corregir A1–A6 (añadir tareas de prueba para FR-008/FR-009, reformular
   `TARIFA_MINUTO` en la spec, actualizar plan/tasks/checklist) — §5.2 paso 4.
2. Repetir `/speckit.analyze` y conservar el informe final en
   `analyze-final.md`.
3. Después, `/speckit-converge` (§5.3) para contrastar el código con los
   criterios ya corregidos.

*Informe guardado antes de aplicar ninguna corrección, como exige §5.2.*
