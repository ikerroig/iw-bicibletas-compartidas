# Registro de hallazgos del analyze — spec-001 (HU-01)

**Enunciado**: §5.2, paso 3. Se ha contrastado cada hallazgo del
[informe inicial](./analyze.md) con el material del proyecto (spec, plan,
tareas, contrato, código y pruebas en navegador) y se ha decidido qué hacer
con él.
**Responsable**: Iker Roig · **Revisor**: Alex Morell · **Fecha**: 2026-11-19

**Estados**: *Corregido* = el documento ya está corregido · *Tareas añadidas* =
la corrección documental está hecha y el código queda para `/speckit-implement`
(§5.3) · *Aceptado* = no requiere cambios, con la justificación indicada.

| ID | Historia / criterio | Gravedad | Evidencia | Decisión | Responsable | Estado |
|----|---------------------|----------|-----------|----------|-------------|--------|
| G1 | HU-01 · FR-011, SC-008 | ALTA | No hay ninguna tarea que mencione un listado; la prueba P2.7 confirma que no existe (solo hay mapa) | **Hallazgo válido.** Se añade la decisión R21 al plan y las tareas T103 (test) y T104 (implementación) | Iker Roig | Tareas añadidas |
| G2 | HU-01 · FR-010 | MEDIA | `MapaRed.tsx` muestra `{error}. Reintentando…`; la prueba P2.5 ve «Sin datos todavía · Error 500. Reintentando…» sin botón. El test existente «si la primera carga falla lo indica y reintenta» (`mapa.test.tsx`) fija el texto antiguo | **Válido, cumplimiento parcial.** El reintento automático existe; faltan el texto y el botón. Se añaden R20, la tarea T101 (**actualizar** el test existente, no uno nuevo) y la T102 | Iker Roig | Tareas añadidas |
| G3 | HU-01 · FR-012 | MEDIA | `RequireAuth` en las rutas y `alNoAutorizado → cerrarSesion` en `api.ts`; P3.3–P3.6 pasan. **El servidor ya tiene test**: `TestEstacionesExigeSesion` (integración, sin token y con token falso → 401); el analyze solo había mirado `spec001_test.go` (contrato). No hay ningún test de frontend del 401 | **Válido en parte, sobrevalorado por el analyze.** El código ya cumple y la parte del servidor ya está probada; la «vuelta al mapa» sale bien porque `/` es el mapa (P3.5). Solo falta el **test de frontend** del 401 durante el polling: R22 y la tarea T106 | Iker Roig | Tareas añadidas |
| G4 | HU-01 · FR-004 (ampliado) | BAJA | En `MapaRed.tsx`, `seleccionada` sale de `red.estaciones()`; P1.4 y P3.2 pasan. **Ya existe** el test «el detalle se actualiza con los datos en vivo de la estación seleccionada» (`mapa.test.tsx`), que cubre el cambio de contadores | **Falso positivo en su mayor parte**: el analyze no encontró ese test. Solo falta el caso del paso a fuera de servicio con el detalle abierto: la tarea T105 **amplía** el test existente | Iker Roig | Tareas añadidas |
| I1 | HU-01 · fase 3 de `tasks.md` | MEDIA | El objetivo decía «Cubre FR-001…FR-009 y SC-001…SC-007» | Se cambia a «FR-001…FR-013 y SC-001…SC-008» | Iker Roig | Corregido |
| I2 | HU-01 · `plan.md` | MEDIA | Solo existía la nota de actualización del 2026-11-03 | Se añaden la nota del 2026-11-19, `ListaEstaciones` al árbol, una fila al re-check de la constitución y R20–R23 en `research.md` | Iker Roig | Corregido |
| I3 | HU-01 · `quickstart.md` | BAJA | No había escenarios para FR-010–FR-012 | Se añaden S1f–S1h y S1i (FR-013) | Iker Roig | Corregido |
| I4 | Transversal · `tasks.md` (fases 4–10) | BAJA | Las casillas de las fases 4–10 están sin marcar, pero `specs/spec-002…008/tasks.md` tienen todas sus tareas hechas (excepto 2 en spec-007) | **Falso positivo en cuanto al trabajo** (no falta nada), **válido en cuanto a la claridad**. No se marcan tareas de otros integrantes; se añade una nota que indica la fuente de verdad del estado | Iker Roig | Corregido |
| A1 | HU-01 · FR-011 («p. ej., pestaña «Lista»») | BAJA | La spec no fija la forma de acceso | **Aceptado.** La spec fija qué se ve y cuándo; la forma de acceso es una decisión de diseño (R21: conmutador «Mapa / Lista»). Que sea comprobable queda asegurado por T103 | Iker Roig | Aceptado |
| H1 | HU-01 · encuadre inicial | — (de las pruebas, no del analyze) | `pruebas-previas.md`: con la ubicación concedida y fuera del área, los 60 marcadores quedan fuera de la pantalla | Era una **decisión funcional pendiente**, así que se volvió a **clarify** antes de tocar el plan (§5.2, paso 4): pregunta 5 → FR-013. Después, R23 y las tareas T107 y T108 | Iker Roig | Tareas añadidas |

## Dependencias con historias de otros integrantes

- **spec-007 (Alberto Copa)** consume `EstacionEstado` (contadores enteros
  también en estaciones fuera de servicio, R18). R20–R23 no cambian el
  contrato, así que no le afecta.
- **spec-002/003/005/006** llaman a `RecalcularEstacion` tras sus cambios; el
  refinamiento no toca ese helper.
- **Sin dependencias nuevas.** No hace falta acuerdo de equipo (principio III).

## Resumen

| | Total | Válidos | Parciales / solo pruebas | Falsos positivos | Aceptados sin cambio |
|---|---|---|---|---|---|
| Hallazgos del analyze | 9 | 5 (G1, G2, I1, I2, I3) | 1 (G3: solo falta un test de frontend) | 2 (G4 casi entero, ya había test; I4 en cuanto al trabajo) | 1 (A1) |
| Hallazgos de las pruebas | 1 | 1 (H1) | — | — | — |

**Lección**: el analyze trabaja sobre `spec.md`, `plan.md` y `tasks.md` y deduce la cobertura de pruebas por palabras clave; no lee los ficheros de test. Por eso dos hallazgos (G3, G4) sobrevaloraban lo que faltaba. Contrastarlos con el código evitó añadir tareas redundantes (se eliminó un test de contrato 401 que ya existía).

Después de las correcciones no queda ningún hallazgo **crítico** ni **alto**
sin resolver en el plano documental. G1–G4 y H1 tienen tareas y pruebas
previstas; su cierre en el código se verificará con `/speckit-converge`
(§5.3).

## Revisión entre compañeros (§5.1 y §5.2) — a rellenar por Alex Morell

> El enunciado pide que el revisor contraste las decisiones del clarify, elija
> **al menos un criterio** de la historia y compruebe que la tarea propuesta
> permitiría verificarlo. Si no hay incidencias reales, se documenta la
> cobertura comprobada, sin inventarse errores.

| Qué se revisó | Criterio / decisión | Cómo se comprobó | Resultado | Fecha |
|---------------|---------------------|------------------|-----------|-------|
| Decisiones del clarify | _(p. ej., pregunta 5 → FR-013)_ | _(lectura de `clarify.md` y la narrativa)_ | _(de acuerdo / discrepancia y motivo)_ | |
| Cobertura de un criterio en la matriz | _(p. ej., FR-010)_ | _(tarea, test y escenario de `docs/matriz-trazabilidad.md`; ejecución del test)_ | | |
| Registro de hallazgos | _(p. ej., G3 / G4, falsos positivos)_ | _(contraste con el código)_ | | |

**Firma del revisor**: _(nombre y fecha)_
