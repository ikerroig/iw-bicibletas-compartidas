# Tasks: Desbloqueo de bicicleta mediante código de anclaje (HU-02)

**Input**: Design documents from `/specs/spec-002/`: [plan.md](./plan.md), [spec.md](./spec.md) (1 user story, P1), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/alquileres.md](./contracts/alquileres.md), [quickstart.md](./quickstart.md)

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: **INCLUIDOS**. El gate de calidad de la constitución exige pruebas automatizadas de las reglas críticas antes de fusionar (T1 y T2).

**Organization**: la HU-02 tiene **una única user story** (US1, P1), que es también el MVP. Setup y Foundational están cubiertas por la base compartida ya fusionada en `main` (PR #1).

**Trazabilidad**: este fichero se generó después del commit `e090f41`, que implementó T033–T039 del [tasks.md global](../spec-001/tasks.md). Las tareas marcadas `[X]` ya están en ese commit, y su correspondencia global se indica en cada una. Las pendientes son los Deltas de [plan.md](./plan.md), que aplican las aclaraciones de la sesión 2026-11-04 y los hallazgos de `/code-review`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

Sin tareas: repositorio, dependencias y entorno están en `main`.

---

## Phase 2: Foundational (Blocking Prerequisites)

Sin tareas: autenticación, `httperr`, `db.EnTransaccion`, `RecalcularEstacion` y esquema `0001` están en `main`. Sin migraciones nuevas.

---

## Phase 3: User Story 1 - Inicio de viaje con código de anclaje (Priority: P1) 🎯 MVP

**Goal**: un usuario con abono vigente introduce el código de un anclaje y obtiene un viaje activo de forma atómica; cualquier otro caso se rechaza con causa y sin efectos.

**Independent Test**: escenarios V1–V8 de [quickstart.md](./quickstart.md) y `go test ./...` con los gates T1 y T2.

### Primera implementación (commit `e090f41`)

- [X] T001 [P] [US1] T1, abono no vigente → 403 `ABONO_VENCIDO` sin efectos, en `backend/tests/integration/spec002_abono_test.go` (global T033)
- [X] T002 [P] [US1] T2, viaje activo y concurrencia → exactamente 1 viaje, y rechazos de anclaje sin efectos, en `backend/tests/integration/spec002_concurrencia_test.go` (global T034)
- [X] T003 [P] [US1] Test de contrato de ambos endpoints en `backend/tests/contract/spec002_test.go` (global T035)
- [X] T004 [US1] Transacción de desbloqueo y consulta del viaje activo en `backend/internal/services/alquiler/service.go` (global T036)
- [X] T005 [US1] Handler y rutas en `backend/internal/handlers/alquiler/handler.go` y `backend/internal/routes/spec_002_routes.go` (global T037)
- [X] T006 [P] [US1] Cliente de API en `frontend/src/services/api/alquiler.ts` (global T038)
- [X] T007 [US1] Página con formulario, causa del rechazo y viaje en curso en `frontend/src/pages/alquiler/Desbloqueo.tsx` (global T039)
- [X] T008 [P] [US1] Tests de la página en `frontend/tests/alquiler/desbloqueo.test.tsx`

### Contrato (Delta D3, D7)

- [X] T009 [US1] En `specs/spec-001/contracts/openapi.yaml`: añadir al esquema `Viaje` las propiedades opcionales `bicicleta_codigo` y `estacion_origen_nombre` (`{ type: [string, 'null'] }`, fuera de `required`); añadir `'401'` a `/alquileres/desbloqueo` y `/alquileres/activo`; cambiar el ejemplo de `codigo_anclaje` de ese endpoint a `E01-A01` y añadir `maxLength: 64`

### Tests de los Deltas (escribir antes y comprobar que fallan)

- [X] T010 [P] [US1] **FR-002 (D1)**: crear `backend/internal/services/alquiler/abono_test.go` con tests de `fechaNegocio(instante time.Time) time.Time`: invierno `2026-11-04T22:59:59Z` → día 4 y `2026-11-04T23:00:00Z` → día 5; verano `2026-07-14T21:59:59Z` → día 14 y `2026-07-14T22:00:00Z` → día 15; con `models.AbonoVigente`, un abono con vencimiento en el día 4 es vigente con el primer instante de invierno y no con el segundo
- [X] T011 [P] [US1] **FR-001 (D2)**: crear `backend/tests/integration/spec002_codigo_test.go`: el código en minúsculas y con espacios alrededor (`"  e<id>-a01 "`) → 201; cadena de 65 caracteres y cadena con `\u0000` → 422 `VALIDACION` sin efectos; 64 caracteres inexistentes → 404 `ANCLAJE_NO_EXISTE`
- [X] T012 [P] [US1] **FR-004 (D5)**: crear `backend/tests/integration/spec002_estacion_test.go`: con una transacción abierta que hace `UPDATE estaciones SET estado = 'fuera_de_servicio'`, un desbloqueo en goroutine espera; tras el commit de esa transacción responde 409 `ANCLAJE_NO_DISPONIBLE` y no crea viaje
- [X] T013 [P] [US1] **SC-003 (D6)**: crear `backend/tests/integration/spec002_postcommit_test.go`: `RecalcularEstacion` con un contexto ya cancelado notifica al suscriptor con un contexto no vencido
- [X] T014 [US1] **FR-008 (D3, D7)**: ampliar `backend/tests/contract/spec002_test.go`: la respuesta 201 y la de `GET /alquileres/activo` incluyen `bicicleta_codigo` y `estacion_origen_nombre` reales; sin token, ambos endpoints devuelven 401 y cumplen el contrato con `llamar`
- [X] T015 [US1] **FR-008**: actualizar `frontend/tests/alquiler/desbloqueo.test.tsx`: la vista del viaje muestra la hora de inicio, `bicicleta_codigo` y `estacion_origen_nombre`, no hace ninguna petición a `/estaciones/{id}` y no muestra `bicicleta_id`
- [X] T028 [P] [US1] **FR-009**: crear `frontend/tests/alquiler/menu.test.tsx`: el rol `usuario` ve «Desbloquear»; `gestor` y los operarios no; con token pero sin usuario guardado tampoco se ve, y «Mapa» sigue visible
- [X] T030 [P] [US1] **SC-003 (D6)**: crear `backend/tests/integration/spec002_latencia_test.go`: un desbloqueo válido responde 201 en menos de 10 s con Redis normal, y también con un cliente Redis contra un `net.Listen` local que acepta conexiones y nunca responde; en ese segundo caso el viaje queda creado
- [X] T031 [P] [US1] **Caso límite de conexión**: crear `backend/tests/integration/spec002_conexion_test.go`: `Desbloquear` con un contexto cancelado devuelve error y no deja viaje parcial ni cambios en anclaje o bicicleta; un reintento posterior responde 201

### Implementación de los Deltas

- [X] T016 [US1] **D1** en `backend/internal/services/alquiler/service.go`: añadir `import _ "time/tzdata"`, una `fechaNegocio` que convierte a `Europe/Madrid` y usarla en `validarUsuario` en lugar de `time.Now().UTC()`, sin modificar `models.AbonoVigente`
- [X] T017 [US1] **D2** en `backend/internal/handlers/alquiler/handler.go`: aplicar `TrimSpace` y `ToUpper`; responder 422 `VALIDACION` si el código queda vacío, supera 64 caracteres o contiene caracteres de control («sin espacios alrededor, en mayúsculas, 1–64 caracteres, sin caracteres de control»)
- [X] T018 [US1] **D5** en `backend/internal/services/alquiler/service.go`: leer la estación con `FOR SHARE`, recuperando también `nombre`
- [X] T019 [US1] **D3** en `backend/internal/services/alquiler/service.go`: crear el tipo `ViajeActivo` que embebe `models.Viaje` con `BicicletaCodigo` (`bicicleta_codigo`) y `EstacionOrigenNombre` (`estacion_origen_nombre`); `Desbloquear` los rellena con lo ya leído en la transacción y `Activo` con un JOIN, sin tocar `models.Viaje`
- [X] T020 [US1] **D6** en `backend/internal/services/estacion/estado.go` y `backend/internal/services/alquiler/service.go`: en `RecalcularEstacion`, notificar `hooks.EstacionCambiada` con un contexto de 3 s derivado de `context.WithoutCancel`, sin cambiar la firma; reducir `recalculoTimeout` a 2 s; activar `ContextTimeoutEnabled` en `backend/internal/cache/client.go`, sin el cual go-redis ignora ese plazo
- [X] T021 [US1] **D3, D8** en `frontend/src/services/api/alquiler.ts`: añadir `bicicleta_codigo` y `estacion_origen_nombre` a un tipo `ViajeActivo` y quitar el parámetro `signal` de `viajeActivo`
- [X] T022 [US1] **FR-008** en `frontend/src/pages/alquiler/Desbloqueo.tsx`: mostrar `bicicleta_codigo` y `estacion_origen_nombre`, y eliminar `createResource` de `detalleEstacion`
- [X] T023 [US1] **D8** en `backend/internal/handlers/alquiler/handler.go`: sustituir las ramas `!ok` inalcanzables por `s, _ := auth.SesionDe(c)`, como en `handlers/auth`
- [X] T029 [US1] **FR-009** en `frontend/src/components/Layout.tsx`: quitar el valor por defecto `?? 'usuario'`. Sin usuario guardado solo se muestran las áreas de `TODOS`: `a.roles === TODOS || (rol !== undefined && a.roles.includes(rol))`

**Checkpoint**: HU-02 cumple FR-001…FR-009 y los tests T010–T015, T028, T030 y T031 pasan.

---

## Phase 4: Correcciones del segundo `/code-review`

**Purpose**: resolver los hallazgos aprobados sobre `05ef865`. Tests antes que implementación (D11–D14 de [research.md](./research.md)).

### Documentación

- [X] T032 [US1] Registrar D11–D14 en `specs/spec-002/research.md` y `specs/spec-002/plan.md`, añadir V9 a `specs/spec-002/quickstart.md` y corregir los códigos de la carrera por una bicicleta en `specs/spec-002/contracts/alquileres.md`

### Tests

- [X] T033 [P] [US1] Calcular `dias()` con la fecha de `Europe/Madrid` en `backend/tests/integration/spec002_abono_test.go`
- [X] T034 [P] [US1] Rehacer `backend/tests/integration/spec002_conexion_test.go`: cancelar con el `INSERT` del viaje en espera y las escrituras previas ya hechas; sin efectos tras el rollback y reintento con 201 (sustituye a T031)
- [X] T035 [P] [US1] **SC-003 (D11)**: crear `backend/tests/integration/spec002_plazo_test.go`: con el anclaje bloqueado por otra transacción, el desbloqueo responde error en menos de 5 s y sin efectos
- [X] T036 [P] [US1] **FR-007 (D12)**: crear `backend/tests/integration/spec002_commit_test.go`: cancelar el contexto al terminar `fn` no impide que `db.EnTransaccion` confirme
- [X] T037 [US1] **FR-007 (D13)**: en `backend/tests/integration/spec002_estacion_test.go`, un desbloqueo espera a una revocación del abono en curso y responde 403 `ABONO_VENCIDO`
- [X] T038 [P] [US1] **FR-008 (D14)**: crear `frontend/tests/alquiler/aviso.test.tsx`: con viaje se ve la hora, el código y el enlace a `/alquiler`; sin viaje, en `/alquiler` o con otro rol no se ve

### Implementación

- [X] T039 [US1] **D11, D13** en `backend/internal/services/alquiler/service.go`: plazo de 4 s para la transacción y `FOR SHARE` al leer el abono
- [X] T040 [US1] **D12** en `backend/internal/db/db.go`: confirmar con `context.WithoutCancel` y plazo de 2 s
- [X] T041 [US1] **D14**: crear `frontend/src/stores/viaje.ts` y usarlo en `frontend/src/pages/alquiler/Desbloqueo.tsx`
- [X] T042 [US1] **D14**: crear `frontend/src/components/AvisoViaje.tsx`, montarlo en `frontend/src/components/Layout.tsx` y añadir el handler por defecto de `/alquileres/activo` en `frontend/tests/handlers.ts`

**Checkpoint**: tests T033–T038 en verde y V9 comprobado en el navegador.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [X] T024 [P] Actualizar la regla de abono de `specs/spec-001/data-model.md` (línea de `CURRENT_DATE`) para indicar la fecha actual en `Europe/Madrid`
- [X] T025 Ejecutar `go vet ./...`, `go test -count=1 ./...` en `backend/` y `npx tsc --noEmit`, `npm run lint`, `npm test` en `frontend/`; repetir 10 veces los tests de concurrencia de `backend/tests/integration/spec002_concurrencia_test.go`
- [X] T026 Ejecutar V1–V8 de `specs/spec-002/quickstart.md` en navegador y API, y anotar el resultado en la PR
- [X] T027 Incluir en la PR los cambios aditivos de `specs/spec-002/contracts/alquileres.md`, la zona horaria del abono (a revisar por spec-004 y spec-008) y la nota de trazabilidad de `specs/spec-002/plan.md`

---

## Phase 6: Refinamiento de HU-02 (2026-11-19, tras el clarify de refinamiento)

> Origen: [refinamiento/pruebas-previas.md](./refinamiento/pruebas-previas.md) y [refinamiento/clarify.md](./refinamiento/clarify.md) (sesión 2026-11-19), enunciado §5.2 paso 1. Responsable: Iker Roig · Revisor: Alex Morell. Deltas D15–D17, D9 y D4 de [plan.md](./plan.md). Cubre FR-002 (ampliado), FR-009 (ampliado), FR-010, FR-011, FR-012 y SC-001 (ampliado). Se ejecuta con `/speckit-implement` y se verifica con `/speckit-converge` (§5.3).

### Contrato

- [ ] T043 [US1] **D15, D16** en `specs/spec-001/contracts/openapi.yaml`: en `/alquileres/desbloqueo`, documentar en el 403 los códigos `SIN_ABONO` y `ABONO_VENCIDO` y sus mensajes; en `/alquileres/activo`, añadir a la respuesta 200 la propiedad requerida `abono` (`vigente: boolean`, `codigo: [string, 'null']` con enum `SIN_ABONO`/`ABONO_VENCIDO`, `mensaje: [string, 'null']`). Cambio aditivo: **avisar al equipo antes de empezar esta tarea** (hallazgo K1), además de dejarlo en la PR (T058)

### Tests (escribir antes y comprobar que fallan donde aplique)

- [ ] T044 [P] [US1] **FR-002, SC-001 (D15)**: crear `backend/internal/services/alquiler/causa_test.go` con tests de `causaAbono(u models.Usuario, hoy time.Time)`: `abono_activo = false` → no vigente, `SIN_ABONO`, «No tienes abono anual»; activo sin fecha → `SIN_ABONO`; activo con fecha `2026-11-09` y hoy `2026-11-19` → `ABONO_VENCIDO`, «Tu abono anual caducó el 09/11/2026»; activo con fecha igual a hoy → vigente, código y mensaje vacíos
- [ ] T045 [US1] **FR-002, SC-001 (D15)**: actualizar `backend/tests/integration/spec002_abono_test.go` (gate T1): sin abono → 403 `SIN_ABONO` «No tienes abono anual»; caducado → 403 `ABONO_VENCIDO` con el mensaje y la fecha exactos; vence hoy → 201; en los rechazos, anclaje, bicicleta y viajes sin cambios. Actualizar también `backend/tests/integration/spec002_estacion_test.go` (T037): la revocación pone `abono_activo = false`, así que ahora espera `SIN_ABONO`
- [ ] T046 [P] [US1] **FR-012 (D4)**: crear `backend/tests/integration/spec002_orden_test.go` con un caso por cada par de reglas consecutivas que fallan a la vez: abono caducado + código inexistente → `ABONO_VENCIDO`; viaje activo + código inexistente → `VIAJE_ACTIVO`; estación fuera de servicio + anclaje vacío → «La estación de este anclaje está fuera de servicio»; anclaje fuera de servicio y vacío → «Este anclaje está fuera de servicio»; anclaje vacío en estación operativa → «Este anclaje no tiene ninguna bicicleta». *Test de caracterización: el orden ya existe (P3.5 de las pruebas previas) y puede pasar desde el principio*
- [ ] T047 [US1] **FR-010 (D16)**: ampliar `backend/tests/contract/spec002_test.go` (tras T043): `GET /alquileres/activo` de un usuario con abono vigente devuelve `abono.vigente = true` con `codigo` y `mensaje` a `null`; el de un usuario caducado, `vigente = false`, `codigo = ABONO_VENCIDO` y un `mensaje` **idéntico** al del 403 del `POST`; ambas respuestas cumplen el contrato
- [ ] T048 [P] [US1] **FR-010 (D16)**: ampliar `frontend/tests/alquiler/desbloqueo.test.tsx`: con MSW devolviendo `abono.vigente = false`, la página muestra el `mensaje` con `role="alert"` antes de escribir nada, y el campo y el botón están desactivados (no se hace ningún `POST`); con `vigente = true`, el formulario está activo. Los rechazos `SIN_ABONO` y `ABONO_VENCIDO` del `POST` muestran su mensaje
- [ ] T049 [P] [US1] **FR-011 (D17)**: crear `frontend/tests/alquiler/resultadoDesconocido.test.tsx` con MSW y temporizadores falsos: (a) `POST` con error de red y `GET /activo` con viaje → «No sabemos si se ha desbloqueado; comprobando…» y después la vista del viaje; (b) `POST` 500 y `GET` sin viaje → «No se ha desbloqueado, inténtalo de nuevo» con el formulario activo; (c) `POST` sin respuesta durante más de 10 s → se cancela y se comprueba; (d) `POST` y `GET` fallan → botón «Volver a comprobar», que al pulsarlo repite solo el `GET`; (e) un 4xx (`ANCLAJE_NO_EXISTE`) muestra su causa sin consultar el viaje. En todos los casos se cuenta **un único** `POST`
- [ ] T050 [P] [US1] **FR-009 (D9)**: crear `frontend/tests/alquiler/guardia.test.tsx`: con sesión de `gestor`, `operario_taller`, `operario_logistica` o con token pero sin usuario guardado, navegar a `/alquiler` lleva a `/` y el formulario no llega a mostrarse; con rol `usuario` se muestra
- [ ] T059 [P] [US1] **FR-001, validación técnica (D2, hallazgo U2)**: ampliar `backend/tests/integration/spec002_codigo_test.go`: código vacío o solo espacios → 422 `VALIDACION` «Introduce el código del anclaje»; 65 caracteres o con `\u0000` → 422 `VALIDACION` «El código no es válido»; en ambos, sin efectos

### Implementación

- [ ] T051 [US1] **D15** en `backend/internal/services/alquiler/service.go`: función `causaAbono` (vigente, código y mensaje con la fecha en `DD/MM/AAAA`), errores `ErrSinAbono` y `ErrAbonoCaducado` (este con la fecha) y uso en `validarUsuario` sin cambiar `models.AbonoVigente`; en `backend/internal/handlers/alquiler/handler.go`, mapearlos a 403 `SIN_ABONO` y 403 `ABONO_VENCIDO` con el mensaje de `causaAbono`; en `backend/internal/httperr/httperr.go` (compartido), añadir la constante `SinAbono = "SIN_ABONO"`. Hace pasar T044 y T045
- [ ] T052 [US1] **D16** en `backend/internal/services/alquiler/service.go` y `backend/internal/handlers/alquiler/handler.go` (tras T051): `Activo` lee el abono del usuario y la respuesta de `GET /alquileres/activo` incluye `abono` calculado con `causaAbono`. Hace pasar T047
- [ ] T053 [US1] **D16** en el frontend: tipo `EstadoAbono` y nueva forma de `viajeActivo()` en `frontend/src/services/api/alquiler.ts`; `frontend/src/stores/viaje.ts` guarda también el abono; `frontend/src/pages/alquiler/Desbloqueo.tsx` muestra el aviso y desactiva el formulario; el handler por defecto de `/alquileres/activo` en `frontend/tests/handlers.ts` (compartido) devuelve `abono: { vigente: true, codigo: null, mensaje: null }`. `AvisoViaje.tsx` debe seguir funcionando igual. **No cambiar la firma de `fijarViajeEnCurso(v: ViajeActivo | null)` ni de `viajeEnCurso()`**: las usa spec-003 (`frontend/src/pages/devolucion/index.tsx` y su test); el abono va en una señal aparte del store. Hace pasar T048
- [ ] T054 [US1] **D17** en `frontend/src/services/api/alquiler.ts` (plazo de 10 s con `AbortController` en `desbloquear`) y `frontend/src/pages/alquiler/Desbloqueo.tsx` (estado «comprobando», consulta única del viaje, mensaje de no desbloqueado y botón «Volver a comprobar»; sin reenvío del `POST`). Hace pasar T049
- [ ] T055 [US1] **D9** en `frontend/src/pages/alquiler/Desbloqueo.tsx`: `<Navigate href="/" />` si `sesion.usuario()?.rol !== 'usuario'`, antes de cargar el viaje. No se toca `routes.tsx`. Hace pasar T050
- [ ] T060 [US1] **D2 (hallazgo U2)** en `backend/internal/handlers/alquiler/handler.go`: separar el mensaje del 422 según el motivo (vacío → «Introduce el código del anclaje»; demasiado largo o con caracteres de control → «El código no es válido»). Va en secuencia con T051 (mismo fichero). Hace pasar T059

### Validación

- [ ] T056 [US1] Validación manual de V1, V7 y V10–V14 de `quickstart.md` en el navegador y por API; registrar en `refinamiento/verificacion-final.md` la entrada, el resultado esperado y el observado **antes** (P2.3, P3.3, P3.5, P3.9 y el caso no probado de pérdida de conexión de `pruebas-previas.md`) y **después** de T051–T055 (enunciado §5.3, paso 4)
- [ ] T057 [US1] Regresión de HU-02: `go vet ./...` y `go test -count=1 ./...` en `backend/`; `npx tsc --noEmit`, `npm run lint` y `npm test` en `frontend/`; repetir V2–V9; guardar las salidas en `refinamiento/evidencias/` (enunciado §5.4, «Regresión»)
- [ ] T058 Comunicar al equipo en la PR los cambios en piezas compartidas, todos aditivos: constante `SinAbono` en `httperr`, `SIN_ABONO` y `abono` en el OpenAPI y el handler por defecto de `frontend/tests/handlers.ts` (const. III)

**Checkpoint**: HU-02 cumple FR-001…FR-012 y SC-001…SC-004; T044–T050 y T059 en verde y V1, V10–V14 comprobados en el navegador.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup y Foundational**: sin tareas, ya en `main`.
- **User Story 1**: T009 → tests T010–T015, T028, T030 y T031 → implementación T016–T023 y T029.
- **Polish**: tras completar la US1.
- **Refinamiento (Phase 6)**: T043 → tests T044–T050 y T059 → implementación T051–T055 y T060 → validación T056–T058. T051 antes de T052 (mismo fichero y depende de `causaAbono`); T051 y T060 tocan `handler.go`: en secuencia; T052 antes de T053 (forma de la respuesta); T053, T054 y T055 tocan `Desbloqueo.tsx`: en secuencia. T044, T046, T048, T049, T050 y T059 se pueden escribir en paralelo.

### User Story Dependencies

- **US1**: única historia. Sin dependencias de otras specs; spec-003 cerrará los viajes que esta crea.
- **Refinamiento 2026-11-19 (revisión de dependencias, §5.2 paso 1)**: spec-003 (Alex) usa `fijarViajeEnCurso` y `viajeEnCurso` de `stores/viaje.ts`; T053 conserva sus firmas. Solo la HU-02 consume `GET /alquileres/activo`, así que el campo `abono` no afecta a nadie más. Ninguna otra spec usa `ABONO_VENCIDO` ni `models.AbonoVigente` (comprobado con `grep` en `backend/` y `frontend/`), y `AbonoVigente` no cambia.

### Within Each User Story

- T009 antes de T014, porque el test de contrato valida contra el OpenAPI.
- T010–T015, T028, T030 y T031 se escriben antes de T016–T023 y T029. T010–T015 y T028 deben fallar hasta su implementación; T030 y T031 son de caracterización y pueden pasar ya, porque protegen un comportamiento existente.
- T016, T018, T019 y T020 tocan `service.go`: en secuencia.
- T017 y T023 tocan `handler.go`: en secuencia.
- T019 antes de T021 y T022.

### Parallel Opportunities

- T010, T011, T012 y T013: ficheros distintos.
- T028, T030 y T031: tests en ficheros propios.
- T024: independiente del resto de su fase.

---

## Parallel Example: User Story 1

```text
Task: "Tests de fecha de negocio en backend/internal/services/alquiler/abono_test.go"
Task: "Tests de código en backend/tests/integration/spec002_codigo_test.go"
Task: "Test de carrera con la estación en backend/tests/integration/spec002_estacion_test.go"
Task: "Test del hook tras el commit en backend/tests/integration/spec002_postcommit_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

La spec solo tiene la US1, y su flujo básico ya existe. El MVP se completa con T009–T023 y T028–T031, porque varias tareas cumplen requisitos de la spec aclarada.

### Incremental Delivery

1. Contrato (T009).
2. Tests previos (T010–T015, T028, T030 y T031).
3. Reglas críticas: abono, código y estación (T016–T018).
4. Viaje activo, tiempos tras el commit y menú (T019–T023 y T029).
5. Validación y PR (T024–T027).
6. Refinamiento del 2026-11-19 (Phase 6): contrato (T043), tests (T044–T050, T059), implementación (T051–T055, T060) y validación (T056–T058).

### Parallel Team Strategy

Spec individual de Iker Roig. T009, T020 y T029 tocan piezas compartidas, por lo que deben comunicarse al equipo (T027).

---

## Notes

- `[X]` indica trabajo ya presente en `e090f41`, no generado después.
- No se reescribe el historial: los documentos y los Deltas van en commits posteriores.
- Numeración propia de esta spec; `T033–T039` del tasks.md global siguen siendo la referencia del plan transversal.
