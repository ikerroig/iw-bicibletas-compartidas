# Research: Historial de viajes y desglose mensual HU-04 (Phase 0)

**Branch**: `spec-004` | **Date**: 2026-10-28 | **Plan**: [plan.md](./plan.md)

Investigación propia de la HU-04. **No repite** las decisiones globales ya resueltas en [specs/spec-001/research.md](../spec-001/research.md) (R1 pila, R2 esquema, R3 Redis, R5 auth, R6 migraciones): esta spec las **adopta tal cual** y solo investiga los puntos abiertos en su `spec.md` (ya con las 5 clarificaciones de la sesión 2026-10-28). La HU-04 es **solo lectura**: no hay concurrencia de escritura ni invalidación de Redis. Resultado: **0 NEEDS CLARIFICATION**.

---

## D1. Consulta del historial: dos sentencias indexadas (página + agregado)

- **Decisión**: `GET /viajes` ejecuta **dos consultas** contra PostgreSQL (sin transacción: lectura pura):
  1. **Página**: `SELECT v.*, eo.nombre, ed.nombre FROM viajes v JOIN estaciones eo ON eo.id = v.estacion_origen_id JOIN estaciones ed ON ed.id = v.estacion_destino_id WHERE v.usuario_id = $1 AND v.estado = 'finalizado' [AND filtro temporal] ORDER BY v.fecha_inicio DESC LIMIT $page_size OFFSET $offset`.
  2. **Agregado del periodo**: `SELECT count(*), COALESCE(sum(coste_total), 0) FROM viajes v WHERE v.usuario_id = $1 AND v.estado = 'finalizado' [AND filtro temporal]` — **sin** `LIMIT/OFFSET`: el sumatorio y el `total` son siempre del periodo completo (FR-003, FR-008, SC-002), independientes de la página visible.
  - `offset = (page − 1) × page_size`.
- **Racional**: el índice `ix_viajes_usuario_fecha (usuario_id, fecha_inicio DESC)` del modelo compartido cubre a la vez el filtro de usuario y el `ORDER BY … DESC`, y el rango temporal se resuelve dentro del mismo índice. Un año de uso habitual son decenas/filas de viajes → `OFFSET` pequeño; SC-004 (<3 s) se cumple con holgura. Dos consultas simples son más legibles y testeables que un `CTE` con window functions.
- **Alternativas consideradas**: `COUNT(*) OVER()` en la misma sentencia (rechazada: obliga a leer la página completa antes de paginar y complica el OFFSET); traer todo y paginar/pensar en el cliente (descartada en la clarificación 1: paginación server-side); caché Redis (rechazada: datos históricos no cambian y añadiría invalidación sin beneficio).

## D2. Filtro temporal `desde`/`hasta` interpretado en Europe/Madrid (FR-002)

- **Decisión**: el contrato global define `desde`/`hasta` con `format: date` (`YYYY-MM-DD`). El backend filtra con:
  `AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $desde AND $hasta`
  (ambos inclusivos). Sin parámetros → sin cláusula → historial completo (FR-003). El frontend construye el rango del mes seleccionado: `desde = primer día del mes`, `hasta = último día del mes` (la UI filtra por mes/año como exige FR-002; la traducción a rango es un detalle de contrato).
- **Racional**: la conversión `AT TIME ZONE 'Europe/Madrid'` se hace **en la BD**, donde vive FR-002 como regla del sistema (const. I): un viaje iniciado el día 31 a las 23:30 hora española cae en el mes que ve el usuario, sea cual sea el huso del servidor o del navegador. `TIMESTAMPTZ` almacena instantes, así que la comparación es exacta y usa `ix_viajes_usuario_fecha`.
- **Alternativas consideradas**: que el cliente envíe el desfase y el backend lo aplique (rechazada: delega una regla de filtrado al cliente y ensucia el contrato); UTC puro (rechazada en la clarificación 2: desplazaría viajes de última hora al mes siguiente respecto a lo que ve el usuario); columna generada/triggers (rechazada: cambio de esquema, const. III).

## D3. Etiqueta "Tramo gratuito (0,00 €)" calculada al leer (FR-004, FR-007)

- **Decisión**: el handler de historial rellena `etiqueta = "Tramo gratuito (0,00 €)"` cuando `coste_total = 0` para cada item `finalizado`. No se persiste: `etiqueta` no es columna (ya lo definió spec-003 en `models.Viaje` como `json:"etiqueta,omitempty"`). El coste y la duración se devuelven **tal cual** los registró el cierre de HU-03 (FR-007): el historial **nunca recalcula**.
- **Racional**: la condición es el propio coste (0,00 €) y no una recomputación de duración ≤ 30 min, de modo que la regla crítica de la constitución I se consume, no se duplica. Un único punto (backend) produce la etiqueta para todo consumidor (HU-04, HU-08).
- **Alternativas consideradas**: que el frontend detecte `coste_total === 0` y lo etiquete (rechazado: const. I — presentación de una regla de negocio en el cliente); columna persistida (rechazada: cambio de esquema innecesario, la etiqueta se deriva).

## D4. Aislamiento por usuario y rol (FR-005)

- **Decisión**: el `usuario_id` de la consulta es **siempre** `auth.UsuarioID(c)` (claim del JWT) y nunca un parámetro de entrada. Autorización: `RequireAuth` **sin** `RequireRole` — cualquier rol autenticado consulta su propio historial (operarios y gestores también pueden ser usuarios del servicio; el spec dice "usuario con sesión iniciada", no "rol usuario"). Sin token → 401.
- **Racional**: FR-005 ("nunca los viajes de otros usuarios") queda garantizado estructuralmente: no existe forma de pedir el historial ajeno. Mínimo privilegio cumplido (solo lo suyo) sin denegar el acceso a roles que sí pueden haber realizado viajes.
- **Alternativas consideradas**: `RequireRole("usuario")` como en `POST /devoluciones` (rechazada: la devolución es una *acción* del ciclo de alquiler restringida a ese rol, mientras que el historial es una *consulta* de los propios viajes sin efectos); parámetro `usuario_id` opcional para gestores (rechazado: ampliaría alcance más allá de la HU-04 y abriría lectura de datos ajenos).

## D5. Contrato de respuesta: `{ total, suma_coste, items }` + nombres de estación aditivos

- **Decisión**: se adopta literalmente la respuesta del OpenAPI global: `200` con `{ total: int, suma_coste: number, items: Viaje[] }` (`total` = viajes de **todo** el periodo — habilita el control de páginas —; `suma_coste` = coste sumado de **todo** el periodo). Cada item es el schema `Viaje` global **más** dos campos aditivos resueltos con JOIN: `estacion_origen_nombre` y `estacion_destino_nombre`.
- **Racional**: FR-001/FR-006 exigen mostrar las estaciones de origen y destino al usuario, y el schema `Viaje` global solo lleva los IDs; sin los nombres el frontend necesitaría un endpoint de estaciones que **spec-001 aún no tiene implementado** (const. V: no puede depender de otra rama). Los campos extra son *aditivos* — ningún consumidor validado contra el OpenAPI deja de funcionar — y siguen el precedente del propio contrato global: `AlertaBalanceo` expone `estacion_id` **y** `estacion_nombre` juntos. No se modifica `openapi.yaml` (clarificación 4).
- **Alternativas consideradas**: traer nombres desde `GET /estaciones` (rechazada: endpoint no implementado y dependencia de otra rama); solo IDs (rechazada: FR-001 no se puede mostrar); anidar objetos `estacion: {id, nombre}` (rechazada: divergencia estructural mayor con el schema `Viaje`).

## D6. Validación de entrada y páginas fuera de rango

- **Decisión**:

  | Entrada | Comportamiento | Fuente |
  |---|---|---|
  | `page` ausente | `1` (default) | OpenAPI global |
  | `page_size` ausente | `50` (default) | OpenAPI global |
  | `page < 1` o `page_size < 1` o `page_size > 200` | `422 ENTRADA_INVALIDA` | límites del contrato global |
  | `desde`/`hasta` con formato no `YYYY-MM-DD` o `desde > hasta` | `422 ENTRADA_INVALIDA` | entrada malformada |
  | `page` válido pero superior a la última página | `200` con `items: []`, `total` y `suma_coste` del periodo completo | clarificación 5 + FR-008 |
  | Periodo sin viajes | `200` con `items: []`, `total: 0`, `suma_coste: 0` | caso límite del spec |

- **Racional**: los límites del contrato se rechazan con el envelope global de errores (mismo patrón que `ENTRADA_INVALIDA` de spec-003), mientras que "fuera de rango" es un estado **válido** de paginación que se responde con lista vacía *sin error* y sumatorio del periodo (la contradicción detectada en clarify se resolvió en la sesión: el sumatorio es propiedad del periodo, no de la página).
- **Alternativas consideradas**: recortar (clamp) valores de paginación inválidos (rechazada: oculta errores del llamador y hace la API impredecible); 404 para página inexistente (rechazada: el spec exige "lista vacía sin error").

## D7. Frontend: filtro mes/año, formato y paginación

- **Decisión**: página `frontend/src/pages/viajes/` con señales `viajes`, `total`, `suma`, `mes`, `anio`, `pagina`, `cargando`, `error`:
  - **Filtro**: dos selects — **mes** libre `1–12` y **año** con las opciones `actual−1`, `actual` y `actual+1` (cubre el historial reciente, el año en curso y meses futuros, que deben ser seleccionables según el caso límite). Estado inicial = **sin filtro** (historial completo, FR-003). Al cambiar mes/año → `pagina = 1` y se envía `desde`/`hasta` (primer/último día del mes, `YYYY-MM-DD`). Selector libre: un mes futuro devuelve lista vacía sin error (caso límite).
  - **Petición**: una llamada a `GET /api/v1/viajes?...` por cambio de filtro o página (cliente en `services/api/viajes.ts`, patrón de `api/devolucion.ts`).
  - **Formato de fecha**: `fecha_inicio` (ISO UTC) se formatea como `dd/mm/aaaa HH:MM` con `toLocaleString('es-ES', { timeZone: 'Europe/Madrid' })` — la misma zona que usa el backend para filtrar, de modo que la fila y el filtro hablan del mismo calendario.
  - **Sumatorio**: se pinta `suma_coste` del response (formato `es-ES`, ej. `4,60 €`); nunca se suma en cliente (const. I y SC-002).
  - **Etiqueta**: se pinta `item.etiqueta` tal cual llega del backend; sin cálculo propio.
  - **Paginación**: "Anterior / Siguiente" + "Página X de Y" (`ceil(total / page_size)`), deshabilitadas en los extremos; si la respuesta llega vacía con `total > 0` (página fuera de rango) se vuelve a la última página válida.
  - **Estados vacíos**: con filtro → "sin viajes en el periodo"; sin filtro → "aún no tienes viajes" (caso límite "usuario sin viajes registrados").
- **Racional**: sigue el patrón de la página de devolución (señales + `Show` + `data-testid` + Tailwind), cumple todos los casos límite del spec y evita estado de lado innecesario.
- **Alternativas consideradas**: estado en URL con el router (rechazada: sobreingeniería para este alcance; ya es mejora futura); traer todos los meses con datos del backend para poblar el selector (rechazada: endpoint extra, fuera de alcance).

## D8. Seed con viajes finalizados (demostración y prueba manual)

- **Decisión**: ampliar `backend/cmd/seed/main.go` con viajes **finalizados** de prueba para `ana@si16.test`, repartidos en al menos dos meses distintos (ej. mes actual y anterior), con: origen/destino distintos, duraciones y costes variados, **al menos uno con `coste_total = 0.00`** (para ver la etiqueta) y **ninguno** que rompa los CHECKs (`fecha_fin`, `estacion_destino_id` y `duracion_minutos` NOT NULL cuando `estado = 'finalizado'`). Se conserva intacto lo ya sembrado (incluido el viaje `en_curso` de ana, que **no** debe aparecer en el historial — caso límite del spec).
- **Racional**: sin estos datos "Mis Viajes" siempre se demostraría vacío (el seed actual solo crea un viaje activo); los tests de integración usan sus propias fixtures con `TRUNCATE`, así que el seed es solo para la validación manual del quickstart. Fichero creado por spec-003 (mismo autor): se amplía sin tocar su contenido existente.
- **Alternativas consideradas**: insertar los viajes solo en el test (rechazada: dejaría la demo manual del quickstart sin datos); crear un script separado (rechazada: el seed es el mecanismo ya establecido).

---

**Resumen**: 8 decisiones (D1–D8), 0 NEEDS CLARIFICATION. D2/D3/D5/D6 materializan las 5 clarificaciones de la sesión 2026-10-28; D1/D4/D7/D8 completan la implementación sin desviarse del plan global. La HU-04 es **solo lectura**: cero migraciones, cero claves Redis nuevas, cero cambios de stack → sin puntos que someter a acuerdo de equipo (principio III).
