# Implementation Plan: Consulta de historial de viajes y desglose mensual (HU-04)

**Branch**: `spec-004` | **Date**: 2026-10-28 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/spec-004/spec.md` (con las 5 clarificaciones de la sesión 2026-10-28) + plan técnico global compartido en [specs/spec-001/plan.md](../spec-001/plan.md) y base de código aportada por [spec-003](../spec-003/plan.md) (rama fusionada en `spec-004`).

**Note**: Este plan **deriva del plan global** (spec-001) y de la base ya implementada por spec-003; no redefine pila, esquema PostgreSQL ni contrato Redis. Solo detalla lo necesario para la HU-04 dentro del marco ya acordado.

## Summary

Mostrar al usuario su historial propio de viajes finalizados ("Mis Viajes") con filtro de mes/año, sumatorio del periodo, etiqueta de tramo gratuito y paginación server-side. Enfoque técnico: **una única lectura** en Go/Gin — `GET /api/v1/viajes` resuelve en dos consultas indexadas la página solicitada (JOIN a `estaciones` para los nombres, orden `fecha_inicio DESC`) y el agregado del periodo (`total`, `suma_coste`), con el filtro temporal interpretado en **Europe/Madrid** en SQL. Frontend SolidJS: página `pages/viajes/` con selects de mes/año, lista, sumatorio, controles de paginación y estados vacíos. **Sin cambios de esquema** ni uso de Redis: se leen las tablas `viajes` y `estaciones` ya definidas en el modelo compartido. Ampliación del seed con viajes finalizados de demostración.

## Technical Context

**Language/Version**: Go 1.22+ (backend); TypeScript 5.x + SolidJS 1.9+ (frontend) — fijado por constitución II y plan global.

**Primary Dependencies**: Gin v1.10, pgx/v5 (ya en `backend/go.mod`, sin dependencias nuevas); Vite 6 + Tailwind CSS. Sin incorporaciones respecto al plan global.

**Storage**: PostgreSQL 16 — **solo DML de lectura** sobre `viajes` y `estaciones` (esquema de `specs/spec-001/data-model.md`, migración `0001_esquema_base.sql`). Redis 7: **no se usa** (la HU-04 no es una función de disponibilidad en vivo).

**Testing**: `go test ./...` (integración del historial contra BD del compose: T4a–T4j del quickstart). Frontend: Vitest + @solidjs/testing-library + stub de `fetch` (patrón establecido por spec-003).

**Target Platform**: Servidor Linux en Docker; SPA web (escritorio y móvil).

**Project Type**: Monorepo web — esta spec aporta 1 endpoint de lectura + 1 página SPA dentro de las raíces `/backend` y `/frontend`.

**Performance Goals**: Historial y filtros en < 3 s con un año completo de uso habitual (SC-004); 2 consultas indexadas por petición.

**Constraints**: Reglas de negocio solo server-side (const. I: sumatorio, etiqueta y aislamiento los produce el backend); stack fijado (const. II); sin cambios de esquema compartido (const. III); cero uso de Redis (la constitución IV aplica a disponibilidad, no a historial); paquetes y rutas propios de spec-004 para no colisionar con las demás ramas (const. V).

**Scope/Scale**: 1 endpoint (`GET /viajes`), 1 servicio (`services/viajes`), 1 página (`pages/viajes`), 1 registro de rutas trivial compartido, ampliación del seed.

**Unknowns**: **0 NEEDS CLARIFICATION.** Las 5 preguntas de `/speckit.clarify` (paginación, zona horaria, fecha visible, contrato global, sumatorio fuera de rango) están resueltas y codificadas en [spec.md § Clarifications](./spec.md); su detalle técnico está en [research.md](./research.md) (D1–D8).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principio (constitución SI-16) | Cumplimiento en este plan | Estado |
|---|---|---|---|
| I | Autoridad del backend en reglas de negocio | El sumatorio del periodo, la etiqueta "Tramo gratuito (0,00 €)" y el filtro temporal (Europe/Madrid) se calculan **solo** en Go/SQL; el frontend pinta lo recibido y nunca suma ni etiqueta por su cuenta. Coste y duración se muestran tal los registró HU-03 (FR-007), sin recálculo. | PASS |
| II | Pila tecnológica acordada | Go/Gin + SolidJS/Tailwind + PostgreSQL: exactamente el stack del plan global, sin incorporaciones. | PASS |
| III | Arquitectura separada con modelo de datos compartido | Código solo en `backend/internal/{services,handlers}/viajes`, `backend/internal/routes/spec_004_*.go` y `frontend/src/pages/viajes/`. **Cero migraciones nuevas**: lectura de `viajes`/`estaciones` con el índice ya existente `ix_viajes_usuario_fecha`. | PASS |
| IV | Disponibilidad en tiempo real ≤ 30 s | **No aplica** a esta spec: el historial no es estado en vivo y no se lee/escribe Redis. El contrato Redis y sus claves quedan intactos. | PASS (N/A) |
| V | Independencia por especificación | Paquetes, página y fichero de rutas propios; los únicos puntos comunes son una línea en `routes/register.go` y una entrada en `frontend/src/routes.tsx` (merges triviales). La spec se prueba con fixtures propias sin depender de otras ramas. | PASS |

**Veredicto del gate**: **PASS** — sin violaciones → *Complexity Tracking* vacío.

### Re-check post-Phase 1 (tras diseñar data-model, contracts y quickstart)

| # | Principio | Resultado del re-chequeo |
|---|---|---|
| I | Autoridad del backend | PASS — [data-model.md](./data-model.md) define V3–V5 (sin recálculo, etiqueta en backend, suma en SQL); quickstart T4c/T4b lo verifica. |
| II | Pila | PASS — los artifacts Phase 1 no introducen tecnologías nuevas. |
| III | Modelo compartido | PASS — [data-model.md](./data-model.md) **no propone ningún cambio** de esquema; los dos campos de nombre en la respuesta son aditivos al contrato (patrón `AlertaBalanceo`) y no tocan `openapi.yaml`. |
| IV | Frescura ≤ 30 s | PASS (N/A) — sin lecturas/escrituras de Redis documentadas en ningún artifact. |
| V | Independencia | PASS — [contracts/viajes.md](./contracts/viajes.md) detalla únicamente el endpoint propio; sin solapes con specs vecinas. |

**Veredicto final**: **PASS (0 violaciones, 0 justificaciones necesarias)**.

## Project Structure

### Documentation (this feature)

```text
specs/spec-004/
├── plan.md              # Este plan
├── research.md          # Phase 0: decisiones D1–D8 de la HU-04
├── data-model.md        # Phase 1: subconjunto del modelo compartido (solo lectura)
├── quickstart.md        # Phase 1: guía de validación end-to-end de la HU-04
├── contracts/
│   └── viajes.md        # Contrato del endpoint GET /viajes y errores
└── tasks.md             # Phase 2 (/speckit.tasks — NO creado por este comando)
```

### Source Code (repository root)

```text
/
├── docker-compose.yml            # Compartido (infra) — ya existe, no se toca
├── .env / .env.example           # Compartido — sin cambios (la HU-04 no añade variables)
│
├── backend/                      # Go + Gin (base aportada por spec-003)
│   ├── cmd/seed/main.go          # Compartido: AMPLIAR con viajes finalizados (research D8)
│   ├── internal/
│   │   ├── models/models.go      # Compartido: no se toca (Viaje ya tiene Etiqueta)
│   │   ├── services/viajes/      # spec-004 — consulta y agregado del historial
│   │   │   ├── service.go        # ListarHistorial: Q1 página + Q2 agregado
│   │   │   └── service_test.go   # Unitarios (etiqueta, offsets, validación)
│   │   ├── handlers/viajes/      # spec-004 — handler HTTP + DTOs
│   │   │   └── handler.go
│   │   └── routes/
│   │       ├── register.go       # Compartido: añadir 1 línea (RegisterSpec004)
│   │       └── spec_004_viajes.go  # GET /viajes (grupo autenticado)
│   └── tests/
│       └── historial_test.go     # Integración T4a–T4j contra la BD del compose
│
└── frontend/                     # SolidJS + Tailwind + Vite (base de spec-003)
    └── src/
        ├── pages/viajes/         # spec-004 — página "Mis Viajes"
        │   ├── index.tsx         # Lista, filtros mes/año, sumatorio, paginación
        │   └── index.test.tsx    # Vitest + stub de fetch
        ├── services/api/viajes.ts  # spec-004 — cliente GET /viajes
        ├── routes.tsx            # Compartido: añadir 1 ruta trivial
        └── index.tsx             # Compartido: añadir 1 enlace de nav
```

**Structure Decision**: se aplica el Option 2 (web application: `/backend` + `/frontend`) del plan global, particionado por spec (principio V). La funcionalidad de la HU-04 vive íntegramente en `services/viajes`, `handlers/viajes`, `routes/spec_004_viajes.go`, `pages/viajes/` y `services/api/viajes.ts`. Los puntos compartidos (`routes/register.go`, `frontend/src/routes.tsx`, `frontend/src/index.tsx`, `cmd/seed/main.go`) reciben adiciones de una línea o bloques nuevos sin modificar lo existente (conflictos de merge triviales).

### Mapa de endpoint (heredado del plan global)

| Spec | Método y ruta | Paquetes propios |
|---|---|---|
| spec-004 | `GET /viajes` (auth: cualquier rol autenticado, siempre el propio historial) | `services/viajes`, `handlers/viajes`, `pages/viajes` |

Detalle de parámetros, respuestas y códigos de error: [contracts/viajes.md](./contracts/viajes.md) (alineado con `specs/spec-001/contracts/openapi.yaml`, tag `spec-004`, operationId `historialViajes`).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Sin violaciones: el Constitution Check inicial y el re-chequeo post-Phase 1 resultaron PASS (tablas superiores). No se registra ninguna complejidad añadida que justificar.
