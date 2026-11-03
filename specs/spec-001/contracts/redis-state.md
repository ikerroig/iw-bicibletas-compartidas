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
| `actualizado_en` | string(RFC3339 UTC, sufijo `Z`, fracción opcional) | Instante del último recálculo; lo expone `GET /estaciones` para el edge case de desconexión (HU-01). El backend escribe nanosegundos de ancho fijo (`…T10:15:32.123456789Z`); se admite también sin fracción (`…T10:15:32Z`). |

- **TTL**: 120 s, refrescado en cada escritura. La expiración es una red de seguridad: si ocurre (p. ej., parada del worker), `GET /estaciones` recalcula desde PostgreSQL y repuebla la clave — nunca se sirven datos sin marca de tiempo (FR-005).
- **Atomicidad**: escrituras con `HSET` + `EXPIRE` en un script Lua por estación (todo el hash se reescribe; es el estado completo de la estación). El script **descarta** la escritura si el hash ya guarda un `actualizado_en` más reciente; compara instantes normalizados, no texto, así que cualquier ancho de fracción es válido. Escribid siempre a través de `cache.GuardarEstados`.

### Claves auxiliares del worker (internas, no consumir desde specs)

- `si16:worker:refresco:ultimo` (string, timestamp del último barrido completo) — observabilidad del worker `refresco_estado`.

### Canal pub/sub (opcional, no vinculante)

- `estacion:estado:cambios`: mensaje `{"estacion_id":N,"actualizado_en":"..."}` emitido tras cada mutación. **No forma parte del contrato consumible** en el MVP (el frontend usa polling); se documenta para futuras mejoras push sin romper nada.

## Frescura (constitución IV)

| Mecanismo | Frecuencia | Garantiza |
|---|---|---|
| Escritura síncrona post-commit (desbloqueo, devolución, avería, traslado, ubicación) | En el acto | Reflejo inmediato para el actor y los paneles de operarios (HU-07 SC-003) |
| Worker `refresco_estado` (recalcula las ~60 estaciones con una consulta agregada) | ≤ 10 s | Respaldo si falla algún camino de escritura: 10 s + 15 s de polling + red/render ≤ 28 s extremo a extremo (HU-01 FR-003, SC-002; presupuesto en R16) |
| Frontend (`stores/`, hook de red) | Polling `GET /estaciones` cada 15 s | ≤ 30 s del commit al pintado; muestra la hora de la última actualización correcta, reintenta si la conexión falla y avisa de «datos desactualizados» a los > 30 s (HU-01 FR-007, R17) |

Los contadores cumplen `bicis_disponibles + anclajes_libres ≤ capacidad_total`: los anclajes con bicicleta averiada no suman en ninguno de los dos (R19). Una estación `fuera_de_servicio` sigue teniendo contadores enteros en Redis y en la API; es la UI de HU-01 la que los oculta (R18).

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
