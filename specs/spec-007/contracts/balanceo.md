# Contract: GET /balanceo/alertas y POST /balanceo/traslados (HU-07)

**Branch**: `spec-007` | **Date**: 2026-10-30 | **Plan**: [plan.md](./plan.md)

Contrato de los dos endpoints de spec-007. Alineados con el OpenAPI global (`specs/spec-001/contracts/openapi.yaml`, tag `spec-007`, operationIds `panelBalanceo` y `confirmarTraslado`): se adoptan **literalmente** sus paths, schemas (`AlertaBalanceo`, `TrasladoRequest`, `EstacionEstado`) y respuestas de error. El fichero global **no se modifica**; este documento detalla la semántica, los mensajes y los casos de validación que el YAML deja abiertos. En caso de divergencia no contemplada prevalece el OpenAPI global y este documento se corrige.

## Autorización (ambos endpoints)

```http
Authorization: Bearer <JWT>     # rol obligatorio: operario_logistica
```

Middleware encadenado `RequireAuth` + `RequireRole("operario_logistica")` (patrón ya usado en la rama). Sin token o token inválido → `401 NO_AUTENTICADO`; token de cualquier otro rol → `403 ROL_NO_PERMITIDO`.

---

## 1. `GET /api/v1/balanceo/alertas`

Panel de logística: lista de estaciones descompensadas **ordenada por urgencia de intervención** (FR-005), solo alertas abiertas (SC-002/SC-004). Sin parámetros.

### Respuestas

#### `200 OK`

```json
{
  "items": [
    {
      "id": 3,
      "estacion_id": 3,
      "estacion_nombre": "Estació Nord",
      "tipo_alerta": "vacia_menos_10",
      "porcentaje_ocupacion": 5.00,
      "severidad": "alta",
      "fecha_generacion": "2026-10-30T09:12:44Z",
      "resuelta": false
    },
    {
      "id": 1,
      "estacion_id": 4,
      "estacion_nombre": "Estació Port",
      "tipo_alerta": "llena_mas_90",
      "porcentaje_ocupacion": 95.00,
      "severidad": "alta",
      "fecha_generacion": "2026-10-30T08:58:01Z",
      "resuelta": false
    }
  ]
}
```

| Campo | Semántica |
|---|---|
| `items` | Alertas con `resuelta = false`, **ya ordenadas por el backend** por urgencia (extremos más alejados del rango `[10, 90]` primero; empates → `fecha_generacion` ASC). El frontend las pinta tal cual (const. I). |
| `estacion_nombre` | Resuelto en backend con JOIN (campo del schema global `AlertaBalanceo`) |
| `porcentaje_ocupacion` | Snapshot de ocupación en la última evaluación del worker (< 60 s = `periodo + T_barrido`, research D6); formatear con `es-ES` |
| `severidad` | Siempre `"alta"` (HU-07) |
| `items: []` | Red equilibrada — **no** es error (caso límite) |

Garantías del `200`:
- **0 duplicados** por (estación, umbral): unicidad garantizada por `ux_alerta_abierta` (FR-003).
- Solo alertas **abiertas**: las resueltas desaparecen del panel automáticamente (FR-004).
- Estaciones `fuera_de_servicio` nunca aparecen (excluidas del cálculo).

#### Errores (envelope global `{ "error": { "codigo", "mensaje" } }`)

| HTTP | `codigo` | Condición | Efectos |
|---|---|---|---|
| 401 | `NO_AUTENTICADO` | Sin JWT o token inválido/expirado | Ninguno |
| 403 | `ROL_NO_PERMITIDO` | Rol distinto de `operario_logistica` | Ninguno |

No hay más códigos: la lectura nunca modifica datos.

### Ejemplos

```http
GET /api/v1/balanceo/alertas                     # panel completo (alertas abiertas, por urgencia)
```

---

## 2. `POST /api/v1/balanceo/traslados`

Confirmación por el operario de la carga/descarga de un lote de bicicletas entre dos estaciones (FR-006). Valida el lote, lo mueve en transacción y devuelve **ambas estaciones ya recalculadas** (FR-007, SC-003).

### Request body

```json
{
  "estacion_origen_id": 4,
  "estacion_destino_id": 3,
  "unidades": 6
}
```

| Campo | Tipo | Obligatorio | Validación |
|---|---|---|---|
| `estacion_origen_id` | integer | sí | distinto de `estacion_destino_id`; estación **existente** y `operativa` |
| `estacion_destino_id` | integer | sí | ídem |
| `unidades` | integer | sí | `≥ 1` (schema global); **≤ bicis disponibles en origen** y **≤ anclajes libres en destino** |

### Respuestas

#### `200 OK`

```json
{
  "origen": {
    "id": 4, "nombre": "Estació Port",
    "latitud": 39.4812, "longitud": -0.3711,
    "capacidad_total": 20, "estado": "operativa",
    "bicis_disponibles": 13, "anclajes_libres": 7,
    "ocupacion_pct": 65.00,
    "actualizado_en": "2026-10-30T10:15:32Z"
  },
  "destino": {
    "id": 3, "nombre": "Estació Nord",
    "latitud": 39.4770, "longitud": -0.3801,
    "capacidad_total": 20, "estado": "operativa",
    "bicis_disponibles": 7, "anclajes_libres": 13,
    "ocupacion_pct": 35.00,
    "actualizado_en": "2026-10-30T10:15:32Z"
  }
}
```

- `origen` y `destino` son el schema global `EstacionEstado`, montados con los valores **recalculados tras el commit** de la transacción (FR-007): la vista del operario refleja los nuevos contadores **en la propia respuesta** (SC-003, "de inmediato").
- El estado global (Redis `estacion:{id}:estado`, frescura ≤ 30 s — constitución IV) se escribe post-commit para ambas estaciones con el helper existente `cache.ActualizarEstadoEstacion`.
- Tras el commit también se **evalúan los umbrales** de ambas estaciones (inline, D1): si el origen queda `< 10 %` o el destino `> 90 %` (o viceversa), la siguiente lectura del panel ya las muestra; si una alerta abierta vuelve a rango `[10, 90]`, queda resuelta.

Garantías del `200`:
- El movimiento es **atómico**: o se aplican las `unidades` a origen y destino, o no se altera nada.
- Las bicicletas trasladadas conservan estado `disponible`; la ocupación se deriva de la fórmula R9 (averiadas excluidas — FR-008).
- Si los datos en pantalla del operario estaban algo desactualizados, el resultado usa los valores **reales bloqueados** de la transacción (caso límite del spec).

#### Errores (envelope global)

| HTTP | `codigo` | Condición | Efectos |
|---|---|---|---|
| 401 | `NO_AUTENTICADO` | Sin JWT o token inválido/expirado | Ninguno |
| 403 | `ROL_NO_PERMITIDO` | Rol distinto de `operario_logistica` | Ninguno |
| 422 | `ENTRADA_INVALIDA` | El **payload** es ilegible o no puede ser interpretado: JSON malformado, campos ausentes o no numéricos, `unidades < 1`, `estacion_origen_id == estacion_destino_id`. La petición ni siquiera llega a evaluarse contra el estado real | Ninguno (validación previa a la transacción) |
| 409 | `ESTACION_SIN_CAPACIDAD` | El payload es **bien formado** pero choca con el **estado actual** del sistema: estación origen o destino **inexistente** o **no operativa**; o el lote supera las bicis disponibles en origen; o supera los anclajes libres en destino (incl. carrera concurrente resuelta con `FOR UPDATE`) | **ROLLBACK íntegro**: las estaciones quedan exactamente igual que antes (caso límite del spec: "se rechaza la confirmación inconsistente sin alterar estaciones") |

Criterio de separación (decisión de diseño, sin cambios en el OpenAPI global): **422 = no puedo interpretar tu petición; 409 = la entiendo, pero el estado actual no permite ejecutarla.** El `422` no está enumerado en el OpenAPI global para este endpoint, pero sigue el envelope ya consolidado por spec-003/spec-004 (`apierr.Codigo`); el `409` mantiene su único `codigo` de negocio global `ESTACION_SIN_CAPACIDAD`. El `mensaje` del `409` distingue la causa concreta (estación inexistente / fuera de servicio / origen sin bicicletas suficientes / destino sin anclajes libres) para que el operario pueda corregir el formulario. Si el equipo prefiere ceñirse al OpenAPI literal, el fallback es mapear el `422` a `409`, perdiendo la distinción (no recomendado).

### Ejemplos

```http
POST /api/v1/balanceo/traslados
{ "estacion_origen_id": 4, "estacion_destino_id": 3, "unidades": 6 }   # 200: estados recalculados

POST /api/v1/balanceo/traslados
{ "estacion_origen_id": 4, "estacion_destino_id": 3, "unidades": 999 } # 409 ESTACION_SIN_CAPACIDAD (origen)

POST /api/v1/balanceo/traslados
{ "estacion_origen_id": 4, "estacion_destino_id": 4, "unidades": 3 }   # 422 ENTRADA_INVALIDA (mismo origen/destino)

POST /api/v1/balanceo/traslados
{ "estacion_origen_id": 4, "estacion_destino_id": 3, "unidades": 0 }   # 422 ENTRADA_INVALIDA (unidades < 1)
```

### Rendimiento (SC-001/SC-003)

- Traslado: 1 transacción corta (bloqueo de 2 filas de estaciones + 2 LIMIT + 2 UPDATE) + 2 refrescos Redis post-commit; objetivo < 10 s (holgura enorme).
- Generación de alertas: inline en el propio traslado; para el resto de mutaciones de la red, el worker garantiza < 60 s (SC-001; condiciones verificables en [research.md](../research.md) D6 — periodo 55 s + barrido medido < 5 s).
- Panel: 1 consulta indexada sobre alertas abiertas; objetivo < 3 s.
