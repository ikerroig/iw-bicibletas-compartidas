# converge-2 — spec-004 (HU-04)

**Comando**: `/speckit.converge` (2ª pasada, §5.3 paso 5) · **Spec Kit**: 1.1.3
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

### Evidencia del «antes» y «después» de las correcciones (§5.3 paso 4)

| ID | Tipo de corrección | Entrada / Condición | Esperado | Observado Antes (30b518a~1) | Observado Después (30b518a) |
|---|---|---|---|---|---|
| **F1** | Documental (contrato) | Inspección de `contracts/viajes.md` frente al backend | Contrato contempla `500 ERROR_INTERNO` si falla BD | Contrato afirmaba: *«No hay más códigos de error»* a pesar de que `service.go:62` devolvía 500 en fallo de base de datos | Fila `500 ERROR_INTERNO` añadida a la tabla y texto acotado a *«códigos de negocio»* ✅ |
| **F2** | Documental (coordinación) | Comparación con `spec-001/contracts/openapi.yaml` | Trazabilidad del código `422` implementado en handler/tests | Discrepancia no documentada (el OpenAPI global omite 422 en `historialViajes`) | Nota de coordinación explícita en `contracts/viajes.md` sin alterar el fichero global (principio III) ✅ |
| **F3** | Cobertura en tests | Petición sin auth o parámetros inválidos en tests de integración | Comprobar HTTP status y `error.codigo` en el envelope JSON | Asertos verificaban únicamente el status HTTP (401/422), sin comprobar `error.codigo` | Helper `codigoError` añadido en `historial_test.go` verificando `NO_AUTENTICADO` y `ENTRADA_INVALIDA` ✅ |
| **F4** | UX / Comportamiento | Usuario selecciona solo Mes = 10, año = «—» | Aviso visible `aviso-filtro-incompleto` y petición sin parámetro `desde=` | Sin aviso: `Unable to find an element by: [data-testid="aviso-filtro-incompleto"]` (se mostraba la lista completa en silencio) | Aviso visible (`«Elige también el año…»`) y petición API sin `desde=` verificada por test ✅ |

## Evaluación final spec ↔ código

- **FR 8/8 · SC 4/4 · AC 6/6** con tarea, prueba y evidencia — sin huecos.
- **Tareas 25/25 `[X]`** (T001–T025, incluyendo T022–T025 en Phase 6: Convergence); ningún `- [ ]` en `tasks.md`.
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
