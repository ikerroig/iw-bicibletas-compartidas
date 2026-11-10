# Data Model: Panel de analítica de uso y exportación CSV (HU-08, subconjunto del modelo compartido)

**Branch**: `spec-008` | **Date**: 2026-11-01 | **Plan**: [plan.md](./plan.md) | **Decisiones**: [research.md](./research.md)

**Este documento NO aplica ningún cambio de esquema.** La fuente única de verdad es el modelo compartido en [specs/spec-001/data-model.md](../spec-001/data-model.md) (`0001_esquema_base.sql`, ya presente en esta rama); la HU-08 usa las tablas **tal como están definidas**. La HU-08 es de **solo lectura**: no escribe ninguna fila de ninguna tabla y no toca Redis. Cualquier divergencia de este documento respecto al modelo compartido se resuelve a favor del global y requiere acuerdo del equipo (constitución III).

## Entidades utilizadas

| Tabla | Uso en la HU-08 | Operación |
|---|---|---|
| `viajes` | Fuente única de la analítica: agregados por hora/día (FR-002) y filas del CSV (FR-003/FR-004); filtro `estado = 'finalizado'` (FR-007) e `ix_viajes_fecha` | **Lectura** |
| `estaciones` | Nombres de origen/destino en las columnas `estacion_origen`/`estacion_destino` del CSV (JOIN) | Lectura |
| `usuarios` | — (solo el rol `gestor` del JWT) | — |

Entidades **no** tocadas: `anclajes`, `bicicletas`, `incidencias_mantenimientos`, `alertas_balanceo`. Sin migraciones nuevas, sin claves Redis nuevas (la HU-08 no consume ni escribe el estado en vivo — const. IV no aplica a esta spec).

## 1. Filtro del rango (FR-001, FR-007 — clarificación 1, D1)

Un único predicado compartido por ambos endpoints y por todas las consultas (D3):

```sql
WHERE v.estado = 'finalizado'
  AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $desde AND $hasta
```

- **Inclusividad**: ambos extremos — un viaje del día `hasta` a las 23:59 hora española **entra**.
- **Zona horaria**: Europe/Madrid, aplicada en SQL (misma técnica y constantes que `services/viajes` de HU-04): `TIMESTAMPTZ` almacena instantes, la conversión es exacta sea cual sea el huso del servidor.
- **Solo finalizados** (FR-007, Assumptions): los viajes `en_curso` no se cuentan hasta su cierre (HU-03); el CHECK del esquema base garantiza que un `finalizado` siempre tiene destino, fin y duración.
- **Validación previa de entrada** (→ `422 ENTRADA_INVALIDA`, D1): `desde`/`hasta` obligatorios, formato estricto `YYYY-MM-DD` y `desde ≤ hasta`. Sin tocar la BD.
- **Rango futuro o vacío**: no hay lógica especial — el predicado no encuentra viajes inexistentes y devuelve 0 en todos los buckets / solo cabecera en el CSV (caso límite del spec).

## 2. Agregados por hora y día de la semana (FR-002 — clarificación 2, D2)

Siempre sobre **`fecha_inicio`** reinterpretada en Europe/Madrid. `por_hora` tiene 24 posiciones (00–23) y `por_dia_semana` 7 posiciones en orden **lunes → domingo** (lunes = índice 0, domingo = índice 6), según el schema global `AnaliticaViajes` (`minItems`/`maxItems` 24 y 7).

```sql
-- Q1 — por_hora: hora local española del inicio (0…23)
SELECT EXTRACT(hour FROM (v.fecha_inicio AT TIME ZONE 'Europe/Madrid'))::int AS hora,
       count(*)::int
  FROM viajes v
 WHERE v.estado = 'finalizado'
   AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $1 AND $2
 GROUP BY 1;

-- Q2 — por_dia_semana: lunes=0 … domingo=6
-- (EXTRACT(dow) de PostgreSQL: 0=domingo … 6=sábado; +6 %7 desplaza a lunes=0)
SELECT ((EXTRACT(dow FROM (v.fecha_inicio AT TIME ZONE 'Europe/Madrid'))::int + 6) % 7) AS dia,
       count(*)::int
  FROM viajes v
 WHERE v.estado = 'finalizado'
   AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $1 AND $2
 GROUP BY 1;

-- Q3 — totales del rango (mismo predicado)
SELECT count(*)::int, COALESCE(sum(v.coste_total), 0)::float8
  FROM viajes v
 WHERE v.estado = 'finalizado'
   AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $1 AND $2;
```

- Los arrays de respuesta se inicializan a `0` en Go y se incrementan con los filas devueltas (D2) → **siempre 24 y 7 posiciones**, buckets sin viajes = 0.
- `total_viajes` se calcula como **suma de `por_hora`** en el propio relleno (imposible que difiera de las barras); `total_coste` viene de Q3 (el coste no se agrupa en barras).
- El mapeo de días está verificado en tests (T8c): domingo a mediodía → índice 6; lunes → 0; sábado → 5.
- Todo el cómputo es server-side (const. I): el frontend pinta los arrays tal cual llegan.

## 3. Filas del CSV (FR-003, FR-004 — clarificación 3, D4)

```sql
SELECT v.id,
       to_char(v.fecha_inicio AT TIME ZONE 'Europe/Madrid', 'YYYY-MM-DD HH24:MI') AS fecha_hora,
       eo.nombre AS estacion_origen,
       COALESCE(ed.nombre, '') AS estacion_destino,
       v.duracion_minutos,
       v.coste_total
  FROM viajes v
  JOIN estaciones eo ON eo.id = v.estacion_origen_id
  LEFT JOIN estaciones ed ON ed.id = v.estacion_destino_id
 WHERE v.estado = 'finalizado'
   AND (v.fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $1 AND $2
 ORDER BY v.fecha_inicio ASC, v.id ASC;
```

- **Mismo predicado** que los agregados (D3) → el CSV del rango contiene exactamente los viajes que suman las barras (FR-006, SC-002).
- `fecha_hora` ya llega formateada en Europe/Madrid (`AAAA-MM-DD HH:MM`) desde SQL; el servicio la emite tal cual.
- `estacion_destino` es NOT NULL por el CHECK de `finalizado`, pero el `LEFT JOIN + COALESCE` deja el CSV robusto ante datos históricos anomalos sin romper la exportación.
- **Formato del fichero** (decisiones de emisión, no de consulta): BOM UTF-8 (`EF BB BF`) manual + coma + comillas RFC 4180 + CRLF (`encoding/csv` con `UseCRLF`); `coste_eur` con **punto decimal** y 2 decimales (p. ej. `2.00` — ver D4); `duracion_minutos` entero; orden garantizado por el `ORDER BY`. Streaming fila a fila (caso "varios años" sin memoria proporcional).
- Rango sin viajes → se emite únicamente la fila de cabecera.

## 4. Respuesta del panel (FR-002 — schema global `AnaliticaViajes`)

Envelope `200` de `GET /analitica/viajes` (espejo del schema global, sin campos extra):

```json
{
  "desde": "2026-10-01",
  "hasta": "2026-10-31",
  "total_viajes": 42,
  "total_coste": 18.40,
  "por_hora":   [0, 0, 0, 0, 0, 0, 2, 5, 7, 4, 2, 1, 3, 2, 1, 0, 2, 6, 5, 1, 1, 0, 0, 0],
  "por_dia_semana": [6, 5, 4, 7, 8, 7, 5]
}
```

- `por_hora[i]` = viajes cuyo inicio local cae en la hora `i` (00–23); `por_dia_semana[i]` = día `i` con **lunes = 0 … domingo = 6**.
- `desde`/`hasta` se devuelven tal como se recibieron (schema global) para que la UI confirme el rango aplicado.
- Rango vacío → todos los arrays a 0, `total_viajes: 0`, `total_coste: 0` (la UI deriva el aviso "sin datos en el periodo" de `total_viajes == 0`).

## Máquina de estados

**No aplica**: la HU-08 no gestiona entidades con ciclo de vida — solo lee viajes ya finalizados (cuyo estado lo gestionan HU-02/HU-03). El "estado" del panel es una función pura del rango consultado y los datos de `viajes`.

## Validaciones de dominio (server-side, const. I)

| # | Validación | Fuente | Comportamiento |
|---|---|---|---|
| V1 | Solo rol `gestor` en ambos endpoints | FR-005, SC-003, OpenAPI | `RequireRole("gestor")` → 403 `ROL_NO_PERMITIDO`; sin token → 401 |
| V2 | `desde`/`hasta` obligatorios y en formato `YYYY-MM-DD` estricto | FR-001, clarificación 1 (patrón HU-04) | `422 ENTRADA_INVALIDA`, sin tocar la BD |
| V3 | `desde ≤ hasta` | clarificación 1 (rango invertido) | `422 ENTRADA_INVALIDA`, sin gráficos ni CSV |
| V4 | Ambos extremos del rango inclusivos, en Europe/Madrid | clarificación 1 | `BETWEEN … ::date` sobre `fecha_inicio` convertida |
| V5 | Agrupación por **fecha de inicio** del viaje en Europe/Madrid; 24 barras 00–23 y 7 barras lunes→domingo | clarificación 2 | `EXTRACT(hour/dow)` tras `AT TIME ZONE`; relleno a 0 en Go |
| V6 | Solo viajes `finalizado` de toda la red | FR-007, Assumptions | `WHERE estado = 'finalizado'` (sin filtro de usuario) |
| V7 | Gráficos y CSV del mismo rango ⇒ mismos viajes y totales | FR-006, SC-002 | Un único `construirFiltro` compartido por agregados y CSV; `total_viajes` = suma de barras |
| V8 | CSV: UTF-8 BOM, coma, RFC 4180, CRLF, columnas del contrato y `fecha_hora` Europe/Madrid | FR-004, R11, clarificación 3 | Emisión con `encoding/csv` + BOM manual; decimal con punto |

## Migraciones

**Ninguna.** La HU-08 no añade tablas, columnas, índices ni restricciones, y no toca Redis. Usa `viajes` y `estaciones` de `0001_esquema_base.sql` con el índice **`ix_viajes_fecha`** ya existente (creado en el esquema base para esta HU) y el JOIN a `estaciones` para los nombres del CSV.

## Cambios en ficheros compartidos — autorizados por el equipo (constitución III) y aplicados

La HU-08 no requiere cambios de esquema. Los dos puntos que afectan a ficheros compartidos fueron **autorizados por el equipo y aplicados** (T019/T020):

1. **`422 ENTRADA_INVALIDA` en el OpenAPI global para `/analitica/viajes` y `/analitica/export`** — añadido a `specs/spec-001/contracts/openapi.yaml` (T019, acuerdo del equipo); el comportamiento 422 sigue el patrón consolidado por spec-003/spec-004/spec-007 y se detalla en [contracts/analitica.md](./contracts/analitica.md).
2. **Dependencia `chart.js` (`^4`) en `frontend/package.json`** — alta y actualización del lockfile aplicadas (T020, acuerdo del equipo); tecnología prevista en el plan global (R14).
El struct `AnaliticaViajes`, el lote de seed y las rutas/nav son adiciones propias de HU-08, ya implementadas en esta rama; no alteran campos existentes ni el esquema compartido.
