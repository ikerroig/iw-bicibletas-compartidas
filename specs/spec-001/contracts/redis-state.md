# Contract: Estado en vivo en Redis (frescura ≤ 30 s)

**Branch**: `spec-001` | **Date**: 2026-10-20 | **Modelo**: [data-model.md](../data-model.md) | **Decisión**: R3/R4 de [research.md](../research.md)

Contrato de claves compartido por las 8 specs (constitución, principio III). Cualquier nueva clave o cambio de formato requiere acuerdo del equipo.

## Claves

### `estacion:{id}:estado` (HASH)

Estado en vivo de una estación, espejo exacto de los cómputos de `data-model.md` (`bicis_disponibles`, `anclajes_libres`, `ocupacion_pct`):

| Campo | Tipo | Descripción |
|---|---|---|
| `bicis_disponibles` | string(int) | Anclajes con bicicleta en estado `disponible` (averiadas y en_taller **excluidas** — HU-01 FR-006, HU-05, HU-07). |
| `anclajes_libres` | string(int) | Anclajes en estado `libre` (un anclaje ocupado por una bicicleta averiada **no** es libre). |
| `ocupacion_pct` | string(float, 2 dec.) | `100 × bicis_disponibles / capacidad_total`; `0` si la estación está `fuera_de_servicio`. |
| `estado` | string | `operativa` \| `fuera_de_servicio` (HU-01: se muestra al usuario). |
| `actualizado_en` | string(RFC3339 UTC) | Instante del último recálculo; lo expone `GET /estaciones` para el edge case de desconexión (HU-01). |

- **TTL**: 120 s, refrescado en cada escritura. La expiración es una red de seguridad: si ocurre (p. ej., parada del worker), `GET /estaciones` recalcula desde PostgreSQL y repuebla la clave — nunca se sirven datos sin marca de tiempo (FR-005).
- **Atomicidad**: escrituras con `HSET` + `EXPIRE` en pipeline (todo el hash se reescribe; es el estado completo de la estación).

### Claves auxiliares del worker (internas, no consumir desde specs)

- `si16:worker:refresco:ultimo` (string, timestamp del último barrido completo) — observabilidad del worker `refresco_estado`.

### Canal pub/sub (opcional, no vinculante)

- `estacion:estado:cambios`: mensaje `{"estacion_id":N,"actualizado_en":"..."}` emitido tras cada mutación. **No forma parte del contrato consumible** en el MVP (el frontend usa polling); se documenta para futuras mejoras push sin romper nada.

## Frescura (constitución IV)

| Mecanismo | Frecuencia | Garantiza |
|---|---|---|
| Escritura síncrona post-commit (desbloqueo, devolución, avería, traslado, ubicación) | En el acto | Reflejo inmediato para el actor y los paneles de operarios (HU-07 SC-003) |
| Worker `refresco_estado` (recalcula las ~60 estaciones con una consulta agregada) | ≤ 15 s | Cota dura de 15 s < 30 s aunque algún camino de escritura falte (HU-01 FR-003, SC-002) |
| Frontend (`stores/`, hook de red) | Polling `GET /estaciones` cada 15 s | ≤ 30 s en pantalla; muestra `actualizado_en` y reintenta si la conexión falla |

## Lectura desde el backend

```text
GET /estaciones:  ids desde PostgreSQL → SCAN MATCH estacion:*:estado → HGETALL en pipeline
                  → clave ausente/expirada: recalcular desde SQL y repoblar
GET /estaciones/{id}:  HGETALL estacion:{id}:estado (mismo fallback)
```

Ningún endpoint sirve contadores leídos de PostgreSQL directamente salvo el fallback anterior: la fuente de lectura en vivo es Redis (HU-01 FR-005, constitución IV).

## Ejemplo

```text
HGETALL estacion:17:estado
  bicis_disponibles = "7"
  anclajes_libres   = "12"
  ocupacion_pct     = "58.33"
  estado            = "operativa"
  actualizado_en    = "2026-10-20T10:15:32Z"
```
