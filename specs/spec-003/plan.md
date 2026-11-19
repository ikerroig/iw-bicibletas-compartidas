# Implementation Plan: Devolución de bicicleta y cálculo de tarificación (HU-03)

**Branch**: `spec-003` | **Date**: 2026-10-27 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/spec-003/spec.md` + plan técnico global compartido en [specs/spec-001/plan.md](../spec-001/plan.md) (contrato de pila, esquema, claves Redis, carpetas y rutas API para las 8 specs).

**Note**: Este plan **deriva del plan global** (spec-001); no redefine pila, esquema PostgreSQL ni contrato Redis. Solo detalla lo necesario para la HU-03 dentro del marco ya acordado.

## Summary

Cerrar el ciclo de alquiler: el usuario fija la bicicleta en un anclaje libre de cualquier estación y el sistema finaliza el viaje registrando hora de llegada, liberando al usuario, anclando la bicicleta y tarificando (0,00 € si ≤ 30 min; solo excedente × `TARIFA_MINUTO` si > 30 min, con redondeo al minuto empezado). Solo el propietario del viaje con rol `usuario` puede cerrarlo (FR-008) y el cierre no exige abono vigente (FR-009). Enfoque técnico: un único endpoint `POST /devoluciones` (Go/Gin) con **una transacción PostgreSQL** que valida el viaje activo con `SELECT … FOR UPDATE`, cierra el viaje, actualiza anclaje y bicicleta, calcula el coste según R8 del plan global y, tras el commit, reescribe la clave Redis `estacion:{id}:estado` de la estación destino (frescura ≤ 30 s, constitución IV). Frontend SolidJS: página `pages/devolucion/` con formulario de código de anclaje y confirmación del importe. Sin cambios de esquema: se usan las tablas `viajes`, `anclajes`, `bicicletas` y `estaciones` ya definidas en el modelo compartido.

## Technical Context

**Language/Version**: Go 1.22+ (backend); TypeScript 5.x + SolidJS 1.9+ (frontend) — fijado por constitución II y plan global.

**Primary Dependencies**: Gin v1.10, pgx/v5, go-redis/v9; Vite 6 + Tailwind CSS. Sin dependencias nuevas respecto al plan global.

**Storage**: PostgreSQL 16 — solo DML sobre `viajes`, `anclajes`, `bicicletas`, `estaciones` (esquema de `specs/spec-001/data-model.md`, migración `0001_esquema_base.sql`). Redis 7 — invalidación/reescritura de `estacion:{id}:estado` post-commit.

**Testing**: `go test ./...` (unitario del cálculo de tarifa + integración contra BD del compose: T3 del gate de calidad). Frontend: Vitest + @solidjs/testing-library + MSW.

**Target Platform**: Servidor Linux en Docker; SPA web (escritorio y móvil).

**Project Type**: Monorepo web — esta spec aporta endpoint REST + página SPA dentro de las raíces `/backend` y `/frontend`.

**Performance Goals**: Devolución completa < 10 s (SC-003); frescura del estado de la estación destino ≤ 30 s tras el cierre (const. IV).

**Constraints**: Reglas de negocio solo server-side (const. I); stack fijado (const. II); sin cambios de esquema compartido (const. III); paquetes y rutas propios de spec-003 para no colisionar con las demás ramas (const. V).

**Scope/Scale**: 1 endpoint (`POST /devoluciones`), 1 servicio (`services/devolucion`), 1 página (`pages/devolucion`), 1 registro de rutas trivial compartido.

**Unknowns**: **0 NEEDS CLARIFICATION.** Todos los puntos (fórmula de tarifa, concurrencia, edge cases, códigos de error, invalidación Redis) se resuelven en [research.md](./research.md), heredando R3/R7/R8 del plan global. Tras la fase de refinamiento (§5.1), la sesión de `/speckit.clarify` del **2026-11-19** resolvió 5 preguntas (tarifa única del servicio, redondeo al minuto empezado, permisos de cierre FR-008, devolución con abono caducado FR-009 y estación sin anclajes) codificadas en [spec.md § Clarifications](./spec.md); D7 (permisos) ya anticipaba FR-008 y el cierre sigue sin consultar el abono (FR-009), por lo que **no hay cambios de arquitectura ni de esquema** derivados del clarify.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principio (constitución SI-16) | Cumplimiento en este plan | Estado |
|---|---|---|---|
| I | Autoridad del backend en reglas de negocio | El cálculo de coste (≤30 min → 0,00 €; excedente × TARIFA_MINUTO), la validación de viaje activo y el cierre único se implementan **solo** en `services/devolucion` (Go), dentro de una transacción, reforzados por los índices únicos parciales del modelo compartido. El frontend solo envía `codigo_anclaje` y pinta el resultado. | PASS |
| II | Pila tecnológica acordada | Go/Gin + SolidJS/Tailwind + PostgreSQL + Redis: exactamente el stack del plan global, sin incorporaciones. | PASS |
| III | Arquitectura separada con modelo de datos compartido | Código solo en `backend/internal/{services,handlers}/devolucion`, `backend/internal/routes/spec_003_*.go` y `frontend/src/pages/devolucion/`. **Cero migraciones nuevas**: se reutiliza `0001_esquema_base.sql` sin alterar (`viajes` ya prevé `fecha_fin`, `duracion_minutos`, `coste_total`, CHECKs de cierre; `anclajes` ya garantiza una bici en un solo anclaje). | PASS |
| IV | Disponibilidad en tiempo real ≤ 30 s | Tras confirmar la transacción de devolución, reescritura síncrona post-commit de `estacion:{id}:estado` de la **estación destino** (y de la origen si el modelo la afecta — ver research D4). El barrido de 15 s del worker global cubre como red de seguridad. | PASS |
| V | Independencia por especificación | Paquetes, página y fichero de rutas propios; el único punto tocado en común es una línea de registro en `routes/register.go` (merge trivial). La spec es testeable de forma aislada con datos seed. | PASS |

**Veredicto del gate**: **PASS** — sin violaciones → *Complexity Tracking* vacío.

### Re-check post-Phase 1 (tras diseñar data-model, contracts y quickstart)

| # | Principio | Resultado del re-chequeo |
|---|---|---|
| I | Autoridad del backend | PASS — cálculo de tarifa y cierre del viaje residen en `services/devolucion`; quickstart define T3 (≤30 min → 0,00 €; >30 min → solo excedente) y los rechazos (sin viaje activo, anclaje ocupado). |
| II | Pila | PASS — artifacts Phase 1 no introducen tecnologías nuevas. |
| III | Modelo compartido | PASS — [data-model.md](./data-model.md) **no propone ningún cambio** de esquema; documenta solo el subconjunto de tablas usado. |
| IV | Frescura ≤ 30 s | PASS — contrato de invalidación Redis documentado en [contracts/](./contracts/); quickstart mide `actualizado_en` tras la devolución. |
| V | Independencia | PASS — [contracts/devoluciones.md](./contracts/devoluciones.md) detalla únicamente el endpoint propio; sin solapes con specs vecinas. |

**Veredicto final**: **PASS (0 violaciones, 0 justificaciones necesarias)**.

## Project Structure

### Documentation (this feature)

```text
specs/spec-003/
├── plan.md              # Este plan
├── research.md          # Phase 0: decisiones de la HU-03 (hereda R3/R7/R8 del global)
├── data-model.md        # Phase 1: subconjunto del modelo compartido usado por la devolución
├── quickstart.md        # Phase 1: guía de validación end-to-end de la HU-03
├── contracts/           # Phase 1
│   └── devoluciones.md  # Contrato del endpoint POST /devoluciones y errores
└── tasks.md             # Phase 2 (/speckit.tasks — NO creado por este comando)
```

### Source Code (repository root)

```text
/
├── docker-compose.yml            # Compartido (infra) — ya existe
├── .env.example                  # Compartido; TARIFA_MINUTO ya previsto
│
├── backend/                      # Go + Gin (estructura del plan global)
│   ├── cmd/server/main.go        # No se toca (registro vía routes)
│   ├── internal/
│   │   ├── config/               # TARIFA_MINUTO ya contemplado en el global
│   │   ├── db/                   # Pool pgx + transacciones (compartido)
│   │   ├── cache/                # Cliente Redis + helper de reescritura de estado
│   │   ├── models/               # Struct Viaje/Anclaje/Bicicleta (compartido)
│   │   ├── services/devolucion/  # spec-003 — TODO el dominio de la HU-03
│   │   │   ├── service.go        # CerrarViaje: transacción, validaciones, tarifa
│   │   │   ├── tarifa.go         # Cálculo puro: segundos → coste (testeable)
│   │   │   └── service_test.go   # Unitarios de tarifa + casos borde
│   │   ├── handlers/devolucion/  # spec-003 — handler HTTP + DTOs de error
│   │   │   └── handler.go
│   │   └── routes/
│   │       ├── register.go       # Compartido: añadir 1 línea (RegistroSpec003)
│   │       └── spec_003_devolucion.go  # POST /devoluciones (grupo autenticado)
│   └── tests/                    # Integración: devolución end-to-end contra BD
│
└── frontend/                     # SolidJS + Tailwind + Vite
    └── src/
        ├── pages/devolucion/     # spec-003 — página de devolución
        │   ├── index.tsx         # Formulario de código de anclaje + resultado
        │   └── index.test.tsx    # Vitest + MSW
        ├── services/api/devolucion.ts  # spec-003 — cliente POST /devoluciones
        └── routes.tsx            # Compartido: añadir 1 ruta trivial
```

**Structure Decision**: se aplica el Option 2 (web application: `/backend` + `/frontend`) del plan global, particionado por spec (principio V). La funcionalidad de la HU-03 vive íntegramente en `services/devolucion`, `handlers/devolucion`, `routes/spec_003_devolucion.go`, `pages/devolucion/` y `services/api/devolucion.ts`. `routes/register.go` y `frontend/src/routes.tsx` reciben una sola línea cada uno (conflicto de merge trivial resoluble aceptando ambos bloques).

### Mapa de endpoint (heredado del plan global)

| Spec | Método y ruta | Paquetes propios |
|---|---|---|
| spec-003 | `POST /devoluciones` (auth: rol `usuario`) | `services/devolucion`, `handlers/devolucion`, `pages/devolucion` |

Detalle de payload, respuestas y códigos de error: [contracts/devoluciones.md](./contracts/devoluciones.md) (alineado con `specs/spec-001/contracts/openapi.yaml`, tag `spec-003`).

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Sin violaciones: el Constitution Check inicial y el re-chequeo post-Phase 1 resultaron PASS (tablas superiores). No se registra ninguna complejidad añadida que justificar.
