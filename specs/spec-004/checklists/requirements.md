# Specification Quality Checklist: Consulta de historial de viajes y desglose mensual

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-20
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Items marked incomplete require spec updates before `/speckit.clarify` or `/speckit.plan`
- Validado 2026-10-20: aclaración de alcance aplicada — "ver mis rutas
  pasadas" = estaciones origen/destino por viaje, sin tracking GPS (FR-006).
- Validado 2026-11-19 (refinamiento fase 4): sin regresiones; los ítems
  siguen en verde tras el clarify 2026-11-19 (AC4–AC6, FR-001 orden,
  FR-005 JWT/401).

## Acceptance Coverage (post-refinamiento, T019 — 2026-11-19)

Cada criterio del spec con el test/verificación que lo demuestra.
Suites: unit `backend/internal/services/viajes/service_test.go` (4),
integración `backend/tests/historial_test.go` T4a–T4k (11, BD del compose),
UI `frontend/src/pages/viajes/index.test.tsx` (6), manual `quickstart.md` §3 (S4).

| Criterio | Verificación |
|---|---|
| **FR-001** lista cronológica con fecha/hora, estaciones, duración, coste y orden DESC | `TestHistorialT4a` (orden `fecha_inicio DESC` estricto + campos + nombres por JOIN) |
| **FR-002** filtro mes/año en Europe/Madrid | `TestHistorialT4bFiltroMes` + `TestHistorialT4gZonaHoraria` (día 31 23:30 → mes esperado) |
| **FR-003** sumatorio del periodo o del historial completo | `TestHistorialT4bFiltroMes` (3,75) + `TestHistorialT4ePaginacion`/`T4fPaginaFueraDeRango` (10,75 sin filtro, idéntico en todas las páginas) |
| **FR-004** viajes de 0,00 € con etiqueta | `TestHistorialT4cEtiqueta` (1 con etiqueta / 4 sin) |
| **FR-005** solo propietario desde JWT, 401 sin sesión, cualquier rol | `TestHistorialT4dAislamientoY401` (aislamiento + 401) + `TestHistorialT4kRoles` (T016: `operario_taller`/`gestor` → 200 propio; parámetro ajeno ignorado) |
| **FR-006** origen/destino por viaje, sin GPS | `TestHistorialT4a` (`estacion_*_nombre`) · Assumptions (alcance) |
| **FR-007** coste y duración = registrados al cierre (HU-003) | `TestHistorialT4a` (campos del response) · S4 p.6 (sin recálculo) |
| **FR-008** paginación server-side 50/200, suma del periodo, fuera de rango sin error | `TestHistorialT4ePaginacion`, `T4fPaginaFueraDeRango`, `T4iEntradaInvalida` |
| **SC-001** 100 % viajes finalizados con campos correctos | T4a + inspección SQL (S4) |
| **SC-002** filtros exactos y sumatorio = suma del periodo | T4b/T4e/T4f + S4 p.2–3 |
| **SC-003** viajes sin coste con la etiqueta | T4c + S4 p.1 |
| **SC-004** < 3 s con un año de uso | `TestHistorialT4jLatencia` (T017: fixture de ≥ 365 viajes finalizados antes de cronometrar) |
| **AC1** lista con campos | T4a |
| **AC2** filtro + sumatorio | T4b |
| **AC3** etiqueta en viaje gratuito | T4c |
| **AC4** cualquier rol autenticado → historial propio; 401 sin sesión | T4d + T4k |
| **AC5** viaje activo ni en lista ni en suma | T4a (total = solo `finalizado`; el `en_curso` de la fixture no cuenta) |
| **AC6** primera carga sin filtro: todo, DESC, suma total | T4a (orden) + T4e/T4f (`suma_coste` sin filtro) |
| UI: lista, filtros, suma, estados vacíos, error, paginación | `pages/viajes/index.test.tsx` (6 tests) |

**Resultado**: 4 unit + 11 integración (Go) + 6 componente (Vitest) en
verde (2026-11-19: +`TestHistorialT4kRoles` T016, T4j ampliado a un año
T017); S4 manual: OK (2026-10-28).
