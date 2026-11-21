# Hallazgos — spec-008 / HU-08

**Fecha**: 2026-11-20 · **Responsable**: Alberto Copa
**Origen**: análisis inicial de [spec.md](../spec.md), [plan.md](../plan.md) y [tasks.md](../tasks.md), con decisión final del equipo sobre el alcance por estación.

| ID | Criterio | Gravedad | Evidencia | Decisión | Responsable | Estado |
|---|---|---|---|---|---|---|
| I1 | FR-002 / US1 | HIGH | La propuesta temporal de ranking no aparecía en la descripción de la historia ni en Independent Test. | Retirar el ranking; mantener los gráficos hora/día que especifica `historia.md`. | Equipo (registro: Alberto Copa) | Cerrado |
| I2 | FR-002 / plan | HIGH | El plan describe dos gráficos y no define métricas por estación. | No ampliar el plan; la decisión final conserva el alcance que ya implementa. | Equipo (registro: Alberto Copa) | Cerrado |
| C1 | Contrato compartido `AnaliticaViajes` | HIGH | El OpenAPI y el struct no incluyen campo de ranking; el equipo no aprobó ampliar el contrato común. | No modificar el OpenAPI global ni el modelo compartido. | Equipo | Cerrado |
| C2 | T004/T007–T009/T014 | HIGH | No había implementación ni pruebas de ranking por estación. | No añadir tareas: el equipo descartó esa ampliación provisional. | Equipo (registro: Alberto Copa) | Cerrado |
| U1 | FR-002 — estaciones con cero salidas | MEDIUM | La composición del ranking no estaba determinada. | Deja de aplicar al retirar el ranking. | Equipo (registro: Alberto Copa) | Cerrado |
| I3 | FR-006 / T8a | MEDIUM | La regla de consistencia solo relaciona gráficos y CSV. | Sin ranking, se conserva la regla original de FR-006 y T8a. | Equipo (registro: Alberto Copa) | Cerrado |

## Resumen

- Abiertos: 0 · Cerrados: 6 · Bloqueados: 0.
- No se modificó el contrato OpenAPI global ni se añadieron tareas de código.
- La especificación vuelve a coincidir con `historia.md`, el plan, el modelo de datos y la implementación vigente.

## Revisión cruzada de la PR (2026-11-21)

Se revisaron `spec.md`, `plan.md`, `tasks.md`, los documentos SSD de esta carpeta, el contrato, la implementación backend/frontend y la trazabilidad de cada criterio.

| Área revisada | Evidencia | Resultado |
|---|---|---|
| FR-001/FR-002/SC-001 | Validación del rango, agregados de 24 horas y 7 días, pruebas T8b/T8c/T8i y respuesta real de un rango anual | Cumple |
| FR-003/FR-004/SC-002/SC-004 | Filtro compartido, CSV con BOM/CRLF, cabecera y costes consistentes; export real con nombres Unicode | Cumple |
| FR-005/SC-003 | Middleware de autenticación y rol gestor; pruebas reales `401`, `403` y `200` | Cumple |
| FR-006/FR-007 | Comparación de agregados y CSV; consulta limitada a viajes finalizados y sin filtro por usuario | Cumple |
| Frontend | 16 archivos y 97 pruebas, build TypeScript/Vite correcto y ruta `/analitica` registrada | Cumple |
| Metodología SSD | `clarify`, `analyze-inicial`, `analyze-final`, `converge-1`, `tasks`, `quickstart` y `trazabilidad` coherentes; no se amplían tareas fuera del alcance | Cumple |

### Resultado

- Hallazgos nuevos: 0.
- Tareas nuevas: 0.
- La implementación satisface los criterios funcionales y de aceptación de la spec-008.
- La única incidencia observada pertenece al entorno de esta revisión: Vite no pudo arrancar por permisos de `esbuild` al resolver la ruta local, aunque las pruebas y el build del frontend pasan.
