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

_(Pendiente.)_

## HU-04 — Historial de viajes y desglose mensual (spec-004) · Alex Morell · Revisor: Iker Roig

Decisiones: `specs/spec-004/research.md` (D1–D8) + clarify 2026-11-19 (C1–C3) · Tareas: `specs/spec-004/tasks.md` (T001–T021).

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
| Regresión y UX del refinamiento (aviso de filtro incompleto, suites) | Clarify C3 (primera carga sin filtro) | T012, T019 | Vitest: «filtro a medias…» + `npm test` (82/82) tras el merge de `main` · checklist con *Acceptance Coverage* | ✅ |

**Cobertura HU-04**: 8 FR + 4 SC = 12 criterios; **12/12 con tarea y
prueba prevista, 12 ✅** (refinamiento 2026-11-19: AC4–AC6 cerrados con
T016–T019; evidencia en `specs/spec-004/refinamiento/converge-2.md`).

## HU-05 — Notificación de averías e inhabilitación (spec-005) · Josep Peiró · Revisor: Alberto Copa

_(Pendiente.)_

## HU-06 — Intervenciones técnicas y taller (spec-006) · Josep Peiró · Revisor: Alberto Copa

_(Pendiente.)_

## HU-07 — Alertas de balanceo de flota (spec-007) · Alberto Copa · Revisor: Josep Peiró

_(Pendiente.)_

## HU-08 — Analítica de uso y exportación CSV (spec-008) · Alberto Copa · Revisor: Josep Peiró

_(Pendiente.)_
