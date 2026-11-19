# converge-1 — spec-003 (HU-03)

**Comando**: `/speckit-converge` (1ª pasada) · **Spec Kit**: 1.1.3
**Fecha**: 2026-11-19 · **Responsable**: Alex Morell
**Funcionalidad activa**: `specs/spec-003` · **Rama**: `spec-003`
**Estado de partida**: implementación completa (T001–T031) y suites en verde.

## Hallazgos

| ID | Gap Type | Severidad | Fuente | Evidencia | Trabajo restante |
|----|----------|-----------|--------|-----------|------------------|
| F1 | partial | MEDIUM | `contracts/devoluciones.md` (touchpoint del plan) | El contrato declara `422 ENTRADA_INVALIDA` para `codigo_anclaje` > 64 chars; `handler.go:36-40` solo valida no-vacío (sin test de longitud) | Validar ≤64 en el handler + test → **T032** |
| F2 | partial | LOW | T027 (checklist) | Fila UI dice «3 tests»; `index.test.tsx` tiene 4 (el 4º añadido en `d8d1626` tras la fusión de specs) | Corregir recuento → **T033** |

**Métricas**: 19 requisitos/criterios (9 FR + 4 SC + 6 AC) + 31 tareas ·
decisiones de plan D1–D7 + touchpoints + contrato + índice
`ux_viajes_usuario_activo` verificado en `0001_esquema_base.sql:70` ·
constitución I–V (II de proceso) · findings: missing 0, partial 2,
contradicts 0, unrequested 0 · severidad: 0 CRITICAL, 0 HIGH, 1 MEDIUM,
1 LOW.

## Verificaciones ejecutadas

- `go vet ./...` → OK
- `go test ./... -count=1` → verde (8 unit tarifa + 8 integración devolución + resto de suites)
- `npm test` → 11 ficheros, 63/63 en verde (incluye 4 tests de `pages/devolucion`)
- Servicio `CerrarViaje`: transacción única con `FOR UPDATE` (anclaje y viaje),
  UPDATE condicionado, bici preserva avería, COMMIT + Redis post-commit
- Ruta: `RequireAuth` + `RequireRole("usuario")` (FR-008) — códigos reales
  `NO_AUTENTICADO`/`ROL_NO_PERMITIDO` (auth.go:127-129)
- Índice de seguridad de concurrencia presente en el esquema compartido

## Resultado

**tasks_appended**: Phase 6 con **T032** y **T033** (append-only sobre
`tasks.md`). Siguiente: `/speckit-implement` y re-ejecutar converge
(`converge-2.md`), esperando `converged`.

*Sin hooks de extensión (`.specify/extensions.yml` no existe).*
