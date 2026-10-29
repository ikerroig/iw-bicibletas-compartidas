# Contract: GET /viajes (HU-04)

**Branch**: `spec-004` | **Date**: 2026-10-28 | **Plan**: [plan.md](./plan.md)

Contrato del único endpoint de spec-004. Alineado con el OpenAPI global (`specs/spec-001/contracts/openapi.yaml`, tag `spec-004`, operationId `historialViajes`): se adoptan **literalmente** sus parámetros (`desde`, `hasta`, `page`, `page_size`) y su respuesta (`total`, `suma_coste`, `items`). Los dos campos de nombre de estación de cada item son **aditivos** al schema `Viaje` (patrón `AlertaBalanceo` del mismo contrato) y se documentan aquí; el fichero global no se modifica. En caso de divergencia no contemplada prevalece el OpenAPI global y este documento se corrige.

## Endpoint

```http
GET /api/v1/viajes?desde=2026-10-01&hasta=2026-10-31&page=1&page_size=50
Authorization: Bearer <JWT>          # cualquier rol autenticado; siempre son SU historial
```

Historial propio de viajes **finalizados** con filtro temporal opcional, paginación y sumatorio del periodo (HU-04).

### Query parameters

| Parámetro | Tipo | Obligatorio | Default | Validación |
|---|---|---|---|---|
| `desde` | `date` (`YYYY-MM-DD`) | no | sin límite inferior | formato estricto; `desde ≤ hasta`; ambos inclusivos; interpretados en **Europe/Madrid** (FR-002) |
| `hasta` | `date` (`YYYY-MM-DD`) | no | sin límite superior | ídem |
| `page` | integer | no | `1` | `≥ 1`; si no → `422 ENTRADA_INVALIDA` |
| `page_size` | integer | no | `50` | `1 … 200`; si no → `422 ENTRADA_INVALIDA` |

- Sin `desde`/`hasta` → historial completo (FR-003).
- Cada parámetro temporal es **opcional e independiente**: si solo viene `desde`, el rango queda abierto por arriba; si solo `hasta`, abierto por abajo (la UI envía siempre ambos juntos o ninguno).
- El filtro mes/año de la UI se envía como `desde` = primer día del mes y `hasta` = último día del mes.

### Respuestas

#### `200 OK`

```json
{
  "total": 7,
  "suma_coste": 4.6,
  "items": [
    {
      "id": 42,
      "usuario_id": 7,
      "bicicleta_id": 15,
      "estacion_origen_id": 3,
      "estacion_destino_id": 21,
      "fecha_inicio": "2026-10-27T10:00:00Z",
      "fecha_fin": "2026-10-27T10:25:00Z",
      "duracion_minutos": 25,
      "coste_total": 0.00,
      "estado": "finalizado",
      "etiqueta": "Tramo gratuito (0,00 €)",
      "estacion_origen_nombre": "Plaza Mayor",
      "estacion_destino_nombre": "Estación Norte"
    }
  ]
}
```

| Campo | Semántica |
|---|---|
| `total` | Nº de viajes de **todo** el periodo (sin paginar) — habilita "Página X de Y" |
| `suma_coste` | Suma de `coste_total` de **todo** el periodo — idéntica en cualquier página (FR-003, SC-002) |
| `items` | La página solicitada, orden `fecha_inicio DESC` (viajes más recientes primero) |
| `etiqueta` | `"Tramo gratuito (0,00 €)"` cuando `coste_total = 0`; `null`/ausente en caso contrario (FR-004) |
| `estacion_origen_nombre` / `estacion_destino_nombre` | Nombres resueltos en el backend con JOIN (aditivos al schema `Viaje`) |

Garantías del `200`:
- Solo viajes con `estado = 'finalizado'` **del usuario del token** — jamás viajes ajenos ni el viaje `en_curso` (FR-005).
- `coste_total` y `duracion_minutos` son los registrados al cierre por HU-03; el backend no recalcula (FR-007).
- **Página fuera de rango** (`page` > última): `items: []` **sin error**, conservando `total` y `suma_coste` del periodo (FR-008, clarificación 5).
- **Periodo sin viajes**: `total: 0`, `suma_coste: 0`, `items: []` (caso límite).

#### Errores (envelope global `{ "error": { "codigo", "mensaje" } }`)

| HTTP | `codigo` | Condición | Efectos |
|---|---|---|---|
| 401 | `NO_AUTENTICADO` | Sin JWT o token inválido/expirado | Ninguno |
| 422 | `ENTRADA_INVALIDA` | `page < 1`, `page_size` fuera de `1…200`, `desde`/`hasta` con formato distinto de `YYYY-MM-DD` o `desde > hasta` | Ninguno |

No hay más códigos de error: la lectura nunca modifica datos y "página fuera de rango" o "periodo vacío" **no** son errores.

### Ejemplos de consulta

```http
GET /api/v1/viajes                          # historial completo, primera página (50)
GET /api/v1/viajes?desde=2026-10-01&hasta=2026-10-31   # octubre 2026 (mes/año de la UI)
GET /api/v1/viajes?page=2&page_size=50      # segunda página; suma_coste sigue siendo el del periodo
GET /api/v1/viajes?desde=2099-01-01&hasta=2099-01-31   # 200, items: [], suma_coste: 0
```

### Rendimiento (SC-004)

Dos consultas indexadas (`ix_viajes_usuario_fecha` + PK de `estaciones`) por petición; sin N+1 (los nombres vienen de la misma sentencia). Objetivo: respuesta en < 3 s para un año completo de uso habitual.
