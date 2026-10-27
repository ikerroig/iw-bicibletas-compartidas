# Specification Quality Checklist: Devolución de bicicleta y cálculo de tarificación

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
- Validado 2026-10-20: la tarifa por minuto excedente queda como configuración
  operativa documentada en Assumptions (valor numérico fuera de alcance).

## Acceptance Coverage (post-implementation, T027 — 2026-10-27)

Cada criterio de aceptación del spec con el test/verificación que lo demuestra.
Suites: unit `backend/internal/services/devolucion/tarifa_test.go`,
integración `backend/tests/devolucion_test.go` (BD+Redis del compose),
UI `frontend/src/pages/devolucion/index.test.tsx`, manual `quickstart.md` §3 (S3).

| Criterio | Verificación |
|---|---|
| **FR-001** finaliza fijando la bici en anclaje libre | `TestDevolucion_ViajeCorto_FinalizaYGratuito` (anclaje → `ocupado` con la bici del viaje) · S3 p.3–4 |
| **FR-002** registra hora de llegada, duración y coste | ídem (`fecha_fin`, `duracion_minutos`, `coste_total`) · `TestDevolucion_RelojInconsistente` (siempre `fecha_fin`) |
| **FR-003** viajes ≤ 30 min → 0,00 € | `TestCalcularCoste_ViajeCortoGratuito` + `TestCalcularCoste_Borde30Minutos` (T3a/T3c) · `TestDevolucion_ViajeCorto_FinalizaYGratuito` · S3 p.3 |
| **FR-004** viajes > 30 min → solo excedente × TARIFA_MINUTO | `TestCalcularCoste_UnMinutoExcedente` + `_45Minutos` + `_DosHoras` (T3d/T3b) · S3 p.6 (45,5 min → 3,20 € = 16 × 0,20) |
| **FR-005** persiste inicio, fin, duración, coste y destino | `TestDevolucion_ViajeCorto_FinalizaYGratuito` (incluye `estacion_destino_id`) · `TestDevolucion_RelojInconsistente` (guardas de duración/coste) |
| **FR-006** un viaje solo se cierra una vez | `TestDevolucion_DobleConfirmacion_SoloUnaEfecto` (T3f: exactamente 1 × 200 + 1 × 409) · S3 p.7 |
| **FR-007** rechazos sin efectos (sin viaje / anclaje ocupado / inexistente / retirado) | `TestDevolucion_SinViajeActivo_RechazaSinEfectos` (T3e) + `TestDevolucion_AnclajesInvalidos_Rechazan` (T3g) · S3 p.7 |
| Edge: bici averiada durante el viaje | `TestDevolucion_BiciAveriadaDuranteViaje_QuedaAncladaYaveriada` (T3h) |
| Edge: reloj inconsistente | `TestDevolucion_RelojInconsistente_CosteNoNegativo` (T3i) |
| **SC-001** 100 % viajes ≤ 30 min con coste 0,00 € | T3a/T3c + integración + S3 p.3 |
| **SC-002** 100 % viajes > 30 min cobrados solo excedente | T3b/T3d + S3 p.6 |
| **SC-003** devolución válida < 10 s | T3j (aserción de latencia en `TestDevolucion_ViajeCorto_FinalizaYGratuito`) |
| **SC-004** usuario libre de inmediato, 0 dobles cierres | T3f + `TestDevolucion_ViajeCorto…` (0 `en_curso`) + S3 p.4 (nuevo viaje aceptado) |
| Constitución IV: frescura de estado ≤ 30 s | S3 p.5 (`actualizado_en` de `estacion:2:estado` post-commit; `anclajes_libres` recalculado R9) |
| UI: formulario, éxito con etiqueta, errores 409 | `index.test.tsx` (3 tests: éxito tramo gratuito, `SIN_VIAJE_ACTIVO`, `ANCLAJE_OCUPADO`) |

**Resultado**: 7 unit + 6 integración (Go) + 3 componente (Vitest) en verde;
S3 manual: todas las comprobaciones OK (2026-10-27).
