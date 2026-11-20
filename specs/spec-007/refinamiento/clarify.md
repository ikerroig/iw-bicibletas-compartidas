# Clarify — registro de decisiones (spec-007 / HU-07)

**Fecha**: 2026-11-20
**Rama**: `refinamiento/spec-007`
**Commit de partida**: `f8197e91a218288ad2facddf079513801f7967d4` (merge-base con `main`, HEAD actual)
**Versión de Spec Kit**: 1.1.3
**HU afectada**: HU-07 — Monitorización de desequilibrios y alertas de balanceo de flota (única historia de spec-007)
**Espec**: [spec.md](../spec.md) · Sesión de clarify: `## Clarifications → Session 2026-11-20`

## Verificación de cobertura (3 casos por historia, HU-07)

| Caso exigido | Estado | Evidencia (pregunta / escenario) |
|--------------|--------|----------------------------------|
| Recorrido normal | Cubierto sin ambigüedad | Escenarios de aceptación 1–3 de US1 + *Independent Test* ya definían el ciclo alerta → panel → confirmación de traslado → recálculo; ninguna pregunta necesaria |
| Entrada inválida o caso límite | Cubierto con pregunta | **Q1** (destino sin capacidad → rechazo) y **Q3** (empate de urgencia en el orden del panel) |
| Permisos o estado | Cubierto con pregunta | **Q2** (acceso solo operario) y **Q4** (estado vacío del panel) |

Resultado: la sesión de clarify de HU-07 cumple lo pedido en la revisión final (4 de 5 preguntas de cuota usadas; el caso normal no requería decisión nueva).

## Preguntas y respuestas

### Q1 — Capacidad del destino en el traslado

- **Pregunta**: When a confirmed transfer would push the destination station above 100% of its capacity, should the system reject it or accept it anyway?
- **Respuesta**: Reject the confirmation; destination must have room for all incoming units.
- **Justificación**: La spec solo rechazaba el exceso de origen; sin esta regla, un destino desbordado corrompería el cómputo de ocupación y las pruebas de aceptación del panel. Rejecting mantiene ambos contadores consistentes con la validación de origen ya existente.
- **HU afectada**: HU-07
- **Criterio cambiado (Dado… cuando… entonces…)**: Dado un traslado confirmado con destino cuya capacidad no admite todas las unidades entrantes, cuando el operario confirma la operación, entonces el sistema rechaza la confirmación completa sin alterar las estaciones de origen ni de destino.
- **Reflejado en spec**: FR-006 (validación de origen y destino) y Edge Cases («Un lote que excedería la capacidad del destino…»).

### Q2 — Acceso al panel de logística

- **Pregunta**: Should access to the logistics panel be restricted so that only the operario role can view and confirm transfers, with other roles denied?
- **Respuesta**: Yes, enforce operario-only access; other roles get denied (403/error view).
- **Justificación**: La spec decía «vista exclusiva del operario» solo como suposición, no como requisito; sin FR explícita, la frontera de seguridad no era verificable en tests. La respuesta añade el requisito sin inventar alcance fuera de HU-07.
- **HU afectada**: HU-07
- **Criterio cambiado (Dado… cuando… entonces…)**: Dado un usuario autenticado sin rol de operario, cuando intenta abrir el panel de logística o confirmar un traslado, entonces el sistema deniega el acceso con vista de error.
- **Reflejado en spec**: FR-009 (nuevo) y Assumptions (acceso denegado a otros roles según FR-009).

### Q3 — Desempate en el orden del panel

- **Pregunta**: When two unbalanced stations are equally critical (e.g., both at 0% occupancy), what tie-break rule should order them in the panel?
- **Respuesta (original)**: Tie-break by station ID (ascending).
- **Respuesta (enmienda 2026-11-20, hallazgo H-001)**: Tie-break by `fecha_generacion` ASC (alerta más antigua primero).
- **Justificación de la enmienda**: la respuesta original entraba en conflicto con data-model.md §3, `handlers/balanceo/handler.go`, quickstart.md (L104, ya validado ✅) y el código implementado, que todos ordenan el empate por `fecha_generacion` ASC. Al ser el criterio ya implementado, testeado y determinista igualmente, se enmienda la spec (opción (b) decidida por Alberto) en lugar de reabrir código; ver hallazgos.md H-001.
- **Justificación**: SC-002 exige un orden concreto del panel; sin desempate determinista, el mismo estado podía producir distintos fallos según la implementación. Cualquiera de los dos criterios es estable y asertible en tests; se conserva el ya validado.
- **HU afectada**: HU-07
- **Criterio cambiado (Dado… cuando… entonces…)**: Dado que dos o más estaciones descompensadas tienen la misma criticidad de urgencia, cuando el operario abre el panel, entonces se muestran ordenadas por `fecha_generacion` ascendente (la alerta más antigua primero).
- **Reflejado en spec**: FR-005 (regla de desempate por `fecha_generacion` ASC) y bullet de Clarifications enmendado.

### Q4 — Estado vacío del panel

- **Pregunta**: When no stations are unbalanced, what should the logistics panel display?
- **Respuesta**: Empty-state message (e.g., "Sin estaciones descompensadas").
- **Justificación**: El caso más habitual de cada turno (red equilibrada) no tenía expectación de aceptación; un mensaje vacío explícito lo distingue de un estado de carga o error y es fácil de asertar en tests.
- **HU afectada**: HU-07
- **Criterio cambiado (Dado… cuando… entonces…)**: Dado que ninguna estación está descompensada, cuando el operario abre su panel de logística, entonces ve un mensaje de estado vacío (p. ej. «Sin estaciones descompensadas») y no una tabla con filas.
- **Reflejado en spec**: US1 → Acceptance Scenarios, escenario 4.

## Acuerdo inicial de descarte (revocado por decisión posterior del equipo)

- **T021** — tabla de auditoría `traslados_balanceo`: la primera decisión registrada el 2026-11-20 fue no crearla.
- **T022** — score de urgencia del panel: la primera decisión registrada el 2026-11-20 fue mantener el cálculo en la consulta, sin índice.

## Enmienda posterior: acuerdo de implementación del equipo (2026-11-20)

El equipo revocó la decisión inicial y aprobó implementar ambas tareas en el esquema compartido, con los ordinales siguientes a las migraciones existentes:

- **T021**: crear `traslados_balanceo` y registrar cada traslado exitoso (incluido el `operario_id`) en la misma transacción que mueve las bicicletas.
- **T022**: crear un índice parcial de expresión sobre el `CASE` de urgencia y `fecha_generacion`; se conserva la consulta y el orden ya implementados.

La implementación queda reflejada en `tasks.md`, `data-model.md`, `spec-001/data-model.md` y `hallazgos.md`. T021/T022 se validaron con la suite completa de Go contra una PostgreSQL temporal aislada; T7d, T7f y T7g comprueban el índice, la auditoría exitosa y la ausencia de auditoría en rechazos.
