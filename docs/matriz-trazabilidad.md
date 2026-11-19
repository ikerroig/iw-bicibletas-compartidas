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
| SC-004 — mapa < 5 s y detalle < 2 s | Plan, *Performance Goals* | T100 | S1e: en local sin limitar, marcadores ≤ 263 ms y detalle 8 ms (`verificacion-final.md`); falta repetirlo con «Fast 4G» | 👁 |
| Regresión de todo lo anterior tras el refinamiento | — | T109, T110 | `go test ./...` (9 ok), `npm test` (81/81), S1–S1i (`refinamiento/verificacion-final.md`) | ✅ |

**Cobertura HU-01**: 13 FR + 8 SC = 21 criterios; **21/21 con tarea y
prueba prevista**. Tras el refinamiento (2026-11-19): **20 ✅** y 1 👁
(SC-004: cumple de sobra en local; falta repetir la medición con «Fast 4G»).

## HU-02 — Desbloqueo de bicicleta por código de anclaje (spec-002) · Iker Roig · Revisor: Alex Morell

_(Pendiente: se rellena tras el clarify y el analyze de spec-002.)_

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

_(Pendiente.)_

## HU-05 — Notificación de averías e inhabilitación (spec-005) · Josep Peiró · Revisor: Alberto Copa

_(Pendiente.)_

## HU-06 — Intervenciones técnicas y taller (spec-006) · Josep Peiró · Revisor: Alberto Copa

_(Pendiente.)_

## HU-07 — Alertas de balanceo de flota (spec-007) · Alberto Copa · Revisor: Josep Peiró

_(Pendiente.)_

## HU-08 — Analítica de uso y exportación CSV (spec-008) · Alberto Copa · Revisor: Josep Peiró

_(Pendiente.)_
