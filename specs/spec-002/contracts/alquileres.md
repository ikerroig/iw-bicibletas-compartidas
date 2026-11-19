# Contract: Alquileres (HU-02)

**Branch**: `spec-002` | **Date**: 2026-11-04 | **Plan**: [plan.md](./plan.md)

Alineado con el OpenAPI global (`specs/spec-001/contracts/openapi.yaml`, tag `spec-002`, operationId `desbloquear` y `viajeActivo`). En caso de divergencia prevalece el OpenAPI y este documento se corrige.

## POST /api/v1/alquileres/desbloqueo

```http
POST /api/v1/alquileres/desbloqueo
Authorization: Bearer <JWT>          # rol: usuario
Content-Type: application/json
```

```json
{ "codigo_anclaje": "E01-A01" }
```

| Campo | Tipo | Obligatorio | Validación |
|---|---|---|---|
| `codigo_anclaje` | string | sí | Sin distinguir mayúsculas ni espacios alrededor; 1–64 caracteres, sin caracteres de control |

### Respuesta `201 Created`

```json
{
  "id": 1,
  "usuario_id": 1,
  "bicicleta_id": 1,
  "bicicleta_codigo": "BICI-0001",
  "estacion_origen_id": 1,
  "estacion_origen_nombre": "Plaça de l'Ajuntament",
  "estacion_destino_id": null,
  "fecha_inicio": "2026-11-04T08:45:56Z",
  "fecha_fin": null,
  "duracion_minutos": null,
  "coste_total": 0,
  "estado": "en_curso"
}
```

`bicicleta_codigo` y `estacion_origen_nombre` son aditivos al esquema `Viaje` y opcionales: los demás endpoints que devuelven `Viaje` no están obligados a incluirlos.

### Errores

Envelope uniforme `{"error":{"codigo","mensaje"}}`. Ningún rechazo tiene efectos.

| Status | Código | Causa |
|---|---|---|
| 401 | `UNAUTENTICADO` | Sin token o token inválido |
| 403 | `ROL_INSUFICIENTE` | Rol distinto de `usuario` |
| 403 | `SIN_ABONO` | Abono inactivo, sin fecha o usuario inexistente. Mensaje: «No tienes abono anual» (D15) |
| 403 | `ABONO_VENCIDO` | Abono activo con fecha de vencimiento anterior a hoy (`Europe/Madrid`). Mensaje: «Tu abono anual caducó el DD/MM/AAAA» (D15) |
| 404 | `ANCLAJE_NO_EXISTE` | Ningún anclaje con ese código |
| 409 | `VIAJE_ACTIVO` | El usuario ya tiene un viaje en curso |
| 409 | `ANCLAJE_NO_DISPONIBLE` | Estación o anclaje fuera de servicio, o anclaje vacío (también si otro usuario acaba de llevarse la bicicleta) |
| 409 | `BICICLETA_NO_DISPONIBLE` | Bicicleta averiada o en taller |
| 422 | `VALIDACION` | Cuerpo inválido o código que no cumple la validación |

Si coinciden varias causas se devuelve una sola, la primera que falla en este orden (FR-012, research D4): abono → viaje activo → anclaje existe → estación en servicio → anclaje en servicio → anclaje con bicicleta → bicicleta disponible.

Un `5xx` o una respuesta que no llega no garantizan que no se haya creado el viaje (research D12). El cliente no debe reenviar el `POST`: debe consultar `GET /alquileres/activo` (FR-011, research D17).

## GET /api/v1/alquileres/activo

```http
GET /api/v1/alquileres/activo
Authorization: Bearer <JWT>          # cualquier rol; solo el viaje propio
```

| Status | Cuerpo |
|---|---|
| 200 | `{"viaje": null, "abono": {…}}` o `{"viaje": { …como el 201… }, "abono": {…}}` |
| 401 | `UNAUTENTICADO` |

`abono` (añadido el 2026-11-19, FR-010, research D16) se calcula en el momento de la consulta con la misma función que decide el rechazo del `POST`, así que el texto coincide:

```json
{ "vigente": false, "codigo": "ABONO_VENCIDO", "mensaje": "Tu abono anual caducó el 09/11/2026" }
```

| Campo | Tipo | Valores |
|---|---|---|
| `vigente` | boolean | `true` si el usuario puede desbloquear por su abono |
| `codigo` | string o null | `null` si `vigente`; si no, `SIN_ABONO` o `ABONO_VENCIDO` |
| `mensaje` | string o null | `null` si `vigente`; si no, el mismo texto que el 403 del `POST` |

Es un aviso informativo: el `POST` vuelve a comprobar el abono en su transacción.

## Cambios en el OpenAPI global

Aditivos, a comunicar al equipo:

- `Viaje`: propiedades opcionales `bicicleta_codigo` y `estacion_origen_nombre`.
- `/alquileres/desbloqueo`: respuestas `401` y `422`; ejemplo `E01-A01`.
- `/alquileres/activo`: respuesta `401`.
- *2026-11-19*: `/alquileres/desbloqueo`, código `SIN_ABONO` documentado en la descripción del 403; `/alquileres/activo`, propiedad `abono` (requerida en la respuesta 200) con el esquema de arriba.
