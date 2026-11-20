# Análisis inicial — spec-008 / HU-08

**Fecha**: 2026-11-20 · **Rama**: `refinamiento/spec-008` · **Commit base**: `f8197e91a218288ad2facddf079513801f7967d4` · **Spec Kit**: `1.1.3`
**Momento del análisis**: después de la propuesta provisional de añadir ranking por estación y antes de la decisión final del equipo de conservar el alcance simple.

## Hallazgos iniciales

| ID | Categoría | Gravedad | Evidencia | Recomendación inicial |
|---|---|---|---|---|
| I1 | Inconsistencia | HIGH | La propuesta de ranking estaba en FR-002/AC3, pero no en la descripción de US1 ni en Independent Test. | Alinear historia y prueba con el ranking o retirar la propuesta. |
| I2 | Inconsistencia | HIGH | `plan.md` solo describe dos gráficos y dos buckets temporales. | Actualizar plan, arquitectura y rendimiento si se mantenía el ranking. |
| C1 | Contrato/cobertura | HIGH | OpenAPI global y `models.AnaliticaViajes` no tienen campo de ranking por estación. | Proponer un campo; obtener aprobación del equipo antes de tocar el OpenAPI global. |
| C2 | Cobertura de tareas y pruebas | HIGH | T004, T007–T009 y T014 solo cubren el modelo, consultas, pruebas y UI existentes por hora/día. | Añadir trabajo y pruebas específicos si se mantenía el ranking. |
| U1 | Ambigüedad | MEDIUM | No estaba decidido si estaciones con cero salidas aparecían en el ranking. | Definir el comportamiento vacío y la composición del ranking. |
| I3 | Consistencia | MEDIUM | FR-006 y T8a solo comparan gráficos con CSV, no el ranking provisional. | Extender la consistencia y las pruebas si se mantenía el ranking. |

## Cobertura inicial

Se revisaron 7 requisitos funcionales, 4 criterios de éxito y 20 tareas. Los requisitos tenían al menos una tarea asociada, pero FR-002 estaba parcialmente cubierto por la ampliación provisional; el ranking no tenía implementación ni pruebas. La cobertura por identificador era 11/11, con una brecha funcional parcial en FR-002.

## Resolución posterior

El equipo eligió no ampliar el contrato compartido y mantener el alcance más simple, coincidente con `historia.md`: gráficos por hora y día de semana, sin ranking por estación. Las propuestas I1–I3 y C1–C2 quedan cerradas por retirada de ese alcance provisional; U1 deja de aplicar. Ver [hallazgos.md](./hallazgos.md) y el análisis final.
