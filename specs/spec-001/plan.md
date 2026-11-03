# Implementation Plan: Plan técnico global y modelo de datos compartido SI-16 (8 especificaciones)

**Branch**: `spec-001` | **Date**: 2026-10-20 (actualizado 2026-11-03 tras `/speckit.clarify`) | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/spec-001/spec.md` + contexto global: `historia.md` (HU-01…HU-08), `.specify/memory/constitution.md` (SI-16 v1.0.1) y la instrucción de plan global del equipo.

**Note**: Este plan es **transversal a las 8 specs**. Aunque el workflow lo genera en `specs/spec-001/`, su contenido (pila, esquema PostgreSQL, claves Redis, carpetas y rutas API) es el contrato compartido que deben seguir spec-001…spec-008. Cada spec generará su propio `tasks.md` derivado de este plan.

## Summary

El SI-16 necesita una única base técnica coherente para las 8 historias: la constitución fija la pila (SolidJS + Tailwind en frontend, Go/Gin en backend, PostgreSQL como modelo único y Redis como estado en vivo con frescura ≤ 30 s) y exige arquitectura separada en `/backend` y `/frontend` con modelo de datos compartido. Este plan define: (1) la pila completa con versiones y un `docker-compose.yml` raíz para levantar PostgreSQL y Redis; (2) el esquema PostgreSQL unificado de 7 tablas (usuarios, estaciones, bicicletas, anclajes, viajes, incidencias_mantenimientos, alertas_balanceo) con restricciones, índices y máquinas de estado que satisfacen los 8 conjuntos de FRs; (3) el contrato de claves Redis `estacion:{id}:estado` actualizadas en < 30 s; y (4) el árbol de carpetas `/backend` + `/frontend` con paquetes y rutas API `/api/v1/...` particionados por spec para que los 4 integrantes trabajen sin colisiones de fusión. El backend es la única autoridad de las reglas críticas (abono, un solo viaje activo, 30 min gratis, bloqueo por avería, umbrales de alerta).

**Actualización 2026-11-03 (aclaraciones de HU-01)**: la cota de 30 s pasa a medirse de extremo a extremo (del commit al pintado), así que el barrido del worker baja de 15 s a 10 s (R16). Se añaden a la UI del mapa el aviso de «datos desactualizados» a los > 30 s sin actualización correcta (R17) y estados visuales accesibles, no solo de color, para «sin bicicletas», «sin anclajes libres» y «fuera de servicio» (este último prevalece y oculta los contadores) (R18). Se confirma que los anclajes con bicicleta averiada no suman en ningún contador (R19). Ningún contrato compartido (API, Redis, esquema) cambia.

## Technical Context

**Language/Version**: Go 1.26+ (backend; mínimo fijado por las dependencias en `backend/go.mod`, antes 1.22+); TypeScript 5.x + SolidJS 1.9+ (frontend).

**Primary Dependencies**: Gin v1.10, pgx/v5 (driver PostgreSQL), go-redis/v9, goose (migraciones), golang-jwt/v5, bcrypt (x/crypto); Vite 6, Tailwind CSS, @solidjs/router, Leaflet + tiles OpenStreetMap (mapa HU-01), Chart.js (gráficos HU-08).

**Storage**: PostgreSQL 16 (esquema único, 7 tablas compartidas) + Redis 7 (hashes de estado en vivo por estación), ambos vía `docker-compose.yml` raíz.

**Testing**: `go test ./...` con base de datos de pruebas en Docker (integración) + Vitest y @solidjs/testing-library con MSW (frontend). Ver [quickstart.md](./quickstart.md).

**Target Platform**: Servidor Linux en Docker; navegadores web de escritorio y móvil (SPA).

**Project Type**: Monorepo web: servicio REST (Go/Gin) + SPA (SolidJS), infraestructura compartida en raíz.

**Performance Goals**: Frescura de disponibilidad ≤ 30 s extremo a extremo, del commit al pintado (const. IV, HU-01 FR-003; peor caso ≈ 28 s según R16); aviso de datos desactualizados a los > 30 s sin actualización correcta (HU-01 FR-007); mapa completo < 5 s; detalle de estación < 2 s; desbloqueo y devolución < 10 s; historial < 3 s; gráficos analítica < 5 s; alerta de balanceo generada < 60 s.

**Constraints**: Reglas de negocio exclusivamente server-side (const. I); stack fijado por constitución (const. II); todo cambio de esquema PostgreSQL/Redis requiere acuerdo previo del equipo (const. III); cada spec debe fusionarse de forma independiente (const. V) → carpetas y rutas particionadas por spec.

**Scale/Scope**: ~60 estaciones, flota de cientos de bicicletas, ~10²–10⁴ viajes, 4 integrantes × 2 specs, 8 áreas funcionales, 7 tablas + ~20 endpoints REST.

**Unknowns**: **Ningún NEEDS CLARIFICATION restante.** Los puntos abiertos (autenticación, estrategia de frescura ≤ 30 s, capa de acceso a datos, cálculo de tarifa, roles, formato CSV, naming de migraciones) se resuelven en [research.md](./research.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principio (constitución SI-16) | Cumplimiento en este plan | Estado |
|---|---|---|---|
| I | Autoridad del backend en reglas de negocio (NON-NEGOTIABLE) | Validación de abono vigente, unicidad de viaje activo, tarifa 30 min, bloqueo por avería y umbrales de alerta se implementan **solo** en servicios Go; restricciones DB de refuerzo (índices únicos parciales, CHECKs). El frontend solo replica UX. | PASS |
| II | Pila tecnológica acordada | SolidJS + Tailwind (frontend), Go + Gin (backend), PostgreSQL, Redis ≤ 30 s: exactamente lo definido, sin sustituciones. | PASS |
| III | Arquitectura separada con modelo de datos compartido | Código solo en `/backend` y `/frontend`; `docker-compose.yml`, docs y specs en raíz (infraestructura, no funcionalidad). Un único esquema PostgreSQL y un único formato de claves Redis documentados en [data-model.md](./data-model.md); adiciones al modelo base se marcan explícitamente para acuerdo de equipo. | PASS |
| IV | Disponibilidad en tiempo real con frescura ≤ 30 s | Estado vivo servido desde Redis (`estacion:{id}:estado`), actualización síncrona en cada mutación + barrido ≤ 10 s; frontend con polling ≤ 15 s, marca de última actualización y aviso de datos desactualizados a los > 30 s; TTL de seguridad con fallback a PostgreSQL. Peor caso extremo a extremo ≈ 28 s (R16). | PASS |
| V | Independencia por especificación | Paquetes backend (`services/`, `handlers/`, `routes/` por spec) y páginas frontend particionados por spec; registro de rutas en un único archivo trivial; migraciones numeradas por acuerdo. Cada spec se desarrolla y prueba sin esperar a las demás. | PASS |

**Gate de calidad (Acuerdos de Colaboración)**: las 5 reglas críticas del principio I se cubren con pruebas automatizadas obligatorias antes de fusionar (mapeadas en [data-model.md](./data-model.md) y validadas en [quickstart.md](./quickstart.md)). Rama por spec, PR revisado por otro integrante: proceso del equipo, fuera del alcance técnico de este plan.

**Veredicto del gate**: **PASS** — sin violaciones → *Complexity Tracking* vacío.

### Re-check post-Phase 1 (tras diseñar data-model, contracts y quickstart)

| # | Principio | Resultado del re-chequeo |
|---|---|---|
| I | Autoridad del backend | PASS — cada regla crítica tiene su endpoint/servicio Go + restricción DB + prueba en quickstart. |
| II | Pila | PASS — artifacts Phase 1 no introducen tecnologías nuevas. |
| III | Modelo compartido | PASS — `data-model.md` es la fuente única; adiciones señaladas (`password_hash`, `estado` en estaciones, índices) requieren acuerdo de equipo antes de migrar. |
| IV | Frescura ≤ 30 s | PASS — contrato Redis en `contracts/redis-state.md` + endpoint `GET /estaciones` con `actualizado_en`; quickstart mide la ventana (S1, S1a–S1d, incluido el camino de respaldo). |
| I (re-check 2026-11-03) | Autoridad del backend | PASS — los estados visuales de R18 son presentación derivada de los contadores del backend; no se toma ninguna decisión de negocio en el cliente. |
| V | Independencia | PASS — `contracts/openapi.yaml` particiona endpoints por tag/spec; propiedad de carpetas sin solapes. |

**Veredicto final**: **PASS (0 violaciones, 0 justificaciones necesarias)**.

## Project Structure

### Documentation (this feature)

```text
specs/spec-001/
├── plan.md              # Este plan (global para las 8 specs)
├── research.md          # Phase 0: decisiones técnicas y alternativas
├── data-model.md        # Phase 1: modelo de datos compartido (PostgreSQL + Redis)
├── quickstart.md        # Phase 1: guía de validación end-to-end
├── contracts/           # Phase 1: contrato de API REST y de claves Redis
│   ├── README.md        # Convenciones y propiedad por spec
│   ├── openapi.yaml     # OpenAPI 3.1 de /api/v1 (18 endpoints)
│   └── redis-state.md   # Formato de claves de estado en vivo
└── tasks.md             # Phase 2 (/speckit.tasks — NO creado por este comando)
```

### Source Code (repository root)

```text
/
├── docker-compose.yml            # PostgreSQL 16 + Redis 7 (infra compartida)
├── .env.example                  # DATABASE_URL, REDIS_ADDR, JWT_SECRET, TARIFA_MINUTO
├── historia.md                   # Fuente única de las 8 HU
├── README.md
│
├── backend/                      # Go + Gin
│   ├── cmd/
│   │   ├── server/main.go        # Punto de entrada HTTP
│   │   ├── migrate/main.go       # Aplicación de migraciones (goose)
│   │   └── seed/main.go          # Datos de prueba (~60 estaciones, roles, flota)
│   ├── internal/
│   │   ├── config/               # Lectura de .env
│   │   ├── db/                   # Pool pgx + helpers de transacción
│   │   ├── cache/                # Cliente Redis + constants de claves
│   │   ├── auth/                 # JWT, bcrypt, middleware de rol
│   │   ├── models/               # Structs compartidos (Usuario, Estacion, Viaje…)
│   │   ├── services/             # UN PAQUETE POR SPEC (dominio + reglas críticas)
│   │   │   ├── estacion/         # spec-001  (Iker)
│   │   │   ├── alquiler/         # spec-002  (Iker)
│   │   │   ├── devolucion/       # spec-003  (Alex)
│   │   │   ├── historial/        # spec-004  (Alex)
│   │   │   ├── incidencia/       # spec-005  (Josep)
│   │   │   ├── taller/           # spec-006  (Josep)
│   │   │   ├── balanceo/         # spec-007  (Alberto)
│   │   │   └── analitica/        # spec-008  (Alberto)
│   │   ├── handlers/             # HTTP handlers, espejo 1:1 con services/
│   │   ├── routes/
│   │   │   ├── register.go       # Único archivo compartido: llama a las 8 funcs
│   │   │   └── spec_00N_*.go     # RegisterSpec00N(rg, deps) — uno por spec
│   │   └── workers/
│   │       ├── refresco_estado/  # Recalcula estacion:{id}:estado cada ≤10 s (R16)
│   │       └── alertas/          # Barrido de umbrales <10% / >90% cada 60 s
│   ├── migrations/               # 0001_esquema_base.sql (acordado) + 000N_spec-00X_*.sql
│   └── tests/                    # integración (DB de pruebas Docker)
│
└── frontend/                     # SolidJS + Tailwind + Vite
    ├── src/
    │   ├── components/           # Compartidos: Layout, MapaBase, tabla, UI atómica
    │   ├── pages/
    │   │   ├── mapa/             # spec-001  (Iker): MapaPage, DetalleEstacion, AvisoDesactualizado, estadoVisual.ts (R17/R18)
    │   │   ├── alquiler/         # spec-002  (Iker)
    │   │   ├── devolucion/       # spec-003  (Alex)
    │   │   ├── historial/        # spec-004  (Alex)
    │   │   ├── incidencias/      # spec-005  (Josep)
    │   │   ├── taller/           # spec-006  (Josep)
    │   │   ├── balanceo/         # spec-007  (Alberto)
    │   │   └── analitica/        # spec-008  (Alberto)
    │   ├── services/             # api.ts (cliente fetch) + api/<spec>.ts por spec
    │   ├── stores/               # sesión, estado de la red en vivo (polling 15 s, ultimaActualizacionOk, desactualizado > 30 s)
    │   └── routes.tsx            # Único archivo compartido: rutas de páginas
    └── tests/
```

### Mapa de rutas API (base `/api/v1`, sin colisiones entre integrantes)

Registro: cada spec expone `RegisterSpec00N(rg *gin.RouterGroup, deps Deps)` en su propio fichero `routes/spec_00N_*.go`; solo `routes/register.go` es compartido (un merge trivial por spec).

| Spec | Integrante | Prefijo / endpoints | Paquetes propios |
|---|---|---|---|
| auth (base) | Iker | `POST /auth/login`, `GET /auth/me` | `auth/` |
| spec-001 | Iker Roig | `GET /estaciones`, `GET /estaciones/{id}` | `services/estacion`, `pages/mapa` |
| spec-002 | Iker Roig | `POST /alquileres/desbloqueo`, `GET /alquileres/activo` | `services/alquiler`, `pages/alquiler` |
| spec-003 | Alex Morell | `POST /devoluciones` | `services/devolucion`, `pages/devolucion` |
| spec-004 | Alex Morell | `GET /viajes` (filtros desde/hasta) | `services/historial`, `pages/historial` |
| spec-005 | Josep Peiró | `POST /incidencias`, `GET /incidencias` | `services/incidencia`, `pages/incidencias` |
| spec-006 | Josep Peiró | `GET /taller/bicicletas/{id}`, `POST /taller/bicicletas/{id}/intervenciones`, `POST /taller/bicicletas/{id}/ubicacion` | `services/taller`, `pages/taller` |
| spec-007 | Alberto Copa | `GET /balanceo/alertas`, `POST /balanceo/traslados` | `services/balanceo`, `pages/balanceo` |
| spec-008 | Alberto Copa | `GET /analitica/viajes`, `GET /analitica/export` | `services/analitica`, `pages/analitica` |

Contrato completo (payloads, códigos de error, roles): [contracts/openapi.yaml](./contracts/openapi.yaml).

**Structure Decision**: monorepo con dos raíces de código (`/backend`, `/frontend`) según principio III; partición interna **por spec** (paquete/página/archivo de rutas propio) según principio V; infraestructura compartida limitada a `docker-compose.yml`, `.env.example` y docs en raíz. Los servicios Go son la única fuente de verdad de las reglas críticas; el frontend consume el estado vivo de Redis vía `GET /estaciones`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Sin violaciones: el Constitution Check inicial y el re-chequeo post-Phase 1 resultaron PASS (tabla superior). No se registra ninguna complejidad añadida que justificar.
