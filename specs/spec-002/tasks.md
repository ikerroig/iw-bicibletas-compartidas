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

## Phase 4: Polish & Cross-Cutting Concerns

- [X] T024 [P] Actualizar la regla de abono de `specs/spec-001/data-model.md` (línea de `CURRENT_DATE`) para indicar la fecha actual en `Europe/Madrid`
- [X] T025 Ejecutar `go vet ./...`, `go test -count=1 ./...` en `backend/` y `npx tsc --noEmit`, `npm run lint`, `npm test` en `frontend/`; repetir 10 veces los tests de concurrencia de `backend/tests/integration/spec002_concurrencia_test.go`
- [X] T026 Ejecutar V1–V8 de `specs/spec-002/quickstart.md` en navegador y API, y anotar el resultado en la PR
- [ ] T027 Incluir en la PR los cambios aditivos de `specs/spec-002/contracts/alquileres.md`, la zona horaria del abono (a revisar por spec-004 y spec-008) y la nota de trazabilidad de `specs/spec-002/plan.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup y Foundational**: sin tareas, ya en `main`.
- **User Story 1**: T009 → tests T010–T015, T028, T030 y T031 → implementación T016–T023 y T029.
- **Polish**: tras completar la US1.

### User Story Dependencies

- **US1**: única historia. Sin dependencias de otras specs; spec-003 cerrará los viajes que esta crea.

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

### Parallel Team Strategy

Spec individual de Iker Roig. T009, T020 y T029 tocan piezas compartidas, por lo que deben comunicarse al equipo (T027).

---

## Notes

- `[X]` indica trabajo ya presente en `e090f41`, no generado después.
- No se reescribe el historial: los documentos y los Deltas van en commits posteriores.
- Numeración propia de esta spec; `T033–T039` del tasks.md global siguen siendo la referencia del plan transversal.
