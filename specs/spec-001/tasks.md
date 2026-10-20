# Tasks: Plan técnico global SI-16 — Bicicletas Compartidas (8 especificaciones)

**Input**: Design documents from `/specs/spec-001/`

**Prerequisites**: [plan.md](./plan.md) ✓, [research.md](./research.md) ✓, [data-model.md](./data-model.md) ✓, [contracts/](./contracts/) ✓, [quickstart.md](./quickstart.md) ✓, specs de las 8 HU (`specs/spec-001…008/spec.md`)

**Tests**: **Incluidos obligatoriamente** — el gate de calidad de la constitución exige pruebas automatizadas de las 5 reglas críticas (T1–T5 de [quickstart.md](./quickstart.md)) antes de fusionar cualquier spec.

**Organization**: Tareas agrupadas por user story (= spec) para implementación y entrega independiente por integrante (constitución, principio V). Las fases 1–2 son la base compartida que hacen posible la independencia.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Puede ejecutarse en paralelo (ficheros distintos, sin dependencias)
- **[Story]**: US1…US8 → HU-01…HU-08

## Mapa de stories → specs → responsables

| Story | Spec / HU | Integrante | Fase |
|---|---|---|---|
| US1 | spec-001 — Consulta de disponibilidad en tiempo real | Iker Roig | 3 🎯 MVP |
| US2 | spec-002 — Desbloqueo por código de anclaje | Iker Roig | 4 |
| US3 | spec-003 — Devolución y tarificación | Alex Morell | 5 |
| US4 | spec-004 — Historial de viajes y desglose mensual | Alex Morell | 6 |
| US5 | spec-005 — Notificación de averías e inhabilitación | Josep Peiró | 7 |
| US6 | spec-006 — Intervenciones técnicas y taller | Josep Peiró | 8 |
| US7 | spec-007 — Alertas de balanceo de flota | Alberto Copa | 9 |
| US8 | spec-008 — Analítica de uso y exportación CSV | Alberto Copa | 10 |

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Estructura raíz e infraestructura compartida (equipo; cualquier integrante puede ejecutarla coordinándose — son ficheros únicos, no de dominio)

- [ ] T001 Crear la estructura raíz según `plan.md`: carpetas `backend/` y `frontend/` con sus árboles (`cmd/`, `internal/`, `migrations/`, `src/`…) y `.gitignore` combinado (Go + Node)
- [ ] T002 Crear `docker-compose.yml` en la raíz: `postgres:16-alpine` (5432, volumen `pgdata`, healthcheck `pg_isready`) y `redis:7-alpine` (6379, `--appendonly yes`, healthcheck `redis-cli ping`); credenciales vía `.env` (R12, aprobado por unanimidad)
- [ ] T003 [P] Crear `.env.example` en la raíz: `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `TARIFA_MINUTO=0.20`, `PUERTO=8080`
- [ ] T004 [P] Inicializar `backend/go.mod` (Go 1.22) con dependencias: gin, pgx/v5, go-redis/v9, golang-jwt/v5, x/crypto, goose
- [ ] T005 [P] Inicializar `frontend/`: Vite + SolidJS + TypeScript + Tailwind (`frontend/package.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/tailwind.config.js`)
- [ ] T006 [P] Configurar lint/format: `backend/.golangci.yml` (gofmt/vet) y `frontend/.eslintrc.cjs` + `frontend/.prettierrc`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Base compartida que MUST completarse ANTES de cualquier user story. Incluye el esquema completo (aprobado por el equipo: `password_hash`, `estaciones.estado` y constraints/índices) y los **helpers compartidos** que hacen independientes las 8 specs.

**⚠️ CRITICAL**: Ninguna historia puede empezar hasta completar esta fase.

- [ ] T007 Escribir `backend/migrations/0001_esquema_base.sql` con **íntegro** el modelo de `data-model.md`: tablas `usuarios`, `estaciones`, `bicicletas`, `anclajes`, `viajes`, `incidencias_mantenimientos`, `alertas_balanceo`; CHECKs, FKs e índices (`ux_viajes_usuario_activo`, `ux_viajes_bicicleta_activa`, `ux_anclaje_bicicleta` vía UNIQUE, `ux_alerta_abierta`, índices de HU-04/HU-08)
- [ ] T008 Implementar `backend/cmd/migrate/main.go` (goose up/down sobre `backend/migrations/`)
- [ ] T009 [P] `backend/internal/config/config.go`: carga `.env` (`DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `TARIFA_MINUTO`, `PUERTO`)
- [ ] T010 [P] `backend/internal/db/db.go`: pool pgx + helper de transacción con `SELECT … FOR UPDATE` (usado por US2/US3/US5/US6/US7)
- [ ] T011 [P] `backend/internal/cache/client.go` + `backend/internal/cache/keys.go`: cliente go-redis y contrato `estacion:{id}:estado` (HSET+EXPIRE 120s, HGETALL, SCAN de snapshot, fallback a PostgreSQL) — `contracts/redis-state.md`
- [ ] T012 [P] `backend/internal/hooks/estacion.go`: registro `OnEstacionCambiada(ids)` invocado tras cada commit que muta estaciones/bicicletas/anclajes; desacopla las specs que mutan (US2/US3/US5/US6/US7) de las que reaccionan (US1 recálculo, US7 umbrales) — clave del principio V
- [ ] T013 `backend/internal/models/`: structs compartidos (`Usuario`, `Estacion`, `Anclaje`, `Bicicleta`, `Viaje`, `Incidencia`, `AlertaBalanceo`), enums de estado y helper `AbonoVigente(u)` (`abono_activo AND fecha_vencimiento_abono >= hoy`, R15)
- [ ] T014 `backend/internal/auth/`: bcrypt, emisión de JWT HS256 (exp 24 h, claims `sub`/`rol`/`nombre`) y middlewares `RequireAuth` / `RequireRole(...)` (R5)
- [ ] T015 [P] `backend/internal/httperr/`: envelope `{error:{codigo,mensaje}}` y códigos de `contracts/README.md` (`ABONO_VENCIDO`, `VIAJE_ACTIVO`, …)
- [ ] T016 `backend/internal/routes/register.go`: router base `/api/v1` y patrón `RegisterSpec00N(rg, deps)` — **único archivo de rutas compartido** (merges triviales)
- [ ] T017 Implementar auth base: `POST /auth/login` y `GET /auth/me` en `backend/internal/handlers/auth/handler.go` + test mínimo de contrato
- [ ] T018 `backend/cmd/server/main.go`: wiring de config, db, cache, hooks, registro de rutas y workers (cada spec añadirá tras fusionar una única línea para su worker)
- [ ] T019 [P] `backend/cmd/seed/main.go`: ~60 estaciones geolocalizadas (incl. alguna `fuera_de_servicio` y extremos de ocupación para S7), anclajes = capacidad, flota (algunas `averiada`), usuarios de **todos los roles** con/sin abono vigente (hashes bcrypt)
- [ ] T020 `backend/internal/services/estacion/estado.go`: `RecalcularEstacion` (consulta agregada de disponibilidad excluyendo averiadas y estaciones fuera de servicio — R9 — + HSET del hash + invocación de `hooks.EstacionCambiada`); **helper compartido** que importarán US2/US3/US5/US6/US7 (propiedad del paquete de Iker)
- [ ] T021 [P] Frontend base: `frontend/src/services/api.ts` (fetch wrapper con Bearer y envelope de error), `frontend/src/stores/sesion.ts` (login/logout, token) y `frontend/src/pages/Login.tsx`
- [ ] T022 `frontend/src/routes.tsx` (**único archivo de rutas compartido**) + `frontend/src/components/Layout.tsx` (navegación por áreas con guard de rol)
- [ ] T023 [P] Configurar Vitest + @solidjs/testing-library + MSW en `frontend/vitest.config.ts` y `frontend/tests/setup.ts` con handlers derivados de `contracts/openapi.yaml`

**Checkpoint**: fundación lista — las 8 stories pueden empezar en paralelo, cada una en su rama `spec-00N`.

---

## Phase 3: User Story 1 — Consulta de disponibilidad en tiempo real (spec-001, HU-01, P1) 🎯 MVP

**Goal**: Mapa en vivo de las ~60 estaciones con contadores fiables desde Redis, frescura ≤ 30 s (Iker Roig).

**Independent Test**: Dar de alta estaciones con estados variados, abrir el mapa, verificar contadores contra la BD real y medir que un cambio se refleja en ≤ 30 s (escenario S1 del quickstart).

### Tests for User Story 1

- [ ] T024 [P] [US1] Test de contrato `GET /estaciones` y `GET /estaciones/{id}` contra `contracts/openapi.yaml` en `backend/tests/contract/spec001_test.go`
- [ ] T025 [P] [US1] Test de integración de frescura en `backend/tests/integration/spec001_frescura_test.go`: mutación → hash `estacion:{id}:estado` actualizado post-commit; clave expirada → fallback a PostgreSQL y repoblado; averiadas excluidas de `bicis_disponibles`; estación `fuera_de_servicio` marcada (FR-005/FR-006)

### Implementation for User Story 1

- [ ] T026 [US1] `backend/internal/workers/refresco_estado/worker.go`: barrido ≤ 15 s de las ~60 estaciones con una única consulta agregada (red de seguridad de frescura) + línea de wiring en `backend/cmd/server/main.go`
- [ ] T027 [US1] `backend/internal/handlers/estacion/handler.go` + `backend/internal/routes/spec_001_routes.go`: snapshot `GET /estaciones` (con `actualizado_en`) y detalle `GET /estaciones/{id}` (con anclajes)
- [ ] T028 [P] [US1] `frontend/src/services/api/estaciones.ts`: tipos `EstacionEstado`/`EstacionDetalle` y cliente
- [ ] T029 [P] [US1] `frontend/src/stores/red.ts`: polling de `GET /estaciones` cada 15 s, reintentos automáticos, marca de última actualización y pausa con pestaña oculta (edge cases HU-01)
- [ ] T030 [US1] `frontend/src/components/MapaBase.tsx`: Leaflet + tiles OpenStreetMap, marcadores con estado por estación
- [ ] T031 [US1] `frontend/src/pages/mapa/MapaRed.tsx`: mapa completo (< 5 s), panel de detalle (< 2 s) con contadores, estado de servicio, `actualizado_en` y aviso de desconexión conservando el último estado conocido
- [ ] T032 [P] [US1] Tests frontend en `frontend/tests/mapa/`: render con datos MSW, estación fuera de servicio, estado vacío y desconexión (Vitest + MSW)

**Checkpoint**: HU-01 funcional y testeable de forma independiente (MVP del sistema).

---

## Phase 4: User Story 2 — Desbloqueo de bicicleta por código de anclaje (spec-002, HU-02, P1)

**Goal**: Inicio de viaje atómico con las reglas críticas de abono y unicidad de viaje activo (Iker Roig).

**Independent Test**: Con usuario de prueba seed, introducir código de anclaje con bici disponible → viaje activo; con abono caducado → 403 con causa; con viaje activo → 409 (escenario S2). Ver `specs/spec-002/spec.md`.

### Tests for User Story 2

- [ ] T033 [P] [US2] **T1 (const. I)**: `backend/tests/integration/spec002_abono_test.go` — desbloqueo con abono caducado/ incoherente → 403 `ABONO_VENCIDO` y cero efectos sobre anclaje, bici y viajes
- [ ] T034 [P] [US2] **T2 (const. I)**: `backend/tests/integration/spec002_concurrencia_test.go` — segundo desbloqueo → 409 `VIAJE_ACTIVO`; dos desbloqueos concurrentes (goroutines) → exactamente 1 viaje creado (red `ux_viajes_usuario_activo`); código inexistente/estación fuera de servicio/anclaje sin bici operativa → 404/409 sin efectos
- [ ] T035 [P] [US2] Test de contrato `POST /alquileres/desbloqueo` y `GET /alquileres/activo` en `backend/tests/contract/spec002_test.go`

### Implementation for User Story 2

- [ ] T036 [US2] `backend/internal/services/alquiler/service.go`: transacción única con `SELECT … FOR UPDATE` del anclaje; valida abono vigente, ausencia de viaje activo, estación operativa, anclaje en servicio y bicicleta `disponible`; crea viaje `en_curso` con `fecha_inicio`; desacopla anclaje (→ `libre`), bici → `en_uso`; tras commit `RecalcularEstacion` + `hooks.EstacionCambiada` (T020)
- [ ] T037 [US2] `backend/internal/handlers/alquiler/handler.go` + `backend/internal/routes/spec_002_routes.go` (403/404/409 con envelope de `httperr`)
- [ ] T038 [P] [US2] `frontend/src/services/api/alquiler.ts`
- [ ] T039 [US2] `frontend/src/pages/alquiler/Desbloqueo.tsx`: entrada de código de anclaje, mensajes de causa legibles, estado de viaje activo actual

**Checkpoint**: HU-02 funcional; US1+US2 = ciclo de apertura de viaje completo.

---

## Phase 5: User Story 3 — Devolución y tarificación (spec-003, HU-03, P1)

**Goal**: Cierre de viaje con 30 min gratuitos y cobro solo del excedente (Alex Morell).

**Independent Test**: Con un viaje activo (fixture SQL en BD, sin depender de spec-002 fusionada), fijar la bici en un anclaje libre y verificar hora de llegada, liberación del usuario, anclaje y coste (≤ 30 min → 0,00 €; > 30 min → solo excedente) — escenario S3.

### Tests for User Story 3

- [ ] T040 [P] [US3] **T3 (const. I)**: `backend/tests/integration/spec003_tarifa_test.go` — 1800 s exactos → `coste_total = 0.00`; 1801 s → 1 minuto cobrado × `TARIFA_MINUTO`; duración `ceil`; duración negativa/reloj inconsistente → jamás coste negativo
- [ ] T041 [P] [US3] `backend/tests/integration/spec003_flujo_test.go` — sin viaje activo → 409 `SIN_VIAJE_ACTIVO` sin tocar el anclaje; anclaje ocupado/inexistente → 409/404; dos confirmaciones casi simultáneas → una sola efectiva; viaje averiado durante el trayecto se cierra correctamente
- [ ] T042 [P] [US3] Test de contrato `POST /devoluciones` en `backend/tests/contract/spec003_test.go`

### Implementation for User Story 3

- [ ] T043 [US3] `backend/internal/services/devolucion/service.go`: transacción con `FOR UPDATE` de viaje y anclaje; cierra viaje (`fecha_fin`, `estacion_destino_id`, `duracion_minutos = ceil(segundos/60)`, `coste_total` según R8 con `TARIFA_MINUTO` de config); ancla la bici (→ anclaje `ocupado`); bici → `disponible` (o `averiada` si la hubo reportada, HU-05); recalculo post-commit (T020)
- [ ] T044 [US3] `backend/internal/handlers/devolucion/handler.go` + `backend/internal/routes/spec_003_routes.go`
- [ ] T045 [P] [US3] `frontend/src/services/api/devolucion.ts`
- [ ] T046 [US3] `frontend/src/pages/devolucion/Devolucion.tsx`: confirmación de anclaje destino, resumen con duración/coste y etiqueta "Tramo gratuito (0,00 €)"; caso estación sin anclajes libres

**Checkpoint**: ciclo de alquiler cerrado (HU-02+HU-03); la flota vuelve íntegra a la red.

---

## Phase 6: User Story 4 — Historial de viajes y desglose mensual (spec-004, HU-04, P1)

**Goal**: "Mis Viajes" con filtros de periodo, sumatorio y etiqueta de tramo gratuito (Alex Morell).

**Independent Test**: Con viajes finalizados repartidos en varios meses (fixtures SQL), abrir "Mis Viajes", verificar campos/orden, aplicar filtro mes/año y comprobar sumatorio y etiqueta — escenario S4.

### Tests for User Story 4

- [ ] T047 [P] [US4] `backend/tests/integration/spec004_test.go` — el rango filtrado devuelve exactamente esos viajes; suma coincide con los listados; solo finalizados y solo del usuario autenticado (aislamiento FR-005); periodo vacío → lista vacía y suma 0,00
- [ ] T048 [P] [US4] Test de contrato `GET /viajes` (paginación, `suma_coste`, `etiqueta`) en `backend/tests/contract/spec004_test.go`

### Implementation for User Story 4

- [ ] T049 [US4] `backend/internal/services/historial/service.go`: consulta paginada sobre `ix_viajes_usuario_fecha` con rango `desde`/`hasta`, suma del periodo y etiqueta "Tramo gratuito (0,00 €)" para coste 0
- [ ] T050 [US4] `backend/internal/handlers/historial/handler.go` + `backend/internal/routes/spec_004_routes.go`
- [ ] T051 [P] [US4] `frontend/src/services/api/historial.ts`
- [ ] T052 [US4] `frontend/src/pages/historial/MisViajes.tsx`: lista cronológica (más recientes primero), filtro mes/año, sumatorio persistente con paginación, estados vacíos

**Checkpoint**: HU-04 funcional; el usuario ve y verifica sus cobros.

---

## Phase 7: User Story 5 — Notificación de averías e inhabilitación (spec-005, HU-05, P1)

**Goal**: Reporte de usuario que inhabilita la bici de inmediato y alimenta el panel de incidencias (Josep Peiró).

**Independent Test**: Con bici operativa anclada (fixture), enviar reporte con descripción y verificar cambio inmediato a `averiada`, exclusión de contadores (consistente con HU-01) e incidencia en el panel de operarios en < 30 s — escenario S5.

### Tests for User Story 5

- [ ] T053 [P] [US5] **T4 (const. I)**: `backend/tests/integration/spec005_inhabilitacion_test.go` — tras el reporte, desbloqueo de la bici → 409 `BICICLETA_NO_DISPONIBLE`; bici ausente de `bicis_disponibles` y del hash Redis; 0 secuencias dejan bici averiada pero alquilable
- [ ] T054 [P] [US5] `backend/tests/integration/spec005_reporte_test.go` — descripción vacía → 422; doble reporte simultáneo → un solo cambio de estado y respuesta `duplicada: true` sin segundo bloqueo; reporte de bici `en_uso` admitido (viaje continúa)
- [ ] T055 [P] [US5] Test de contrato `POST /incidencias` y `GET /incidencias` + 403 para rol no `operario_taller` en `backend/tests/contract/spec005_test.go`

### Implementation for User Story 5

- [ ] T056 [US5] `backend/internal/services/incidencia/service.go`: crea incidencia `pendiente` (`usuario_reporta_id` de sesión, `fecha_reporte`), bici → `averiada` en la misma transacción, idempotencia, recálculo post-commit (T020)
- [ ] T057 [US5] `backend/internal/handlers/incidencia/handler.go` + `backend/internal/routes/spec_005_routes.go`
- [ ] T058 [P] [US5] `frontend/src/services/api/incidencias.ts`
- [ ] T059 [US5] `frontend/src/pages/incidencias/ReportarAveria.tsx`: formulario de usuario (tipo_fallo, descripción obligatoria) con confirmación de registro
- [ ] T060 [US5] `frontend/src/pages/incidencias/PanelIncidencias.tsx`: vista `operario_taller` de pendientes ordenadas por fecha de reporte

**Checkpoint**: HU-05 funcional; la seguridad de la flota opera ya en el mapa de US1.

---

## Phase 8: User Story 6 — Intervenciones técnicas y taller (spec-006, HU-06, P1)

**Goal**: Ciclo de vida de reparación con parte firmado y reaparición controlada en el mapa (Josep Peiró).

**Independent Test**: Con bicicleta averiada con reporte (fixture), abrir ficha como operario, registrar parte marcándola operativa, verificar intervención firmada y que solo reaparece en el mapa tras confirmar ubicación — escenario S6.

### Tests for User Story 6

- [ ] T061 [P] [US6] `backend/tests/integration/spec006_parte_test.go` — campos obligatorios ausentes → 422; parte firmado con `operario_id` + `fecha_reparacion`; **todas** las incidencias abiertas de la bici → `resuelta`; mantenimiento preventivo inserta incidencia `resuelta` con `tipo_fallo = 'otro'`; cancelar sin guardar no altera nada
- [ ] T062 [P] [US6] `backend/tests/integration/spec006_ubicacion_test.go` — anclaje ocupado/inexistente → 409/404; bici `disponible` sin anclaje **no** figura en ningún contador hasta ubicar (FR-007); tras ubicar, visible en Redis en ≤ 30 s (consistencia con HU-01)
- [ ] T063 [P] [US6] Test de contrato de `GET/POST /taller/bicicletas/{id}…` + 403 para rol no `operario_taller` en `backend/tests/contract/spec006_test.go`

### Implementation for User Story 6

- [ ] T064 [US6] `backend/internal/services/taller/service.go`: ficha (bicicleta + historial de incidencias con el reporte original de HU-05 visible); cerrar reparación (UPDATE de todas las incidencias abiertas → `resuelta` con piezas/notas; bici → `disponible`); ubicar (transacción con `FOR UPDATE` del anclaje libre → `ocupado`, bici anclada; recálculo post-commit — T020)
- [ ] T065 [US6] `backend/internal/handlers/taller/handler.go` + `backend/internal/routes/spec_006_routes.go`
- [ ] T066 [P] [US6] `frontend/src/services/api/taller.ts`
- [ ] T067 [US6] `frontend/src/pages/taller/FichaBicicleta.tsx`: reporte original, historial, formulario de parte (piezas o notas obligatorios, checkbox "marcar operativa")
- [ ] T068 [US6] `frontend/src/pages/taller/UbicarBicicleta.tsx`: confirmación de ubicación en anclaje con resultado visible

**Checkpoint**: HU-06 funcional; el ciclo avería → reparación → disponibilidad está completo.

---

## Phase 9: User Story 7 — Alertas de balanceo de flota (spec-007, HU-07, P1)

**Goal**: Detección automática de desequilibrios (< 10 % / > 90 %) y panel/logística para el operario de calle (Alberto Copa).

**Independent Test**: Con estaciones seed en extremos de ocupación, verificar generación de alertas altas en < 60 s, orden del panel sin duplicados, resolución automática y recálculo inmediato tras confirmar un traslado — escenario S7.

### Tests for User Story 7

- [ ] T069 [P] [US7] **T5 (const. I)**: `backend/tests/integration/spec007_umbrales_test.go` — 9,99 % → `vacia_menos_10`; 10,00 % y 90,00 % → sin alerta (umbrales estrictos); 90,01 % → `llena_mas_90`; estación `fuera_de_servicio` excluida; averiadas fuera del cómputo
- [ ] T070 [P] [US7] `backend/tests/integration/spec007_alertas_test.go` — unicidad por (estación, umbral) con `ux_alerta_abierta`; vuelta a rango [10, 90] → `resuelta`; re-cruzar abre fila nueva; 0 alertas obsoletas
- [ ] T071 [P] [US7] `backend/tests/integration/spec007_traslados_test.go` — lote superior a disponibles en origen → 409 sin alterar nada; traslado válido → ambas estaciones recalculadas de inmediato en Redis
- [ ] T072 [P] [US7] Test de contrato `GET /balanceo/alertas` y `POST /balanceo/traslados` + 403 para rol no `operario_logistica` en `backend/tests/contract/spec007_test.go`

### Implementation for User Story 7

- [ ] T073 [US7] `backend/internal/services/balanceo/umbrales.go`: `EvaluarEstacion` (crear alerta si cruza umbral y no hay abierta; resolver si vuelve a rango); se registra en `hooks.EstacionCambiada` (T012) vía `init()` de su paquete — sin tocar ficheros de otras specs
- [ ] T074 [US7] `backend/internal/workers/alertas/worker.go`: barrido de toda la red cada 60 s (garantía SC-001 < 1 min) + línea de wiring en `backend/cmd/server/main.go`
- [ ] T075 [US7] `backend/internal/services/balanceo/traslados.go`: valida unidades ≤ disponibles en origen y estaciones operativas; mueve las bicis en transacción; recalcula ambas estaciones y evalúa umbrales
- [ ] T076 [US7] `backend/internal/handlers/balanceo/handler.go` + `backend/internal/routes/spec_007_routes.go` (panel ordenado por urgencia: extremos primero)
- [ ] T077 [P] [US7] `frontend/src/services/api/balanceo.ts`
- [ ] T078 [US7] `frontend/src/pages/balanceo/PanelBalanceo.tsx`: lista de alertas abiertas por urgencia con severidad/fecha y formulario de traslado (origen, destino, unidades) con estado recalculado de ambas estaciones

**Checkpoint**: HU-07 funcional; la red se autorregula con logística informada.

---

## Phase 10: User Story 8 — Analítica de uso y exportación CSV (spec-008, HU-08, P1)

**Goal**: Gráficos por hora/día de la semana y CSV consistente para el gestor municipal (Alberto Copa).

**Independent Test**: Con viajes finalizados de toda la red en varias semanas (fixtures), seleccionar rango, verificar gráficos y descargar CSV con los mismos totales; probar denegación a no gestores — escenario S8.

### Tests for User Story 8

- [ ] T079 [P] [US8] `backend/tests/integration/spec008_agregados_test.go` — agregados por hora (24 casillas) y día de la semana (7) sobre viajes finalizados de toda la red; rango vacío → ceros sin error; rango futuro acotado a hoy
- [ ] T080 [P] [US8] `backend/tests/integration/spec008_csv_test.go` — CSV = exactamente los viajes del rango con los mismos totales que los agregados (FR-006); BOM UTF-8 + comillas RFC 4180; nombres con tildes/ñ sin corrupción; solo cabecera en rango vacío
- [ ] T081 [P] [US8] Test de contrato `GET /analitica/viajes` y `GET /analitica/export` + **403** para roles `usuario`/`operario_*` en `backend/tests/contract/spec008_test.go`

### Implementation for User Story 8

- [ ] T082 [US8] `backend/internal/services/analitica/service.go`: agregados SQL (`date_part` sobre `ix_viajes_fecha`) por rango, devolviendo `por_hora[24]`, `por_dia_semana[7]` y totales
- [ ] T083 [US8] `backend/internal/services/analitica/csv.go`: streaming CSV UTF-8 con BOM, comas, CRLF, columnas `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur` y `Content-Disposition`; usa **el mismo filtro/servicio** que los gráficos
- [ ] T084 [US8] `backend/internal/handlers/analitica/handler.go` + `backend/internal/routes/spec_008_routes.go`
- [ ] T085 [P] [US8] `frontend/src/services/api/analitica.ts`
- [ ] T086 [US8] `frontend/src/components/ChartBarras.tsx`: wrapper Chart.js aislado para barras interactivas
- [ ] T087 [US8] `frontend/src/pages/analitica/PanelAnalitica.tsx` + `frontend/src/pages/analitica/ExportarCSV.tsx`: selector de rango, gráficos hora (24) y día de semana (7) con estado "sin datos en el periodo", y descarga del CSV del mismo rango

**Checkpoint**: las 8 historias están implementadas de forma independiente y funcional.

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Validación global del sistema y endurecimiento (todo el equipo, tras fusionar las specs)

- [ ] T088 [P] Ejecutar los escenarios **S1–S8** de [quickstart.md](./quickstart.md) end-to-end en local (docker compose + seed + backend + frontend) y registrar resultados por spec en la PR correspondiente
- [ ] T089 Re-ejecutar el gate **T1–T5** (reglas críticas, const. I) en local/CI antes de cada fusión de spec — checklist obligatoria de PR
- [ ] T090 [P] Optimización de rendimiento: `EXPLAIN ANALYZE` de las consultas calientes (snapshot de estaciones, historial paginado, agregados de analítica) contra objetivos de `plan.md` (< 5 s mapa, < 3 s historial, < 5 s analítica)
- [ ] T091 [P] Security hardening: rate limiting en `POST /auth/login` y `POST /alquileres/desbloqueo`, cabeceras HTTP (CSP, X-Content-Type-Options), validación de entrada en todos los handlers
- [ ] T092 [P] Actualizar `README.md` de la raíz: arranque (`docker compose up -d`, migrate, seed, server, dev), enlaces a `specs/spec-001…008/` y roles seed
- [ ] T093 Checkpoint de esquema (const. III): confirmar con el equipo que ninguna spec añadió migraciones nuevas tras `0001`; si alguna lo hizo, verificar ordinal acordado y comunicación previa
- [ ] T094 [P] Revisión de constitución por PR (5 principios + reparto de historias) por un integrante distinto del autor de cada spec

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias — puede empezar ya
- **Foundational (Phase 2)**: depende de Setup; **bloquea TODAS las user stories**
- **User Stories (Phases 3–10)**: todas dependen de la fase fundacional; después **son paralelas entre sí** (ramas `spec-001…008`, principio V)
- **Polish (Phase 11)**: depende de las specs deseadas fusionadas; T088–T094 escalan con cada fusión

### User Story Dependencies (independencia verificable)

- **US1 (spec-001, Iker)**: arranca tras Phase 2; sin dependencias de otras stories (el estado inicial lo aporta el seed)
- **US2 (spec-002, Iker)**: arranca tras Phase 2; crea viajes que US3 cerrará, pero US3 **no** la requiere fusionada (sus tests usan fixtures SQL de viaje `en_curso`)
- **US3 (spec-003, Alex)**: arranca tras Phase 2; independiente vía fixture de viaje activo y anclaje libre
- **US4 (spec-004, Alex)**: arranca tras Phase 2; independiente vía fixtures de viajes finalizados
- **US5 (spec-005, Josep)**: arranca tras Phase 2; independiente vía fixture de bicicleta anclada; consume el helper `RecalcularEstacion` (T020)
- **US6 (spec-006, Josep)**: arranca tras Phase 2; independiente vía fixture de incidencia abierta; consume T020
- **US7 (spec-007, Alberto)**: arranca tras Phase 2; independiente vía fixtures de estaciones en extremos de ocupación; se engancha a `hooks.EstacionCambiada` (T012) sin modificar ficheros ajenos
- **US8 (spec-008, Alberto)**: arranca tras Phase 2; independiente vía fixtures de viajes finalizados de toda la red

### Within Each User Story

1. Tests (T1–T5 y contrato) escritos **antes** de la implementación y fallando
2. Servicio (lógica de negocio y transacciones) → handlers + `routes/spec_00N_routes.go` → cliente frontend → páginas → tests frontend
3. Recálculo de estaciones post-commit siempre vía `RecalcularEstacion` (T020) — nunca duplicar cómputos
4. Historia completa y validada (su escenario S_n del quickstart) antes de fusionar

### Parallel Opportunities (estrategia del equipo de 4)

- Fases 1–2 en conjunto (ficheros únicos, coordinados); tras el checkpoint de Phase 2, los 4 integrantes arrancan **en paralelo**: Iker US1→US2, Alex US3→US4, Josep US5→US6, Alberto US7→US8
- Dentro de cada story, las tareas `[P]` (tests, modelos frontend, clientes API) pueden ejecutarse juntas
- Los únicos puntos de coordinación tras la fundación son de **una línea** cada uno: `backend/cmd/server/main.go` (workers de US1/US7) y `frontend/src/routes.tsx` (páginas) — conflictos de merge triviales
- Los paquetes backend de cada story pertenecen a su spec; ningún fichero de dominio se comparte entre ramas

---

## Parallel Example: User Story 2 (Iker, tras completar US1 o en paralelo con ella)

```text
Task: "T033 T1 test abono caducado en backend/tests/integration/spec002_abono_test.go"
Task: "T034 T2 test concurrencia en backend/tests/integration/spec002_concurrencia_test.go"
Task: "T035 contrato en backend/tests/contract/spec002_test.go"
# → los tres en paralelo (ficheros distintos); luego:
Task: "T036 services/alquiler/service.go (transacción + FOR UPDATE)"
```

---

## Implementation Strategy

### MVP First (US1 solamente)

1. Completar Phase 1: Setup
2. Completar Phase 2: Foundational (**CRÍTICO — bloquea todas las stories**)
3. Completar Phase 3 (US1, spec-001)
4. **PARAR y VALIDAR**: escenario S1 del quickstart + frescura ≤ 30 s medido
5. Demo/despliegue si está listo (mapa en vivo = valor demostrable)

### Incremental Delivery (por spec y rama)

1. Setup + Foundational → base lista
2. + US1 → S1 → PR revisada → fusión (MVP)
3. + US2 → S2 → PR → fusión … hasta US8 → S8
4. Cada historia añade valor sin romper las anteriores (fixtures SQL para no esperar ramas ajenas)

### Parallel Team Strategy (4 integrantes)

1. El equipo completo completa Setup + Foundational (o Iker lidera la fundación en spec-001 y el resto espera el checkpoint — es corto)
2. Tras el checkpoint: Iker US1/US2 · Alex US3/US4 · Josep US5/US6 · Alberto US7/US8, en ramas propias
3. Cada historia se implementa, prueba y fusiona por PR revisada por otro integrante, en orden P1

---

## Notes

- `[P]` = ficheros distintos, sin dependencias → paralelo seguro
- `[Story]` = trazabilidad US_n ↔ HU_n ↔ spec-00N ↔ integrante
- Gate de calidad: los tests T1–T5 (const. I) son **obligatorios** en las fases 4, 5, 7 y 9 y se re-ejecutan en T089
- Esquema compartido: 0001 es la única migración aportada; cambios futuros exigen ordinal acordado (T093, const. III)
- Commit tras cada tarea o grupo lógico; parar en cada Checkpoint para validar la historia de forma independiente
- Evitar: tareas vagas, conflictos de fichero entre stories, dependencias que rompan la independencia (el hook `hooks/estacion.go` + helpers T011/T020 existen precisamente para evitarlas)
