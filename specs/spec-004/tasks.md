# Tasks: Consulta de historial de viajes y desglose mensual (HU-04)

**Input**: Design documents from `/specs/spec-004/` — [plan.md](./plan.md), [spec.md](./spec.md) (1 user story, P1), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/viajes.md](./contracts/viajes.md), [quickstart.md](./quickstart.md)

**Prerequisites**: plan.md ✅, spec.md ✅ (5 clarificaciones integradas), research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: **INCLUIDOS** — el gate de calidad de la constitución (Acuerdos de Colaboración) exige pruebas automatizadas de las reglas críticas que cubre esta spec (coste 0,00 € etiquetado como "Tramo gratuito", sumatorio exacto, aislamiento por usuario) antes de fusionar; T4a–T4j de [quickstart.md](./quickstart.md) las define.

**Organization**: La HU-04 tiene **una única user story** (US1, P1) que es también el MVP de esta spec. Tareas agrupadas en Setup → Foundational → US1 → Polish.

**Base de código**: la rama `spec-004` ya contiene `backend/` y `frontend/` tras fusionar `spec-003` (verificado en verde el 2026-10-28), por lo que las fases de inicialización están heredadas y se marcan cumplidas. Si al ejecutar detectas que algo falta, recrea la tarea y márcala.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirmar la base heredada de spec-003 (ya fusionada en esta rama)

- [X] T001 Base monorepo heredada de spec-003: `backend/` (Go/Gin, auth JWT, migraciones, `POST /devoluciones`) y `frontend/` (SolidJS + Vite + Tailwind + Vitest) en la rama `spec-004` tras la fusión fast-forward `3997ca8` — `go test ./...` y `npm test` en verde (2026-10-28)
- [X] T002 [P] Entorno de ejecución verificado: `docker compose up -d` (Postgres 16 + Redis 7), `go run ./cmd/migrate up`, `go run ./cmd/seed`, `go run ./cmd/server` (:8080) y `npm install` + `npm run dev` (:5173); login `POST /auth/login` → `200` (2026-10-28)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Datos de demostración sin los cuales la US1 no puede validarse manualmente (quickstart S4)

**⚠️ CRITICAL**: La US1 puede implementarse en paralelo, pero el escenario manual S4 no se ejecuta hasta completar esta fase

- [X] T003 [P] Ampliar `backend/cmd/seed/main.go` con viajes **finalizados** de prueba para `ana@si16.test` repartidos en ≥ 2 meses (alguno con `coste_total = 0.00` para la etiqueta, otros con coste > 0), respetando los CHECKs (`fecha_fin`, `estacion_destino_id`, `duracion_minutos` NOT NULL en `finalizado`) y **sin tocar** el viaje `en_curso` existente ni los usuarios — research D8

**Checkpoint**: con el seed ampliado, `GET /viajes` y la página "Mis Viajes" tendrán datos reales para el escenario manual

---

## Phase 3: User Story 1 - "Mis Viajes" con filtros de periodo (Priority: P1) 🎯 MVP

**Goal**: El usuario ve su historial cronológico de viajes finalizados (fecha/hora, estaciones por nombre, duración, coste), filtra por mes/año, ve el sumatorio del periodo y la etiqueta "Tramo gratuito (0,00 €)" en los viajes de coste 0, con paginación server-side.

**Independent Test**: Con un usuario con viajes repartidos en varios meses (fixtures), `GET /api/v1/viajes` devuelve la lista cronológica con campos correctos y solo lo propio; el filtro `desde`/`hasta` devuelve exactamente ese mes con su sumatorio; los viajes de 0,00 € llevan etiqueta; paginación no altera el sumatorio. En la UI (`/viajes`), los mismos checks con filtros y estados vacíos. Ver [quickstart.md](./quickstart.md) T4a–T4j y S4.

### Tests for User Story 1 ⚠️ (primero, deben FALLAR antes de implementar)

- [X] T004 [P] [US1] Unit tests del servicio en `backend/internal/services/viajes/service_test.go`: derivación de `etiqueta = "Tramo gratuito (0,00 €)"` cuando `coste_total = 0` y ausencia cuando > 0; cálculo `offset = (page−1)×page_size`; validación de `page ≥ 1`, `page_size ∈ [1,200]`, `desde ≤ hasta` y formato `YYYY-MM-DD` (research D6)
- [X] T005 [P] [US1] Tests de integración en `backend/tests/historial_test.go` con fixture propia (`TRUNCATE … RESTART IDENTITY CASCADE` + estaciones + `ana` con viajes finalizados en 2 meses, uno con coste 0,00, uno `en_curso` y viajes de otro usuario), montando `routes.Register` con `httptest` (patrón de `tests/devolucion_test.go`): T4a lista/orden/campos/nombres, T4b filtro de mes con suma exacta, T4c etiqueta, T4d aislamiento + 401, T4e paginación con suma idéntica entre páginas, T4f página fuera de rango, T4g timezone (día 31 a las 23:30 Europe/Madrid), T4h periodo vacío, T4i `422 ENTRADA_INVALIDA`, T4j latencia < 3 s
- [X] T006 [P] [US1] Test de la página en `frontend/src/pages/viajes/index.test.tsx` (Vitest + `@solidjs/testing-library` + `vi.stubGlobal("fetch", …)` — patrón de `pages/devolucion/index.test.tsx`, NO MSW): render con datos → filas con fecha `dd/mm/aaaa HH:MM`, nombres de estación, duración, coste, etiqueta y sumatorio; cambio de filtro → petición con `desde`/`hasta` correctos; estado vacío "sin viajes en el periodo"; error de red → mensaje con `role="alert"`; siguiente página → `page=2`

### Implementation for User Story 1

- [X] T007 [US1] Implementar el servicio `backend/internal/services/viajes/service.go` (`New(pool)` + `ListarHistorial(ctx, usuarioID, desde, hasta, page, pageSize)`): Q1 página con JOIN a `estaciones` y `ORDER BY fecha_inicio DESC LIMIT/OFFSET`, Q2 agregado `count`/`sum` del periodo sin paginar, filtro `(fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN`, etiqueta en el DTO — data-model.md + research D1/D2/D3 (depende de T004)
- [X] T008 [US1] Implementar el handler `backend/internal/handlers/viajes/handler.go` (`Handler(d deps.Deps) gin.HandlerFunc`): parseo y validación de `desde`/`hasta`/`page`/`page_size` (defaults 1/50; inválidos → `apierr` 422 `ENTRADA_INVALIDA`), extracción de usuario con `auth.UsuarioID(c)`, respuesta `200 { total, suma_coste, items }` — contracts/viajes.md (depende de T007)
- [X] T009 [US1] Registrar la ruta: crear `backend/internal/routes/spec_004_viajes.go` con `RegisterSpec004(rg *gin.RouterGroup, d deps.Deps)` → `rg.GET("/viajes", auth.RequireAuth(d.JWTSecret), viajes.Handler(d))` y añadir **una línea** `RegisterSpec004(rg, d)` en `backend/internal/routes/register.go` (fichero compartido) (depende de T008)
- [X] T010 [P] [US1] Crear el cliente `frontend/src/services/api/viajes.ts`: interfaces `ViajeHistorial` y `HistorialResponse`, función `listarViajes({desde, hasta, page, pageSize})` usando `api.request` de `frontend/src/services/api.ts` — contracts/viajes.md
- [X] T011 [US1] Implementar la página `frontend/src/pages/viajes/index.tsx`: señales (`viajes`, `total`, `suma`, `mes`, `anio`, `pagina`, `cargando`, `error`), selects de mes (1–12 libre) y año (`actual−1 … actual+1`, research D7) que construyen `desde`/`hasta` (primer/último día), lista cronológica con `data-testid`, sumatorio `suma_coste` formateado `es-ES`, etiqueta tal cual del backend, controles "Anterior/Siguiente" + "Página X de Y", estados vacíos ("sin viajes en el periodo" / "aún no tienes viajes"), formato de fecha con `timeZone: 'Europe/Madrid'`, Tailwind + roles ARIA (depende de T010, T006)
- [X] T012 [US1] Publicar la ruta: añadir la entrada `{ path: "/viajes", component: MisViajes }` en `frontend/src/routes.tsx` y el enlace `<A href="/viajes">Mis viajes</A>` en el nav de `frontend/src/index.tsx` (ficheros compartidos, una entrada cada uno) (depende de T011)

**Checkpoint**: US1 completa y testeable — `go test ./...` y `npm test` en verde; escenario manual S4 operativo con el seed de T003

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Gate de calidad, verificación de SC y documentación de cierre

- [X] T013 [P] Gate de calidad constitución I + estática: re-ejecutar T4c (coste 0,00 € → etiqueta en historial) y T4b/T4e (sumatorio exacto) en `go test ./...`; `go vet ./...`, `gofmt -l` (sin novedades reales: los avisos por CRLF del working tree no cuentan), `npm test` y `npm run build` (typecheck TS)
- [X] T014 Ejecutar el escenario manual S4 y la validación SC-001…SC-004 de [quickstart.md](./quickstart.md) contra `localhost:5173` + API; registrar resultado en las notas de [quickstart.md](./quickstart.md)
- [X] T015 [P] Revisar coherencia final spec ↔ código: barrido de FR-001…FR-008 y SC-001…SC-004 contra lo implementado (entrada de `/speckit.converge`) y actualizar los checkboxes de este `tasks.md`

---

## Phase 5: Refinamiento (fase 4 — analyze 2026-11-19)

**Purpose**: Cerrar los huecos de verificación detectados en
`refinamiento/analyze-inicial.md` (A1–A6): el clarify 2026-11-19 añadió
AC4–AC6 que la implementación no prueba por completo.

- [ ] T016 [US1] Test de integración **AC4/roles** en `backend/tests/historial_test.go` (`TestHistorialT4kRoles`): token `operario_taller` y token `gestor` → `200` con su propio historial (aislamiento por JWT); con `?usuario_id=<id ajeno>` el parámetro se ignora y sigue siendo el historial propio — *hallazgo A1*
- [ ] T017 [US1] Ampliar la fixture de `TestHistorialT4jLatencia` a ~365 viajes finalizados (un año de uso habitual) antes de cronometrar y asertar `< 3 s` — *hallazgo A2 (SC-004 / quickstart T4j)*
- [X] T018 [US1] Quickstart: añadir fila **T4k** (AC4: cualquier rol autenticado ve su propio historial) y mapear AC5→T4a, AC6→T4a+T4e en la tabla SC — *hallazgo A3*
- [ ] T019 [US1] Crear la tabla *Acceptance Coverage* en `specs/spec-004/checklists/requirements.md` (FR/SC/AC → test que lo demuestra, incl. T4k) — *hallazgo A4*
- [X] T020 [US1] plan.md: nota en Unknowns sobre la sesión de clarify 2026-11-19 (C1–C3) remitiendo a `refinamiento/clarify.md` — *hallazgo A5*
- [X] T021 [US1] clarify.md: renombrar decisiones D6–D13 → C1–C8 para eliminar la colisión con la numeración D1–D8 de `research.md` — *hallazgo A6*

**Checkpoint**: `go test ./...` + `npm test` en verde con T016/T017;
analyze final sin hallazgos abiertos

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: heredada de spec-003, completada
- **Foundational (Phase 2)**: sin dependencias técnicas de la US1 (T003 en paralelo posible), pero **bloquea el escenario manual S4**
- **User Story 1 (Phase 3)**: requiere Setup ✅; los tests (T004–T006) se escriben ANTES que la implementación
- **Polish (Phase 4)**: depende de T003 + toda la US1
- **Refinamiento (Phase 5)**: documentada en el analyze 2026-11-19; T016/T017 reutilizan el patrón de montaje de `historial_test.go` (T005) y T019 cierra el checklist

### Within User Story 1

- T004, T005, T006 son paralelizables (archivos distintos) y deben FALLAR antes de implementar
- Backend: T007 → T008 → T009 (cadena de dependencia de compilación)
- Frontend: T010 → T011 → T012 (cadena de dependencia de compilación)
- Backend y frontend avanzan en paralelo tras los tests

### Parallel Example: User Story 1

```text
# Arranque (3 archivos distintos, en paralelo):
T004  tests unitarios  backend/internal/services/viajes/service_test.go
T005  tests integración backend/tests/historial_test.go
T006  test página       frontend/src/pages/viajes/index.test.tsx

# Implementación (dos cadenas en paralelo):
Backend: T007 → T008 → T009
Frontend: T010 → T011 → T012
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Setup (heredado ✅) + T003 seed
2. Tests T004–T006 → verlos FALLAR
3. Backend T007–T009 → `go test ./...` en verde
4. Frontend T010–T012 → `npm test` en verde
5. **STOP y VALIDAR**: escenario S4 del quickstart
6. Polish T013–T015

## Notes

- [P] tasks = different files, no dependencies
- Cada entrada en `routes/register.go`, `frontend/src/routes.tsx` y `frontend/src/index.tsx` es de **una línea/entrada** para que los merges con otras ramas sigan siendo triviales (const. V)
- No migraciones, no claves Redis nuevas, no dependencias npm/go nuevas
- Commits por fase (convención del repo): `docs: spec-004 …` / `feat: spec-004 …`
