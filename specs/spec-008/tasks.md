# Tasks: Panel de analítica de uso y exportación de datos a CSV (HU-08)

**Input**: Design documents from `/specs/spec-008/` — [plan.md](./plan.md), [spec.md](./spec.md) (1 user story, P1), [research.md](./research.md) (D1–D9), [data-model.md](./data-model.md) (V1–V8), [contracts/analitica.md](./contracts/analitica.md), [quickstart.md](./quickstart.md)

**Prerequisites**: plan.md ✅, spec.md ✅ (checklist validado + 3 clarificaciones 2026-11-01), research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: **INCLUIDOS (obligatorios)** — [quickstart.md](./quickstart.md) define T8a–T8i como gate de verificación de esta spec (consistencia gráficos↔CSV FR-006/SC-002, bordes del rango inclusivo, denegación de roles SC-003, formato CSV SC-004); se escriben ANTES de implementar.

**Organization**: La HU-08 tiene **una única user story** (US1, P1) que es también el MVP de esta spec. Tareas agrupadas en Setup → Foundational → US1 → Polish → **Bloqueadas**.

**Base de código**: la rama `spec-008` ya contiene `backend/` y `frontend/` tras los avances de spec-003/spec-004/spec-007 (auth JWT con `RequireRole`, `apierr`, `services/viajes` con filtro Europe/Madrid, `services/balanceo`, `workers/alertas`, páginas devolución/historial/balanceo). **No recrear nada de eso.** La HU-08 añade solo: `services/analitica`, `handlers/analitica`, `routes/spec_008_analitica.go`, `pages/analitica/`, `services/api/analitica.ts`, el struct `AnaliticaViajes` y el lote de viajes de demo en el seed. Si al ejecutar detectas que algo de la base falta, recrea la tarea y márcala.

**Esquema compartido**: la HU-08 **no necesita migraciones** — solo LEE `viajes` y `estaciones` con el índice `ix_viajes_fecha` ya existente; sin claves Redis nuevas. Las tareas que tocan el **OpenAPI global** y **`frontend/package.json`** están en la **Phase 5: Bloqueadas** y NO se ejecutan sin acuerdo del equipo (const. III); los cambios aditivos de models/seed/rutas/nav son tareas revisables en el PR (patrón de spec-007).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1)
- **[BLOCKED]**: No ejecutar hasta que el equipo apruebe el cambio de fichero compartido (Phase 5)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirmar la base heredada de spec-003/spec-004/spec-007 (ya en esta rama) sin recrearla

- [ ] T001 Base monorepo heredada de spec-003/spec-004/spec-007 verificada en la rama `spec-008`: `backend/` (Go/Gin, auth JWT + `RequireAuth`/`RequireRole`, `apierr` con envelope `{ error: { codigo, mensaje } }`, `services/viajes` con filtro `AT TIME ZONE 'Europe/Madrid'`, `services/balanceo`, `workers/alertas`, migración `0001_esquema_base.sql` con `viajes`/`estaciones` e `ix_viajes_fecha`) y `frontend/` (SolidJS + Vite + Tailwind + Vitest, `services/api.ts`, páginas devolución/historial/balanceo) — comprobar `go test ./...` y `npm test` en verde antes de empezar; nada de este listado se reimplementa
- [ ] T002 [P] Entorno de ejecución verificado: `docker compose up -d` (Postgres 16 + Redis 7), `go run ./cmd/migrate up`, `go run ./cmd/seed`, `go run ./cmd/server` (:8080) y `npm install` + `npm run dev` (:5173); login `POST /auth/login` con `gestor@si16.test` → `200` (usuario ya existente en el seed)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Datos de demostración y struct compartido sin los cuales la US1 no puede validarse ni compilarse

**⚠️ CRITICAL**: La US1 puede escribir sus tests en paralelo, pero ni el escenario manual S8 ni la integración con datos reales se ejecutan hasta completar esta fase

- [ ] T003 [P] Ampliar `backend/cmd/seed/main.go` **añadiendo** (sin tocar lo existente: estaciones E1/E2/Nord/Port, viaje activo de Ana, viajes HU-04, usuarios) un lote de viajes **finalizados de demo HU-08**: ~40 viajes de `ana@si16.test` con bicicleta nueva `B-200` (`disponible`, sin anclaje), repartidos deterministamente en las **últimas 4 semanas** con patrón de horas locales `{7, 8, 9, 12, 17, 18, 20, 22}` y distintos días de la semana; duraciones y costes coherentes con la tarifa HU-03 (≤ 30 min → `0.00`; si no, excedente × 0,20); cumplir el CHECK de `finalizado` (destino, fin y duración NOT NULL); insertados **después** de los viajes HU-04 (no alteran sus ids) — research D7
- [ ] T004 [P] Añadir a `backend/internal/models/models.go` el struct `AnaliticaViajes` (campos `Desde`, `Hasta` string `YYYY-MM-DD`; `TotalViajes int`; `TotalCoste float64`; `PorHora` y `PorDiaSemana` slices de enteros con JSON arrays de 24 y 7 posiciones), espejo JSON del schema global `AnaliticaViajes` de `specs/spec-001/contracts/openapi.yaml` — data-model.md §4; **sin tocar los structs existentes** (Viaje, Estacion, AlertaBalanceo, TrasladoRequest, EstacionEstado)

**Checkpoint**: con el seed ampliado y el struct, `GET /analitica/viajes` y la exportación tendrán datos y tipos para S8 y los tests

---

## Phase 3: User Story 1 - "Analítica de la red con exportación" (Priority: P1) 🎯 MVP

**Goal**: El gestor municipal filtra por un rango de fechas (inclusivo, Europe/Madrid, validado server-side → 422 si invertido/inválido) y recibe gráficos interactivos con el volumen de viajes finalizados de toda la red por hora del día (00–23) y día de la semana (lunes→domingo, agrupados por **fecha de inicio**), más la descarga del CSV del **mismo rango** (UTF-8 BOM, comas, RFC 4180, CRLF, columnas `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur`) con totales idénticos a los gráficos por construcción; acceso exclusivo del rol `gestor` (401/403).

**Independent Test**: Con el lote de seed (T003) y fixtures propias en los tests de integración: (a) rango con datos → `GET /api/v1/analitica/viajes` con `por_hora` de 24 posiciones y `por_dia_semana` de 7 (lunes=0…domingo=6), horas/días locales correctos; (b) `GET /api/v1/analitica/export` del mismo rango → CSV cuyas filas suman exactamente `total_viajes` y `total_coste`; (c) roles: sin token → 401, `usuario`/`operario_*` → 403, `gestor` → 200; rango invertido → 422 en ambos; rango vacío/futuro → 200 con ceros y CSV de solo cabecera. Ver [quickstart.md](./quickstart.md) T8a–T8i y S8.

### Tests for User Story 1 ⚠️ (primero, deben FALLAR antes de implementar)

- [ ] T005 [P] [US1] Unit tests del servicio en `backend/internal/services/analitica/service_test.go` (sin BD): `ValidarEntrada` — ausentes/mal formato (no `YYYY-MM-DD` estricto, p. ej. `01/10/2026` o `2026-13-40`) / `desde > hasta` → `422 ENTRADA_INVALIDA`; válidos → parámetros parseados; relleno de buckets inicializados a **0** (24 y 7 posiciones siempre presentes); mapeo de día de semana: domingo → índice **6**, lunes → **0**, sábado → **5** — research D1/D2 + data-model.md V2/V3/V5
- [ ] T006 [P] [US1] Unit tests del escritor CSV en `backend/internal/services/analitica/csv_test.go` (bytes a `bytes.Buffer`): primeros 3 bytes = BOM `EF BB BF`; cabecera exacta `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur`; finales de línea CRLF; separador coma; nombre con ñ/tildes pasa intacto; campo con coma comillado (RFC 4180); `coste_eur` con **punto** decimal y 2 decimales (`3.00`); rango sin filas → solo cabecera — research D4 + contracts/analitica.md §2
- [ ] T007 [P] [US1] Tests de integración en `backend/tests/analitica_test.go` con fixture propia (`TRUNCATE … RESTART IDENTITY CASCADE` + usuarios `gestor`/`usuario`/`operario_logistica` + viajes finalizados con horas y días conocidos, montando `routes.Register` con `httptest` — patrón de `tests/historial_test.go`): T8a consistencia (mismo rango → `total_viajes` == filas de datos del CSV, `total_coste` == suma de `coste_eur`, `por_hora[i]` == filas en la hora `i`); T8b bordes inclusivos (viaje día 15 23:30 Europe/Madrid cuenta con `hasta=15`, no con `hasta=14`; día 1 00:30 cuenta con `desde=1`); T8c agrupación (horas locales 00–23 en `por_hora`; domingo→6, lunes→0, sábado→5 en `por_dia_semana`); T8d viaje `en_curso` excluido; T8e rango vacío y rango futuro → `200` con arrays a 0 + `total_viajes: 0` y CSV **solo cabecera**; T8f `422` en ambos endpoints con fechas ausentes/mal formateadas/invertidas; T8g roles (401 sin token; 403 `usuario`/`operario_taller`/`operario_logistica`; 200 `gestor`); T8h CSV completo (BOM, columnas, CRLF, orden `fecha_inicio` ASC, nombre de cabecera `Content-Disposition: attachment; filename=viajes_AAAAMMDD_AAAAMMDD.csv`); T8i latencia `GET /analitica/viajes` con rango de un año cronometrado **< 5 s** — quickstart.md §2
- [ ] T008 [P] [US1] Test de la página en `frontend/src/pages/analitica/index.test.tsx` (Vitest + `@solidjs/testing-library` + `vi.stubGlobal("fetch", …)` — patrón de `pages/viajes/index.test.tsx`, NO MSW; **mockear también el wrapper de gráfico** `GraficoBarras` para no requerir chart.js en Vitest): render con datos → dos gráficos con 24/7 etiquetas y totales visibles; `total_viajes: 0` → aviso **"sin datos en el periodo"**; click en "Descargar CSV" → `fetch` con `Authorization` y descarga vía Blob/`<a download>`; doble clic en "Descargar CSV" → cada acción produce **un único archivo consistente** (el botón está deshabilitado mientras `descargando`, evitando envíos superpuestos); respuesta `403` → mensaje con `role="alert"`; `422` → mensaje de rango inválido — contracts/analitica.md + research D6

### Implementation for User Story 1

- [ ] T009 [US1] Implementar el servicio `backend/internal/services/analitica/service.go`: constante `zonaHoraria = "Europe/Madrid"`; `ValidarEntrada(desde, hasta string)` con parseo estricto `2006-01-02` (copiar el criterio de `services/viajes.ParsearFechaParam` **sin importar ni modificar** ese paquete — independencia por spec, const. V) y `desde ≤ hasta` → `apierr.New(422, "ENTRADA_INVALIDA", …)`; `construirFiltro(p)` → **un único** `WHERE v.estado = 'finalizado' AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $1 AND $2` con args compartidos (FR-006); `Agregados(ctx, p)` → Q1 `EXTRACT(hour …)` GROUP BY, Q2 `((EXTRACT(dow …)::int + 6) % 7)` GROUP BY, Q3 `count(*) + COALESCE(sum(coste_total),0)`, zero-fill de arrays 24/7 en Go y `total_viajes` = suma de `por_hora` — data-model.md §1/§2 + research D1–D3 (depende de T004)
- [ ] T010 [US1] Implementar la exportación `backend/internal/services/analitica/csv.go` (mismo paquete, reutiliza `construirFiltro` de T009): `ExportarCSV(ctx, p, w io.Writer)` escribe a mano el BOM `EF BB BF`, crea `csv.Writer` con `UseCRLF = true` y coma por defecto, emite la cabecera de 6 columnas, itera la consulta streaming (`ORDER BY v.fecha_inicio ASC, v.id ASC`, JOIN `estaciones` origen + destino, `to_char(… 'YYYY-MM-DD HH24:MI')` para `fecha_hora`, `%.2f` para `coste_eur`) y cierra/`Flush` con manejo de error; helpers `NombreFichero(desde, hasta)` → `viajes_AAAAMMDD_AAAAMMDD.csv` — data-model.md §3 + research D4 (depende de T009)
- [ ] T011 [US1] Implementar los handlers `backend/internal/handlers/analitica/handler.go` (`HandlerAgregados(d deps.Deps)` y `HandlerExport(d deps.Deps)`): agregados → lee `desde`/`hasta` de la query, `ValidarEntrada` (→ `422`), `Agregados`, `200` con `models.AnaliticaViajes` eco de fechas; export → `ValidarEntrada` **antes** de escribir cabeceras, luego `Content-Type: text/csv; charset=utf-8` + `Content-Disposition: attachment; filename=…` y streaming de `ExportarCSV`; en error se responde el envelope JSON global **sin** cabeceras de descarga; mapear `*apierr.Error` al status/código — contracts/analitica.md §1/§2 (depende de T009, T010)
- [ ] T012 [US1] Registrar las rutas: crear `backend/internal/routes/spec_008_analitica.go` con `RegisterSpec008(rg *gin.RouterGroup, d deps.Deps)` → `rg.GET("/analitica/viajes", auth.RequireAuth(d.JWTSecret), auth.RequireRole("gestor"), analitica.HandlerAgregados(d))` y `rg.GET("/analitica/export", …)` (mismo encadenamiento de middlewares), y añadir **una línea** `RegisterSpec008(rg, d)` en `backend/internal/routes/register.go` (fichero compartido, sin modificar las líneas existentes) — contracts/analitica.md §autorización + research D5 (depende de T011)
- [ ] T013 [P] [US1] Crear el cliente `frontend/src/services/api/analitica.ts`: interface `AnaliticaViajes` (espejo del backend), función `consultarAnalitica(desde, hasta)` vía `api.request` de `frontend/src/services/api.ts` (patrón de `services/api/viajes.ts`) y función `descargarCSV(desde, hasta)` que hace `fetch` autenticado de `/analitica/export` (Authorization del mismo mecanismo que `api.request`), crea un `Blob` y dispara la descarga con un `<a download>` temporal (el JWT **no** viaja en la URL) — contracts/analitica.md + research D6 (cadena frontend, paralela con backend)
- [ ] T014 [US1] Implementar la página `frontend/src/pages/analitica/index.tsx` con wrapper aislado `frontend/src/pages/analitica/GraficoBarras.tsx` (componente fino sobre **chart.js** que recibe etiquetas + datos): señales (`desde`, `hasta` con default **últimos 30 días**, `datos`, `cargando`, `error`, `descargando`); al aplicar → `consultarAnalitica` y pintado de **tal cual** los arrays del backend (const. I): "Viajes por hora del día" (eje 00–23) y "Viajes por día de la semana" (eje lunes…domingo), totales (`total_viajes`, `total_coste` es-ES) y aviso **"sin datos en el periodo"** cuando `total_viajes == 0`; validación cliente `desde ≤ hasta` (el 422 del backend se muestra igualmente); botón "Descargar CSV" **deshabilitado mientras `descargando`** (edge del doble clic: cada acción descarga un único archivo consistente) → `descargarCSV`; 403 → mensaje de permiso con `role="alert"`; Tailwind + `data-testid` (patrón de `pages/viajes/index.tsx`) — research D6. **Nota**: `GraficoBarras.tsx` importa `chart.js`, cuya alta en `frontend/package.json` está **pendiente de acuerdo del equipo** (Phase 5, T020) — hasta su aprobación la página puede desarrollarse y sus tests (T008) pasan mockeando el wrapper; el `npm run build` completo la requiere (depende de T008, T013)
- [ ] T015 [US1] Publicar la página: añadir la entrada `{ path: "/analitica", component: PanelAnalitica }` en `frontend/src/routes.tsx` y el enlace `<A href="/analitica">Analítica</A>` en el nav de `frontend/src/index.tsx` (ficheros compartidos, una entrada cada uno, sin modificar las existentes) (depende de T014)

**Checkpoint**: US1 completa y testeable — `go test ./...` y `npm test` en verde; escenario manual S8 (API + página) operativo con el seed de T003

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Gate de calidad, verificación de SC y documentación de cierre

- [ ] T016 [P] Gate de calidad + estática: re-ejecutar T8a/T8b/T8c/T8f/T8g (consistencia, bordes del rango, agrupación, 422, roles) en `go test ./...`; `go vet ./...`, `gofmt -l` (solo cuentan avisos reales, no CRLF del working tree), `npm test` y `npm run build` (typecheck TS — requiere chart.js instalado, ver T020)
- [ ] T017 Ejecutar el escenario manual S8 y la validación SC-001…SC-004 de [quickstart.md](./quickstart.md) contra `localhost:5173` + API (login `gestor@si16.test`, agregados de 30 días, descarga y apertura del CSV con tildes/ñ, rango futuro → "sin datos en el periodo", rango invertido → 422, `ana@si16.test` → 403); registrar resultado en la sección 5 de [quickstart.md](./quickstart.md)
- [ ] T018 [P] Revisar coherencia final spec ↔ código: barrido de FR-001…FR-007, clarificaciones 2026-11-01 (inclusividad Europe/Madrid, 00–23 lunes→domingo por inicio, CSV R11) y SC-001…SC-004 contra lo implementado (entrada de `/speckit.converge`), verificar que **no** se tocó `specs/spec-007/`, el plan global ni código de otras specs, y actualizar los checkboxes de este `tasks.md`

---

## Phase 5: Bloqueadas — pendientes de acuerdo del equipo (constitución III)

**⚠️ NO EJECUTAR** sin acuerdo previo del equipo: estas tareas tocan **ficheros compartidos** por las 8 specs. El MVP de la HU-08 está completo y validado sin ellas (Phase 3–4, con el wrapper de gráfico mockeado en tests); se documentaron como `⚠ PENDIENTE DE ACUERDO DEL EQUIPO` en [data-model.md](./data-model.md) y [contracts/analitica.md](./contracts/analitica.md).

- [ ] T019 [BLOCKED] *(requiere acuerdo de equipo)* Añadir la respuesta `422 ENTRADA_INVALIDA` a `/analitica/viajes` y `/analitica/export` en `specs/spec-001/contracts/openapi.yaml` (hoy solo enumeran `200` y `403`), reflejando el comportamiento que la rama ya consolida (patrón spec-003/004/007); hasta entonces [contracts/analitica.md](./contracts/analitica.md) es la fuente de detalle — data-model.md § "PENDIENTES" punto 1
- [ ] T020 [BLOCKED] *(requiere acuerdo de equipo)* Añadir la dependencia `"chart.js": "^4"` a `frontend/package.json` (+ `npm install`), tecnología **ya designada globalmente** para los gráficos HU-08 en R14 del plan global; desbloquea el import de `frontend/src/pages/analitica/GraficoBarras.tsx` y el `npm run build` de T016 — data-model.md § "PENDIENTES" punto 2 + research D6

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: heredada de spec-003/spec-004/spec-007, verificada
- **Foundational (Phase 2)**: T003 y T004 en paralelo; **T003 bloquea el escenario manual S8** y la integración con datos reales
- **User Story 1 (Phase 3)**: requiere Setup ✅; los tests (T005–T008) se escriben ANTES que la implementación; T010 depende de T009 (mismo paquete, `construirFiltro`); T012 depende de T011; T015 depende de T014
- **Polish (Phase 4)**: depende de toda la US1 (T016/T017 requieren T020 aprobada para el build con chart.js)
- **Bloqueadas (Phase 5)**: sin dependencia de ejecución — **solo** de la aprobación del equipo; T019 nunca bloquea el MVP; T020 desbloquea el build del gráfico real (los tests de página la eluden mockeando el wrapper)

### Within User Story 1

- T005, T006, T007, T008 son paralelizables (archivos distintos) y deben FALLAR antes de implementar
- Backend: T009 → T010 → T011 → T012 (cadena de dependencia de compilación, mismo `construirFiltro`)
- Frontend: T013 → T014 → T015 (cadena); T013 en paralelo con la cadena backend
- Backend y frontend avanzan en paralelo tras los tests

### Parallel Example: User Story 1

```text
# Arranque (4 archivos distintos, en paralelo):
T005  unit servicio   backend/internal/services/analitica/service_test.go
T006  unit CSV        backend/internal/services/analitica/csv_test.go
T007  integración     backend/tests/analitica_test.go
T008  test página     frontend/src/pages/analitica/index.test.tsx

# Implementación (dos cadenas en paralelo):
Backend:  T009 → T010 → T011 → T012
Frontend: T013 → T014 → T015
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Setup (heredada ✅) + Foundational (T003 seed lote HU-08, T004 struct `AnaliticaViajes`)
2. Tests T005–T008 → verlos FALLAR
3. Backend T009–T012 → `go test ./...` en verde
4. Frontend T013–T015 → `npm test` en verde (wrapper de gráfico mockeado mientras T020 esté bloqueada)
5. **STOP y VALIDAR**: escenario S8 del quickstart vía API (y con la página una vez aprobada T020)
6. Polish T016–T018 (T016 completo tras aprobación de T020)
7. Phase 5 queda **aparcada** hasta el acuerdo del equipo (no forma parte del MVP funcional; solo T020 condiciona el build definitivo del gráfico)

## Notes

- [P] tasks = different files, no dependencies
- Cada entrada en `routes/register.go` y en el nav/rutas del frontend es de **una línea/entrada** para que los merges con otras ramas sigan siendo triviales (const. V); `models.go` y `cmd/seed/main.go` reciben bloques **puramente aditivos** (patrón revisable de spec-007)
- No migraciones, no claves Redis, sin dependencias npm/go nuevas en las fases 1–4 (chart.js queda a la espera del acuerdo en T020)
- La HU-08 es de **solo lectura**: ningún endpoint toca `viajes`, `estaciones` ni Redis con escrituras
- Un único `construirFiltro` por petición alimenta agregados y CSV: es la garantía FR-006/SC-002 "por construcción" — no duplicar el predicado en ningún fichero
- Commits por fase (convención del repo): `docs: spec-008 …` / `feat: spec-008 …`
