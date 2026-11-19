# Informe de analyze — spec-001 (HU-01: Consulta de disponibilidad en tiempo real)

**Responsable**: Iker Roig
**Fecha**: 2026-11-19
**Comando**: `/speckit-analyze` (solo lectura)
**Artefactos revisados**: `spec.md` (tras el clarify del 2026-11-19), `plan.md`,
`tasks.md`, `contracts/openapi.yaml`, `quickstart.md` y la constitución SI-16
v1.0.1. Además se ha mirado el código del frontend, solo para saber qué
requisitos nuevos ya están cubiertos en la práctica.

## Contexto

Este analyze se ejecuta **después** del clarify de refinamiento, que añadió el
FR-010, el FR-011, el FR-012 y el SC-008 y amplió el FR-004 cuando la HU-01 ya
estaba implementada. Era de esperar que el análisis encontrara requisitos sin
tareas; precisamente para eso está el paso siguiente, `/speckit-converge`.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| G1 | Cobertura | ALTA | spec.md FR-011, SC-008; tasks.md fase 3 | El listado de texto de estaciones no tiene ninguna tarea ni está implementado: la única forma de consultar es el mapa | Añadir tareas para un componente de listado en `pages/mapa/` (reutiliza el store `red.ts` y `estadoVisual`, sin cambiar el contrato) y su test |
| G2 | Cobertura | MEDIA | spec.md FR-010; `MapaRed.tsx`; T029/T031 | La primera carga fallida se gestiona solo en parte: se muestra «{error}. Reintentando…» con reintento automático, pero no el texto «No se pudo cargar la disponibilidad» ni el botón «Reintentar». Ninguna tarea lo pide | Añadir una tarea de UI y un test (MSW con error sin estado previo → mensaje, botón y sin marcadores) |
| G3 | Cobertura | MEDIA | spec.md FR-012; T014, T021, `services/api.ts` | El backend ya exige sesión (`security` global en OpenAPI, `RequireAuth`) y el frontend cierra sesión con un 401, pero **no devuelve al usuario al mapa** tras iniciar sesión. Tampoco hay un test de contrato que compruebe el 401 en `GET /estaciones` | Añadir un test de contrato del 401 en `/estaciones` y `/estaciones/{id}`, y una tarea de «volver a la ruta de origen» después del login |
| G4 | Cobertura | BAJA | spec.md FR-004 (ampliado); `MapaRed.tsx` | El detalle abierto ya se actualiza solo (`seleccionada` sale de `red.estaciones()`), pero ningún test lo garantiza | Añadir a T032 un caso que cambie los datos con el detalle abierto, incluido el paso a fuera de servicio |
| I1 | Inconsistencia | MEDIA | tasks.md fase 3 («Cubre FR-001…FR-009 y SC-001…SC-007») | El objetivo de la fase ha quedado desfasado tras el clarify | Actualizarlo a FR-001…FR-012 y SC-001…SC-008 al añadir las tareas nuevas |
| I2 | Inconsistencia | MEDIA | plan.md (Resumen y estructura de `pages/mapa/`) | El plan solo recoge la actualización del 2026-11-03; no menciona el listado, el error de primera carga ni el retorno tras el login | Añadir una nota de actualización del 2026-11-19 (sin cambios de contrato, esquema ni Redis) y el componente de listado en el árbol |
| I3 | Inconsistencia | BAJA | quickstart.md S1–S1e | No hay escenarios de validación para el FR-010, el FR-011 ni el FR-012 | Añadir los escenarios S1f (primera carga fallida), S1g (listado con las teselas bloqueadas) y S1h (sesión caducada → login → mapa) |
| I4 | Inconsistencia | BAJA | tasks.md fases 4–11 | El `tasks.md` de spec-001 es global y deja sin marcar tareas que ya están implementadas en las specs 002–008 (por ejemplo T047, T062 y T080–T087) | Indicar en las notas que el estado de las fases 4–10 se sigue en el `tasks.md` de cada spec, o marcarlas |
| A1 | Ambigüedad | BAJA | spec.md FR-011 («accesible en todo momento (p. ej., pestaña «Lista»)») | No se fija la forma de acceso (pestaña, conmutador o sección) | Aceptable: se decide en el plan o las tareas; no afecta a las pruebas mientras sea accesible desde la vista del mapa |

No hay hallazgos de duplicación, y tampoco de ambigüedad en los criterios
cuantitativos: SC-002, SC-004 y SC-005 tienen cifras y condiciones de medida.

## Cobertura de requisitos

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|-----------|---------------|--------|-------|
| FR-001 | Sí | T027, T030, T031, T032 | |
| FR-002 | Sí | T020, T027, T031 | |
| FR-003 | Sí | T025, T026, T029, T099 | Peor caso ≈ 28 s (R16) |
| FR-004 | Sí (parcial) | T027, T031 | Falta un test del detalle que se actualiza solo (G4) |
| FR-005 | Sí | T011, T025 | |
| FR-006 | Sí | T020, T025, T031 | |
| FR-007 | Sí | T029, T097, T098 | |
| FR-008 | Sí | T030, T031, T095, T096 | |
| FR-009 | Sí | T030, T031, T095, T096 | |
| FR-010 | Parcial | T029, T031 | Faltan el texto y el botón exigidos (G2) |
| FR-011 | **No** | — | G1 |
| FR-012 | Parcial | T014, T017, T021 | Faltan el retorno al mapa y el test del 401 (G3) |
| SC-001 | Sí | T032 | |
| SC-002 | Sí | T025, T099 | |
| SC-003 | Sí | T025 | |
| SC-004 | Sí | T100 | Validación manual (S1e) |
| SC-005 | Sí | T098 | |
| SC-006 | Sí | T096, T032 | |
| SC-007 | Sí | T025, T032 | |
| SC-008 | **No** | — | G1 |

## Constitución

**No hay conflictos.** El FR-012 refuerza el principio I, porque la autorización
se valida en el servidor. El FR-011 reutiliza `GET /estaciones` y el store en
vivo, así que respeta el principio IV (frescura ≤ 30 s) y no toca el esquema
ni Redis (principio III). Ningún cambio añade tecnologías (principio II).

## Tareas sin requisito asociado

Ninguna en la fase 3. Las fases 1–2 son base compartida y quedan justificadas
por el plan.

## Métricas

- Requisitos en total: 20 (12 FR + 8 SC)
- Tareas de la fase 3 (US1): 15 (T024–T032 y T095–T100)
- Cobertura: 18/20 = **90 %** (requisitos con al menos una tarea; 3 de ellos solo en parte)
- Ambigüedades: 1
- Duplicaciones: 0
- Problemas críticos: 0

## Próximos pasos

- No hay problemas críticos. El hallazgo G1 (ALTA) es trabajo pendiente nuevo,
  no una contradicción entre artefactos.
- Ejecutar `/speckit-converge` en la spec-001 para añadir a `tasks.md` las
  tareas de G1–G4 y corregir I1.
- Actualizar a mano `plan.md` (I2) y `quickstart.md` (I3) con la nota del
  2026-11-19.
- Después, `/speckit-implement` sobre las tareas nuevas.
