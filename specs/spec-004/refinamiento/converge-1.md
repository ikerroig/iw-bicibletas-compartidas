# converge-1 — spec-004 (HU-04)

**Comando**: `/speckit.converge` (1ª pasada, §5.3 paso 2) · **Spec Kit**: 1.1.3
**Fecha**: 2026-11-19 · **Responsable**: Alex Morell
**Funcionalidad activa**: `specs/spec-004` · **Rama**: `spec-004`
**Estado de partida**: 21/21 tareas `[X]`; informe
[analyze-final](./analyze-final.md) con 0 hallazgos abiertos.

## Artefactos cargados

`spec.md` · `plan.md` · `tasks.md` (T001–T021) · `contracts/viajes.md` ·
`checklists/requirements.md` · `refinamiento/{clarify,analyze-*.md}` ·
`research.md` · constitución del repo.

## Inventario de intención → código

| Área | Tareas | Verificación |
|---|---|---|
| Entorno + seed | T002, T003 | `cmd/seed/main.go`: 5 viajes finalizados de Ana (research D8) + viaje `en_curso` intacto |
| Servicio de lectura | T004, T005 | `services/viajes/service.go`: Q2 agregado + Q1 página con JOIN, `ORDER BY fecha_inicio DESC`, `estado='finalizado'`, `AT TIME ZONE 'Europe/Madrid'`, etiqueta server-side (`EtiquetaParaCoste`) |
| Handler + validación | T006 | `handlers/viajes/handler.go`: `ValidarEntrada` (page ≥1, page_size 1–200, fechas estrictas, desde ≤ hasta → 422 `ENTRADA_INVALIDA`); usuario **solo** del JWT (`auth.UsuarioID`) |
| Ruta | T007 | `routes/spec_004_viajes.go`: `GET /viajes` + `RequireAuth` únicamente (cualquier rol, AC4) |
| Contrato | T008 | `contracts/viajes.md` vs handler: parámetros, envelope, garantías 200, errores |
| UI Mis Viajes | T009–T011 | `pages/viajes/index.tsx`: estado inicial **sin filtro** (`mes=""`,`anio=""` → C3/AC6), `dd/mm/aaaa HH:MM` Europe/Madrid, etiqueta del backend, estados vacíos, `role="alert"`, paginación 50 |
| Test UI | T012 | `index.test.tsx` (6→7 tests) |
| Validación backend | T013–T015 | `historial_test.go` T4a–T4i |
| Refinamiento | T016–T021 | T4k roles (AC4), T4j con ≥365 viajes (SC-004), checklist Acceptance Coverage, plan/clarify/quickstart actualizados |

## Evaluación spec/plan/tasks ↔ código

- **FR 8/8**, **SC 4/4**, **AC 6/6** con tarea y prueba — sin huecos.
- Contrato ↔ implementación: coinciden en parámetros, validación, envelope,
  garantías (`total`/`suma_coste` del periodo entero, página fuera de rango
  sin error) y en el 401 documentado.
- Constitución: I (docs+tests) PASS · II (PR revisada) — pendiente de abrir
  PR · III (sin cambios de OpenAPI/modelo) PASS — ver F2 · IV N/A · V PASS.
- Suites: Go verde (`go vet` + `go test ./...`), frontend **64/64**.

## Hallazgos

| ID | Categoría | Gravedad | Descripción | Remediación |
|----|-----------|----------|-------------|-------------|
| F1 | Contrato desactualizado | **MEDIA** | `viajes.md` decía «No hay más códigos de error», pero el servicio devuelve `500 ERROR_INTERNO` en fallo de BD (`service.go:62`) | Contrato: fila `500` en la tabla + redacción acotada a «códigos de negocio» |
| F2 | OpenAPI global incompleto | BAJA (coordinación) | `historialViajes` en `spec-001/contracts/openapi.yaml` lista solo `200`/`401`; falta el `422` que sí implementa y testea el código (T4i) | Nota de coordinación en el contrato; **no se toca** el fichero global (const. III → acuerdo del equipo) |
| F3 | Tests sin aserción de código | BAJA | T4d/T4i comprobaban solo el status (401/422), no `error.codigo` del envelope | Helper `codigoError` + aserciones `NO_AUTENTICADO` y `ENTRADA_INVALIDA` |
| F4 | UX: filtro a medias | BAJA | Elegir solo mes (o solo año) no filtraba ni avisaba: la lista seguía completa con el selector marcado | Aviso `aviso-filtro-incompleto` en la barra de filtros + test nuevo |

**CRITICAL/HIGH**: 0. **Nuevos en esta pasada**: F1–F4 (los A1–A6 del
analyze estaban cerrados o con tarea ejecutada).

## Resultado

**NO convergido en converge-1** — 4 hallazgos de remediación inmediata
(1 media + 3 bajas; F2 solo documental). Remediaciones F1/F3/F4 aplicadas
en el ciclo actual; F2 queda como ítem de coordinación del equipo.
Pendiente: segunda pasada `converge-2` tras aplicarlas.
