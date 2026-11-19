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

- [X] T001 Crear la estructura raíz según `plan.md`: carpetas `backend/` y `frontend/` con sus árboles (`cmd/`, `internal/`, `migrations/`, `src/`…) y `.gitignore` combinado (Go + Node)
- [X] T002 Crear `docker-compose.yml` en la raíz: `postgres:16-alpine` (5432, volumen `pgdata`, healthcheck `pg_isready`) y `redis:7-alpine` (6379, `--appendonly yes`, healthcheck `redis-cli ping`); credenciales vía `.env` (R12, aprobado por unanimidad)
- [X] T003 [P] Crear `.env.example` en la raíz: `DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `TARIFA_MINUTO=0.20`, `PUERTO=8080`
- [X] T004 [P] Inicializar `backend/go.mod` (Go 1.22) con dependencias: gin, pgx/v5, go-redis/v9, golang-jwt/v5, x/crypto, goose
- [X] T005 [P] Inicializar `frontend/`: Vite + SolidJS + TypeScript + Tailwind (`frontend/package.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/tailwind.config.js`)
- [X] T006 [P] Configurar lint/format: `backend/.golangci.yml` (gofmt/vet) y `frontend/.eslintrc.cjs` + `frontend/.prettierrc` — *Hecho con ESLint 9 en formato «flat» (`frontend/eslint.config.js`) en lugar de `.eslintrc.cjs`, que ESLint 9 ya no usa por defecto.*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Base compartida que MUST completarse ANTES de cualquier user story. Incluye el esquema completo (aprobado por el equipo: `password_hash`, `estaciones.estado` y constraints/índices) y los **helpers compartidos** que hacen independientes las 8 specs.

**⚠️ CRITICAL**: Ninguna historia puede empezar hasta completar esta fase.

- [X] T007 Escribir `backend/migrations/0001_esquema_base.sql` con **íntegro** el modelo de `data-model.md`: tablas `usuarios`, `estaciones`, `bicicletas`, `anclajes`, `viajes`, `incidencias_mantenimientos`, `alertas_balanceo`; CHECKs, FKs e índices (`ux_viajes_usuario_activo`, `ux_viajes_bicicleta_activa`, `ux_anclaje_bicicleta` vía UNIQUE, `ux_alerta_abierta`, índices de HU-04/HU-08)
- [X] T008 Implementar `backend/cmd/migrate/main.go` (goose up/down sobre `backend/migrations/`)
- [X] T009 [P] `backend/internal/config/config.go`: carga `.env` (`DATABASE_URL`, `REDIS_ADDR`, `JWT_SECRET`, `TARIFA_MINUTO`, `PUERTO`)
- [X] T010 [P] `backend/internal/db/db.go`: pool pgx + helper de transacción con `SELECT … FOR UPDATE` (usado por US2/US3/US5/US6/US7)
- [X] T011 [P] `backend/internal/cache/client.go` + `backend/internal/cache/keys.go`: cliente go-redis y contrato `estacion:{id}:estado` (HSET+EXPIRE 120s, HGETALL, SCAN de snapshot, fallback a PostgreSQL) — `contracts/redis-state.md` — *El fallback a PostgreSQL vive en `services/estacion` (que es quien sabe calcular); `cache` solo lee y escribe el contrato. La lectura usa `HGETALL` en pipeline por id (sin `SCAN`), y la escritura un script Lua que no pisa un estado más reciente.*
- [X] T012 [P] `backend/internal/hooks/estacion.go`: publicador de eventos con `hooks.Suscribir(fn func(ctx, ids []int64))` y `hooks.EstacionCambiada(ctx, ids)`. **Flujo único, sin ciclos**: las specs que mutan (US2/US3/US5/US6/US7) llaman tras el commit **solo** a `estacion.RecalcularEstacion(ctx, ids)` (T020); esta reescribe Redis y, al final, llama a `hooks.EstacionCambiada(ctx, ids)` para avisar a los suscriptores (hoy US7 umbrales). Ningún suscriptor MUST volver a llamar a `RecalcularEstacion` (evita bucles); un error de un suscriptor se registra en el log y no corta a los demás — clave del principio V
- [X] T013 `backend/internal/models/`: structs compartidos (`Usuario`, `Estacion`, `Anclaje`, `Bicicleta`, `Viaje`, `Incidencia`, `AlertaBalanceo`), enums de estado y helper `AbonoVigente(u)` (`abono_activo AND fecha_vencimiento_abono >= hoy`, R15)
- [X] T014 `backend/internal/auth/`: bcrypt, emisión de JWT HS256 (exp 24 h, claims `sub`/`rol`/`nombre`) y middlewares `RequireAuth` / `RequireRole(...)` (R5)
- [X] T015 [P] `backend/internal/httperr/`: envelope `{error:{codigo,mensaje}}` y códigos de `contracts/README.md` (`ABONO_VENCIDO`, `VIAJE_ACTIVO`, …)
- [X] T016 `backend/internal/routes/register.go`: router base `/api/v1` y patrón `RegisterSpec00N(rg, deps)` — **único archivo de rutas compartido** (merges triviales)
- [X] T017 Implementar auth base: `POST /auth/login` y `GET /auth/me` en `backend/internal/handlers/auth/handler.go` + test mínimo de contrato
- [X] T018 `backend/cmd/server/main.go`: wiring de config, db, cache, hooks, registro de rutas y workers (cada spec añadirá tras fusionar una única línea para su worker)
- [X] T019 [P] `backend/cmd/seed/main.go`: ~60 estaciones geolocalizadas (incl. alguna `fuera_de_servicio` y extremos de ocupación para S7), anclajes = capacidad, flota (algunas `averiada`), usuarios de **todos los roles** con/sin abono vigente (hashes bcrypt)
- [X] T020 `backend/internal/services/estacion/estado.go`: `RecalcularEstacion` (consulta agregada R9/R19: `bicis_disponibles` = anclajes con bicicleta en estado `disponible`; `anclajes_libres` = anclajes en estado `libre`; un anclaje con bicicleta `averiada`/`en_taller` **no suma en ninguno de los dos**; `ocupacion_pct = 0` si la estación está `fuera_de_servicio`, pero sus contadores enteros se calculan y guardan igualmente; **no** validar `bicis_disponibles + anclajes_libres = capacidad_total` + HSET del hash + al final, notificación `hooks.EstacionCambiada(ctx, ids)` (T012); firma `RecalcularEstacion(ctx, ids []int64) error`, es el **único** punto de entrada que llaman las specs tras el commit). Si la escritura en Redis falla tras el commit, registrar en el log y **no** revertir la transacción (lo corrige el barrido, R16); **helper compartido** que importarán US2/US3/US5/US6/US7 (propiedad del paquete de Iker)
- [X] T021 [P] Frontend base: `frontend/src/services/api.ts` (fetch wrapper con Bearer y envelope de error), `frontend/src/stores/sesion.ts` (login/logout, token) y `frontend/src/pages/Login.tsx`
- [X] T022 `frontend/src/routes.tsx` (**único archivo de rutas compartido**) + `frontend/src/components/Layout.tsx` (navegación por áreas con guard de rol)
- [X] T023 [P] Configurar Vitest + @solidjs/testing-library + MSW en `frontend/vitest.config.ts` y `frontend/tests/setup.ts` con handlers derivados de `contracts/openapi.yaml`

**Checkpoint**: fundación lista — las 8 stories pueden empezar en paralelo, cada una en su rama `spec-00N`.

---

## Phase 3: User Story 1 — Consulta de disponibilidad en tiempo real (spec-001, HU-01, P1) 🎯 MVP

**Goal**: Mapa en vivo de las ~60 estaciones con contadores fiables desde Redis, frescura ≤ 30 s extremo a extremo (del commit al pintado), aviso de «datos desactualizados» y estados visuales accesibles «sin bicicletas», «sin anclajes libres» y «fuera de servicio» (Iker Roig). Cubre FR-001…FR-013 y SC-001…SC-008 (FR-010…FR-013 y SC-008 añadidos el 2026-11-19 en el clarify de refinamiento).

**Independent Test**: Dar de alta estaciones con estados variados, abrir el mapa, verificar contadores contra la BD real y medir que un cambio se refleja en ≤ 30 s; cortar la conexión y ver el aviso a los > 30 s; comprobar los estados sin bicis, sin anclajes y fuera de servicio (escenarios S1 y S1a–S1d del quickstart).

**Orden de ejecución dentro de la fase** (las tareas T095–T100 se añadieron el 2026-11-03 tras `/speckit.clarify`; conservan la numeración del resto del equipo): T024, T025, T099, T096, T098 (tests, fallando) → T026, T027 → T028, T029, T095 → T030, T097 → T031 → T032 → T100 (validación de rendimiento). **Refinamiento 2026-11-19** (T101–T110, enunciado §5.2): T101, T103, T105, T106, T107 (tests, fallando donde aplique) → T102, T104, T108 → T109, T110 (validación manual y regresión).

### Tests for User Story 1

- [X] T024 [P] [US1] Test de contrato `GET /estaciones` y `GET /estaciones/{id}` contra `contracts/openapi.yaml` en `backend/tests/contract/spec001_test.go`
- [X] T025 [P] [US1] Test de integración de frescura en `backend/tests/integration/spec001_frescura_test.go`: mutación → hash `estacion:{id}:estado` actualizado post-commit; clave expirada → fallback a PostgreSQL y repoblado; averiadas excluidas de `bicis_disponibles`; estación `fuera_de_servicio` marcada (FR-005/FR-006). Además (R19): un anclaje con bici `averiada` no cuenta como libre; una estación con todos sus anclajes ocupados por averiadas da `bicis_disponibles = 0` y `anclajes_libres = 0`; se acepta `bicis_disponibles + anclajes_libres < capacidad_total`; un anclaje `fuera_de_servicio` sin bici **no** cuenta como libre (SC-003); dos cambios sucesivos en la misma estación → el hash termina con el estado final más reciente (caso límite de la spec); `RecalcularEstacion` notifica una sola vez a un suscriptor de prueba y no se reentra (T012); una estación `fuera_de_servicio` sigue devolviendo contadores enteros en la API (el contrato `EstacionEstado` no cambia, R18)
- [X] T099 [P] [US1] Test de integración del camino de respaldo en `backend/tests/integration/spec001_latencia_test.go` (R16, FR-003, SC-002): `UPDATE` directo de `anclajes`/`bicicletas`/`estaciones.estado` en PostgreSQL **sin** llamar a `RecalcularEstacion` → el hash `estacion:{id}:estado` refleja el cambio en el siguiente barrido del worker. Para que el test sea determinista, el worker acepta el intervalo como parámetro (en producción 10 s, T026): el test lo arranca con 1 s y espera hasta 1 s + 2 s de margen; aparte, un test unitario comprueba que la constante de producción es `10 * time.Second`; con Redis caído durante la escritura síncrona, la transacción SQL queda confirmada y el barrido repara el hash al volver Redis

### Implementation for User Story 1

- [X] T026 [US1] `backend/internal/workers/refresco_estado/worker.go`: barrido **cada 10 s** (R16; antes 15 s) de las ~60 estaciones con una única consulta agregada (red de seguridad de frescura; con polling de 15 s deja el peor caso extremo a extremo en ≈ 28 s < 30 s) + escritura de `si16:worker:refresco:ultimo` + línea de wiring en `backend/cmd/server/main.go`; constante `IntervaloBarrido = 10 * time.Second` en el paquete y constructor `Nuevo(deps, intervalo time.Duration)` para que los tests usen un intervalo corto (T099)
- [X] T027 [US1] `backend/internal/handlers/estacion/handler.go` + `backend/internal/routes/spec_001_routes.go`: snapshot `GET /estaciones` (con `actualizado_en`) y detalle `GET /estaciones/{id}` (con anclajes) — FR-001, FR-002, FR-004
- [X] T028 [P] [US1] `frontend/src/services/api/estaciones.ts`: tipos `EstacionEstado`/`EstacionDetalle` y cliente
- [X] T029 [P] [US1] `frontend/src/stores/red.ts`: polling de `GET /estaciones` cada 15 s, reintentos automáticos en el siguiente ciclo **sin borrar el último estado conocido**, pausa con pestaña oculta (al volver a mostrarla, petición inmediata; mientras esa petición está en curso y durante un máximo de 2 s no se muestra el aviso, para evitar un parpadeo si la respuesta llega bien) y, según R17: señal `ultimaActualizacionOk` con el **reloj del cliente** al recibir una respuesta 2xx; temporizador de 1 s que expone `desactualizado = ahora − ultimaActualizacionOk > 30 s` (estrictamente mayor); `desactualizado` vuelve a `false` en cuanto llega una respuesta correcta (FR-007)
- [X] T095 [P] [US1] `frontend/src/pages/mapa/estadoVisual.ts`: función pura `estadoVisual(e: EstacionEstado)` con prioridad R18: (1) `estado = 'fuera_de_servicio'` → `'fuera_de_servicio'`; (2) `bicis_disponibles = 0` y `anclajes_libres = 0` → `'sin_bicis_y_sin_anclajes'`; (3) `bicis_disponibles = 0` → `'sin_bicis'`; (4) `anclajes_libres = 0` → `'sin_anclajes'`; (5) si no → `'normal'`. Exporta también, para cada estado, su icono, su etiqueta corta («Fuera de servicio», «Sin bicicletas», «Sin anclajes libres», ambos) y su clase Tailwind de color (FR-008, FR-009)
- [X] T030 [US1] `frontend/src/components/MapaBase.tsx`: Leaflet + tiles OpenStreetMap; marcadores `L.divIcon` que reciben el estado de `estadoVisual` (T095) y muestran **icono + texto corto además del color** (FR-008: la distinción MUST NOT depender solo del color) con `aria-label` descriptivo; todos los marcadores, también los de estaciones fuera de servicio, son seleccionables (FR-009); geolocalización (caso límite de la spec): si el usuario concede permiso, `navigator.geolocation` centra el mapa en su posición; si lo deniega, no está disponible o tarda más de 5 s, el mapa se queda encuadrado en toda la red (`fitBounds` de las estaciones) sin errores bloqueantes; el componente es genérico, la lógica de estados vive en `pages/mapa/`. *Nota 2026-11-19: el centrado en el usuario queda acotado por FR-013 (solo si está dentro del área de la red); lo implementa T108, que sustituye esta parte de T030*
- [X] T097 [US1] `frontend/src/pages/mapa/AvisoDesactualizado.tsx`: muestra siempre «Actualizado a las HH:MM:SS» (hora local de `ultimaActualizacionOk`) y, cuando `desactualizado` (T029), un banner visible con `role="status"`, icono y texto «Datos desactualizados» (no solo color); no oculta ni vacía los contadores (FR-007, SC-005)
- [X] T031 [US1] `frontend/src/pages/mapa/MapaRed.tsx`: mapa completo (< 5 s) con `AvisoDesactualizado` (T097) y panel de detalle (< 2 s) `frontend/src/pages/mapa/DetalleEstacion.tsx`: estación operativa → contadores de bicis disponibles y anclajes libres; un contador a 0 se muestra como «0 · Sin bicicletas» / «0 · Sin anclajes libres» (FR-008); **solo estos dos contadores, sin desglose de averiadas** (FR-006); estación `fuera_de_servicio` → texto «Fuera de servicio» y «—» en lugar de ambos contadores, sin ninguna acción de recogida ni devolución (FR-009); conserva el último estado conocido si falla la conexión
- [X] T032 [P] [US1] Tests frontend en `frontend/tests/mapa/`: render con datos MSW, estado vacío y desconexión (Vitest + MSW); `DetalleEstacion` operativa con 0 bicis, con 0 anclajes y con ambos a 0 (textos de FR-008), fuera de servicio con «—» y sin acciones (FR-009); marcadores con texto/`aria-label` distinto por estado (no solo clase de color); marcador fuera de servicio seleccionable; se pintan **todas** las estaciones devueltas por la API (SC-001); geolocalización denegada (mock de `navigator.geolocation` con error) → el mapa se muestra encuadrado en la red sin error
- [X] T096 [P] [US1] Tests unitarios de `estadoVisual` en `frontend/tests/mapa/estadoVisual.test.ts`: tabla con todas las combinaciones de `estado` × (`bicis_disponibles` 0/>0) × (`anclajes_libres` 0/>0), incluido fuera de servicio con 0 bicis (debe ganar `'fuera_de_servicio'`) y suma de contadores < capacidad (R18, R19)
- [X] T098 [P] [US1] Tests del aviso en `frontend/tests/mapa/desactualizado.test.ts` con temporizadores falsos de Vitest + MSW devolviendo error de red: a los 30 s exactos **no** hay aviso; a los 31 s aparece «Datos desactualizados» y los contadores siguen en pantalla; la primera respuesta 2xx lo quita y actualiza «Actualizado a las…» (FR-007, SC-005)
- [x] T100 [US1] Validación de rendimiento SC-004 (escenario S1e de `specs/spec-001/quickstart.md`): con el seed de ~60 estaciones, Chrome DevTools en perfil «Fast 4G» y caché vacía, medir con la pestaña Performance el tiempo hasta que todos los marcadores están pintados (< 5 s) y desde el clic en un marcador hasta que el detalle muestra los datos (< 2 s); registrar los tiempos en la PR de spec-001. Si no se cumple, optimizar `GET /estaciones` (T027) o el pintado de marcadores (T030) antes de fusionar

### Refinamiento de User Story 1 (2026-11-19, tras el clarify de refinamiento)

> Origen: `refinamiento/clarify.md` (sesión 2026-11-19) y `refinamiento/analyze.md` (hallazgos G1–G4, H1). Responsable: Iker Roig · Revisor: _(compañero)_. Ninguna tarea cambia la API, el esquema ni Redis.

- [ ] T101 [P] [US1] Actualizar el test «si la primera carga falla lo indica y reintenta» de `frontend/tests/mapa/mapa.test.tsx` (FR-010, escenario 4; hoy espera «No se pudo conectar con el servidor»): MSW devuelve error **sin datos previos** → se ve «No se pudo cargar la disponibilidad» con `role="alert"`, hay un botón «Reintentar» y ningún marcador; al pulsar «Reintentar» se hace una petición nueva en el acto; cuando MSW responde 2xx, aparecen los marcadores y el mensaje desaparece. El test existente «si la conexión falla conserva el último estado conocido» **no** se toca: con datos previos se mantiene R17
- [ ] T102 [US1] Implementar R20 en `frontend/src/pages/mapa/MapaRed.tsx` (y `frontend/src/stores/red.ts` si hace falta exponer el estado «sin datos + error»): sustituir «Sin datos todavía · Error N. Reintentando…» por el mensaje y el botón de FR-010, que llama a `refrescar()`; hace pasar T101
- [ ] T103 [P] [US1] Test del listado en `frontend/tests/mapa/listaEstaciones.test.tsx` (FR-011, SC-008, escenario 5): el conmutador «Mapa / Lista» es accesible por teclado; el listado contiene **todas** las estaciones devueltas por la API con el mismo texto de estado (`estadoVisual`) y los mismos contadores que los `aria-label` de los marcadores; una estación fuera de servicio aparece con «Fuera de servicio» y «—»; tras un nuevo polling con datos cambiados, el listado se actualiza; pulsar una fila abre `DetalleEstacion`; con Leaflet sin teselas (mock de error de carga) el listado sigue funcionando
- [ ] T104 [US1] Implementar R21: `frontend/src/pages/mapa/ListaEstaciones.tsx` (filas ordenadas por nombre, cada una un botón con nombre, etiqueta de estado y contadores o «—») y conmutador «Mapa / Lista» con `role="tablist"` en `MapaRed.tsx`, compartiendo la selección con el detalle; hace pasar T103
- [ ] T105 [P] [US1] Ampliar el test existente «el detalle se actualiza con los datos en vivo de la estación seleccionada» de `frontend/tests/mapa/mapa.test.tsx` (FR-004 ampliado, escenario 6) con el **paso a fuera de servicio**: el siguiente polling devuelve `estado: 'fuera_de_servicio'` → el detalle abierto muestra «Fuera de servicio» y «—», sin cerrarse ni perder la selección. *El caso de los contadores ya está cubierto; el comportamiento ya existe (P1.4 y P3.2 de `refinamiento/pruebas-previas.md`)*
- [ ] T106 [P] [US1] Test de sesión caducada en `frontend/tests/mapa/sesionCaducada.test.tsx` (FR-012, escenario 7): con la sesión iniciada y el mapa abierto, MSW responde 401 al siguiente polling → se llama a `cerrarSesion()` y se navega a `/login`; tras un login correcto se vuelve a `/` (vista de disponibilidad). *La parte del servidor ya la cubre `TestEstacionesExigeSesion` en `backend/tests/integration/spec001_frescura_test.go` (sin token y con token falso → 401)*
- [ ] T107 [P] [US1] Tests de encuadre en `frontend/tests/mapa/encuadre.test.tsx` (FR-013, escenario 8; hallazgo H1) con mock de `navigator.geolocation`: (a) posición **fuera** del área y datos que llegan **después** de la posición → `data-encuadre="red"`; (b) posición fuera y datos **antes** → `red`; (c) posición **dentro** del área → `usuario` (equivale al test existente «con permiso de ubicación centra el mapa en el usuario», que debe seguir pasando); (d) permiso denegado → `red`
- [ ] T108 [US1] Implementar R23 en `frontend/src/components/MapaBase.tsx`: área de la red = `L.latLngBounds(estaciones).pad(0.2)`; guardar la posición si llega antes que los datos y decidir cuando lleguen; centrar en el usuario solo si `area.contains(posicion)`; en otro caso, `fitBounds` de la red; hace pasar T107 sin romper los tests de geolocalización de `mapa.test.tsx`
- [ ] T109 [US1] Validación manual de los escenarios **S1f–S1i** de `quickstart.md` en el navegador; registrar en `refinamiento/` la entrada, el resultado esperado y el observado **antes** (P2.5, P2.7 y H1 de `pruebas-previas.md`) y **después** de T102, T104 y T108 (enunciado §5.3, paso 4)
- [ ] T110 [US1] Regresión de HU-01: ejecutar `go test ./...` y `npm test` y repetir S1–S1e; guardar la salida en `refinamiento/` (enunciado §5.4, «Regresión»)

**Checkpoint**: HU-01 funcional y testeable de forma independiente (MVP del sistema).

---

## Phase 4: User Story 2 — Desbloqueo de bicicleta por código de anclaje (spec-002, HU-02, P1)

**Goal**: Inicio de viaje atómico con las reglas críticas de abono y unicidad de viaje activo (Iker Roig).

**Independent Test**: Con usuario de prueba seed, introducir código de anclaje con bici disponible → viaje activo; con abono caducado → 403 con causa; con viaje activo → 409 (escenario S2). Ver `specs/spec-002/spec.md`.

### Tests for User Story 2

- [X] T033 [P] [US2] **T1 (const. I)**: `backend/tests/integration/spec002_abono_test.go` — desbloqueo con abono caducado/ incoherente → 403 `ABONO_VENCIDO` y cero efectos sobre anclaje, bici y viajes
- [X] T034 [P] [US2] **T2 (const. I)**: `backend/tests/integration/spec002_concurrencia_test.go` — segundo desbloqueo → 409 `VIAJE_ACTIVO`; dos desbloqueos concurrentes (goroutines) → exactamente 1 viaje creado (red `ux_viajes_usuario_activo`); código inexistente/estación fuera de servicio/anclaje sin bici operativa → 404/409 sin efectos
- [X] T035 [P] [US2] Test de contrato `POST /alquileres/desbloqueo` y `GET /alquileres/activo` en `backend/tests/contract/spec002_test.go`

### Implementation for User Story 2

- [X] T036 [US2] `backend/internal/services/alquiler/service.go`: transacción única con `SELECT … FOR UPDATE` del anclaje; valida abono vigente, ausencia de viaje activo, estación operativa, anclaje en servicio y bicicleta `disponible`; crea viaje `en_curso` con `fecha_inicio`; desacopla anclaje (→ `libre`), bici → `en_uso`; tras commit `estacion.RecalcularEstacion(ctx, ids)` (T020), que ya notifica el hook — no llamar a `hooks.EstacionCambiada` directamente
- [X] T037 [US2] `backend/internal/handlers/alquiler/handler.go` + `backend/internal/routes/spec_002_routes.go` (403/404/409 con envelope de `httperr`)
- [X] T038 [P] [US2] `frontend/src/services/api/alquiler.ts`
- [X] T039 [US2] `frontend/src/pages/alquiler/Desbloqueo.tsx`: entrada de código de anclaje, mensajes de causa legibles, estado de viaje activo actual

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

- [ ] T073 [US7] `backend/internal/services/balanceo/umbrales.go`: `EvaluarEstacion` (crear alerta si cruza umbral y no hay abierta; resolver si vuelve a rango); se suscribe con `hooks.Suscribir(...)` (T012) vía `init()` de su paquete — sin tocar ficheros de otras specs y sin llamar a `RecalcularEstacion` desde el suscriptor (evita bucles)
- [ ] T074 [US7] `backend/internal/workers/alertas/worker.go`: barrido de toda la red cada 60 s (garantía SC-001 < 1 min) + línea de wiring en `backend/cmd/server/main.go`
- [ ] T075 [US7] `backend/internal/services/balanceo/traslados.go`: valida unidades ≤ disponibles en origen y estaciones operativas; mueve las bicis en transacción; tras el commit llama a `estacion.RecalcularEstacion(ctx, [origen, destino])` (T020), que dispara la evaluación de umbrales de T073 vía hook (no evaluarlos dos veces)
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

- [ ] T088 [P] Ejecutar los escenarios **S1–S8** (incluidos S1a–S1d de HU-01) de [quickstart.md](./quickstart.md) end-to-end en local (docker compose + seed + backend + frontend) y registrar resultados por spec en la PR correspondiente
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
- **US7 (spec-007, Alberto)**: arranca tras Phase 2; independiente vía fixtures de estaciones en extremos de ocupación; se suscribe con `hooks.Suscribir` (T012) sin modificar ficheros ajenos
- **US8 (spec-008, Alberto)**: arranca tras Phase 2; independiente vía fixtures de viajes finalizados de toda la red

### Within Each User Story

1. Tests (T1–T5 y contrato) escritos **antes** de la implementación y fallando
2. Servicio (lógica de negocio y transacciones) → handlers + `routes/spec_00N_routes.go` → cliente frontend → páginas → tests frontend
3. Recálculo de estaciones post-commit siempre vía `RecalcularEstacion` (T020), que notifica a los suscriptores del hook (T012) — nunca duplicar cómputos ni llamar al hook directamente
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
4. **PARAR y VALIDAR**: escenarios S1 y S1a–S1d del quickstart + frescura ≤ 30 s medida de extremo a extremo (también por el camino de respaldo)
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
- IDs T095–T100 (US1) añadidos el 2026-11-03 al final de la numeración para no renumerar las tareas del resto del equipo; su orden real está al inicio de la Phase 3
- **Estado de las fases 4–10**: este `tasks.md` es el desglose global del plan, pero cada integrante ha llevado el estado real de su historia en el `tasks.md` de su propia spec (`specs/spec-002…008/tasks.md`). Las casillas sin marcar de esas fases aquí **no** indican trabajo pendiente; la fuente de verdad del estado es el `tasks.md` de cada spec (hallazgo I4 del analyze de refinamiento, 2026-11-19)
- IDs T101–T110 (US1) añadidos el 2026-11-19 en la fase de analyze del refinamiento (enunciado §5.2, pasos 1 y 4), a partir de los hallazgos de `refinamiento/analyze.md`; se ejecutan con `/speckit-implement` y se verifican con `/speckit-converge` (§5.3)
- Cambios en helpers compartidos (T012 flujo único del hook, T020 «no revertir el commit si falla Redis»): comunicarlos al equipo en la PR de spec-001 (const. III y acuerdos de colaboración), aunque no cambien esquema ni claves
- Gate de calidad: los tests T1–T5 (const. I) son **obligatorios** en las fases 4, 5, 7 y 9 y se re-ejecutan en T089
- Esquema compartido: 0001 es la única migración aportada; cambios futuros exigen ordinal acordado (T093, const. III)
- Commit tras cada tarea o grupo lógico; parar en cada Checkpoint para validar la historia de forma independiente
- Evitar: tareas vagas, conflictos de fichero entre stories, dependencias que rompan la independencia (el hook `hooks/estacion.go` + helpers T011/T020 existen precisamente para evitarlas)
