# Implementation Plan: Desbloqueo de bicicleta mediante código de anclaje (HU-02)

**Branch**: `spec-002` | **Date**: 2026-11-04 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/spec-002/spec.md` + plan técnico global compartido en [specs/spec-001/plan.md](../spec-001/plan.md) (pila, esquema, claves Redis, carpetas y rutas API para las 8 specs).

**Note**: Este plan **deriva del plan global** (spec-001); no redefine pila, esquema PostgreSQL ni contrato Redis. Solo detalla lo necesario para la HU-02.

**Trazabilidad del proceso**: este plan se generó **después** de una primera implementación (commit `e090f41`), hecha a partir de las tareas T033–T039 del `tasks.md` global, sin `clarify`, `plan`, `tasks` ni `analyze` propios de esta spec. Se completa ahora el flujo de Spec Kit, y este plan es el diseño vigente: lo que difiere de `e090f41` se marca como **Delta** y se corregirá desde el `tasks.md` de esta spec.

## Summary

Abrir el ciclo de alquiler: un usuario con abono vigente introduce el código de un anclaje y el sistema crea su viaje activo, desacopla la bicicleta y actualiza el estado en vivo. Enfoque técnico: `POST /alquileres/desbloqueo` con **una transacción PostgreSQL** que bloquea el anclaje, valida las reglas críticas y escribe; `ux_viajes_usuario_activo` es la red de seguridad frente a desbloqueos simultáneos, y tras el commit se recalcula `estacion:{id}:estado`. `GET /alquileres/activo` devuelve el viaje en curso. Frontend SolidJS: página `pages/alquiler/`. Sin cambios de esquema.

## Technical Context

**Language/Version**: Go 1.26+ (backend); TypeScript 5.x + SolidJS 1.9+ (frontend), según el plan global.

**Primary Dependencies**: Gin, pgx/v5, go-redis/v9; Vite 6 + Tailwind CSS. Sin dependencias nuevas; `time/tzdata` es de la biblioteca estándar (research D1).

**Storage**: PostgreSQL 16, solo DML sobre `viajes`, `anclajes`, `bicicletas` y lectura de `usuarios`, `estaciones`. Redis 7, reescritura de `estacion:{id}:estado` tras el commit mediante `RecalcularEstacion`.

**Testing**: `go test ./...` (integración contra la BD del compose: gates T1 y T2; contrato con kin-openapi). Frontend: Vitest + @solidjs/testing-library + MSW.

**Target Platform**: Servidor Linux en Docker; SPA web (escritorio y móvil).

**Project Type**: Monorepo web: 2 endpoints REST + 1 página SPA.

**Performance Goals**: Desbloqueo válido confirmado en < 10 s (SC-003), también con Redis degradado (research D6); frescura ≤ 30 s (const. IV).

**Constraints**: Reglas solo server-side (const. I); pila fijada (const. II); sin cambios de esquema (const. III); paquetes y rutas propios (const. V).

**Scale/Scope**: 2 endpoints, 1 servicio (`services/alquiler`), 1 handler, 1 página (`pages/alquiler`).

**Unknowns**: **0 NEEDS CLARIFICATION.** Las 5 aclaraciones de la sesión 2026-11-04 y los puntos técnicos se resuelven en [research.md](./research.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principio (constitución SI-16) | Cumplimiento en este plan | Estado |
|---|---|---|---|
| I | Autoridad del backend en reglas de negocio | Abono vigente, viaje único, disponibilidad y rol se validan solo en Go dentro de la transacción, con `ux_viajes_usuario_activo` y `ux_viajes_bicicleta_activa` como refuerzo. El frontend envía el código y muestra la causa. | PASS |
| II | Pila tecnológica acordada | Go/Gin + SolidJS/Tailwind + PostgreSQL + Redis, sin incorporaciones. | PASS |
| III | Arquitectura separada con modelo de datos compartido | **Cero migraciones** y ninguna clave Redis nueva. El contrato OpenAPI compartido recibe cambios aditivos (D3, D7) que se comunican al equipo. | PASS |
| IV | Disponibilidad en tiempo real ≤ 30 s | Reescritura síncrona tras el commit; el barrido de 10 s del worker cubre los fallos de Redis. | PASS |
| V | Independencia por especificación | Paquetes, página y rutas propios; solo una línea en cada punto común (`register.go`, `routes.tsx`, `Layout.tsx`). | PASS |

**Veredicto del gate**: **PASS**, sin violaciones.

### Re-check post-Phase 1 (tras diseñar data-model, contracts y quickstart)

| # | Principio | Resultado del re-chequeo |
|---|---|---|
| I | Autoridad del backend | PASS: la fecha del abono y la normalización del código se resuelven en el servidor; [quickstart.md](./quickstart.md) cubre T1 y T2. |
| II | Pila | PASS: los artefactos no introducen tecnologías nuevas. |
| III | Modelo compartido | PASS: [data-model.md](./data-model.md) no propone cambios de esquema. |
| IV | Frescura ≤ 30 s | PASS: V8 de quickstart comprueba el contador tras desbloquear. |
| V | Independencia | PASS: `bicicleta_codigo` y `estacion_origen_nombre` no tocan `models.Viaje`; el nombre coincide con el que usa spec-004, y el cambio en `RecalcularEstacion` (D6) no altera su firma. |

**Veredicto final**: **PASS (0 violaciones, 0 justificaciones necesarias)**.

## Project Structure

### Documentation (this feature)

```text
specs/spec-002/
├── spec.md              # Especificación con aclaraciones
├── plan.md              # Este plan
├── research.md          # Phase 0: decisiones de la HU-02 (hereda R5/R7 del global)
├── data-model.md        # Phase 1: subconjunto del modelo compartido
├── quickstart.md        # Phase 1: guía de validación
├── contracts/
│   └── alquileres.md    # Contrato de los 2 endpoints
└── tasks.md             # Phase 2 (/speckit-tasks; no lo crea este comando)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── services/alquiler/service.go     # Transacción y reglas críticas
│   ├── services/estacion/estado.go      # Delta D6: contexto propio del hook
│   ├── handlers/alquiler/handler.go     # HTTP y traducción de causas
│   ├── routes/spec_002_routes.go        # RegisterSpec002
│   └── routes/register.go               # 1 línea compartida
└── tests/
    ├── integration/spec002_*_test.go    # Gates T1 y T2
    └── contract/spec002_test.go

frontend/
├── src/services/api/alquiler.ts
├── src/pages/alquiler/Desbloqueo.tsx
├── src/routes.tsx, src/components/Layout.tsx   # 1 línea compartida en cada uno
└── tests/alquiler/desbloqueo.test.tsx
```

**Structure Decision**: se usa el árbol del plan global sin añadir raíces. `specs/spec-001/contracts/openapi.yaml` sigue siendo la referencia de la API; si difiere de [contracts/alquileres.md](./contracts/alquileres.md), prevalece el OpenAPI y se corrige este documento.

## Deltas respecto a `e090f41`

| Delta | Origen | Cambio |
|---|---|---|
| D1 | FR-002 | La vigencia del abono se calcula con la fecha de `Europe/Madrid`, no UTC |
| D2 | FR-001 | Código sin distinguir mayúsculas; entrada inválida (NUL, > 64) → 422 |
| D3 | FR-008 | Respuesta con `bicicleta_codigo` y `estacion_origen_nombre`; la página deja de mostrar el id y de pedir el detalle de la estación |
| D5 | FR-004 | `FOR SHARE` al leer la estación |
| D6 | SC-003 | Presupuesto de tiempo propio para recálculo y hook |
| D7 | Contrato | `401` declarado y ejemplo real de código |
| D8 | Mantenimiento | Eliminar ramas inalcanzables y el `signal` sin usar |
| D9 | FR-009 | Quitar el rol por defecto del menú |

## Complexity Tracking

Sin violaciones de la constitución; no aplica.
