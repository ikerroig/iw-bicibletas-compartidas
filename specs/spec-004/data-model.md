# Data Model: Historial de viajes y desglose mensual (HU-04, subconjunto del modelo compartido)

**Branch**: `spec-004` | **Date**: 2026-10-28 | **Plan**: [plan.md](./plan.md) | **Decisiones**: [research.md](./research.md)

**Este documento NO define esquema nuevo.** La fuente única de verdad es el modelo compartido en [specs/spec-001/data-model.md](../spec-001/data-model.md); la HU-04 **solo lee** tablas ya existentes en `0001_esquema_base.sql`. Cualquier divergencia de este documento respecto al modelo compartido se resuelve a favor del global y requiere acuerdo del equipo (constitución III).

## Entidades consultadas (solo lectura)

| Tabla | Uso en la HU-04 | Campos leídos |
|---|---|---|
| `viajes` | Listado paginado + agregado del periodo | `id`, `usuario_id`, `estacion_origen_id`, `estacion_destino_id`, `fecha_inicio`, `fecha_fin`, `duracion_minutos`, `coste_total`, `estado` |
| `estaciones` | JOIN para mostrar nombres en cada fila (FR-001/FR-006) | `id`, `nombre` (origen y destino) |
| `usuarios` | — (solo el `usuario_id` del JWT) | — |

Entidades **no** tocadas: `bicicletas`, `anclajes`, `incidencias_mantenimientos`, `alertas_balanceo`, y **Redis completo** (la HU-04 no lee ni escribe estado en vivo — no es una función de disponibilidad).

## Consultas (detalle de las dos sentencias de [research D1](./research.md))

**Q1 — Página del historial** (orden cronológico descendente, viajes finalizados propios):

```sql
SELECT v.id, v.usuario_id, v.estacion_origen_id, v.estacion_destino_id,
       v.fecha_inicio, v.fecha_fin, v.duracion_minutos, v.coste_total, v.estado,
       eo.nombre AS estacion_origen_nombre,
       ed.nombre AS estacion_destino_nombre
  FROM viajes v
  JOIN estaciones eo ON eo.id = v.estacion_origen_id
  JOIN estaciones ed ON ed.id = v.estacion_destino_id
 WHERE v.usuario_id = $1
   AND v.estado = 'finalizado'
   [ AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $2 AND $3 ]
 ORDER BY v.fecha_inicio DESC
 LIMIT $4 OFFSET $5;
```

**Q2 — Agregado del periodo completo** (siempre sin `LIMIT/OFFSET` — FR-003/FR-008):

```sql
SELECT count(*)::int AS total,
       COALESCE(sum(v.coste_total), 0)::float8 AS suma_coste
  FROM viajes v
 WHERE v.usuario_id = $1
   AND v.estado = 'finalizado'
   [ AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $2 AND $3 ];
```

- **Índice usado**: `ix_viajes_usuario_fecha (usuario_id, fecha_inicio DESC)` cubre filtro de usuario + `ORDER BY` + rango temporal. El JOIN a `estaciones` es por PK.
- `coste_total` es `NUMERIC(10,2)` → se escanea como `float8` en el agregado (mismo mapeo que ya usa spec-003 en los viajes).
- **`total`** = nº de viajes de **todo** el periodo (habilita "Página X de Y"); **`suma_coste`** = suma de **todo** el periodo (SC-002).

## Proyección de respuesta (item del historial)

El item = schema `Viaje` global (`specs/spec-001/contracts/openapi.yaml`) **+** dos campos aditivos de nombre (patrón `AlertaBalanceo`, ver [research D5](./research.md)):

| Campo | Origen | Notas |
|---|---|---|
| `id`, `usuario_id`, `estacion_origen_id`, `estacion_destino_id` | `viajes` | tal cual |
| `fecha_inicio`, `fecha_fin` | `viajes` | ISO-8601 UTC; la UI formatea `dd/mm/aaaa HH:MM` en Europe/Madrid |
| `duracion_minutos` | `viajes` | registrado en el cierre (HU-03), **no** se recalcula (FR-007) |
| `coste_total` | `viajes` | registrado en el cierre (HU-03), **no** se recalcula (FR-007) |
| `estado` | `viajes` | siempre `finalizado` (FR-005) |
| `etiqueta` | derivada | `"Tramo gratuito (0,00 €)"` cuando `coste_total = 0` (FR-004, [research D3](./research.md)) |
| `estacion_origen_nombre` | JOIN `estaciones` | **aditivo** al schema global |
| `estacion_destino_nombre` | JOIN `estaciones` | **aditivo** al schema global; `NOT NULL` en la práctica porque el CHECK de cierre exige destino en viajes finalizados |

Envelope de la respuesta: `{ total, suma_coste, items: [...] }` (OpenAPI global, operación `historialViajes`).

## Filtros y semántica temporal (FR-002)

- `desde`/`hasta`: fechas `YYYY-MM-DD` **inclusivas**, interpretadas en `Europe/Madrid` mediante `AT TIME ZONE` en SQL ([research D2](./research.md)).
- Sin parámetros → historial completo del usuario.
- El mes/año seleccionado en la UI se traduce a `desde = día 1` y `hasta = último día del mes`; un mes futuro o vacío devuelve `items: []` sin error (casos límite).

## Máquina de estados relevante

```text
viaje:  en_curso ──cierre (HU-03)──> finalizado   ← ÚNICO estado listado por HU-04
        en_curso ──(HU-04)──────────  NO se lista   (caso límite: viaje activo)
```

La HU-04 **no transiciona estados**: solo observa `estado = 'finalizado'`. El viaje `en_curso` del usuario (si existe) jamás aparece en el historial hasta que HU-03 lo cierre.

## Validaciones de dominio (server-side, const. I)

| # | Validación | Fuente | Comportamiento |
|---|---|---|---|
| V1 | El historial es siempre el del usuario del JWT | FR-005 | `usuario_id` nunca es entrada del request; sin token → 401 |
| V2 | Solo viajes `finalizado` | FR-005, caso "viaje activo" | `WHERE estado = 'finalizado'` — el activo no se lista |
| V3 | Coste y duración sin recalcular | FR-007 | Select puro de las columnas cerradas por HU-03 |
| V4 | Etiqueta derivada del coste | FR-004 | `coste_total = 0` → `etiqueta` (nunca calculada en el cliente) |
| V5 | Sumatorio del periodo completo | FR-003, FR-008, SC-002 | Q2 sin `LIMIT/OFFSET`; idéntico en cualquier página |
| V6 | Entrada de paginación y fechas | clarificaciones 1/4/5 | fuera de rango → `200` vacío con suma del periodo; malformada → `422 ENTRADA_INVALIDA` |

## Migraciones

**Ninguna.** La HU-04 no añade tablas, columnas, índices ni restricciones, y no toca Redis. Se usa `0001_esquema_base.sql` tal cual (el índice `ix_viajes_usuario_fecha` ya existe). No hay nada que someter a acuerdo de equipo bajo el principio III.
