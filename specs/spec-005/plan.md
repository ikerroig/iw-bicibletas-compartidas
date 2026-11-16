# Plan: spec-005

## Alcance técnico

### Backend

- Añadir una migración para guardar `estacion_id` y garantizar una incidencia
  abierta por bicicleta.
- Implementar `services/incidencias`, con reporte transaccional y consulta del
  panel pendiente.
- Implementar `handlers/incidencias` con validación de descripción, tipo,
  bicicleta y estación.
- Validar que una bicicleta anclada se reporte desde su estación real; las
  bicicletas en viaje activo no tienen anclaje actual y mantienen el flujo
  permitido por HU-05.
- Registrar `POST /api/v1/averias`, `POST /api/v1/incidencias`,
  `GET /api/v1/incidencias` y `GET /api/v1/incidencias/pendientes`.
- Proteger los reportes con el rol `usuario` y el panel con
  `operario_taller`.

### Frontend

- Crear el cliente API de averías.
- Crear la pantalla de reporte con código de bicicleta, estación, tipo y
  descripción obligatoria.
- Crear el panel de incidencias pendientes y publicar ambas pantallas en el
  menú según el rol.
- Actualizar automáticamente el panel del taller en menos de 30 segundos,
  manteniendo también el botón de actualización manual.

## Flujo principal

```text
usuario -> POST /averias
        -> bloquear bicicleta
        -> insertar incidencia pendiente
        -> bicicleta = averiada
        -> commit
        -> recalcular estación / Redis
        -> confirmación al usuario

operario_taller -> GET /incidencias -> incidencias abiertas ordenadas por fecha
```

## Trazabilidad

| Requisito | Código previsto | Verificación |
| --- | --- | --- |
| FR-001, FR-006 | cliente, handler y `POST /averias` | formulario y respuesta 201/200 |
| FR-002, FR-003, FR-007 | `services/incidencias` | transacción, bloqueo y reintento |
| FR-004 | `services/estacion` reutilizado | mapa/contadores tras el reporte |
| FR-005 | `GET /incidencias` y panel | lista con estación, descripción y fecha |

## Decisiones de cierre

- La correspondencia entre bicicleta anclada y estación se comprueba dentro de
  la misma transacción que registra la incidencia. Si no coincide, la API
  devuelve `409 BICICLETA_EN_OTRA_ESTACION` y no cambia el estado de la
  bicicleta.
- El panel de incidencias vuelve a consultar el backend cada 10 segundos y
  conserva la actualización manual.
