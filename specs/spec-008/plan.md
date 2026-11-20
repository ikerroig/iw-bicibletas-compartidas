# Implementation Plan: Panel de analítica de uso y exportación de datos a CSV (HU-08)

**Branch**: `spec-008` | **Date**: 2026-11-01 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/spec-008/spec.md` (checklist completado, 0 NEEDS CLARIFICATION; 3 clarificaciones del 2026-11-01: rango inclusivo Europe/Madrid, agrupación por inicio 00–23 / lunes–domingo, CSV R11; y decisión de alcance del 2026-11-20: conservar los gráficos horarios y semanales sin añadir ranking por estación ni modificar el contrato compartido) + `historia.md` (HU-08) + plan técnico global compartido en [specs/spec-001/plan.md](../spec-001/plan.md) + modelo compartido en [specs/spec-001/data-model.md](../spec-001/data-model.md) y contratos en [specs/spec-001/contracts/](../specs/spec-001/contracts/) (R11 CSV, R14 analítica/Chart.js), sobre la base ya implementada por spec-003/spec-004/spec-007 presente en esta rama.

**Note**: Este plan **deriva del plan global** (spec-001) y de la base aportada por spec-003/spec-004/spec-007; no redefine pila, esquema PostgreSQL ni contrato Redis. Solo detalla lo necesario para la HU-08 dentro del marco ya acordado. El alcance es exactamente la HU-08: no se incorporan funcionalidades de otras historias ni se modifican specs/spec-007 ni el plan global.

## Summary

Exponer al gestor municipal la analítica de uso de la red sobre viajes finalizados: el backend calcula, para un rango `desde`/`hasta` (inclusivo, Europe/Madrid, validado server-side → `422` si inválido), los agregados del volumen de viajes **por hora del día (00–23)** y **por día de la semana (lunes→domingo)** agrupados sobre la **fecha de inicio**, y los sirve en `GET /analitica/viajes` (schema global `AnaliticaViajes`); en paralelo, `GET /analitica/export` entrega el histórico filtrado del **mismo rango** como CSV en streaming con BOM UTF-8, comas, RFC 4180, CRLF y las columnas del contrato global `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur` (R11), de modo que gráficos y CSV son idénticos **por construcción** (un único predicado compartido, FR-006). Ambos endpoints exigen rol `gestor` (`RequireAuth` + `RequireRole`, 401/403). Frontend SolidJS: página `pages/analitica/` con selector de rango (por defecto últimos 30 días), dos gráficos de barras **Chart.js** (R14 — alta en `frontend/package.json` autorizada y aplicada en T020), estado "sin datos en el periodo" y descarga autenticada vía `fetch` + Blob. **Cero migraciones nuevas y cero claves Redis nuevas**: se leen `viajes` y `estaciones` con el índice `ix_viajes_fecha` ya existente. El `422` de rango inválido está documentado en el contrato propio y en el OpenAPI global (T019, autorizada y aplicada).

## Technical Context

**Language/Version**: Go 1.22+ (backend); TypeScript 5.x + SolidJS 1.9+ (frontend) — fijado por constitución II y plan global.

**Primary Dependencies**: Gin, pgx/v5 (ya en `backend/go.mod`, sin dependencias Go nuevas); Vite 6 + Tailwind CSS + **chart.js** (R14 del plan global; incorporada en T020, autorizada por el equipo). Sin sustituciones respecto al plan global.

**Storage**: PostgreSQL 16 — **solo lectura** sobre `viajes` (filtro `estado='finalizado'`, `ix_viajes_fecha`) y `estaciones` (JOIN de nombres en el CSV), esquema de `specs/spec-001/data-model.md` (`0001_esquema_base.sql`, ya en esta rama). **Redis no se usa**: la HU-08 no consume ni escribe estado en vivo (const. IV no aplica).

**Testing**: `go test ./...` (integración contra BD del compose: T8a–T8i del quickstart, fixtures propias, patrón de `historial_test.go`/`balanceo_test.go` — sin kin-openapi, research D9). Frontend: Vitest + @solidjs/testing-library + stub de `fetch`.

**Target Platform**: Servidor Linux en Docker; SPA web (escritorio y móvil).

**Project Type**: Monorepo web — esta spec aporta 2 endpoints + 1 servicio + 1 página SPA dentro de las raíces `/backend` y `/frontend`. Sin workers (la analítica se calcula bajo demanda, R14).

**Performance Goals**: Gráficos < 5 s para rangos de hasta un año (SC-001 — 3 consultas pequeñas indexadas, presupuesto verificado por test cronometrado); CSV en streaming sin memoria proporcional al rango (caso "varios años").

**Constraints**: Reglas de negocio solo server-side (const. I: filtro Europe/Madrid, agregados, totales y validación de rango se producen en Go/SQL; el frontend pinta los arrays); stack fijado (const. II); sin cambios de esquema compartido (const. III — los puntos que tocan ficheros compartidos se marcan explícitamente); paquetes y rutas propios de spec-008 para no colisionar con otras ramas (const. V).

**Scope/Scale**: 2 endpoints (`GET /analitica/viajes`, `GET /analitica/export`), 1 servicio (`services/analitica`), 1 página (`pages/analitica`), 3 entradas de integración de una línea en 3 ficheros compartidos (`routes/register.go`, frontend `routes.tsx` e `index.tsx`) + bloques aditivos en `models/models.go` y `cmd/seed/main.go` + 1 dependencia frontend (`chart.js`, aplicada en T020).

**Unknowns**: **0 NEEDS CLARIFICATION** — el spec está validado (checklist [requirements.md](./checklists/requirements.md)); la revisión del 2026-11-20 confirmó el alcance de `historia.md` y no requiere ampliar el schema `AnaliticaViajes`. El OpenAPI global define ambos endpoints, sus parámetros, el schema `AnaliticaViajes` y las respuestas `422` (T019). Las decisiones de diseño están en [research.md](./research.md) (D1–D9).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principio (constitución SI-16) | Cumplimiento en este plan | Estado |
|---|---|---|---|
| I | Autoridad del backend en reglas de negocio | El filtro del rango (inclusivo, Europe/Madrid, `desde ≤ hasta` → 422), la agrupación por hora/día de la fecha de inicio, los totales y el contenido del CSV se calculan **solo** en Go/SQL (`services/analitica`); el frontend pinta los arrays y descarga el fichero ya generado. Validación de rol server-side (403). | PASS |
| II | Pila tecnológica acordada | Go/Gin + SolidJS/Tailwind + PostgreSQL: exactamente el stack del plan global. Chart.js no es una incorporación libre: es la tecnología designada **globalmente** para los gráficos HU-08 (R14). | PASS |
| III | Arquitectura separada con modelo de datos compartido | Código solo en `backend/internal/{services,handlers}/analitica`, `backend/internal/routes/spec_008_*.go` y `frontend/src/pages/analitica/`. **Cero migraciones**: lectura de `viajes`/`estaciones` con `ix_viajes_fecha` ya existente. Cero claves Redis. La documentación `422` en el OpenAPI global y el alta de `chart.js` fueron **autorizadas por el equipo y aplicadas** (T019/T020); no se modifica ningún fichero de spec-007 ni el plan global. | PASS |
| IV | Disponibilidad en tiempo real ≤ 30 s | **No aplica a esta spec**: la analítica es histórica (viajes finalizados, R11/R14); no lee ni escribe el estado en vivo de Redis. El contrato `estacion:{id}:estado` queda intacto. | N/A |
| V | Independencia por especificación | Paquetes, página y fichero de rutas propios; los puntos comunes son líneas de una entrada cada uno (`register.go`, `routes.tsx`, `index.tsx`) o bloques puramente aditivos (`models.go`, `seed`). Las pruebas usan fixtures propias. La HU-08 se desarrolla y prueba sin esperar a otras ramas. | PASS |

**Veredicto del gate**: **PASS** — sin violaciones → *Complexity Tracking* vacío.

### Re-check post-Phase 1 (tras diseñar data-model, contracts y quickstart)

| # | Principio | Resultado del re-chequeo |
|---|---|---|
| I | Autoridad del backend | PASS — [data-model.md](./data-model.md) define V1–V8 (validación de rango, filtro, agrupación, consistencia, formato CSV); quickstart T8a–T8f verifican consistencia, bordes del rango, exclusiones y 422. |
| II | Pila | PASS — los artifacts Phase 1 no introducen tecnologías nuevas (Chart.js ya designado en R14). |
| III | Modelo compartido | PASS — [data-model.md](./data-model.md) confirma cero cambios de esquema; los cambios compartidos de OpenAPI/dependencia fueron autorizados por el equipo y aplicados (T019/T020). Las adiciones de modelos, seed, rutas y navegación de HU-08 ya están aplicadas. |
| IV | Frescura ≤ 30 s | N/A — sin uso de Redis ni estado en vivo; los artifacts no lo requieren. |
| V | Independencia | PASS — [contracts/analitica.md](./contracts/analitica.md) detalla únicamente los 2 endpoints propios, alineados con el OpenAPI global sin modificarlo; fixtures de tests y seed propios. |

**Veredicto final**: **PASS (0 violaciones, 0 justificaciones necesarias)**.

## Project Structure

### Documentation (this feature)

```text
specs/spec-008/
├── plan.md              # Este plan
├── research.md          # Phase 0: decisiones D1–D9 de la HU-08
├── data-model.md        # Phase 1: subconjunto del modelo compartido (solo lectura de viajes/estaciones)
├── quickstart.md        # Phase 1: guía de validación end-to-end de la HU-08 (T8a–T8i, S8)
├── contracts/
│   └── analitica.md     # Contratos de GET /analitica/viajes y GET /analitica/export
└── tasks.md             # Phase 2 (/speckit.tasks — NO creado por este comando)
```

### Source Code (repository root)

```text
/
├── docker-compose.yml            # Compartido (infra) — ya existe, no se toca
├── .env / .env.example           # Compartido — sin cambios (la HU-08 no añade variables)
│
├── backend/                      # Go + Gin (base aportada por spec-003/spec-004/spec-007)
│   ├── cmd/seed/main.go          # Compartido: AMPLIAR con el lote de viajes HU-08 (research D7)
│   ├── internal/
│   │   ├── models/models.go      # Compartido: añadir struct AnaliticaViajes (espejo del OpenAPI)
│   │   ├── services/analitica/   # spec-008 — dominio HU-08
│   │   │   ├── service.go        # ValidarEntrada (422), construirFiltro, Agregados (Q1/Q2/Q3, D1–D3)
│   │   │   ├── service_test.go   # Unitarios (parseo/validación de rango, relleno de buckets, mapeo dow)
│   │   │   ├── csv.go            # Escritura CSV: BOM, coma, CRLF, columnas, fecha_hora Europe/Madrid (D4)
│   │   │   └── csv_test.go       # Unitarios (BOM, cabecera, decimal con punto, comillas RFC 4180)
│   │   ├── handlers/analitica/   # spec-008 — handlers HTTP (JSON + streaming CSV)
│   │   │   └── handler.go
│   │   └── routes/
│   │       ├── register.go       # Compartido: añadir 1 línea (RegisterSpec008)
│   │       └── spec_008_analitica.go  # GET /analitica/viajes, GET /analitica/export (rol gestor)
│   └── tests/
│       └── analitica_test.go     # Integración T8a–T8i contra la BD del compose (fixtures propias)
│
└── frontend/                     # SolidJS + Tailwind + Vite (base de spec-003/spec-004/spec-007)
    ├── package.json              # Compartido: dependencia chart.js (R14 — T020 autorizada y aplicada)
    └── src/
        ├── pages/analitica/      # spec-008 — panel del gestor
        │   ├── index.tsx         # Selector de rango + 2 gráficos Chart.js + descarga + estados
        │   └── index.test.tsx    # Vitest + stub de fetch
        ├── services/api/analitica.ts  # spec-008 — clientes GET agregados / GET export (Blob)
        ├── routes.tsx            # Compartido: añadir 1 ruta trivial (/analitica)
        └── index.tsx             # Compartido: añadir 1 enlace de nav ("Analítica")
```

**Structure Decision**: se aplica el Option 2 (web application: `/backend` + `/frontend`) del plan global, particionado por spec (principio V). La funcionalidad de la HU-08 vive íntegramente en `services/analitica`, `handlers/analitica`, `routes/spec_008_analitica.go`, `pages/analitica/` y `services/api/analitica.ts`. Los puntos compartidos (`routes/register.go`, `frontend/src/routes.tsx`, `frontend/src/index.tsx`, `models/models.go`, `cmd/seed/main.go`, `frontend/package.json`) reciben adiciones de una línea o bloques nuevos sin modificar lo existente (conflictos de merge triviales). No hay workers ni uso de Redis en esta spec.

### Mapa de endpoints (heredado del plan global)

| Spec | Método y ruta | Rol | Paquetes propios |
|---|---|---|---|
| spec-008 | `GET /analitica/viajes` | `gestor` | `services/analitica`, `handlers/analitica`, `pages/analitica` |
| spec-008 | `GET /analitica/export` | `gestor` | ídem |

Detalle de payloads, formato CSV y códigos de error: [contracts/analitica.md](./contracts/analitica.md) (alineado con `specs/spec-001/contracts/openapi.yaml`, tag `spec-008`, operationIds `analiticaViajes` y `exportarCSV`).

## Integraciones compartidas

La documentación de `422 ENTRADA_INVALIDA` en el OpenAPI global (T019) y el alta de `chart.js` en `frontend/package.json`/lockfile (T020) fueron **autorizadas por el equipo y aplicadas**. No alteran esquema ni Redis. Las integraciones aditivas de modelos, seed, rutas y navegación ya están aplicadas y quedan sujetas a la revisión normal del PR.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Sin violaciones: el Constitution Check inicial y el re-chequeo post-Phase 1 resultaron PASS (tablas superiores). No se registra ninguna complejidad añadida que justificar.
