# Contract: POST /devoluciones (HU-03)

**Branch**: `spec-003` | **Date**: 2026-10-27 | **Plan**: [plan.md](./plan.md)

Contrato del único endpoint de spec-003. Alineado con el OpenAPI global (`specs/spec-001/contracts/openapi.yaml`, tag `spec-003`, operationId `devolver`); en caso de divergencia prevalece el OpenAPI global y este documento se corrige.

## Endpoint

```http
POST /api/v1/devoluciones
Authorization: Bearer <JWT>          # rol: usuario
Content-Type: application/json
```

Cierra el viaje activo del usuario autenticado anclando la bicicleta en el anclaje indicado y lo tarifica.

### Request body

```json
{ "codigo_anclaje": "AN-00456" }
```

| Campo | Tipo | Obligatorio | Validación |
|---|---|---|---|
| `codigo_anclaje` | string | sí | no vacío tras trim, ≤ 64 chars; formato no validado (la existencia se resuelve en BD) |

### Respuestas

#### `200 OK` — viaje finalizado

Schema `Viaje` global:

```json
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
  "etiqueta": "Tramo gratuito (0,00 €)"
}
```

Efectos garantizados (todos dentro de una transacción):
- `viajes`: `fecha_fin`, `duracion_minutos`, `coste_total`, `estacion_destino_id` escriturados y `estado = 'finalizado'`;
- `anclajes`: el anclaje pasa a `ocupado` con `bicicleta_id` de la bicicleta del viaje;
- `bicicletas`: estado pasa a `disponible` (o se conserva `averiada` si fue reportada durante el viaje);
- usuario queda libre (sin viaje `en_curso`) — puede desbloquear de inmediato;
- Redis: `estacion:{id}:estado` de la estación destino reescrita post-commit (`actualizado_en` nuevo).

#### Errores (envelope global `{ "error": { "codigo", "mensaje" } }`)

| HTTP | `codigo` | Condición | Efectos |
|---|---|---|---|
| 401 | `NO_AUTENTICADO` | Sin JWT o token inválido/expirado | Ninguno |
| 403 | `ROL_NO_PERMITIDO` | JWT válido con rol distinto de `usuario` | Ninguno |
| 422 | `ENTRADA_INVALIDA` | `codigo_anclaje` ausente, vacío o > 64 chars | Ninguno |
| 409 | `ANCLAJE_NO_ENCONTRADO` | `codigo_anclaje` inexistente | Ninguno (viaje sigue activo) |
| 409 | `ANCLAJE_OCUPADO` | Anclaje con bicicleta alojada | Ninguno (viaje sigue activo) |
| 409 | `ANCLAJE_FUERA_DE_SERVICIO` | Anclaje retirado de la red | Ninguno (viaje sigue activo) |
| 409 | `SIN_VIAJE_ACTIVO` | El usuario no tiene viaje `en_curso` | Ninguno (anclaje intacto) |
| 409 | `VIAJE_YA_CERRADO` | Segunda confirmación casi simultánea del mismo viaje | Ninguno (solo la primera devolución tiene efecto) |

**Invariante**: cualquier respuesta distinta de `200` deja la base de datos **exactamente** como estaba (rollback total; edge case "cálculo falla o queda sin confirmar": nunca un cierre parcial ni usuario liberado sin registro).

### Concurrencia

- El endpoint es seguro ante llamadas concurrentes del mismo usuario o hacia el mismo anclaje: `SELECT … FOR UPDATE` sobre anclaje y viaje + UPDATE condicionado por `estado = 'en_curso'` + índice único parcial `ux_viajes_usuario_activo` como red de seguridad a nivel de BD.
- Exactamente una de N llamadas simultáneas devuelve `200`; el resto reciben `409` con código correspondiente.

### Tarificación (resumen; detalle en [data-model.md](./data-model.md))

| Duración | `coste_total` |
|---|---|
| ≤ 30:00 (≤ 1800 s) | `0.00` + `etiqueta: "Tramo gratuito (0,00 €)"` |
| > 30:00 | `ceil((segundos − 1800) / 60) × TARIFA_MINUTO` |

`TARIFA_MINUTO` es configuración operativa del servidor (env); el cliente nunca calcula el importe (constitución I).
