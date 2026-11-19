# Informe de clarify — spec-001 (HU-01: Consulta de disponibilidad en tiempo real)

**Responsable**: Iker Roig · **Revisor**: _(compañero que contrasta las decisiones)_
**Comando**: `/speckit-clarify` · **Spec Kit**: 1.1.3
**Spec**: [`specs/spec-001/spec.md`](../spec.md)
**Commit de referencia** (antes del refinamiento): `44b5050`
**Diferencia entre la spec anterior y la revisada**: `git diff 44b5050 -- specs/spec-001/spec.md`

La spec-001 ha pasado por dos sesiones de clarificación. La primera se hizo
antes del `/speckit-plan`. La segunda se hizo dentro de la fase de
refinamiento (enunciado §5.1), con la funcionalidad ya implementada.

**Fuente de las justificaciones**: la descripción narrativa del sistema
(`SI-16-bicicletas-compartidas-descripcion-narrativa.pdf`, en adelante
«narrativa») y la constitución del proyecto. Cuando la narrativa no determina
una única respuesta, se indica como **decisión del grupo** y se explica el
motivo. Ninguna opción se ha elegido porque la aplicación ya funcionase así:
en la sesión del 2026-11-19, tres de las cinco respuestas obligan a cambiar el
código (FR-010, FR-011 y FR-013).

## Sesión 2026-11-03 (antes del plan)

| # | Pregunta | Respuesta | Justificación | Criterios |
|---|----------|-----------|---------------|-----------|
| 1 | ¿Desde y hasta cuándo se mide el límite de 30 s? | Extremo a extremo: desde el commit en el servidor hasta que el contador aparece en la pantalla | La narrativa fija el retraso «respecto a lo que ocurre en la estación», es decir, lo que ve el usuario. Medirlo solo en el servidor dejaría fuera el refresco del cliente | FR-003, SC-002 |
| 2 | ¿Cuándo se avisa de que los datos están desactualizados? | Cuando pasan más de 30 s desde la última actualización correcta; los contadores siguen visibles | Decisión del grupo derivada de la narrativa: si se tolera un máximo de 30 s, pasado ese tiempo el dato ya no cumple y el usuario debe saberlo. Se mantienen los contadores porque un dato viejo marcado es más útil que una pantalla vacía | FR-007, SC-005 |
| 3 | ¿Cómo se distinguen «sin bicicletas» y «sin anclajes libres»? | Estado visual propio en el marcador y en el detalle, que no depende solo del color | La narrativa pide saber «dónde coger una bicicleta como dónde dejarla»; son dos decisiones distintas y cada una necesita su señal. Que no dependa solo del color es una decisión del grupo por accesibilidad | FR-008, SC-006 |
| 4 | ¿Qué se ve de una estación fuera de servicio? | Estado «Fuera de servicio» que prevalece, se puede seleccionar y muestra «—» en lugar de contadores | La narrativa pide «un mapa con todas las estaciones», así que no se oculta. Mostrar contadores induciría a planificar sobre una estación inutilizable (decisión del grupo) | FR-004, FR-009, SC-007 |
| 5 | ¿Cómo cuenta un anclaje ocupado por una bicicleta averiada? | No cuenta como disponible ni como libre, sin desglose | La narrativa: la averiada «deja de contar como disponible en el mapa». Tampoco es un anclaje libre, porque físicamente no se puede dejar una bici en él | FR-006, SC-003 |

## Sesión 2026-11-19 (refinamiento, §5.1)

**Pruebas previas**: [pruebas-previas.md](./pruebas-previas.md), con 17
comprobaciones en tres situaciones (habitual, límite o inválida, y permisos o
estado). Se ejecutaron después de las preguntas 1–4 y antes de la 5, que
surge de una de ellas (hallazgo H1). Este desorden respecto al §5.1 queda
registrado aquí.

| # | Pregunta | Respuesta | Justificación | Criterios | ¿Cambia el código? |
|---|----------|-----------|---------------|-----------|--------------------|
| 1 | ¿Qué se ve si la **primera carga** falla y no hay datos previos? | Mensaje «No se pudo cargar la disponibilidad», botón «Reintentar» y reintento automático, sin marcadores ni contadores | La narrativa pide información «prácticamente al día», y un mapa vacío sin explicación se confundiría con «no hay estaciones». Recuerda que el sistema actual «genera muchas reclamaciones», así que el mensaje debe ser claro. El botón se añade (decisión del grupo) para que el usuario no dependa solo del reintento automático | FR-010 | Sí (P2.5 falla) |
| 2 | ¿Hay un **listado de texto** además del mapa? | Sí: listado siempre disponible con nombre, estado y contadores | La narrativa pide un mapa, pero su finalidad es saber dónde hay bicis y anclajes. Un listado da la misma información a quien no puede usar el mapa (accesibilidad) o cuando falla el proveedor externo de teselas. No amplía el alcance: son los mismos datos de la HU-01 (decisión del grupo) | FR-011, SC-008 | Sí (P2.7 falla) |
| 3 | ¿El **detalle abierto** se actualiza solo? | Sí, en ≤ 30 s, sin cerrarse ni perder la selección | El límite de 30 s de la narrativa se aplica a toda la información que ve el usuario, y el detalle es donde decide adónde ir. Un detalle congelado incumpliría ese límite | FR-004 | No (P1.4 y P3.2 ya cumplían) |
| 4 | ¿Qué pasa **sin sesión** o si la sesión caduca? | El servidor rechaza la consulta; la aplicación lleva al login y después vuelve al mapa | La narrativa no lo dice explícitamente. **Decisión del grupo**: la aplicación es para usuarios con abono («hay que tener un abono anual activo») y la constitución (principio I) exige que el control esté en el servidor. Se descartó hacer el mapa público (opción B) porque abre los datos sin necesidad y crea dos modos de uso que probar | FR-012 | No (P3.3–P3.6 ya cumplían) |
| 5 | ¿Dónde se centra el mapa si el usuario está **fuera del área de la red**? | Solo en el usuario si está dentro del área; si está fuera, se encuadra toda la red | La narrativa: el usuario mira el mapa «antes de salir de casa» para ver «dónde hay bicicletas». Un mapa centrado lejos de las estaciones no responde a eso (hallazgo H1: las 60 estaciones quedaban fuera de la pantalla). Se mantiene el centrado en el usuario cuando está en la ciudad porque le ayuda a ver las estaciones cercanas. Se descartó la opción B (no centrar nunca) porque pierde esa ventaja | FR-013 | Sí (H1) |

**Historia afectada**: todas las preguntas afectan solo a la HU-01. Ninguna
cambia el contrato de la API, el esquema de PostgreSQL ni las claves de Redis,
así que no afectan a historias de otros integrantes (constitución, principio
III).

**Decisión que explica el alumno** (§5.1, finalización): _(elegir al menos
una; se proponen la 2 o la 5 de la sesión del 2026-11-19, porque son las que
más se apartan de la implementación original)_

## Cobertura tras la sesión 2026-11-19

| Categoría | Estado |
|-----------|--------|
| Alcance funcional y objetivos | Clara |
| Modelo de dominio y datos | Clara |
| Flujo de interacción y UX | Resuelta (detalle en vivo, primera carga fallida, encuadre inicial) |
| Estados de error, vacío y carga | Resuelta (FR-010) |
| Accesibilidad | Resuelta (FR-011) |
| Rendimiento | Clara (SC-004) |
| Fiabilidad y frescura | Clara (FR-003, FR-007) |
| Seguridad (autenticación) | Resuelta (FR-012) |
| Dependencias externas (mapa de fondo, geolocalización) | Resuelta (FR-011, FR-013) |
| Casos límite y concurrencia | Clara |
| Terminología | Clara |
| Criterios de aceptación medibles | Clara (escenarios de aceptación 4–8 en formato Dado/Cuando/Entonces) |
| Observabilidad (logs y métricas) | Pendiente, poco impacto: no lo pide el enunciado |

## Checklist de calidad

`checklists/requirements.md`: 16/16 → 16/16. Ningún ítem cambia de estado.
Los requisitos nuevos se escribieron sin detalles de implementación.

## Siguiente paso

Fase de analyze (§5.2): actualizar `plan.md`, `tasks.md` y `quickstart.md`
con FR-010…FR-013, hacer el registro de hallazgos del informe inicial
[analyze.md](./analyze.md) y repetir el analyze.
