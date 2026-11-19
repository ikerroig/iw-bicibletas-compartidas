# Informe de converge 1 — spec-001 (HU-01)

**Responsable**: Iker Roig · **Revisor**: _(compañero)_
**Fecha**: 2026-11-19 · **Comando**: `/speckit-converge` (Spec Kit 1.1.3)
**Estado de partida**: commit `49f5233` (implement de T101–T108). Tests: 78/78
del frontend y todos los paquetes del backend en verde
([evidencias/tests-despues.txt](./evidencias/tests-despues.txt)).
**Alcance**: HU-01, es decir, `spec.md`, `plan.md`, `research.md` (R16–R23) y
`tasks.md` (fase 3 y los helpers de la fase 2 que usa: T011, T014, T020 y
T021), contrastados con el código de `backend/internal/{handlers,routes,services}/estacion`,
`backend/internal/{cache,workers/refresco_estado}`, `frontend/src/{stores/red.ts,components/MapaBase.tsx,pages/mapa/}`
y sus tests.

> Enunciado §5.3, paso 2: se distinguen trabajo ausente (*missing*),
> cumplimiento parcial (*partial*), comportamiento contradictorio
> (*contradicts*) y código sin justificación (*unrequested*). Una tarea marcada
> como hecha **no** se ha tomado como prueba de cumplimiento: se ha leído el
> código de cada criterio.

## Hallazgos

| ID | Tipo | Severidad | Origen | Evidencia | Tarea |
|----|------|-----------|--------|-----------|-------|
| F1 | partial | MEDIA | FR-011, escenario 5 | `ListaEstaciones.tsx` usa `<For each={ordenadas()}>`. `<For>` de Solid identifica cada fila por la referencia del objeto, y `red.ts` hace `setEstaciones(s.estaciones)` con objetos recién leídos del JSON en cada refresco, así que **cada 15 s se vuelven a crear todas las filas** y quien navega con teclado o lector de pantalla pierde el foco. El listado es precisamente la alternativa accesible del FR-011. *El mapa no tiene este problema: `MapaBase` sigue los marcadores por `id`.* Comprobado leyendo el código; la extensión del navegador no estaba disponible. La T111 empieza con un test que reproduce el fallo | T111 |
| F2 | partial | MEDIA | FR-010, FR-007 | Sin ninguna carga correcta, pasados 30 s `AvisoDesactualizado.tsx` muestra «Datos desactualizados · Mostramos los últimos datos conocidos; reintentando…» **a la vez** que el mensaje del FR-010, cuando no hay ningún dato que mostrar. El test «sin ninguna carga correcta, avisa igualmente pasados 30 s» (`desactualizado.test.tsx`), escrito con R17 antes de existir el FR-010, da por bueno ese comportamiento | T112 |
| F3 | partial | BAJA | FR-011 (accesibilidad) | Con la vista «Mapa» activa, la pestaña «Lista» declara `aria-controls="panel-lista"`, pero `panel-lista` solo existe con la vista «Lista» (`<Show when={vista() === 'lista'}>`). La referencia ARIA apunta a un elemento que no existe | T113 |
| F4 | unrequested | BAJA | plan R18 | `MapaRed.tsx` incluye una `Leyenda` de estados en el panel lateral que no aparece en la spec, el plan ni las tareas. Es coherente con el FR-008 («no solo color»), pero no está justificada en los documentos | T114 |

## Revisado sin hallazgos

| Ámbito | Elementos | Comprobación |
|--------|-----------|--------------|
| Requisitos funcionales | FR-001…FR-013 | FR-001–FR-009 y FR-012: código y tests de la fase 3 y P1–P3 de las pruebas previas. FR-010, FR-011 y FR-013: T101–T108, salvo F1–F3 |
| Criterios de éxito | SC-001…SC-008 | SC-004 es de validación manual (T100, S1e); se repetirá en la T110 |
| Escenarios de aceptación | 1…8 | Cada uno tiene un test o escenario S1x asociado (ver `docs/matriz-trazabilidad.md`) |
| Decisiones del plan | R3/R4 (TTL de 120 s, `cache/keys.go`), R16 (`IntervaloBarrido = 10 * time.Second`), R17, R18, R19, R20, R21, R22, R23 | Valores y comportamiento conforme a lo escrito; R22: `RequireAuth` en `routes/spec_001_routes.go` |
| Contrato | `GET /estaciones` (200/401) y `GET /estaciones/{id}` (200/401/404) | Un id no numérico responde 404 `NO_ENCONTRADO`, dentro de lo que define el contrato |
| Constitución | Principios I–V | Sin incumplimientos: los contadores se calculan solo en el backend (I), no hay tecnologías nuevas (II), no cambian el esquema ni Redis (III), frescura ≤ 30 s (IV) y los cambios no salen de los paquetes de spec-001 (V) |

## Métricas

- Requisitos y criterios revisados: 21 (13 FR + 8 SC) y 8 escenarios de aceptación
- Decisiones del plan revisadas: 8 bloques (R3/R4, R16–R23)
- Principios de la constitución: 5
- Hallazgos por tipo: 0 *missing* · 3 *partial* · 0 *contradicts* · 1 *unrequested*
- Hallazgos por severidad: 0 críticos · 0 altos · 2 medios · 2 bajos

## Resultado

**`tasks_appended`**: 4 tareas (T111–T114) en la **Phase 12: Convergence** de
`tasks.md`. Ninguna procede de un requisito ambiguo, así que no hace falta
volver a clarify (§5.3, paso 3).

## Siguiente paso

`/speckit-implement` sobre T111–T114, empezando por los tests que reproducen F1
y F2. Después, `/speckit-converge` otra vez (`converge-2.md`) y, cuando salga
limpio, T109 (antes y después en el navegador) y T110 (regresión).
