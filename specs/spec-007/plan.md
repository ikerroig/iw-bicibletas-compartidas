# Implementation Plan: Monitorización de desequilibrios y alertas de balanceo de flota (HU-07)

**Branch**: `spec-007` | **Date**: 2026-10-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/spec-007/spec.md` (checklist completado, 0 NEEDS CLARIFICATION; sesión de clarify 2026-11-20 con 4 decisiones — Q1 destino validado, Q2 acceso operario, Q3 empate `fecha_generacion` ASC [enmienda H-001], Q4 estado vacío — ver [refinamiento/clarify.md](./refinamiento/clarify.md)) + `historia.md` (HU-07) + plan técnico global compartido en [specs/spec-001/plan.md](../spec-001/plan.md) + modelo compartido en [specs/spec-001/data-model.md](../spec-001/data-model.md) y contratos en [specs/spec-001/contracts/](../spec-001/contracts/), sobre la base ya implementada por spec-003/spec-004 presente en esta rama.

**Note**: Este plan **deriva del plan global** (spec-001) y de la base aportada por spec-003/spec-004; no redefine pila, esquema PostgreSQL ni contrato Redis. Solo detalla lo necesario para la HU-07 dentro del marco ya acordado. El alcance es exactamente la HU-07: no se incorporan funcionalidades de otras historias ni se tocan specs/spec-004.

## Summary

Detectar y gestionar desequilibrios de ocupación de la red: el backend evalúa, para cada estación `operativa`, la ocupación según la fórmula R9 del modelo compartido (`bicis_disponibles / capacidad_total`, averiadas excluidas — FR-001/FR-008) y genera automáticamente una alerta de prioridad alta con umbrales estrictos (`< 10 %` → `vacia_menos_10`, `> 90 %` → `llena_mas_90`), resolviéndola automáticamente al volver al rango `[10, 90]` (FR-002…FR-004). La generación es de doble mecanismo (R10): evaluación **inline** tras cada traslado y un **worker `alertas`** (cada 55 s — desviación mínima de R10, research D6 —, arrancado desde `cmd/server/main.go`) que cubre cualquier otra mutación de la red (devoluciones, averías, taller) sin tocar el código de otras specs (const. V). El operario de logística consume dos endpoints propios, ya definidos en el OpenAPI global: `GET /balanceo/alertas` (panel ordenado por urgencia, FR-005) y `POST /balanceo/traslados` (confirmación de lote origen→destino en transacción con bloqueo, auditoría persistente y recálculo inmediato, FR-006/FR-007). Frontend SolidJS: página `pages/balanceo/` con la lista de alertas y el formulario de traslado. **Dos migraciones aprobadas por el equipo**: `0004_spec-007_traslados.sql` crea la auditoría y `0005_spec-007_score.sql` añade el índice de expresión parcial para el orden por urgencia; Redis sigue usando `estacion:{id}:estado` sin cambios.

## Technical Context

**Language/Version**: Go 1.22+ (backend); TypeScript 5.x + SolidJS 1.9+ (frontend) — fijado por constitución II y plan global.

**Primary Dependencies**: Gin v1.10, pgx/v5, go-redis/v9 (ya en `backend/go.mod`, sin dependencias nuevas); Vite 6 + Tailwind CSS. Sin incorporaciones respecto al plan global.

**Storage**: PostgreSQL 16 — DML sobre `alertas_balanceo` (INSERT/UPDATE), `estaciones`, `anclajes` y `bicicletas` (esquema de `specs/spec-001/data-model.md`, migración `0001_esquema_base.sql` ya presente en esta rama). Redis 7 — reutiliza `estacion:{id}:estado` vía el helper existente `cache.ActualizarEstadoEstacion` (post-commit del traslado); **sin claves nuevas**.

**Testing**: `go test ./...` (integración contra BD del compose: T7a–T7h del quickstart). Frontend: Vitest + @solidjs/testing-library + stub de `fetch` (patrón establecido por spec-003/spec-004).

**Target Platform**: Servidor Linux en Docker; SPA web (escritorio y móvil).

**Project Type**: Monorepo web — esta spec aporta 2 endpoints + 1 worker + 1 página SPA dentro de las raíces `/backend` y `/frontend`.

**Performance Goals**: Alerta generada < 60 s (SC-001 = periodo de 55 s + barrido < 5 s verificado — research D6, desviación mínima de R10); panel < 3 s; traslado < 10 s con recálculo inmediato en la respuesta y frescura ≤ 30 s en el estado global (SC-003, const. IV).

**Constraints**: Reglas de negocio solo server-side (const. I: umbrales, unicidad y validación del traslado se producen en Go + restricción DB `ux_alerta_abierta`); stack fijado (const. II); cambios del esquema compartido limitados a las dos migraciones aprobadas por el equipo (const. III); paquetes y rutas propios de spec-007 para no colisionar con otras ramas (const. V).

**Scope/Scale**: 2 endpoints (`GET /balanceo/alertas`, `POST /balanceo/traslados`), 1 worker (`alertas`), 1 servicio (`services/balanceo`), 1 página (`pages/balanceo`), 2 entradas de una línea en ficheros compartidos (`routes/register.go`, `cmd/server/main.go`) + 2 en el frontend (`routes.tsx`, `index.tsx`), ampliación del seed con estaciones en extremos.

**Unknowns**: **0 NEEDS CLARIFICATION** — el spec está validado (checklist [requirements.md](./checklists/requirements.md)) y el OpenAPI global ya define ambos endpoints, `AlertaBalanceo` y `TrasladoRequest`. Las decisiones de diseño están en [research.md](./research.md) (D1–D8). El clarify de 2026-11-20 no añade incertidumbre: Q1/Q2/Q4 ya estaban cubiertas por este plan (V1, V7, T011, T013, T008/T016) y Q3 se resolvió (H-001) a favor del orden ya implementado (`urgencia DESC, fecha_generacion ASC`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principio (constitución SI-16) | Cumplimiento en este plan | Estado |
|---|---|---|---|
| I | Autoridad del backend en reglas de negocio | Umbrales estrictos (<10 % / >90 %), unicidad por (estación, umbral), resolución automática y validación del traslado (unidades, estaciones operativas, capacidad) se implementan **solo** en `services/balanceo`; refuerzo en BD con `ux_alerta_abierta`. El frontend solo pinta el panel y envía el formulario. | PASS |
| II | Pila tecnológica acordada | Go/Gin + SolidJS/Tailwind + PostgreSQL + Redis: exactamente el stack del plan global, sin incorporaciones. | PASS |
| III | Arquitectura separada con modelo de datos compartido | Código solo en los paquetes de HU-07. El equipo aprobó previamente las migraciones secuenciales `0004_spec-007_traslados.sql` y `0005_spec-007_score.sql`; se añaden la tabla de auditoría y un índice parcial de expresión, sin cambiar claves Redis. | PASS |
| IV | Disponibilidad en tiempo real ≤ 30 s | El traslado escribe `estacion:{id}:estado` de ambas estaciones vía `cache.ActualizarEstadoEstacion` post-commit (helper existente, patrón de la devolución); el worker `alertas` no toca Redis. El contrato Redis queda intacto. | PASS |
| V | Independencia por especificación | Paquetes, worker, página y fichero de rutas propios; los puntos comunes son líneas de una entrada cada uno (`register.go`, `main.go`, `routes.tsx`, `index.tsx`). El worker `alertas` cubre las mutaciones de otras specs **sin modificar su código**. Las pruebas usan fixtures propias. | PASS |

**Veredicto del gate**: **PASS** — sin violaciones → *Complexity Tracking* vacío.

### Re-check post-Phase 1 (tras diseñar data-model, contracts y quickstart)

| # | Principio | Resultado del re-chequeo |
|---|---|---|
| I | Autoridad del backend | PASS — [data-model.md](./data-model.md) define V1–V8 (umbrales, unicidad, exclusión, validaciones del traslado); quickstart T7a/T7b/T7g verifican generación, rechazo y no-duplicados. |
| II | Pila | PASS — los artifacts Phase 1 no introducen tecnologías nuevas. |
| III | Modelo compartido | PASS — el cambio queda limitado a las migraciones `0004_spec-007_traslados.sql` y `0005_spec-007_score.sql`, aprobadas por el equipo el 2026-11-20 y documentadas en [data-model.md](./data-model.md). |
| IV | Frescura ≤ 30 s | PASS — el traslado reutiliza `cache.ActualizarEstadoEstacion` (post-commit); [contracts/balanceo.md](./contracts/balanceo.md) lo garantiza en la respuesta y el estado global. |
| V | Independencia | PASS — [contracts/balanceo.md](./contracts/balanceo.md) detalla únicamente los 2 endpoints propios, alineados con el OpenAPI global sin modificarlo. |

**Veredicto final**: **PASS (0 violaciones, 0 justificaciones necesarias)**.

## Project Structure

### Documentation (this feature)

```text
specs/spec-007/
├── plan.md              # Este plan
├── research.md          # Phase 0: decisiones D1–D8 de la HU-07
├── data-model.md        # Phase 1: subconjunto del modelo compartido (alertas + traslado)
├── quickstart.md        # Phase 1: guía de validación end-to-end de la HU-07
├── contracts/
│   └── balanceo.md      # Contratos de GET /balanceo/alertas y POST /balanceo/traslados
└── tasks.md             # Phase 2 (/speckit.tasks — NO creado por este comando)
```

### Source Code (repository root)

```text
/
├── docker-compose.yml            # Compartido (infra) — ya existe, no se toca
├── .env / .env.example           # Compartido — sin cambios (la HU-07 no añade variables)
│
├── backend/                      # Go + Gin (base aportada por spec-003/spec-004)
│   ├── cmd/seed/main.go          # Compartido: AMPLIAR con estaciones en extremos (research D8)
│   ├── cmd/server/main.go        # Compartido: añadir 1 línea de arranque del worker alertas
│   ├── internal/
│   │   ├── models/models.go      # Compartido: añadir structs AlertaBalanceo/Traslado (espejo del OpenAPI)
│   │   ├── services/balanceo/    # spec-007 — dominio HU-07
│   │   │   ├── umbrales.go       # EvaluarEstacion: crea/resuelve alertas (R9 + umbrales estrictos)
│   │   │   ├── umbrales_test.go  # Unitarios (umbrales, unicidad, resolución, exclusión)
│   │   │   ├── traslados.go      # ConfirmarTraslado: transacción con FOR UPDATE
│   │   │   └── traslados_test.go # Unitarios (validaciones de entrada)
│   │   ├── handlers/balanceo/    # spec-007 — handlers HTTP + DTOs
│   │   │   └── handler.go
│   │   ├── workers/alertas/      # spec-007 — barrido de umbrales cada 55 s (research D6)
│   │   │   └── worker.go
│   │   └── routes/
│   │       ├── register.go       # Compartido: añadir 1 línea (RegisterSpec007)
│   │       └── spec_007_balanceo.go  # GET /balanceo/alertas, POST /balanceo/traslados
│   └── tests/
│       └── balanceo_test.go      # Integración T7a–T7h contra la BD del compose
│
└── frontend/                     # SolidJS + Tailwind + Vite (base de spec-003/spec-004)
    └── src/
        ├── pages/balanceo/       # spec-007 — panel de logística
        │   ├── index.tsx         # Lista de alertas por urgencia + formulario de traslado
        │   └── index.test.tsx    # Vitest + stub de fetch
        ├── services/api/balanceo.ts  # spec-007 — clientes GET alertas / POST traslados
        ├── routes.tsx            # Compartido: añadir 1 ruta trivial
        └── index.tsx             # Compartido: añadir 1 enlace de nav
```

**Structure Decision**: se aplica el Option 2 (web application: `/backend` + `/frontend`) del plan global, particionado por spec (principio V). La funcionalidad de la HU-07 vive íntegramente en `services/balanceo`, `handlers/balanceo`, `workers/alertas`, `routes/spec_007_balanceo.go`, `pages/balanceo/` y `services/api/balanceo.ts`. Los puntos compartidos (`routes/register.go`, `cmd/server/main.go`, `frontend/src/routes.tsx`, `frontend/src/index.tsx`, `cmd/seed/main.go`, `models/models.go`) reciben adiciones de una línea o bloques nuevos sin modificar lo existente (conflictos de merge triviales). El worker `refresco_estado` (spec-001) **no** forma parte de esta spec.

### Mapa de endpoints (heredado del plan global)

| Spec | Método y ruta | Rol | Paquetes propios |
|---|---|---|---|
| spec-007 | `GET /balanceo/alertas` | `operario_logistica` | `services/balanceo`, `handlers/balanceo`, `pages/balanceo` |
| spec-007 | `POST /balanceo/traslados` | `operario_logistica` | ídem |

Detalle de payloads, respuestas y códigos de error: [contracts/balanceo.md](./contracts/balanceo.md) (alineado con `specs/spec-001/contracts/openapi.yaml`, tag `spec-007`, operationIds `panelBalanceo` y `confirmarTraslado`).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Sin violaciones: el Constitution Check inicial y el re-chequeo post-Phase 1 resultaron PASS (tablas superiores). No se registra ninguna complejidad añadida que justificar.
