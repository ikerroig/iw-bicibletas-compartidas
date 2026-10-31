# Tasks: Monitorización de desequilibrios y alertas de balanceo de flota (HU-07)

**Input**: Design documents from `/specs/spec-007/` — [plan.md](./plan.md), [spec.md](./spec.md) (1 user story, P1), [research.md](./research.md) (D1–D8), [data-model.md](./data-model.md) (V1–V8), [contracts/balanceo.md](./contracts/balanceo.md), [quickstart.md](./quickstart.md)

**Prerequisites**: plan.md ✅, spec.md ✅ (checklist validado), research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: **INCLUIDOS (obligatorios)** — el gate de calidad de la constitución (Acuerdos de Colaboración) exige pruebas automatizadas de las reglas críticas que cubre esta spec (umbrales estrictos <10 % / >90 % = regla T5 del [quickstart global](../spec-001/quickstart.md), unicidad de alertas, rechazo del traslado inconsistente) antes de fusionar; T7a–T7h de [quickstart.md](./quickstart.md) las definen.

**Organization**: La HU-07 tiene **una única user story** (US1, P1) que es también el MVP de esta spec. Tareas agrupadas en Setup → Foundational → US1 → Polish → **Bloqueadas**.

**Base de código**: la rama `spec-007` ya contiene `backend/` y `frontend/` tras los avances de spec-003/spec-004 (auth JWT con `RequireRole`, `cache.ActualizarEstadoEstacion`, `apierr`, `POST /devoluciones`, `GET /viajes`, páginas devolución/historial). **No recrear nada de eso.** La HU-07 añade solo: `services/balanceo`, `handlers/balanceo`, `workers/alertas`, `routes/spec_007_balanceo.go`, `pages/balanceo/`, `services/api/balanceo.ts` y el seed de estaciones en extremos. Si al ejecutar detectas que algo de la base falta, recrea la tarea y márcala.

**Esquema compartido**: la HU-07 **no necesita migraciones** — usa `alertas_balanceo`, `estaciones`, `anclajes`, `bicicletas` y Redis `estacion:{id}:estado` tal como existen. Las únicas tareas que tocarían el modelo compartido (`traslados_balanceo`, índice de score) están en la **Phase 5: Bloqueadas** y NO se ejecutan sin acuerdo del equipo (const. III).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1)
- **[BLOCKED]**: No ejecutar hasta que el equipo apruebe el cambio de esquema compartido (Phase 5)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirmar la base heredada de spec-004 (ya en esta rama) sin recrearla

- [X] T001 Base monorepo heredada de spec-003/spec-004 verificada en la rama `spec-007`: `backend/` (Go/Gin, auth JWT + `RequireRole`, `apierr`, `cache.ActualizarEstadoEstacion`, migración `0001_esquema_base.sql` con `alertas_balanceo`, `POST /devoluciones`, `GET /viajes`) y `frontend/` (SolidJS + Vite + Tailwind + Vitest, páginas devolución/historial) — comprobar `go test ./...` y `npm test` en verde antes de empezar; nada de este listado se reimplementa
- [X] T002 [P] Entorno de ejecución verificado: `docker compose up -d` (Postgres 16 + Redis 7), `go run ./cmd/migrate up`, `go run ./cmd/seed`, `go run ./cmd/server` (:8080) y `npm install` + `npm run dev` (:5173); login `POST /auth/login` con `logistica@si16.test` → `200` (usuario ya existente en el seed)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Datos de demostración y structs compartidos sin los cuales la US1 no puede validarse ni compilarse

**⚠️ CRITICAL**: La US1 puede escribir sus tests en paralelo, pero ni el escenario manual S7 ni la integración se ejecutan hasta completar esta fase

- [X] T003 [P] Ampliar `backend/cmd/seed/main.go` **añadiendo** (sin tocar lo existente: E1/E2, viajes de Ana, usuarios) dos estaciones con sus anclajes y bicicletas — **E-vacía**: capacidad 20 con 1 bici disponible anclada → 5 %; **E-llena**: capacidad 20 con 19 bicis disponibles ancladas → 95 %; insertadas **después** de las actuales (ids ≥ 3) para no alterar los ids de los fixtures HU-03/HU-04; cumplir CHECKs de `anclajes` (ocupado ⇒ bicicleta; libre ⇒ sin bicicleta) — research D8
- [X] T004 [P] Añadir a `backend/internal/models/models.go` los structs `AlertaBalanceo` (campos `ID`, `EstacionID`, `TipoAlerta`, `PorcentajeOcupacion` NUMERIC→float64, `Severidad`, `FechaGeneracion`, `Resuelta` + `EstacionNombre` resuelto por JOIN) y `TrasladoRequest` (`EstacionOrigenID`, `EstacionDestinoID`, `Unidades int`), espejo JSON de los schemas globales `AlertaBalanceo` y `TrasladoRequest` de `specs/spec-001/contracts/openapi.yaml` — data-model.md §2/§4; sin tocar los structs existentes (Viaje, Estacion…)

**Checkpoint**: con el seed ampliado y los structs, `GET /balanceo/alertas` y el traslado tendrán datos y tipos para S7 y los tests

---

## Phase 3: User Story 1 - "Panel de alertas de balanceo para el operario de calle" (Priority: P1) 🎯 MVP

**Goal**: El sistema vigila la ocupación de cada estación (fórmula R9, averiadas excluidas, umbrales estrictos <10 % / >90 %), genera alertas de prioridad alta sin duplicados, las resuelve al volver a rango `[10, 90]`, ofrece al operario de logística el panel ordenado por urgencia y le permite confirmar traslados de lote que recalculan ambas estaciones de inmediato (respuesta + Redis post-commit).

**Independent Test**: Con las estaciones en extremos del seed (T003) y fixtures propias en los tests de integración: (a) mover ocupaciones por debajo del 10 % y por encima del 90 % → alertas altas generadas (inline y por worker < 60 s), 0 duplicados por (estación, umbral); (b) `GET /api/v1/balanceo/alertas` → lista ordenada por urgencia, solo abiertas, 403/401 según rol; (c) `POST /api/v1/balanceo/traslados` con lote factible → `200` con ambos estados recalculados en la respuesta y Redis actualizado; lote inconsistente → `409` sin alterar nada. Ver [quickstart.md](./quickstart.md) T7a–T7h y S7.

### Tests for User Story 1 ⚠️ (primero, deben FALLAR antes de implementar)

- [X] T005 [P] [US1] Unit tests del servicio de umbrales en `backend/internal/services/balanceo/umbrales_test.go` (cómputo puro sin BD o con helpers mockeados): 9,99 % → `vacia_menos_10`; **10,00 % y 90,00 % → sin alerta** (umbrales estrictos); 90,01 % → `llena_mas_90`; estación `fuera_de_servicio` excluida; bici `averiada`/`en_taller` fuera del cómputo y bici `disponible` sin anclaje no cuenta (FR-008); score de urgencia: 0 % ordena antes que 9 %, 100 % antes que 91 % — data-model.md §1/§3 + research D2/D3 (regla T5 constitución I)
- [X] T006 [P] [US1] Unit tests de validación de traslado en `backend/internal/services/balanceo/traslados_test.go`: `unidades < 1` → `422 ENTRADA_INVALIDA`; origen == destino → `422`; JSON/estructura inválida → `422`; ninguna de estas validaciones altera datos — contracts/balanceo.md §2 + data-model.md V6
- [X] T007 [P] [US1] Tests de integración en `backend/tests/balanceo_test.go` con fixture propia (`TRUNCATE … RESTART IDENTITY CASCADE` + estaciones en extremos + operario `operario_logistica` + usuario distractor, montando `routes.Register` con `httptest` — patrón de `tests/historial_test.go`): T7a umbrales estrictos con INSERT/UPDATE de anclajes; T7b unicidad (evaluar 2× → 1 alerta) y resolución automática al volver a `[10, 90]` + reapertura como fila nueva; T7c exclusiones (`fuera_de_servicio`, solo averiadas ancladas); T7d panel (orden por urgencia, `estacion_nombre`, `items: []` con red equilibrada); T7e roles (401 sin token, 403 en `usuario`/`gestor`/`operario_taller`); T7f traslado válido → 200 con `origen`/`destino` recalculados + Redis post-commit + alertas inline; T7g traslado inconsistente (999 unidades, destino sin huecos, estación inexistente/fuera de servicio) → 409 con recuentos idénticos antes/después; T7h latencia del worker (barrido completo con el volumen de estaciones del proyecto cronometrado **< 5 s** + latencia mutación→alerta < periodo de test + 5 s; la garantía de SC-001 depende de cumplir ese límite medido) — quickstart.md §2
- [X] T008 [P] [US1] Test de la página en `frontend/src/pages/balanceo/index.test.tsx` (Vitest + `@solidjs/testing-library` + `vi.stubGlobal("fetch", …)` — patrón de `pages/viajes/index.test.tsx`, NO MSW): render del panel con alertas → filas en el orden del backend con etiqueta "Vacía (< 10 %)"/"Llena (> 90 %)", porcentaje `es-ES` y fecha; estado vacío "Sin estaciones descompensadas"; error de red → `role="alert"`; envío del formulario de traslado → POST con `estacion_origen_id`/`estacion_destino_id`/`unidades` y pintado de las dos tarjetas `EstacionEstado` de la respuesta; recarga del panel tras el 200 — contracts/balanceo.md + research D7

### Implementation for User Story 1

- [X] T009 [US1] Implementar el servicio de umbrales `backend/internal/services/balanceo/umbrales.go`: `New(pool)` + `EvaluarEstacion(ctx, estacionID)`: consulta R9 (capacidad/estado + bicis `disponible` ancladas), creación con `INSERT … ON CONFLICT (estacion_id, tipo_alerta) WHERE NOT resuelta DO NOTHING`, resolución (`UPDATE … SET resuelta = TRUE WHERE estacion_id = $1 AND NOT resuelta`) y refresco del snapshot `porcentaje_ocupacion` de las alertas abiertas; sin acción para estaciones `fuera_de_servicio` — data-model.md §1/§2 + research D1/D2 (depende de T004, T005)
- [X] T010 [US1] Implementar el worker `backend/internal/workers/alertas/worker.go`: `Run(ctx, pool)` en goroutine con evaluación **inmediata al arrancar** + ticker de **55 s** (constante `periodoBarrido`, inyectable en tests); por cada estación `operativa` llama a `EvaluarEstacion` (una estación con error no detiene el barrido del resto; log mínimo); registrar la duración de cada barrido con **warning si ≥ 5 s** (observabilidad del presupuesto, no garantía de latencia); sin dependencias nuevas (solo pgx) — research D1/D6 (depende de T009)
- [X] T011 [US1] Implementar el servicio de traslado `backend/internal/services/balanceo/traslados.go`: `ConfirmarTraslado(ctx, req)` en **una transacción**: validación de entrada (unidades ≥ 1, origen ≠ destino → `apierr.New(422, "ENTRADA_INVALIDA", …)`), `SELECT … FROM estaciones WHERE id = ANY($1) ORDER BY id FOR UPDATE`, validación de estado/lote (existentes, `operativa`, unidades ≤ bicis disponibles en origen y ≤ anclajes libres en destino → `apierr.New(409, "ESTACION_SIN_CAPACIDAD", …)` con mensaje que distingue la causa), movimiento de anclajes (origen → `libre`/NULL; destino → `ocupado`/bicicleta; bicis conservan `disponible`); **post-commit**: `cache.ActualizarEstadoEstacion` de ambas + `EvaluarEstacion` de ambas; devuelve los `EstacionEstado` recalculados — data-model.md §4 + research D4 (depende de T006, T009)
- [X] T012 [US1] Implementar los handlers `backend/internal/handlers/balanceo/handler.go` (`HandlerPanel(d deps.Deps)` y `HandlerTraslado(d deps.Deps)`): panel → `200 { items: [...] }` con la consulta ordenada por urgencia (`CASE` score DESC, `fecha_generacion` ASC) y JOIN de `estacion_nombre`; traslado → parseo `TrasladoRequest`, mapeo de `*apierr.Error` al status/código correspondiente, `200 { origen, destino }` — contracts/balanceo.md + data-model.md §3 (depende de T009, T011)
- [X] T013 [US1] Registrar las rutas: crear `backend/internal/routes/spec_007_balanceo.go` con `RegisterSpec007(rg *gin.RouterGroup, d deps.Deps)` → `rg.GET("/balanceo/alertas", auth.RequireAuth(d.JWTSecret), auth.RequireRole("operario_logistica"), handlers.Panel(d))` y `rg.POST("/balanceo/traslados", …)` (mismo encadenamiento de middlewares), y añadir **una línea** `RegisterSpec007(rg, d)` en `backend/internal/routes/register.go` (fichero compartido, sin modificar las líneas existentes) — contracts/balanceo.md §autorización + research D5 (depende de T012)
- [X] T014 [US1] Arrancar el worker: añadir **una línea** en `backend/cmd/server/main.go` — `go workers_alertas.Run(ctx, pool)` tras crear el pool (context cancelable con el apagado; fichero compartido, sin tocar el resto) — research D6 (depende de T010)
- [X] T015 [P] [US1] Crear el cliente `frontend/src/services/api/balanceo.ts`: interfaces `AlertaBalanceo` y `TrasladoResultado` (con `origen`/`destino` del schema `EstacionEstado`), funciones `listarAlertas()` y `confirmarTraslado({ estacion_origen_id, estacion_destino_id, unidades })` usando `api.request` de `frontend/src/services/api.ts` (patrón de `services/api/viajes.ts`) — contracts/balanceo.md
- [X] T016 [US1] Implementar la página `frontend/src/pages/balanceo/index.tsx`: señales (`alertas`, `cargando`, `error`, `origen`, `destino`, `unidades`, `resultado`, `enviando`); panel que pinta los items **tal cual llegan** (orden del backend — const. I) con etiqueta de `tipo_alerta`, porcentaje `es-ES`, severidad y fecha Europe/Madrid; estados vacío ("Sin estaciones descompensadas — la red está equilibrada") y error `role="alert"`; botón "Actualizar"; formulario de traslado (ids + unidades ≥ 1) que tras el `200` muestra las dos tarjetas de `EstacionEstado` y recarga el panel; Tailwind + `data-testid` (patrón de `pages/viajes/index.tsx`) — research D7 (depende de T008, T015)
- [X] T017 [US1] Publicar la página: añadir la entrada `{ path: "/balanceo", component: Balanceo }` en `frontend/src/routes.tsx` y el enlace `<A href="/balanceo">Balanceo</A>` en el nav de `frontend/src/index.tsx` (ficheros compartidos, una entrada cada uno, sin modificar las existentes) (depende de T016)

**Checkpoint**: US1 completa y testeable — `go test ./...` y `npm test` en verde; escenario manual S7 operativo con el seed de T003

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Gate de calidad constitución I, verificación de SC y documentación de cierre

- [X] T018 [P] Gate de calidad constitución I + estática: re-ejecutar T7a/T7b (umbrales estrictos, unicidad, resolución) y T7g (traslado rechazado sin efectos) en `go test ./...`; `go vet ./...`, `gofmt -l` (solo cuentan avisos reales, no CRLF del working tree), `npm test` y `npm run build` (typecheck TS)
- [X] T019 Ejecutar el escenario manual S7 y la validación SC-001…SC-004 de [quickstart.md](./quickstart.md) contra `localhost:5173` + API (login `logistica@si16.test`, panel ordenado, traslado 4→3, resolución automática, 409/422, roles 403, `HGETALL estacion:4:estado`); registrar resultado en la sección 5 de [quickstart.md](./quickstart.md)
- [X] T020 [P] Revisar coherencia final spec ↔ código: barrido de FR-001…FR-008 y SC-001…SC-004 contra lo implementado (entrada de `/speckit.converge`), verificar que **no** se tocó `specs/spec-004/` ni código de otras specs, y actualizar los checkboxes de este `tasks.md`

---

## Phase 5: Bloqueadas — pendientes de acuerdo del equipo (constitución III)

**⚠️ NO EJECUTAR** sin acuerdo previo del equipo: estas tareas crearían una migración nueva (`000N_spec-007_*.sql` con ordinal acordado, R6) sobre el **esquema PostgreSQL compartido** por las 8 specs. El MVP de la HU-07 está completo y validado sin ellas (Phase 3–4); se documentaron como `⚠ PENDIENTE DE ACUERDO DEL EQUIPO` en [data-model.md](./data-model.md).

- [ ] T021 [BLOCKED] [US1] *(requiere acuerdo de equipo)* Crear migración `backend/migrations/000N_spec-007_traslados.sql` con la tabla opcional de auditoría `traslados_balanceo` (FK estaciones origen/destino, `unidades`, `fecha`, `operario_id` FK usuarios) **solo si** el equipo aprueba la propuesta de data-model.md § "PENDIENTE": afectaría a `cmd/seed` y al INSERT dentro de `ConfirmarTraslado` (`backend/internal/services/balanceo/traslados.go`)
- [ ] T022 [BLOCKED] *(requiere acuerdo de equipo)* Materializar el score de urgencia del panel (columna generada o índice expression sobre `alertas_balanceo`) en `backend/migrations/000N_spec-007_score.sql` **solo si** el equipo lo aprueba; ajustaría la consulta de orden de `backend/internal/handlers/balanceo/handler.go` — data-model.md § "PENDIENTE" punto 2

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: heredada de spec-004, verificada
- **Foundational (Phase 2)**: T003 y T004 en paralelo; **T003 bloquea el escenario manual S7** y la integración con datos reales
- **User Story 1 (Phase 3)**: requiere Setup ✅; los tests (T005–T008) se escriben ANTES que la implementación; T013 depende de T012, T014 depende de T010
- **Polish (Phase 4)**: depende de toda la US1
- **Bloqueadas (Phase 5)**: sin dependencia de ejecución — **solo** de la aprobación del equipo; nunca bloquean el MVP

### Within User Story 1

- T005, T006, T007, T008 son paralelizables (archivos distintos) y deben FALLAR antes de implementar
- Backend: T009 → T010 y T009 → T011 → T012 → T013 (cadena de dependencia de compilación); T014 en paralelo con T011–T013 tras T010
- Frontend: T015 → T016 → T017 (cadena); T015 en paralelo con la cadena backend
- Backend y frontend avanzan en paralelo tras los tests

### Parallel Example: User Story 1

```text
# Arranque (4 archivos distintos, en paralelo):
T005  unit umbrales    backend/internal/services/balanceo/umbrales_test.go
T006  unit traslados   backend/internal/services/balanceo/traslados_test.go
T007  integración      backend/tests/balanceo_test.go
T008  test página      frontend/src/pages/balanceo/index.test.tsx

# Implementación (dos cadenas en paralelo):
Backend:  T009 → T010 → T011 → T012 → T013   (T014 tras T010, en paralelo con T011–T013)
Frontend: T015 → T016 → T017
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Setup (heredada ✅) + Foundational (T003 seed extremos, T004 models)
2. Tests T005–T008 → verlos FALLAR
3. Backend T009–T014 → `go test ./...` en verde
4. Frontend T015–T017 → `npm test` en verde
5. **STOP y VALIDAR**: escenario S7 del quickstart
6. Polish T018–T020
7. Phase 5 queda **aparcada** hasta el acuerdo del equipo (no forma parte del MVP)

## Notes

- [P] tasks = different files, no dependencies
- Cada entrada en `routes/register.go`, `cmd/server/main.go`, `frontend/src/routes.tsx` y `frontend/src/index.tsx` es de **una línea/entrada** para que los merges con otras ramas sigan siendo triviales (const. V)
- No migraciones, no claves Redis nuevas, no dependencias npm/go nuevas en las fases 1–4
- El worker cubre las mutaciones de otras specs (devoluciones ya presentes; desbloqueos/averías/taller futuros) **sin modificar su código** — la HU-07 es independiente (const. V)
- Commits por fase (convención del repo): `docs: spec-007 …` / `feat: spec-007 …`
