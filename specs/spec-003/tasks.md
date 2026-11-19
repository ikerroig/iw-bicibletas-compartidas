# Tasks: Devolución de bicicleta y cálculo de tarificación (HU-03)

**Input**: Design documents from `/specs/spec-003/` — [plan.md](./plan.md), [spec.md](./spec.md) (1 user story, P1), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/devoluciones.md](./contracts/devoluciones.md), [quickstart.md](./quickstart.md)

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: **INCLUIDOS** — el gate de calidad de la constitución (Acuerdos de Colaboración) exige pruebas automatizadas de las reglas críticas antes de fusionar; T3a–T3j de [quickstart.md](./quickstart.md) las define.

**Organization**: La HU-03 tiene **una única user story** (US1, P1) que es también el MVP de esta spec. Tareas agrupadas en Setup → Foundational → US1 → Polish.

**Base de código**: el repositorio aún no contiene `backend/` ni `frontend/`; las fases 1–2 arrancan el monorepo según el [plan global](../spec-001/plan.md). Si al fusionar ya existen estos cimientos (de la rama que los aporte), los tasks de esas capas se marcan como ya cumplidos y se reutilizan en lugar de duplicarlos.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Inicializar las dos raíces de código y la configuración de entorno

- [X] T001 [P] Initialize Go module `si16/backend` in `backend/go.mod` with deps: gin v1.10, pgx/v5, pgxpool, go-redis/v9, golang-jwt/v5, golang.org/x/crypto (bcrypt), github.com/pressly/goose/v3
- [X] T002 [P] Scaffold SolidJS + Vite 6 + Tailwind app in `frontend/` (`frontend/package.json`, `frontend/vite.config.ts`, `frontend/tailwind.config.js`, `frontend/index.html`, `frontend/src/index.tsx`)
- [X] T003 [P] Create `/.env.example` at repo root with `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `TARIFA_MINUTO=0,20` (requisito de `docker-compose.yml` y de la tarificación HU-03)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Infraestructura base SIN la cual la US1 no puede implementarse ni probarse de forma independiente (constitución V)

**?? CRITICAL**: Ninguna tarea de la US1 empieza hasta completar esta fase

- [X] T004 Create env config loader in `backend/internal/config/config.go` (lee `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `TARIFA_MINUTO` — este último como `float64`, default 0.20)
- [X] T005 [P] Create pgx pool + transaction helper in `backend/internal/db/db.go` (pool global y `WithTx(ctx, fn)` que hace BEGIN/COMMIT/ROLLBACK — la US1 exige una única transacción, research D2)
- [X] T006 [P] Create shared schema migration in `backend/migrations/0001_esquema_base.sql` tomando el DDL de [data-model de spec-001](../spec-001/data-model.md) (tablas `usuarios`, `estaciones`, `bicicletas`, `anclajes`, `viajes` + índices `ux_viajes_usuario_activo`, `ux_viajes_bicicleta_activa`, `ix_anclajes_estacion`) — **sin adiciones**: la HU-03 no cambia esquema; si el fichero ya existe al fusionar, NO se toca (propiedad spec-001, principio III)
- [X] T007 Create migration runner in `backend/cmd/migrate/main.go` (goose `up` contra `DATABASE_URL`)
- [X] T008 [P] Create minimal seed in `backend/cmd/seed/main.go`: ≥2 estaciones, anclajes (`AN-00001`… con `codigo_anclaje`), bicicletas (una `disponible` anclada, una `averiada`), usuarios de prueba (uno con abono vigente y otro caducado, rol `usuario`) — necesario para probar la US1 sin dependencias de otras ramas
- [X] T009 [P] Create shared structs in `backend/internal/models/models.go` (`Viaje`, `Anclaje`, `Bicicleta`, `Estacion`, `Usuario` con los campos exactos del modelo compartido)
- [X] T010 [P] Create minimal auth in `backend/internal/auth/` (`jwt.go` con `GenerarToken`/middleware `RequireAuth` + `RequireRole`, endpoint `POST /auth/login` en `backend/internal/auth/login.go`) — base mínima para probar el endpoint autenticado de la US1; si al fusionar ya existe el auth global de spec-001, se elimina este y se reutiliza el suyo
- [X] T011 [P] Create Redis state writer in `backend/internal/cache/estado.go` (reescritura de `estacion:{id}:estado` con campos `bicis_disponibles`, `anclajes_libres`, `ocupacion_pct`, `estado`, `actualizado_en`, TTL 120 s, fórmula R9 — la US1 la llama post-commit, research D4)
- [X] T012 Create HTTP server entrypoint in `backend/cmd/server/main.go` + shared router skeleton in `backend/internal/routes/register.go` (función `Register(rg *gin.RouterGroup, deps Deps)` que hoy solo llama al auth; cada spec añadirá su `RegisterSpec00N` en su propio fichero)
- [X] T013 [P] Create typed fetch client in `frontend/src/services/api.ts` (lee token de `localStorage`, header `Authorization: Bearer`, maneja el envelope `{ error: { codigo, mensaje } }`)

**Checkpoint**: infraestructura lista — `docker compose up -d`, `go run ./cmd/migrate up`, `go run ./cmd/seed`, `go run ./cmd/server` y `npm run dev` arrancan; login devuelve JWT

---

## Phase 3: User Story 1 - Fin de viaje y tarifa (Priority: P1) ? MVP

**Goal**: El usuario fija la bicicleta en un anclaje libre de cualquier estación y el sistema cierra el viaje (hora de llegada, liberación del usuario, bicicleta anclada) tarificando 0,00 € si ≤ 30 min o solo el excedente × `TARIFA_MINUTO` si > 30 min.

**Independent Test**: Con un viaje activo de duración controlada, `POST /devoluciones` con un `codigo_anclaje` libre devuelve `200` con viaje `finalizado` y coste correcto (T3a–T3d); rechaza sin viaje activo/anclaje ocupado sin efectos (T3e, T3g) y solo la primera de dos confirmaciones simultáneas tiene efecto (T3f). Ver [quickstart.md](./quickstart.md).

### Tests for User Story 1 ?? (primero, deben FALLAR antes de implementar)

- [X] T014 [P] [US1] Unit tests for tarifa in `backend/internal/services/devolucion/tarifa_test.go`: T3a 25 min → 0.00; T3c exactamente 1800 s → 0.00; T3d 1801 s × TARIFA 0.20 → 0.20; T3b 45 min → 3.00; segundos negativos → piso 0.00 (spec edge "reloj inconsistente")
- [X] T015 [P] [US1] Integration tests for cierre in `backend/tests/devolucion_test.go` (contra BD del compose): T3e sin viaje activo → 409 `SIN_VIAJE_ACTIVO` y anclaje intacto; T3g anclaje ocupado/inexistente → 409 y viaje sigue `en_curso`; T3f 2 devoluciones concurrentes → exactamente 1 × 200 + 1 × 409 `VIAJE_YA_CERRADO`; T3h bici averiada durante viaje → 200 y bici queda anclada + `averiada`; T3i `fecha_fin < fecha_inicio` forzada → coste ≥ 0; T3j latencia: cronometrar la devolución válida y asertar < 10 s (SC-003)

### Implementation for User Story 1

- [X] T016 [US1] Implement pure tariff calculation in `backend/internal/services/devolucion/tarifa.go`: `CalcularCoste(segundos int64, tarifaMinuto float64) (coste float64)` — ≤ 1800 → 0; si no `ceil((segundos−1800)/60) × tarifaMinuto`, piso 0 (research D1, FR-003/FR-004)
- [X] T017 [US1] Implement closing service in `backend/internal/services/devolucion/service.go`: `CerrarViaje(ctx, usuarioID, codigoAnclaje)` con la transacción única y el orden de research D2 — `FOR UPDATE` anclaje (valida `libre` → V2), `FOR UPDATE` viaje `en_curso` (→ V1), validación de reloj (V4), cálculo con T016, `UPDATE viajes … WHERE estado='en_curso'` (rows=0 → `VIAJE_YA_CERRADO`, V3), `UPDATE anclajes` a `ocupado`, `UPDATE bicicletas` preservando `averiada` (V5), COMMIT; todos los errores con los códigos de research D3
- [X] T018 [US1] Add post-commit Redis invalidation in `backend/internal/services/devolucion/service.go` (tras COMMIT, reescribe `estacion:{id}:estado` de la estación destino vía `cache.ActualizarEstadoEstacion`, research D4; fallo de Redis no revierte el cierre)
- [X] T019 [P] [US1] Implement HTTP handler in `backend/internal/handlers/devolucion/handler.go`: valida `codigo_anclaje` (D6 → 422 `ENTRADA_INVALIDA`), llama al servicio, responde `200` con schema `Viaje` incluyendo `etiqueta: "Tramo gratuito (0,00 €)"` cuando `coste_total = 0` (D5) y el envelope de error correspondiente (D3)
- [X] T020 [US1] Register route in `backend/internal/routes/spec_003_devolucion.go`: `RegisterSpec003(rg *gin.RouterGroup, deps Deps)` montando `POST /devoluciones` con `RequireAuth` + `RequireRole("usuario")` (research D7) y añadir su llamada en `backend/internal/routes/register.go` (único punto de merge compartido)
- [X] T021 [P] [US1] Create API client in `frontend/src/services/api/devolucion.ts`: `devolver(codigoAnclaje)` → `POST /devoluciones` tipado contra el contrato [contracts/devoluciones.md](./contracts/devoluciones.md)
- [X] T022 [US1] Implement page in `frontend/src/pages/devolucion/index.tsx`: formulario de `codigo_anclaje`, estado de envío, pantalla de éxito con duración/coste/etiqueta "Tramo gratuito (0,00 €)", y manejo de los 409/422 mostrando `error.mensaje`; registrar la ruta en `frontend/src/routes.tsx` (segundo punto de merge compartido)
- [X] T023 [P] [US1] Component test in `frontend/src/pages/devolucion/index.test.tsx` (Vitest + @solidjs/testing-library; **stub global de fetch en lugar de MSW**: el único msw del registro local, 3.0.3, marca subpaths de Node como `"browser": null` y es irresoluble en el pipeline jsdom de Vitest 5 + Vite 6 — misma cobertura: éxito ≤ 30 min pinta la etiqueta; 409 `SIN_VIAJE_ACTIVO` y 409 `ANCLAJE_OCUPADO` pintan el mensaje sin cambiar de vista)

**Checkpoint**: US1 completa y testeable en solitario — devolución funcional de extremo a extremo (UI → API → transacción → Redis)

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Verificación de calidad y validación contra el spec

- [X] T024 [P] Run backend test suite in `backend/`: `go vet ./...` y `go test ./...` (unit + integración T3a–T3i) hasta dejarlo en verde
- [X] T025 [P] Run frontend test suite in `frontend/`: `npm test` (Vitest) hasta dejarlo en verde
- [X] T026 Execute manual validation per [quickstart.md](./quickstart.md) §3 (escenario S3): devolución ≤ 30 min → 0,00 €; > 30 min → solo excedente; usuario puede desbloquear de inmediato; `actualizado_en` de la estación destino nuevo en ≤ 30 s (constitución IV); rechazos sin efectos
- [X] T027 Update `specs/spec-003/checklists/requirements.md` marcando cada criterio de aceptación del spec como cubierto (con puntero al test/verificación que lo demuestra)

---

## Phase 5: Refinamiento (fase 4 — analyze 2026-11-19)

**Purpose**: Cerrar los huecos de verificación detectados en
`refinamiento/analyze-inicial.md` (A1, A2, A4, A7): los criterios nuevos del
clarify (FR-008/FR-009) están implementados pero sin pruebas.

- [X] T028 [P] [US1] Test de integración de **permisos** en `backend/tests/devolucion_test.go` (FR-008, AC4): sin sesión → `401 NO_AUTENTICADO`; token con rol `operario_taller` → `403 ROL_NO_PERMITIDO`; en ambos casos el viaje sigue `en_curso` y el anclaje intacto (montaje `httptest` + `routes.Register`, patrón del propio fichero) — *hallazgo A1*
- [X] T029 [P] [US1] Test de integración de **abono caducado** en `backend/tests/devolucion_test.go` (FR-009, AC5): usuario con `abono_activo = false` y `fecha_vencimiento_abono` pasada, con viaje activo → `POST /devoluciones` devuelve `200`, viaje `finalizado` con `coste_total` registrado — *hallazgo A2*
- [X] T030 [P] [US1] Aserto de **redondeo fraccionario > 1 min** en `backend/internal/services/devolucion/tarifa_test.go` (AC6): 46:30 → 17 minutos excedentes × tarifa (no 16) — *hallazgo A7*
- [X] T031 [US1] Ampliar la tabla *Acceptance Coverage* de `specs/spec-003/checklists/requirements.md` con las filas de FR-008, FR-009, AC4–AC6 y el nuevo aserto de T030, apuntando al test que lo demuestra — *hallazgo A4*

**Checkpoint**: `go test ./...` en verde con las 4 tareas nuevas; el
analyze final (`refinamiento/analyze-final.md`) sin hallazgos abiertos

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias — arranca de inmediato
- **Foundational (Phase 2)**: depende de Setup — **bloquea** la US1
- **US1 (Phase 3)**: depende de Foundational; es la única story (MVP = feature completa)
- **Polish (Phase 4)**: depende de la US1 completa
- **Refinamiento (Phase 5)**: documentada en el analyze 2026-11-19; T028–T030 reutilizan el patrón de montaje de T015 y T014, T031 cierra el checklist

### Within User Story 1

- Tests (T014, T015) **primero**, en verde no-verde (deben FALLAR) → luego implementación
- Servicio (T016 → T017 → T018) antes que handler (T019) y ruta (T020)
- Cliente API (T021) y página (T022) pueden avanzar en paralelo con el backend tras definir el contrato (ya definido en contracts/)
- Test de componente (T023) al final de la story

### Parallel Opportunities

- **Phase 1**: T001 ∥ T002 ∥ T003
- **Phase 2**: T005 ∥ T006 ∥ T008 ∥ T009 ∥ T010 ∥ T011 ∥ T013
- **Phase 3**: T014 ∥ T015 (tests) y, tras el contrato: T019 ∥ T021 ∥ T023
- **Phase 4**: T024 ∥ T025

---

## Parallel Example: User Story 1

```text
# Tests (en paralelo, ambos deben fallar primero):
Task: "Unit tests for tarifa in backend/internal/services/devolucion/tarifa_test.go"
Task: "Integration tests for cierre in backend/tests/devolucion_test.go"

# Tras definir contrato (ya existente) — backend y frontend en paralelo:
Task: "Implement HTTP handler in backend/internal/handlers/devolucion/handler.go"
Task: "Create API client in frontend/src/services/api/devolucion.ts"
Task: "Component test in frontend/src/pages/devolucion/index.test.tsx"
```

---

## Implementation Strategy

### MVP First (US1 = feature completa)

1. Completar Phase 1: Setup
2. Completar Phase 2: Foundational (CRITICAL — bloquea la story)
3. Completar Phase 3: US1 con tests primero
4. **PARAR Y VALIDAR**: quickstart §3 (S3 + T3)
5. Phase 4: polish y checklist

### Incremental Delivery

1. Setup + Foundational → infraestructura viva (compose, migrate, seed, server, login)
2. T014–T018 → núcleo de cierre y tarifa verificado por tests
3. T019–T020 → endpoint REST completo
4. T021–T023 → flujo de usuario completo en la UI
5. T024–T027 → gate de calidad y validación

---

## Notes

- [P] = ficheros distintos, sin dependencias entre sí
- La story US1 es el 100 % del alcance de spec-003 (el spec no define más historias)
- Commit tras cada tarea o grupo lógico; la rama `spec-003` nunca se fusiona directamente a `main` (PR revisado por otro integrante)
- Puntos de merge compartidos: `backend/internal/routes/register.go` y `frontend/src/routes.tsx` — una línea cada uno
- No modificar `docker-compose.yml`, `0001_esquema_base.sql` (si existe) ni el OpenAPI global sin acuerdo del equipo (principio III)
