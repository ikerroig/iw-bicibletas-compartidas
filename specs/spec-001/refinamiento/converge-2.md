# Informe de converge 2 — spec-001 (HU-01)

**Responsable**: Iker Roig · **Revisor**: _(compañero)_
**Fecha**: 2026-11-19 · **Comando**: `/speckit-converge` (segunda ejecución,
enunciado §5.3 paso 5: «Repetir converge tras las correcciones»)
**Estado de partida**: commit `a42bbbf` (T111–T114). Tests: 81/81 del frontend
([evidencias/tests-despues-convergencia.txt](./evidencias/tests-despues-convergencia.txt)),
backend en verde. **Informe anterior**: [converge-1.md](./converge-1.md).

## Hallazgos de converge 1

| ID | Tarea | Evidencia del cierre | Estado |
|----|-------|----------------------|--------|
| F1 — el listado pierde el foco en cada refresco | T111 | `ListaEstaciones.tsx` itera sobre los `id` (`idsOrdenados`) y lee cada estación de `porId()`. Test «un refresco actualiza la fila enfocada sin volver a crearla ni perder el foco»: **falla antes** (la fila enfocada se queda con «7 bicicletas» porque ya no está en la página, [tests-antes-convergencia.txt](./evidencias/tests-antes-convergencia.txt)) y **pasa después** | Cerrado |
| F2 — «Datos desactualizados» sin datos | T112 | `AvisoDesactualizado.tsx` solo muestra el aviso si `ultimaActualizacionOk !== null`. Hay dos tests (con el arnés y con `MapaRed`, a los 31 s sin datos) que fallan antes y pasan después; el test de R17 con datos previos sigue pasando | Cerrado |
| F3 — `aria-controls` sin panel | T113 | Los dos `tabpanel` existen siempre; el inactivo lleva `hidden` (`[hidden] { display: none }` del preflight de Tailwind 3, así que no recibe el foco). El test de `aria-controls` falla antes y pasa después | Cerrado |
| F4 — leyenda sin justificar | T114 | Justificada en `research.md` (R18): explica los iconos con el texto de la spec y refuerza el «no solo color» del FR-008. Se mantiene | Cerrado (justificado) |

## Nueva revisión

Se han vuelto a revisar los 21 criterios (13 FR + 8 SC), los 8 escenarios de
aceptación, las decisiones R3/R4 y R16–R23 y los principios I–V de la
constitución contra el código actual. **No hay hallazgos nuevos.** En concreto
se ha comprobado que T111–T113 no rompen:

- el aviso de datos desactualizados **con** datos previos (FR-007, test de 30/31 s);
- la selección compartida entre listado, mapa y detalle (FR-004, FR-011);
- la frescura del listado, que sigue leyendo el mismo store (FR-003, SC-008).

## Resultado

**✅ `converged`**: la implementación satisface la spec, el plan y las tareas
de HU-01. `tasks.md` no se modifica en esta ejecución.

> Un resultado limpio del converge **no** sustituye las pruebas ni la
> demostración funcional (§5.3). Quedan pendientes: **T109** (rehacer en el
> navegador P2.5, P2.7 y H1 para tener el «después», escenarios S1f–S1i) y
> **T110** (regresión: `go test`, `npm test` y S1–S1e).

## Métricas de la iteración de convergencia

| | Converge 1 | Converge 2 |
|---|---|---|
| Hallazgos | 4 (0 críticos, 0 altos, 2 medios, 2 bajos) | **0** |
| Tareas añadidas | 4 (T111–T114) | 0 |
| Tests del frontend | 78/78 | 81/81 |
