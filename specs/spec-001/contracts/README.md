# Contracts: API REST y estado en vivo SI-16

**Branch**: `spec-001` | **Date**: 2026-10-20 | **Plan**: [plan.md](../plan.md) | **Modelo**: [data-model.md](../data-model.md)

Contratos de las interfaces que el sistema expone. Son **el acuerdo de integración entre las 8 specs y los 4 integrantes**: cualquier cambio requiere acuerdo del equipo (constitución, principio III).

## Archivos

| Archivo | Contenido |
|---|---|
| [openapi.yaml](./openapi.yaml) | OpenAPI 3.1 de todo `/api/v1`: 18 endpoints agrupados por `tag` = spec, con payloads, códigos de error y roles requeridos. |
| [redis-state.md](./redis-state.md) | Formato de las claves de estado en vivo (`estacion:{id}:estado`), escritura, frescura y fallback. |

## Convenciones globales

- **Base**: `http://localhost:8080/api/v1` (dev). Todas las respuestas `application/json; charset=utf-8` salvo la exportación CSV (`text/csv`).
- **Autenticación**: `Authorization: Bearer <JWT>` (HS256, exp 24 h; claims `sub`, `rol`, `nombre`). Sin cabecera o token inválido → `401 UNAUTENTICADO`. Rol insuficiente → `403 ROL_INSUFICIENTE`.
- **Envelope de error** (uniforme en todo el sistema):

  ```json
  { "error": { "codigo": "ABONO_VENCIDO", "mensaje": "Su abono anual no está vigente" } }
  ```

  Códigos de negocio usados: `ABONO_VENCIDO` (403), `VIAJE_ACTIVO` (409), `SIN_VIAJE_ACTIVO` (409), `ANCLAJE_NO_EXISTE` (404), `ANCLAJE_NO_DISPONIBLE` (409), `ANCLAJE_OCUPADO` (409), `BICICLETA_NO_DISPONIBLE` (409), `ESTACION_SIN_CAPACIDAD` (409), `VALIDACION` (422).
- **Fecha/hora**: RFC3339 UTC. **Moneda**: euros con 2 decimales (`coste_total`).
- **Propiedad por spec** (evita colisiones): cada tag corresponde al paquete de backend y a la página de frontend de la spec indicada en el mapa de rutas de `plan.md`; los endpoints de `auth` son la base compartida (spec-001).
- **Regla**: los clientes (frontend de cualquier spec) deben tratar los campos desconocidos como tolerantes (forward-compatible) y basarse en los `tag`/`operationId` del contrato para sus tests con MSW/kin-openapi.
