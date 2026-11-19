# converge-2 — spec-004 (HU-04)

**Comando**: `/speckit.converge` (2ª pasada, §5.3 pasos 6–7) · **Spec Kit**: 1.1.3
**Fecha**: 2026-11-19 · **Responsable**: Alex Morell
**Funcionalidad activa**: `specs/spec-004` · **Rama**: `spec-004`
**Estado de partida**: remediaciones F1/F3/F4 aplicadas (commit
`30b518a`); F2 documentado; hallazgos previos en
[converge-1](./converge-1.md).

## Verificación de remediaciones

| ID | Remediación esperada | Verificación | Estado |
|----|----------------------|--------------|--------|
| F1 | Fila `500 ERROR_INTERNO` + redacción acotada en `contracts/viajes.md` | Tabla de errores con 401/422/500; nota sobre envelope global | ✅ |
| F2 | Nota de coordinación en el contrato (sin tocar el OpenAPI global) | Nota presente; `spec-001/contracts/openapi.yaml` **sin cambios** (const. III) | ✅ (ítem de equipo) |
| F3 | `codigoError` + aserciones en T4d/T4i | `NO_AUTENTICADO` en 401 y `ENTRADA_INVALIDA` en los 6 casos 422; `go test ./...` y `go vet` verdes | ✅ |
| F4 | Aviso de filtro incompleto + test | `aviso-filtro-incompleto` en `index.tsx`; test «filtro a medias» verifica que no se envía `desde=`; `npm test` **64/64** | ✅ |

## Evaluación final spec ↔ código

- **FR 8/8 · SC 4/4 · AC 6/6** con tarea, prueba y evidencia — sin huecos.
- **Tareas 21/21 `[X]`** (T001–T021); ningún `- [ ]` en `tasks.md`.
- Contrato ↔ handler ↔ tests ↔ UI: consistentes (parámetros, validación,
  envelope, códigos de error, garantías 200, orden DESC, etiqueta backend,
  estado inicial sin filtro C3).
- Constitución: I PASS · II — PR pendiente de abrir por Alex · III PASS
  (OpenAPI global intacto; F2 solo documentado) · IV N/A · V PASS.
- Suites: `go vet` OK · `go test ./...` verde (integración con compose) ·
  `npm test` 64/64.
- Análisis previos: clarify (3 decisiones C1–C3), analyze inicial/final
  (A1–A6 cerrados o ejecutados), sin CRITICAL/HIGH abiertos.

**Hallazgos nuevos**: 0.

## Resultado

**✅ CONVERGIDO** — spec-004 cumple spec/plan/tasks/contract/checklists con
código y tests verdes; F2 queda como único ítem de coordinación (acuerdo
de equipo para añadir el `422` al OpenAPI global).

**Siguientes pasos**: push de la rama `spec-004`, PR hacia `main` abierta
por Alex (revisión cruzada con Iker, const. II), y registro en
`docs/matriz-trazabilidad.md` (HU-03/HU-04).
