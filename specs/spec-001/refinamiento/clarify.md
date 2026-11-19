# Informe de clarify — spec-001 (HU-01: Consulta de disponibilidad en tiempo real)

**Responsable**: Iker Roig
**Comando**: `/speckit-clarify`
**Spec**: [`specs/spec-001/spec.md`](../spec.md)

La spec-001 ha pasado por dos sesiones de clarificación. La primera se hizo
antes del `/speckit-plan`. La segunda se hizo con la funcionalidad ya
implementada, dentro de la fase de refinamiento: sus respuestas añaden
requisitos nuevos que pueden no estar cubiertos por el código actual, y eso
lo detectarán los pasos siguientes (`/speckit-analyze` y `/speckit-converge`).

## Sesión 2026-11-03 (antes del plan)

5 preguntas, todas aceptadas.

| # | Pregunta | Respuesta | Secciones afectadas |
|---|----------|-----------|---------------------|
| 1 | ¿Desde y hasta cuándo se mide el límite de 30 s? | Extremo a extremo: desde el commit en el servidor hasta que el contador aparece en la pantalla (cliente abierto y con conexión) | FR-003, SC-002 |
| 2 | ¿Cuándo se avisa de que los datos están desactualizados? | Cuando pasan más de 30 s desde la última actualización correcta; los contadores siguen visibles junto con la hora | FR-007, SC-005, casos límite |
| 3 | ¿Cómo se distinguen en el mapa las estaciones «sin bicicletas» y «sin anclajes libres»? | Estado visual propio en el marcador y en el detalle, que no depende solo del color | FR-008, SC-006, casos límite |
| 4 | ¿Qué se ve de una estación fuera de servicio? | Estado «Fuera de servicio», que prevalece sobre los demás, se puede seleccionar y su detalle muestra «—» en lugar de contadores | FR-004, FR-009, SC-007 |
| 5 | ¿Cómo cuenta un anclaje ocupado por una bicicleta averiada? | No cuenta como disponible ni como libre, sin desglose; su suma puede ser menor que la capacidad | FR-006, SC-003 |

## Sesión 2026-11-19 (refinamiento)

4 preguntas, las 4 con la opción recomendada. No se hizo una quinta porque
lo que quedaba pendiente tenía poco impacto.

| # | Pregunta | Respuesta | Cambios en la spec |
|---|----------|-----------|--------------------|
| 1 | ¿Qué se ve si la **primera carga** falla y no hay datos previos? | Mensaje «No se pudo cargar la disponibilidad», botón «Reintentar» y reintento automático, sin marcadores ni contadores | **FR-010** (nuevo) y un caso límite |
| 2 | ¿Hay un **listado de texto** además del mapa (accesibilidad y caída del mapa de fondo)? | Sí: listado siempre disponible con nombre, estado y contadores, que funciona aunque el mapa de fondo no cargue | **FR-011** y **SC-008** (nuevos), y un caso límite |
| 3 | ¿El **detalle abierto** se actualiza solo? | Sí, en ≤ 30 s (contadores y estado), sin cerrarse ni perder la selección | **FR-004** ampliado y un caso límite |
| 4 | ¿Qué pasa **sin sesión** o si la sesión caduca? | El servidor rechaza la consulta; la aplicación redirige al inicio de sesión y después devuelve al usuario al mapa | **FR-012** (nuevo), un caso límite y la suposición ajustada |

## Cobertura tras la sesión 2026-11-19

| Categoría | Estado |
|-----------|--------|
| Alcance funcional y objetivos | Clara |
| Modelo de dominio y datos | Clara |
| Flujo de interacción y UX | Resuelta (detalle en vivo, primera carga fallida) |
| Estados de error, vacío y carga | Resuelta (FR-010) |
| Accesibilidad | Resuelta (FR-011) |
| Rendimiento | Clara (SC-004) |
| Fiabilidad y frescura | Clara (FR-003, FR-007) |
| Seguridad (autenticación) | Resuelta (FR-012) |
| Dependencias externas (mapa de fondo) | Resuelta (FR-011) |
| Casos límite y concurrencia | Clara |
| Terminología | Clara |
| Criterios de aceptación medibles | Clara |
| Observabilidad (logs y métricas) | Pendiente, poco impacto: no lo pide el enunciado |
| Centrado inicial del mapa sin geolocalización | Pendiente, poco impacto: decisión visual que no afecta a datos ni a pruebas |

## Checklist de calidad

`checklists/requirements.md`: 16/16 → 16/16. Ningún ítem cambia de estado.
Los requisitos nuevos se escribieron sin detalles de implementación
(«servidor» y «aplicación» en lugar de backend/frontend, y sin nombrar
tecnologías).

## Impacto esperado en la implementación

Los requisitos FR-010, FR-011 y FR-012, y la ampliación del FR-004, se han
añadido después de implementar. El siguiente paso es `/speckit-analyze` para
comprobar que spec, plan y tasks siguen siendo coherentes, y después
`/speckit-converge` para añadir a `tasks.md` el trabajo que falte.
