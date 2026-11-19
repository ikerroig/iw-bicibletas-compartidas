# converge-2 — spec-003 (HU-03)

**Comando**: `/speckit-converge` (2ª pasada) · **Spec Kit**: 1.1.3
**Fecha**: 2026-11-19 · **Responsable**: Alex Morell
**Funcionalidad activa**: `specs/spec-003` · **Rama**: `spec-003`
**Estado de partida**: Phase 6 implementada (T032 `9dbed49`).

## Resultado

**✅ Converged — la implementación satisface el spec, el plan y los tasks.**

`tasks.md` sin cambios (0 hallazgos accionables; sin nueva fase).

## Verificación de los hallazgos de converge-1

| ID | Estado | Evidencia |
|----|--------|-----------|
| F1 (≤64 chars, contrato) | **Resuelto** | `handler.go` valida longitud en runas tras trim → `422 ENTRADA_INVALIDA`; `TestDevolucion_CodigoLargo_ValidacionContrato` (65 ASCII/Unicode → 422 sin efectos · 64 ASCII y 33 multibyte `Ñ` → pasa a 409 `ANCLAJE_NO_ENCONTRADO`) en verde |
| F2 (recuento UI del checklist) | **Resuelto** | Fila UI del checklist: «4 tests» con los nombres; línea Resultado actualizada a 8 unit + 9 integración + 4 componente |

### Evidencia reproducible antes / después de la corrección F1 (T032, §5.3 paso 4)

Para la corrección de código de F1 (T032), se reproduce el fallo previo ejecutando el test `TestDevolucion_CodigoLargo_ValidacionContrato` sobre la versión inmediatamente anterior a la corrección (`9dbed49~1`):

| Caso / Entrada | Resultado Esperado | Resultado Observado Antes (`9dbed49~1`) | Resultado Observado Después (`9dbed49` / actual) |
|----------------|-------------------|------------------------------------------|--------------------------------------------------|
| `codigo_anclaje` de 65 caracteres (`strings.Repeat("A", 65)`) | `422 ENTRADA_INVALIDA` (sin efectos sobre viaje ni anclaje) | `409 ANCLAJE_NO_ENCONTRADO` ❌ (se delegaba al servicio sin validar longitud per contrato) | `422 ENTRADA_INVALIDA` ✅ (`codigo_anclaje no puede superar 64 caracteres`) |
| `codigo_anclaje` de 64 caracteres (`strings.Repeat("A", 64)`) | `409 ANCLAJE_NO_ENCONTRADO` (supera validación y busca el anclaje) | `409 ANCLAJE_NO_ENCONTRADO` ✅ | `409 ANCLAJE_NO_ENCONTRADO` ✅ |
| `codigo_anclaje` multibyte de 33 «Ñ» (66 bytes, 33 caracteres Unicode ≤ 64) | `409 ANCLAJE_NO_ENCONTRADO` (longitud en caracteres válida per contrato) | `409 ANCLAJE_NO_ENCONTRADO` (sin validación) | `409 ANCLAJE_NO_ENCONTRADO` ✅ (`utf8.RuneCountInString` cuenta caracteres, no bytes) |

> **Nota sobre T028–T030**: Para las tareas T028 (permisos FR-008 / AC4), T029 (abono caducado FR-009 / AC5) y T030 (redondeo fraccionario AC6), el código preexistente ya cumplía la funcionalidad esperada por omisión o diseño; se añadieron exclusivamente suites de pruebas automatizadas y asertos de caracterización para blindar el comportamiento frente a regresiones, por lo que no existía fallo de implementación previo que corregir.

## Inventario re-verificado (2ª pasada)

- **Requisitos**: 9 FR (MUST) + 4 SC + 6 AC → todos con prueba; FR-008/FR-009
  ahora con T028/T029; AC6 con T030; contrato ≤64 con T032.
- **Plan**: D1–D7, touchpoints (ruta/handler/servicio/tarifa/cliente/página),
  transacción `FOR UPDATE` + UPDATE condicionado + Redis post-commit,
  índice `ux_viajes_usuario_activo` en `0001_esquema_base.sql:70`.
- **Constitución**: I (cliente no calcula el importe), III (sin cambios de
  esquema ni de OpenAPI), IV (frescura Redis testada), V (suites
  automatizadas) → PASS; II satisfecho a nivel de proceso (PR revisado).
- **Tasks**: T001–T033 todos verificados con evidencia (no solo checkbox).
- **OpenAPI vs contrato local**: el `devolver` global lista 200/409
  representativos; el contrato local documenta además 401/403/422 —
  omisión, no contradicción (precedente: 422 ya existía en T019 sin estar
  en el OpenAPI). No es hallazgo; cualquier alineación del OpenAPI global
  requiere acuerdo del equipo (const. III).

**Métricas**: 19 requisitos/criterios · 33 tasks · 5 principios constitución ·
findings: 0 (missing 0, partial 0, contradicts 0, unrequested 0) ·
0 CRITICAL / 0 HIGH / 0 MEDIUM / 0 LOW.

## Verificaciones de cierre (2026-11-19)

- `go vet ./...` → OK
- `go test ./... -count=1` → verde (8 unit tarifa + 9 integración devolución + todas las suites)
- `npm test` → 11 ficheros, 63/63 verde

## Siguiente

- Abrir PR `spec-003` → `main` (revisado por otro integrante, constitución II).
- Repetir el ciclo clarify → analyze → converge en `spec-004` (HU-04).

*Sin hooks de extensión (`.specify/extensions.yml` no existe).*
