# Quickstart: Desbloqueo por código de anclaje (HU-02)

**Branch**: `spec-002` | **Date**: 2026-11-04 | **Plan**: [plan.md](./plan.md)

Guía de validación. El arranque del entorno está en la [guía global](../spec-001/quickstart.md); este escenario amplía su S2.

## Requisitos

- `docker compose up -d`, migraciones y seed aplicados (el seed **borra** los datos de esa BD).
- Backend y frontend en marcha.
- Usuarios seed, contraseña `password`: `ana@si16.test` (abono vigente), `bruno@si16.test` (caducado), `carla@si16.test` (sin abono), `gestor@si16.test` (rol gestor).

## Tests automáticos

```text
cd backend  && go test ./...
cd frontend && npm test
```

Los gates T1 y T2 están en `backend/tests/integration/spec002_*_test.go`.

## Validación manual

Usa en cada caso un anclaje del tipo indicado; se localizan en el detalle de estación del mapa o con SQL.

| # | Requisito | Pasos | Esperado |
|---|---|---|---|
| V1 | FR-002 | Bruno y Carla desbloquean un anclaje con bicicleta | 403 `ABONO_VENCIDO`, mensaje de causa, sin cambios |
| V2 | FR-001, FR-005, FR-008 | Ana teclea el código en minúsculas con espacios (` e01-a01 `) | 201; la página muestra hora de inicio, código de bicicleta y estación de origen |
| V3 | FR-006 | Ana recarga la página | Sigue viendo el mismo viaje, sin formulario |
| V4 | FR-003 | Con viaje activo, Ana llama a la API con otro código | 409 `VIAJE_ACTIVO`, sin un segundo viaje |
| V5 | FR-004 | Otro usuario con abono prueba un código inexistente, una bici averiada y una estación fuera de servicio | 404 o 409 con causa, sin cambios |
| V6 | FR-009 | El gestor no ve «Desbloquear» en el menú y llama a la API | 403 `ROL_INSUFICIENTE` |
| V7 | FR-001 | Código con 65 caracteres o con `\u0000` | 422 `VALIDACION`, nunca 500 |
| V9 | FR-008 | Con viaje activo, Ana cierra sesión, vuelve a entrar y queda en el mapa | Aviso con hora de inicio y código de bicicleta; «Ver» lleva a `/alquiler`; desaparece al entrar otro usuario sin viaje |
| V8 | Const. IV | Mirar `bicis_disponibles` de la estación antes y después de V2 | Baja en 1 y `anclajes_libres` sube en 1, en ≤ 30 s |

## Vencimiento del abono (FR-002)

Lo cubren los tests: un abono que vence hoy según la hora de España sigue vigente hasta las 23:59:59 y deja de estarlo a las 00:00:00, comprobado con un instante fijo en ambos lados de la medianoche, en horario de invierno y de verano.

## Consulta de control

```sql
SELECT usuario_id, count(*) FROM viajes
WHERE estado = 'en_curso' GROUP BY usuario_id HAVING count(*) > 1;
```

Debe devolver 0 filas (SC-002).

## Referencias

- Contrato: [contracts/alquileres.md](./contracts/alquileres.md)
- Reglas y transiciones: [data-model.md](./data-model.md)
