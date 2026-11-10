# Contract: GET /analitica/viajes y GET /analitica/export (HU-08)

**Branch**: `spec-008` | **Date**: 2026-11-01 | **Plan**: [plan.md](./plan.md)

Contrato de los dos endpoints de spec-008. Alineados con el OpenAPI global (`specs/spec-001/contracts/openapi.yaml`, tag `spec-008`, operationIds `analiticaViajes` y `exportarCSV`): se adoptan literalmente sus paths, parámetros (`desde`/`hasta`, `format: date`, obligatorios) y schemas (`AnaliticaViajes`; `text/csv` para la exportación). Este documento detalla la semántica, los mensajes y el mapeo de columnas; el `422` queda documentado en el propio YAML global desde T019 (acuerdo del equipo). En caso de divergencia no contemplada prevalece el OpenAPI global y este documento se corrige.

## Autorización (ambos endpoints)

```http
Authorization: Bearer <JWT>     # rol obligatorio: gestor
```

Middleware encadenado `RequireAuth` + `RequireRole("gestor")` (patrón ya usado en la rama por spec-007). Sin token o token inválido → `401 NO_AUTENTICADO`; token de cualquier otro rol (`usuario`, `operario_taller`, `operario_logistica`) → `403 ROL_NO_PERMITIDO` (FR-005, SC-003: el 100 % de los intentos no gestor se deniega).

## Parámetros de consulta (ambos endpoints)

| Parámetro | Tipo | Obligatorio | Validación |
|---|---|---|---|
| `desde` | `date` (`YYYY-MM-DD`) | sí | formato estricto; `desde ≤ hasta`; **ambos inclusivos**, interpretados en **Europe/Madrid** (FR-001, clarificación 1) |
| `hasta` | `date` (`YYYY-MM-DD`) | sí | ídem |

Criterio semántico: el rango delimita la **fecha de inicio** del viaje (hora española): entra todo viaje finalizado cuyo inicio local cae en `[desde, hasta]` inclusive. Rangos futuros o vacíos **no son error** (datos inexistentes → ceros / solo cabecera). Ver [data-model.md §1](../data-model.md).

---

## 1. `GET /api/v1/analitica/viajes`

Agregados del volumen de viajes finalizados de **toda la red** en el rango: 24 barras por hora del día (00–23) y 7 barras por día de la semana (lunes → domingo), más los totales (FR-002, FR-007, clarificación 2).

### Respuestas

#### `200 OK`

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

| Campo | Semántica |
|---|---|
| `desde`/`hasta` | Eco del rango aplicado (schema global) |
| `por_hora` | **Exactamente 24** enteros: índice `i` = viajes iniciados en la hora local `i` (00–23, Europe/Madrid); buckets sin viajes = `0` |
| `por_dia_semana` | **Exactamente 7** enteros: índice `0` = lunes … índice `6` = domingo (Europe/Madrid); buckets sin viajes = `0` |
| `total_viajes` | Suma de `por_hora` (por construcción, FR-006) |
| `total_coste` | Suma de `coste_total` de los mismos viajes, euros con punto decimal (p. ej. `18.40`) |

Garantías del `200`:
- **Mismo filtro que el CSV** (FR-006, SC-002): ambos endpoints comparten el mismo predicado construido una vez por petición — el total de viajes del rango es idéntico en gráficos y exportación.
- Solo viajes `finalizado` de toda la red (FR-007); los activos no aparecen hasta su cierre (HU-03).
- Rango sin viajes → arrays todos a `0`, `total_viajes: 0`, `total_coste: 0` (**no** es error; la UI muestra "sin datos en el periodo").

#### Errores (envelope global `{ "error": { "codigo", "mensaje" } }`)

| HTTP | `codigo` | Condición | Efectos |
|---|---|---|---|
| 401 | `NO_AUTENTICADO` | Sin JWT o token inválido/expirado | Ninguno |
| 403 | `ROL_NO_PERMITIDO` | Rol distinto de `gestor` | Ninguno |
| 422 | `ENTRADA_INVALIDA` | `desde` o `hasta` ausentes, con formato distinto de `YYYY-MM-DD` estricto (p. ej. `2026-13-40` o `01/10/2026`) o `desde > hasta` | Ninguno (validación previa a la consulta) |

### Ejemplos

```http
GET /api/v1/analitica/viajes?desde=2026-10-01&hasta=2026-10-31   # agregados del mes
GET /api/v1/analitica/viajes?desde=2026-11-01&hasta=2026-12-31   # rango futuro → 200 con ceros
GET /api/v1/analitica/viajes?desde=2026-10-31&hasta=2026-10-01   # 422 ENTRADA_INVALIDA
```

---

## 2. `GET /api/v1/analitica/export`

CSV con el **histórico filtrado** del mismo rango (FR-003): exactamente los viajes que suman los gráficos, en streaming (FR-004, R11, clarificación 3).

### Cabeceras de respuesta

```http
Content-Type: text/csv; charset=utf-8
Content-Disposition: attachment; filename=viajes_20261001_20261031.csv
```

El nombre lleva `desde` y `hasta` sin guiones (`viajes_AAAAMMDD_AAAAMMDD.csv`).

### Formato del cuerpo

- **BOM UTF-8** (`EF BB BF`) como primeros 3 bytes — imprescindible para que Excel/LibreOffice reconozcan UTF-8 (SC-004, tildes/ñ).
- Separador **coma**; comillas **RFC 4180** solo cuando el campo lo requiere (coma, comillas o salto de línea); finales de línea **CRLF**.
- **Columnas** (en este orden, cabecera siempre presente):

```csv
id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur
1201,2026-10-05 08:15,Estació Universitat,Estació Mercat,12,0.00
1202,2026-10-05 17:40,Estació Mercat,Estació Nord,45,3.00
```

| Columna | Contenido |
|---|---|
| `id` | Identificador del viaje (`viajes.id`) |
| `fecha_hora` | `fecha_inicio` en **Europe/Madrid**, formato `AAAA-MM-DD HH:MM` |
| `estacion_origen` | Nombre literal de la estación de origen (puede contener tildes/ñ) |
| `estacion_destino` | Nombre literal de la estación de destino |
| `duracion_minutos` | Entero registrado al cierre (HU-03; `ceil(segundos/60)`) |
| `coste_eur` | Decimal con **punto** y 2 decimales, p. ej. `0.00`, `3.00` (el punto es el único compatible con separador de columna coma — RFC 4180) |

- **Orden**: `fecha_inicio` ASC (empate → `id` ASC).
- **Rango vacío**: únicamente la fila de cabecera (caso límite del spec).
- **Streaming**: la primera fila viaja en cuanto la BD la produce; el consumo de memoria no crece con el tamaño del rango (caso "varios años").

#### Errores

Mismos que el endpoint de agregados: `401 NO_AUTENTICADO`, `403 ROL_NO_PERMITIDO`, `422 ENTRADA_INVALIDA` (mismas condiciones). En caso de error **no** se envían cabeceras de descarga: el error usa el envelope JSON global estándar.

### Ejemplos

```http
GET /api/v1/analitica/export?desde=2026-10-01&hasta=2026-10-31
# 200 text/csv — BOM + cabecera + filas de octubre ordenadas por inicio

GET /api/v1/analitica/export?desde=2027-01-01&hasta=2027-12-31
# 200 text/csv — solo BOM + cabecera (rango sin viajes aún)
```

---

## Consistencia gráficos ↔ CSV (FR-006, SC-002)

Ambos endpoints construyen **una sola vez** el predicado del rango y lo reutilizan en todas sus consultas ([research D3](../research.md)); `total_viajes` se deriva de la suma de las propias barras. Por construcción, para cualquier rango válido:

```text
total_viajes(gráficos) == número de filas de datos del CSV
total_coste(gráficos)  == suma de coste_eur de las filas del CSV
por_hora[i]            == filas del CSV cuya fecha_hora cae en la hora i
por_dia_semana[i]      == filas del CSV cuya fecha_hora cae en el día i (lunes=0)
```

Cualquier discrepancia es defecto (caso límite del spec) y se verifica en los tests T8a–T8c del [quickstart](../quickstart.md).

## Rendimiento (SC-001)

- Agregados: 3 consultas pequeñas indexadas en `ix_viajes_fecha` (Q1/Q2/Q3 de [data-model.md](../data-model.md)); presupuesto **< 5 s** para rangos de hasta un año (verificado por test cronometrado).
- Exportación: 1 consulta streaming con el mismo índice; sin presupuesto específico más allá de la apertura correcta del fichero (SC-004).

## `422 ENTRADA_INVALIDA` en el OpenAPI global (T019 — autorizada y aplicada)

El OpenAPI global (`specs/spec-001/contracts/openapi.yaml`) documenta desde T019 la respuesta `422` en `/analitica/viajes` y `/analitica/export` mediante la respuesta compartida `components/responses/Error` (envelope global), reflejando el `422 ENTRADA_INVALIDA` que devuelve la implementación ante rangos ausentes, inválidos o invertidos — patrón consolidado en spec-003/spec-004/spec-007 (entrada ilegible vs. estado del sistema). El detalle de condiciones y mensajes vive en este contrato.
