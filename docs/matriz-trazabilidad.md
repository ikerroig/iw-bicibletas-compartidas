# Matriz de trazabilidad — SI-16 (8 historias)

**Enunciado**: §5.2, paso 5. Cada fila relaciona una historia y un criterio
con la decisión del plan, las tareas que lo cubren y la prueba prevista. Cada
integrante rellena las filas de sus dos historias, y **otro integrante** las
revisa: comprueba al menos un criterio por historia y que la tarea propuesta
permitiría verificarlo.

**Revisión cruzada** (acordada el 2026-11-19): **Iker Roig ↔ Alex Morell** y
**Josep Peiró ↔ Alberto Copa**. Cada pareja revisa las dos historias del otro.
Las historias de cada pareja van seguidas en el flujo del sistema
(desbloqueo → devolución → historial; avería → taller → balanceo y
analítica), lo que también sirve para los recorridos entre integrantes del
§5.4. Para que la revisión sea comprobable, el revisor deja por escrito qué
criterio concreto comprobó y cómo, no solo «revisado».

**Leyenda de pruebas**: `Go:` test de backend (`backend/tests/…`) · `Vitest:`
test de frontend (`frontend/tests/…`) · `S1x`: escenario manual de
`specs/spec-001/quickstart.md` · `P…`: comprobación de
`specs/spec-001/refinamiento/pruebas-previas.md` (evidencia «antes»).
**Estado**: ✅ cubierto y probado · 🛠 tareas añadidas en el refinamiento,
pendientes de implementar · 👁 solo validación manual.

## HU-01 — Consulta de disponibilidad en tiempo real (spec-001) · Iker Roig · Revisor: Alex Morell

Decisiones: `specs/spec-001/research.md` (Rn) · Tareas: `specs/spec-001/tasks.md` (Tnnn).

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001, SC-001 — mapa con todas las estaciones | R9 (snapshot desde Redis), `GET /estaciones` | T027, T030, T031 | Vitest: «pinta TODAS las estaciones de la API (SC-001)» · S1 · P1.1 | ✅ |
| FR-002 — contadores por estación | R9 (`RecalcularEstacion`, fuente única) | T020, T027, T031 | Go: `TestMutacionActualizaHashTrasCommit` · Vitest: `DetalleEstacion` · P1.2 | ✅ |
| FR-003, SC-002 — ≤ 30 s de extremo a extremo | R3, R16 (barrido de 10 s + polling de 15 s, peor caso ≈ 28 s) | T025, T026, T029, T099 | Go: `TestBarridoReflejaMutacionesSinRecalculoSincrono`, `TestRedisCaidoNoRevierteElCommitYElBarridoRepara` · S1, S1d · P1.3 (≈ 16 s) | ✅ |
| FR-004 — detalle seleccionable y en vivo | R18; el detalle sale del store (R22) | T027, T031, **T105** | Vitest: «el detalle se actualiza con los datos en vivo…» (+ paso a fuera de servicio, T105) · P1.4, P3.2 | ✅ |
| FR-005 — datos en vivo, no estáticos | R3, R4 (TTL de 120 s + fallback a PostgreSQL) | T011, T025 | Go: `TestClaveExpiradaRecalculaDesdePostgresYRepuebla`, `TestSnapshotFuncionaConRedisCaido` | ✅ |
| FR-006, SC-003 — averiadas fuera de los contadores | R9, R19 | T020, T025, T031 | Go: `TestAveriadasNoSumanEnNingunContador`, `TestTodasAveriadasDaCeroYCero` · Vitest: estadoVisual (suma < capacidad) · P2.2, P2.3 | ✅ |
| FR-007, SC-005 — aviso de datos desactualizados | R17 | T029, T097, T098 | Vitest: `desactualizado.test.tsx` (30 s sin aviso, 31 s con aviso) · S1a | ✅ |
| FR-008, SC-006 — «sin bicicletas» y «sin anclajes libres» | R18 (`estadoVisual`, icono + texto) | T030, T031, T095, T096 | Vitest: `estadoVisual.test.ts`, `DetalleEstacion` 0/0 · S1b · P2.1, P2.2 | ✅ |
| FR-009, SC-007 — fuera de servicio | R18 (prevalece, «—», sin acciones) | T025, T030, T031, T095 | Go: `TestFueraDeServicioDevuelveEnterosEnLaAPI` · Vitest: «fuera de servicio → «—»…», «marcador seleccionable» · S1c · P3.1, P3.2 (11,7 s) | ✅ |
| FR-010 — primera carga fallida | R20 | **T101**, **T102**, **T112** | Vitest: primera carga (T101) y sin datos > 30 s (T112) · S1f ✅ (antes P2.5 ❌) | ✅ |
| FR-011, SC-008 — listado textual | R21 | **T103**, **T104**, **T111**, **T113** | Vitest: `listaEstaciones.test.tsx` (incl. foco y `aria-controls`) · S1g ✅ (antes P2.7 ❌) | ✅ |
| FR-012 — acceso sin sesión o con sesión caducada | R5, R22 | T014, T021, **T106** | Go: `TestEstacionesExigeSesion` (401) · Vitest: `sesionCaducada.test.tsx` (T106) · S1h ✅ · P3.3–P3.6 | ✅ |
| FR-013 — encuadre si el usuario está fuera del área | R23 | **T107**, **T108** | Vitest: `encuadre.test.tsx` (posición antes/después de los datos, dentro/fuera) · S1i ✅ (antes H1 ❌) | ✅ |
| SC-004 — mapa < 5 s y detalle < 2 s | Plan, *Performance Goals* | T100 | S1e con «Fast 4G» y caché vacía (`verificacion-final.md`, `evidencias/s1e-fast4g/`): marcadores ≤ 0,85 s en producción y ≤ 4,02 s en desarrollo; detalle ≤ 73 ms; con la ubicación denegada, 60/60 marcadores en la vista | ✅ |
| Regresión de todo lo anterior tras el refinamiento | — | T109, T110 | `go test ./...` (9 ok), `npm test` (81/81), S1–S1i (`refinamiento/verificacion-final.md`) | ✅ |

**Cobertura HU-01**: 13 FR + 8 SC = 21 criterios; **21/21 con tarea y
prueba prevista**. Tras el refinamiento (2026-11-19): **21 ✅** (SC-004
medido también con el perfil «Fast 4G», en `verificacion-final.md`).

## HU-02 — Desbloqueo de bicicleta por código de anclaje (spec-002) · Iker Roig · Revisor: Alex Morell

Decisiones: `specs/spec-002/research.md` (Dn) · Tareas: `specs/spec-002/tasks.md` (Tnnn) · Escenarios manuales: `Vn` de `specs/spec-002/quickstart.md` · Evidencia «antes»: `Pn.n` de `specs/spec-002/refinamiento/pruebas-previas.md`. Los tests de backend están en `backend/tests/integration/spec002_*_test.go`, salvo que se indique otro sitio.

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001 — código sin distinguir mayúsculas ni espacios | D2 | T011, T017 | Go: `TestCodigoSinDistinguirMayusculasNiEspacios` · V2 · P2.8 | ✅ |
| FR-001 (validación técnica) — mensaje de código mal formado | D2 (mensajes, hallazgo U2) | T011, T017, **T059**, **T060** | Go: `TestCodigoInvalidoEsErrorDeValidacion` + ampliación T059 (vacío vs. «El código no es válido») · V7 ✅ (antes P2.3: mensaje engañoso) | ✅ |
| FR-002, SC-001 — abono vigente y causa concreta | D1, **D15** | T001, T010, T016, **T044**, **T045**, **T051** | Go: `TestFechaNegocio`, `TestAbonoVenceALaMedianocheDeEspana`, `TestDesbloqueoConAbonoQueVenceHoy`, `TestDesbloqueoRechazaAbonoNoVigente` (T045: `SIN_ABONO` / `ABONO_VENCIDO` con fecha), `causa_test.go` (T044) · V1 ✅ (antes P3.3: mismo mensaje) | ✅ |
| FR-003, SC-002 — un solo viaje activo | R7 global, `ux_viajes_usuario_activo` | T002 | Go: `TestSegundoDesbloqueoRechazado`, `TestDesbloqueosConcurrentesMismoUsuario` · V4 · P2.9, P3.6 | ✅ |
| FR-004, SC-004 — anclaje existe, en servicio y con bici disponible | D4, D5 | T002, T012, T018 | Go: `TestDesbloqueoRechazosDeAnclajeSinEfectos`, `TestDesbloqueoCodigoInexistente`, `TestDesbloqueoEsperaAlCambioDeEstacion`, `TestDesbloqueosConcurrentesMismoAnclaje` · V5 · P2.1, P2.4–P2.7, P2.10 | ✅ |
| FR-005 — desacoplar, hora de inicio y viaje | Plan, *Summary* (transacción única) | T004 | Go: `TestDesbloqueoValidoCreaViajeYActualizaEstado` · V2, V8 · P1.2, P1.6 | ✅ |
| FR-006 — viaje asociado hasta la devolución | D14, `stores/viaje.ts` | T004, T007, T041 | Vitest: «muestra el viaje activo en lugar del formulario» · V3 · P1.4 | ✅ |
| FR-007 — atomicidad | D11, D12, D13 | T002, T036, T037, T039, T040 | Go: `TestConexionPerdidaNoDejaViajeParcial`, `TestCommitSobreviveALaCancelacion`, `TestDesbloqueoEsperaALaRevocacionDelAbono`, `TestDesbloqueoTienePlazoAnteBloqueoRetenido` | ✅ |
| FR-008 — datos del viaje al desbloquear y al volver | D3, D14 | T014, T015, T019, T021, T022, T038, T041, T042 | Go: `TestContratoAlquileres` · Vitest: «muestra inicio, código de bicicleta y estación…», `aviso.test.tsx` · V2, V9 · P1.2, P1.4, P1.5 | ✅ |
| FR-009 — solo el rol `usuario`; menú y acceso por URL | D9 (corregida) | T028, T029, **T050**, **T055** | Go: `TestDesbloqueoRequiereRolUsuario` · Vitest: `menu.test.tsx` + `guardia.test.tsx` (T050) · V6, V13 ✅ (antes P3.9: formulario visible) | ✅ |
| FR-010 — aviso del abono al abrir la pantalla | **D16** | **T047**, **T048**, **T052**, **T053** | Go: `TestContratoAlquileres` ampliado (T047, mismo texto que el 403) · Vitest: `desbloqueo.test.tsx` ampliado (T048) · V10, V11 ✅ (antes P3.2: el aviso solo llegaba al enviar) | ✅ |
| FR-011 — resultado desconocido del desbloqueo | **D17** (+ D12) | **T049**, **T054** | Vitest: `resultadoDesconocido.test.tsx` (red caída, 5xx, > 10 s, consulta fallida, un único `POST`) · V12a, V12b ✅ (antes: no probado) | ✅ |
| FR-012 — una sola causa, en orden fijo | D4 | **T046** | Go: `spec002_orden_test.go` (T046, caracterización) · V14 ✅ · P3.5 | ✅ |
| SC-003 — desbloqueo en < 10 s | D6, D11 | T013, T030, T035 | Go: `TestDesbloqueoEnMenosDe10s`, `TestDesbloqueoConRedisDegradado`, `TestHookRecibeContextoVigenteTrasCancelacion` · P1.3 (24–32 ms) | ✅ |
| Regresión de todo lo anterior tras el refinamiento | — | **T056**, **T057** | `go test ./...` (todo ok), `npm test` (96/96), V1–V14 (`specs/spec-002/refinamiento/verificacion-final.md`) | ✅ |

**Cobertura HU-02**: 12 FR + 4 SC = 16 criterios; **16/16 con tarea y
prueba prevista**. Tras implementar el refinamiento (2026-11-19, commit
`058b3bd`): **16 ✅**, incluida la mejora técnica del mensaje de FR-001 (U2).
`/speckit-converge` (§5.3): la primera pasada añadió T061–T063 (FR-011: plazo en
la consulta de comprobación) y la segunda dio **converged**
(`specs/spec-002/refinamiento/converge-2.md`).

## HU-03 — Devolución y tarificación (spec-003) · Alex Morell · Revisor: Iker Roig

Decisiones: `specs/spec-003/research.md` (D1–D7) · Tareas: `specs/spec-003/tasks.md` (T001–T033).

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001, FR-002 — devolución: bici fijada en anclaje libre, hora de llegada, usuario libre | D2 (transacción única con `FOR UPDATE` y orden de cierre) | T017, T019 | Go: `TestDevolucion_ViajeCorto_FinalizaYGratuito` · S3 | ✅ |
| FR-003, SC-001 — ≤ 30 min → 0,00 € | D1 (`CalcularCoste`, umbral 1800 s) | T014, T016 | Go: `TestCalcularCoste_ViajeCortoGratuito`, `TestCalcularCoste_Borde30Minutos` · S3 | ✅ |
| FR-004, SC-002 — solo el excedente × tarifa única, minuto empezado (ceil) | D1 (`ceil((s−1800)/60) × TARIFA_MINUTO`) | T016, **T030** | Go: `TestCalcularCoste_45Minutos`, `TestCalcularCoste_DosHoras`, `TestCalcularCoste_FraccionesDeExcedente` (46:30 → 3,40) · S3 | ✅ |
| FR-005 — persiste inicio, llegada, duración y coste (entrada de la HU-04) | D2 (`UPDATE viajes` en la misma transacción) | T017 | Go: asertos de `fecha_fin`/`duracion_minutos`/`coste_total` en `TestDevolucion_ViajeCorto_FinalizaYGratuito` | ✅ |
| FR-006 — un solo cierre por viaje; bici en un solo anclaje | D2 (V3: `rows=0` → `VIAJE_YA_CERRADO`) | T015, T017 | Go: `TestDevolucion_DobleConfirmacion_SoloUnaEfecto` (2 concurrentes → 1×200 + 1×409) | ✅ |
| FR-007 — sin viaje activo → 409 y anclaje intacto | D3 (códigos de error) | T015, T017 | Go: `TestDevolucion_SinViajeActivo_RechazaSinEfectos`, `TestDevolucion_AnclajesInvalidos_Rechazan` | ✅ |
| FR-008 — solo rol `usuario` propietario: 401 sin sesión / 403 otro rol, sin efectos | D7 (rol en la ruta) | T020, **T028** | Go: `TestDevolucion_Permisos_SinSesionYRolNoUsuario` | ✅ |
| FR-009 — el abono caducado no impide devolver | Clarify 2026-11-19 (AC5) | **T029** | Go: `TestDevolucion_AbonoCaducado_PuedeDevolver` | ✅ |
| SC-003 — devolución válida < 10 s | Plan, *Performance Goals* | T015 (T3j), T026 | Go: latencia en `TestDevolucion_ViajeCorto_FinalizaYGratuito` · S3 | ✅ |
| SC-004 — desbloqueo inmediato posterior; 0 dobles cierres | D2 + D4 (invalidación Redis post-commit) | T015, T018, T026 | Go: T3f · S3 (`actualizado_en` nuevo) | ✅ |
| Regresión del refinamiento (contrato ≤64, UI, suites) | Contract parcial (converge F1) | **T032**, T033, T024, T025 | Go: `TestDevolucion_CodigoLargo_ValidacionContrato` · `go test ./...` + `npm test` (81/81) tras el merge de `main` | ✅ |

**Cobertura HU-03**: 9 FR + 4 SC = 13 criterios; **13/13 con tarea y
prueba prevista, 13 ✅** (refinamiento 2026-11-19: AC4/AC5/AC6 cerrados con
T028–T031; evidencia en `specs/spec-003/refinamiento/converge-2.md`).

## HU-04 — Historial de viajes y desglose mensual (spec-004) · Alex Morell · Revisor: Iker Roig

Decisiones: `specs/spec-004/research.md` (D1–D8) + clarify 2026-11-19 (C1–C3) · Tareas: `specs/spec-004/tasks.md` (T001–T025).

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001 — lista cronológica con fecha, estaciones, duración, coste; orden DESC | D1 (Q1 con JOIN a estaciones) | T005, T006 | Go: `TestHistorialT4a` (orden estricto + nombres) · Vitest: «render con datos…» | ✅ |
| FR-002 — filtro mes/año interpretado en Europe/Madrid | D2 (`AT TIME ZONE 'Europe/Madrid'`) | T005 | Go: `TestHistorialT4bFiltroMes`, `TestHistorialT4gZonaHoraria` | ✅ |
| FR-003 — sumatorio del periodo o del historial completo | D1 (Q2 agregado sin paginar) | T005 | Go: T4b (3,75), T4e/T4f (10,75 en cualquier página) · Vitest: sumatorio | ✅ |
| FR-004 — etiqueta «Tramo gratuito (0,00 €)» calculada en el backend | D3 (`EtiquetaParaCoste`, solo `coste = 0`) | T004, T005 | Go: `TestHistorialT4cEtiqueta` · Vitest: etiqueta tal cual | ✅ |
| FR-005 — solo el historial propio desde el JWT; 401 sin sesión; cualquier rol | D4 (aislamiento por usuario + `RequireAuth`) | T007, **T016** | Go: `TestHistorialT4dAislamientoY401` (401 `NO_AUTENTICADO`), `TestHistorialT4kRoles` (operario/gestor → 200 propio; parámetro ajeno ignorado) | ✅ |
| FR-006 — origen y destino por viaje, sin GPS (alcance del clarify) | Assumptions + D5 (campos aditivos) | T005 | Go: T4a (`estacion_*_nombre`) · S4 | ✅ |
| FR-007 — coste y duración tal como los registró la HU-03; sin recálculo | D3 (solo lectura) | T004 | Go: T4a · S4 (sin recálculo) | ✅ |
| FR-008 — paginación server-side 50/200, suma del periodo, fuera de rango sin error | D6 (offset + página vacía con `total`/`suma`) | T005, T006 | Go: `TestHistorialT4ePaginacion`, `T4fPaginaFueraDeRango`, `T4iEntradaInvalida` | ✅ |
| SC-001 — 100 % de viajes finalizados con campos correctos | D1 (filtro `estado='finalizado'`) | T004 | Go: T4a (el `en_curso` de la fixture no cuenta) | ✅ |
| SC-002 — filtros exactos y sumatorio = suma del periodo | D1, D2 | T005 | Go: T4b/T4e/T4f · S4 | ✅ |
| SC-003 — viajes sin coste muestran la etiqueta | D3 | T004 | Go: T4c · S4 | ✅ |
| SC-004 — < 3 s con un año completo de uso | D1 (dos consultas indexadas, sin N+1) | **T017** | Go: `TestHistorialT4jLatencia` con fixture de ≥ 365 viajes (0,5 s medidos) | ✅ |
| Regresión, UX y convergencia (contrato 500, aserciones error.codigo, aviso de filtro incompleto, suites) | Clarify C3 + Converge F1–F4 | T012, T019, **T022–T025** | Vitest: «filtro a medias…» + `npm test` (82/82) tras el merge de `main` · Go: T4d/T4i con `error.codigo` · checklist y contrato actualizados | ✅ |

**Cobertura HU-04**: 8 FR + 4 SC = 12 criterios; **12/12 con tarea y
prueba prevista, 12 ✅** (refinamiento y convergencia 2026-11-19: AC4–AC6 y F1–F4 cerrados con
T016–T025; evidencia en `specs/spec-004/refinamiento/converge-2.md`).

## HU-05 — Notificación de averías e inhabilitación (spec-005) · Josep Peiró · Revisor: Alberto Copa

Decisiones: `specs/spec-005/research.md` (C-001–C-004) · Tareas: `specs/spec-005/tasks.md` · Escenario: `specs/spec-005/quickstart.md`.

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001 — reporte con descripción y estación real | Servicio valida bicicleta, sesión y estación; una bicicleta en viaje no exige estación anclada | T004, T005, T007, T011 | Go: `TestAveriaRechazaEstacionDistinta` · recorrido de reporte en la demo | ✅ |
| FR-002 — cambio inmediato a averiada | Transacción con bloqueo de la bicicleta y actualización atómica | T004, T009, T013 | Go: `TestAveriaBloqueaYEsIdempotente` | ✅ |
| FR-003 — averiada no alquilable ni disponible | El estado `averiada` impide el desbloqueo y se excluye del recuento | T004, T009, T013 | Go: `TestAveriaBloqueaYEsIdempotente` · quickstart HU-05 | ✅ |
| FR-004 — exclusión del mapa y contadores | Recalcular la estación tras confirmar el reporte | T004, T009, T013, T018 | Go: `TestAveriadasNoSumanEnNingunContador`, `TestTodasAveriadasDaCeroYCero` | ✅ |
| FR-005 — incidencia pendiente visible al taller | Persistir incidencia y exponer listado solo a taller | T004, T005, T008, T012, T018 | Go: `TestAveriaPanelSoloTaller` · demo HU-05 → panel de taller | ✅ |
| FR-006 — confirmación al usuario | El formulario presenta el mensaje devuelto por la API | T005, T007, T010, T014 | Recorrido manual y demo: confirmación tras «Registrar avería» | ✅ |
| FR-007 — reporte repetido idempotente | Índice de incidencia abierta y tratamiento explícito del reintento | T002, T004, T009, T013 | Go: `TestAveriaBloqueaYEsIdempotente` | ✅ |
| FR-008 — refresco del panel < 30 s | Refresco automático cada 10 s y botón «Actualizar» | T008, T012, T018 | Quickstart HU-05 · demo con incidencia visible inmediatamente | ✅ |
| SC-001 — 100 % de reportes cambian el estado | La mutación y la incidencia comparten transacción | T004, T009, T013 | Go: `TestAveriaBloqueaYEsIdempotente` | ✅ |
| SC-002 — panel < 30 s y 0 alquileres de averiadas | Polling de 10 s y validación server-side del alquiler | T004, T008, T012, T018 | Go: panel solo taller · quickstart · demo funcional | ✅ |
| SC-003 — mapas/listas excluyen averiadas | Una única definición de bicicleta operativa | T004, T013, T018 | Go: tests de contadores de HU-01 | ✅ |
| SC-004 — 0 estados averiada/disponible inconsistentes | Bloqueo, transacción e idempotencia | T002, T004, T009, T013 | Go: `TestAveriaBloqueaYEsIdempotente` | ✅ |

**Cobertura HU-05**: 8 FR + 4 SC = **12/12 con tarea y prueba prevista; 12 ✅**.

## HU-06 — Intervenciones técnicas y taller (spec-006) · Josep Peiró · Revisor: Alberto Copa

Decisiones: `specs/spec-006/research.md` (C-001–C-005) · Tareas: `specs/spec-006/tasks.md` · Escenario: `specs/spec-006/quickstart.md`.

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001 — ficha con reporte e historial | Consulta conjunta de bicicleta, incidencias e intervenciones | T003, T008–T010, T017 | Go: `TestTallerFichaReparacionYUbicacion` · Vitest de taller · demo | ✅ |
| FR-002 — parte con piezas, notas y operario | Validación de campos y autor obtenido del JWT | T004, T007–T010, T017 | Go: `TestTallerValidaRolYCampos` · Vitest de taller | ✅ |
| FR-003 — cambio a operativa e intervención firmada | Guardado transaccional del parte, fecha y operario | T004, T005, T009, T011 | Go: `TestTallerFichaReparacionYUbicacion` | ✅ |
| FR-004 — historial completo | No se sobrescriben intervenciones; la ficha devuelve todo el historial | T003, T004, T009 | Go: ficha tras guardar el parte · demo funcional | ✅ |
| FR-005 — ubicación en anclaje libre y disponibilidad | Reparación y ubicación son pasos separados; la ubicación recalcula la estación | T006, T008–T010, T017 | Go: `TestTallerFichaReparacionYUbicacion` · Vitest · demo | ✅ |
| FR-006 — incidencia atendida deja de estar pendiente | El parte resuelve las incidencias abiertas de la bicicleta | T004, T005, T009 | Go: `TestTallerFichaReparacionYUbicacion` | ✅ |
| FR-007 — operativa sin ubicación no disponible | La bicicleta no cuenta hasta confirmar un anclaje | T004, T006, T009, T017 | Go: reparación y ubicación en dos pasos · quickstart HU-06 | ✅ |
| SC-001 — 100 % de fichas muestran el reporte original | La ficha conserva y presenta la incidencia de HU-05 | T003, T009, T017 | Go + demo: reporte visible antes de reparar | ✅ |
| SC-002 — 100 % de reparaciones con autor y fecha | Operario del JWT y fecha de servidor obligatorios | T004, T007, T009 | Go: `TestTallerFichaReparacionYUbicacion` | ✅ |
| SC-003 — disponible ≤ 30 s tras ubicar | Recalcular estación después del commit | T006, T009, T017 | Go + quickstart + demo: ubicación confirmada | ✅ |
| SC-004 — 0 operativas con incidencia pendiente | Parte e incidencia se actualizan en la misma transacción | T004, T005, T009 | Go: incidencia resuelta tras reparación | ✅ |

**Cobertura HU-06**: 7 FR + 4 SC = **11/11 con tarea y prueba prevista; 11 ✅**.

## HU-07 — Alertas de balanceo de flota (spec-007) · Alberto Copa · Revisor: Josep Peiró

Decisiones: `specs/spec-007/research.md` (D1–D8) · Tareas: `specs/spec-007/tasks.md` · Pruebas T7a–T7h: `specs/spec-007/quickstart.md`.

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001 — porcentaje de ocupación | Bicicletas disponibles ancladas / capacidad total | T005, T009 | T7a, T7c | ✅ |
| FR-002 — alertas estrictas <10 % / >90 % | Evaluación inline y worker periódico | T005, T007, T009, T010 | T7a, T7h | ✅ |
| FR-003 — una alerta abierta por estación/umbral | Índice parcial y `ON CONFLICT DO NOTHING` | T007, T009 | T7b | ✅ |
| FR-004 — resolución automática al normalizar | Evaluación resuelve alertas fuera de condición | T007, T009 | T7b · S7 | ✅ |
| FR-005 — panel ordenado por urgencia | Score descendente y antigüedad ascendente en empate | T007, T012, T016, T022 | T7d · demo HU-07 | ✅ |
| FR-006 — traslado validado y atómico | Bloqueo de estaciones, capacidad y auditoría en transacción | T006, T007, T011, T021 | T7f, T7g · demo éxito/error | ✅ |
| FR-007 — recálculo inmediato y global ≤ 30 s | Respuesta recalculada y Redis post-commit | T011, T016, T019 | T7f · S7 · demo | ✅ |
| FR-008 — averiadas excluidas | La consulta solo cuenta disponibles ancladas | T005, T007, T009 | T7c | ✅ |
| FR-009 — acceso exclusivo de logística | Autenticación y `RequireRole("operario_logistica")` | T007, T013, T016 | T7e | ✅ |
| SC-001 — alerta en < 60 s | Worker cada 55 s con barrido < 5 s; inline en traslado | T007, T010, T019 | T7h · S7 | ✅ |
| SC-002 — panel ordenado, sin duplicados/obsoletas | Consulta de abiertas y orden único | T007, T008, T016, T022 | T7b, T7d · demo | ✅ |
| SC-003 — traslado visible inmediatamente y global ≤ 30 s | Respuesta con ambos estados + Redis | T011, T019 | T7f · S7 | ✅ |
| SC-004 — 0 alertas obsoletas | Resolución automática en cada evaluación | T007, T009, T019 | T7b · S7 | ✅ |

**Cobertura HU-07**: 9 FR + 4 SC = **13/13 con tarea y prueba prevista; 13 ✅**.

## HU-08 — Analítica de uso y exportación CSV (spec-008) · Alberto Copa · Revisor: Josep Peiró

Decisiones y evidencia: `specs/spec-008/refinamiento/trazabilidad.md` · Tareas: `specs/spec-008/tasks.md`.

| Criterio | Decisión del plan | Tareas | Prueba prevista | Estado |
|----------|-------------------|--------|-----------------|--------|
| FR-001 — rango inclusivo Europe/Madrid | Validación backend y filtro por fecha local de inicio | T005, T007, T009, T014 | `service_test.go`, T8b, T8f | ✅ |
| FR-002 — 24 horas y 7 días | El servidor calcula buckets y el frontend pinta los arrays | T004, T007–T009, T014, T018, T020 | T8c · tests de página · demo | ✅ |
| FR-003 — CSV del mismo rango | Exportación backend y descarga autenticada | T006–T007, T010–T013, T021 | T8a, T8e, T8h · demo | ✅ |
| FR-004 — formato CSV interoperable | BOM UTF-8, RFC 4180, CRLF, columnas y nombre fijados | T006–T007, T010–T011, T017 | Tests CSV · T8h · S8 | ✅ |
| FR-005 — acceso exclusivo de gestor | `RequireAuth` y `RequireRole("gestor")` en ambas rutas | T007, T011–T012, T014, T021 | T8g · demo de 403 y CSV autenticado | ✅ |
| FR-006 — gráficos y CSV consistentes | Un filtro común alimenta agregados y exportación | T007, T009–T011, T016 | T8a · S8 | ✅ |
| FR-007 — viajes finalizados de toda la red | Consultas sin usuario y con estado finalizado | T007, T009–T010 | T8d | ✅ |
| SC-001 — gráficos < 5 s hasta un año | Consultas indexadas y prueba cronometrada | T007, T016–T017 | T8i · quickstart | ✅ |
| SC-002 — CSV completo y consistente | Predicado compartido y comparación automatizada | T007, T010–T011, T016–T017 | T8a · S8 | ✅ |
| SC-003 — 100 % de accesos no gestor denegados | Autorización server-side | T007, T012, T016–T017 | T8g · demo funcional | ✅ |
| SC-004 — CSV sin pérdida de caracteres | Serialización Unicode con BOM | T006–T007, T017 | T8h · comprobación manual S8 | ✅ |

**Cobertura HU-08**: 7 FR + 4 SC = **11/11 con tarea y prueba prevista; 11 ✅**.
