# Data Model: Monitorización de desequilibrios y alertas de balanceo (HU-07, subconjunto del modelo compartido)

**Branch**: `spec-007` | **Date**: 2026-10-30 | **Plan**: [plan.md](./plan.md) | **Decisiones**: [research.md](./research.md)

**Este documento NO aplica ningún cambio de esquema.** La fuente única de verdad es el modelo compartido en [specs/spec-001/data-model.md](../spec-001/data-model.md) (`0001_esquema_base.sql`, ya presente en esta rama); la HU-07 usa las tablas y claves Redis **tal como están definidas**. Cualquier divergencia de este documento respecto al modelo compartido se resuelve a favor del global y requiere acuerdo del equipo (constitución III). Las únicas propuestas que tocan el modelo compartido se recogen al final marcadas `⚠ PENDIENTE DE ACUERDO DEL EQUIPO` y **no se implementan** en esta spec.

## Entidades utilizadas

| Tabla / clave | Uso en la HU-07 | Operación |
|---|---|---|
| `alertas_balanceo` | Alerta de balanceo: creación, resolución, snapshot de ocupación, panel | **Lectura/escritura** |
| `estaciones` | Cómputo de ocupación (capacidad, estado); bloqueo `FOR UPDATE` en traslado; nombres del panel | Lectura + bloqueo |
| `anclajes` | Recuento de bicis disponibles y anclajes libres; **reubicación** de bicis en el traslado | Lectura/escritura (DML) |
| `bicicletas` | Exclusión de averiadas/en_taller del cómputo (FR-008); selección del lote | Lectura |
| `usuarios` | — (solo el rol `operario_logistica` del JWT) | — |
| Redis `estacion:{id}:estado` | Estado recalculado de origen y destino tras el traslado (FR-007, const. IV) | Escritura vía `cache.ActualizarEstadoEstacion` (helper existente) |

Entidades **no** tocadas: `viajes`, `incidencias_mantenimientos`. Sin migraciones nuevas, sin claves Redis nuevas.

## 1. Fórmula de ocupación (FR-001, FR-008 — R9 del modelo compartido)

Misma definición que ya usan Redis y el modelo global; el worker y el inline la implementan con esta consulta (por estación):

```sql
SELECT e.capacidad_total, e.estado,
       (SELECT COUNT(*)
          FROM anclajes a
          JOIN bicicletas b ON b.id = a.bicicleta_id
         WHERE a.estacion_id = e.id AND b.estado = 'disponible') AS bicis_disponibles
  FROM estaciones e
 WHERE e.id = $1;
```

- `ocupacion_pct = 100 × bicis_disponibles / capacidad_total` — solo si `e.estado = 'operativa'`; si no, la estación **se excluye** del cálculo y no genera alertas (caso límite del spec).
- Bicicletas `averiada` o `en_taller` **no cuentan** aunque ocupen anclaje (FR-008 / HU-05), porque el JOIN exige `b.estado = 'disponible'`.
- Bicicleta `disponible` **sin anclaje** (reparada en taller, HU-06) no pertenece a ninguna estación → no cuenta en ningún sitio.
- `capacidad_total` tiene `CHECK (> 0)` en el esquema base → no hay división por cero; "sin capacidad computable" no es alcanzable salvo estación `fuera_de_servicio`, ya excluida.

## 2. Creación y resolución de alertas (FR-002…FR-004 — R10)

Umbrales **estrictos** (caso límite del spec: 10,00 % y 90,00 % exactos **no** alertan):

| Condición (estación `operativa`) | Acción |
|---|---|
| `ocupacion < 10` | Crear `tipo_alerta = 'vacia_menos_10'` si no hay alerta **abierta** de ese (estación, umbral); resolver la abierta de `'llena_mas_90'` si existiera (una estación no puede estar en ambos extremos a la vez) |
| `ocupacion > 90` | Crear `'llena_mas_90'` si no hay abierta de ese umbral; resolver la `'vacia_menos_10'` abierta si existiera |
| `10 ≤ ocupacion ≤ 90` | Resolver (→ `resuelta = true`) **todas** las alertas abiertas de la estación (FR-004) |
| `estado <> 'operativa'` | Sin acción (excluida — [research D2](./research.md)) |

```sql
-- Creación (el índice único parcial garantiza FR-003 aunque haya doble carril):
INSERT INTO alertas_balanceo (estacion_id, tipo_alerta, porcentaje_ocupacion)
VALUES ($1, $2, $3::numeric)
ON CONFLICT (estacion_id, tipo_alerta) WHERE NOT resuelta DO NOTHING;

-- Resolución al volver a rango normal:
UPDATE alertas_balanceo
   SET resuelta = TRUE
 WHERE estacion_id = $1 AND NOT resuelta;

-- Refresco del snapshot de urgencia (D3) en cada barrido del worker:
UPDATE alertas_balanceo a
   SET porcentaje_ocupacion = $nuevo::numeric
 WHERE a.estacion_id = $1 AND NOT a.resuelta;
```

- **Unicidad (FR-003, SC-002)**: `ux_alerta_abierta (estacion_id, tipo_alerta) WHERE NOT resuelta` ya existe en `0001_esquema_base.sql`; la creación usa `ON CONFLICT … DO NOTHING` → **0 duplicados** por diseño y por índice. Al resolver y re-cruzar el umbral se abre una fila **nueva** (la resuelta queda como histórico).
- **Origen de la evaluación**: inline tras cada traslado (origen y destino) + worker `alertas` cada 55 s sobre toda la red (D1, research D6) → cubre también devoluciones y futuras mutaciones de otras specs sin modificar su código.
- `severidad` usa el DEFAULT `'alta'` del esquema (HU-07: prioridad alta por definición); `fecha_generacion` el DEFAULT `now()`.

## 3. Panel del operario (FR-005, SC-002 — `GET /balanceo/alertas`)

```sql
SELECT a.id, a.estacion_id, e.nombre AS estacion_nombre,
       a.tipo_alerta, a.porcentaje_ocupacion, a.severidad,
       a.fecha_generacion, a.resuelta,
       CASE WHEN a.tipo_alerta = 'vacia_menos_10'
            THEN 10 - a.porcentaje_ocupacion
            ELSE     a.porcentaje_ocupacion - 90 END AS urgencia
  FROM alertas_balanceo a
  JOIN estaciones e ON e.id = a.estacion_id
 WHERE NOT a.resuelta
 ORDER BY urgencia DESC, a.fecha_generacion ASC;
```

- **Orden por urgencia** (FR-005): extremos más alejados del rango `[10, 90]` primero (0 % antes que 9 %; 100 % antes que 91 %); empates → alerta más antigua. `ix_alertas_abiertas` cubre el filtro de abiertas.
- Solo alertas **abiertas** (`resuelta = false`); las resueltas salen del panel automáticamente (SC-004, 0 obsoletas). El `porcentaje_ocupacion` mostrado es el snapshot refrescado por el worker (< 60 s = `periodo + T_barrido`, D3/D6).

## 4. Traslado de lote (FR-006, FR-007, SC-003 — `POST /balanceo/traslados`)

Secuencia dentro de **una transacción** (D4):

```sql
-- 1. Bloqueo de estaciones (orden de id → sin deadlocks):
SELECT id, estado FROM estaciones
 WHERE id = ANY($1::bigint[])   -- [origen, destino]
 ORDER BY id FOR UPDATE;

-- 2. Selección del lote en origen (disponibles ancladas) y huecos en destino:
SELECT a.id, a.bicicleta_id
  FROM anclajes a
  JOIN bicicletas b ON b.id = a.bicicleta_id
 WHERE a.estacion_id = $origen AND a.estado = 'ocupado' AND b.estado = 'disponible'
 ORDER BY a.id LIMIT $unidades FOR UPDATE OF a;

SELECT id FROM anclajes
 WHERE estacion_id = $destino AND estado = 'libre'
 ORDER BY id LIMIT $unidades FOR UPDATE;

-- 3. Movimiento (si ambos recuentos ≥ unidades; si no → 409, ROLLBACK):
UPDATE anclajes SET estado = 'libre', bicicleta_id = NULL
 WHERE id = ANY($ids_origen);                    -- origen queda vacío
UPDATE anclajes SET estado = 'ocupado', bicicleta_id = bici_i
 WHERE id = ids_destino_i;                       -- por cada bicicleta
```

- Las bicicletas **conservan** `estado = 'disponible'`; solo cambian de anclaje/estación → la ocupación de ambas estaciones se deriva sola de R9 (punto 1).
- **Validaciones previas** (server-side, const. I): `unidades ≥ 1`; origen ≠ destino (→ `422`); ambas estaciones existen y están `operativa`; `unidades ≤ bicis disponibles en origen` y `unidades ≤ anclajes libres en destino` (si no → `409 ESTACION_SIN_CAPACIDAD`, sin alterar nada). Detalle en [contracts/balanceo.md](./contracts/balanceo.md).
- **Post-commit**: `cache.ActualizarEstadoEstacion(origen)` y `(destino)` (escritura de `estacion:{id}:estado`, TTL 120 s — contrato Redis intacto) + evaluación inline de umbrales de ambas (punto 2). La respuesta `200` devuelve `EstacionEstado` de origen y destino con los valores ya recalculados (FR-007, SC-003).

## Máquina de estados de la alerta

```text
(no existe) ──cruza umbral (<10 % o >90 %)──> abierta (resuelta = false)
abierta ──vuelve a [10, 90]──> resuelta (resuelta = true)   ← final; sale del panel
abierta ──empeora dentro del mismo extremo──> sigue abierta (snapshot refrescado)
resuelta ──re-cruza el umbral──> (nueva fila abierta; la resuelta no se reabre)
```

- **Unicidad**: a lo sumo **una** fila abierta por (estación, umbral) — garantizado por `ux_alerta_abierta`.
- **Alternancia rápida** (caso límite del spec): el panel siempre refleja el estado actual; el worker y el inline cierran/reabren sin duplicados.

## Validaciones de dominio (server-side, const. I)

| # | Validación | Fuente | Comportamiento |
|---|---|---|---|
| V1 | Solo rol `operario_logistica` en ambos endpoints | HU-07 (panel del operario), OpenAPI | `RequireRole` → 403 `ROL_NO_PERMITIDO`; sin token → 401 |
| V2 | Umbrales estrictos `< 10` / `> 90` | FR-002, caso límite | 10,00 % y 90,00 % exactos no generan alerta |
| V3 | Una única alerta abierta por (estación, umbral) | FR-003, SC-002 | `ON CONFLICT` sobre `ux_alerta_abierta` |
| V4 | Resolución automática al volver a `[10, 90]` | FR-004, SC-004 | UPDATE `resuelta = true`; fuera del panel |
| V5 | Estaciones `fuera_de_servicio` excluidas; averiadas fuera del cómputo | caso límite, FR-008 | `WHERE estado = 'operativa'`; JOIN exige bici `disponible` |
| V6 | Traslado: entrada válida (`unidades ≥ 1`, origen ≠ destino) | OpenAPI (`minimum: 1`), coherencia | `422 ENTRADA_INVALIDA` sin efectos |
| V7 | Traslado: estaciones existentes/operativas y lote factible | OpenAPI (409), caso límite | `409 ESTACION_SIN_CAPACIDAD`, ROLLBACK íntegro |
| V8 | Recálculo usa los valores confirmados reales (bloqueados) | caso límite (datos desactualizados) | `FOR UPDATE` + cómputo dentro de la transacción |

## Migraciones

**Ninguna.** La HU-07 no añade tablas, columnas, índices ni restricciones, y no toca el formato de Redis. Usa `0001_esquema_base.sql` (tabla `alertas_balanceo` con `ux_alerta_abierta` e `ix_alertas_abiertas`, tablas `estaciones`/`anclajes`/`bicicletas` con sus CHECKs e índices) y `cache.ActualizarEstadoEstacion` tal cual.

## ⚠ PENDIENTE DE ACUERDO DEL EQUIPO (constitución III) — propuestas NO aplicadas

Las siguientes propuestas tocarían el **esquema compartido** (nuevas migraciones PostgreSQL). Quedan documentadas aquí como candidatas y **no se implementan** hasta que el equipo las apruebe; la HU-07 es perfectamente implementable y testeable sin ellas:

1. **`traslados_balanceo` (auditoría de operaciones)** — el `spec.md` (Key Entities) menciona la entidad "Operación de traslado" (origen, destino, unidades, fecha), pero el modelo compartido actual **no tiene tabla para ella** y el OpenAPI no la expone. Propuesta opcional:
   - ¿Qué aportaría? histórico firmado de cada confirmación de lote (quién, cuándo, qué estación origen/destino, cuántas unidades) para auditoría y futura analítica (HU-08 podría querer repartir traslados).
   - ¿Impacto? nueva migración `000N_spec-007_traslados.sql` con ordinal **acordado** (R6), INSERT en la transacción del traslado, y decisión sobre si el DELETE/limpieza afecta a seeds y tests de otras specs.
   - Sin esta tabla, el efecto del traslado queda reflejado en los anclajes (estado físico de la red) y en las alertas, que es lo que exigen FR-006/FR-007 y todo el alcance validado de la HU-07.
2. **Índice adicional en `alertas_balanceo`** — la consulta del panel (punto 3) ordena por un `CASE` calculado; con el volumen real (~60 estaciones, pocas alertas abiertas) no se necesita índice propio, pero si el equipo prefiere materializar el score, requeriría acuerdo (columna generada o índice expression).

**Decisión de esta spec**: MVP sin cambios de esquema (todo respaldado por el modelo y los contratos existentes). Cualquier incorporación de los puntos anteriores se somete a acuerdo del equipo **antes** de crear la migración (const. III y Acuerdos de Colaboración).
