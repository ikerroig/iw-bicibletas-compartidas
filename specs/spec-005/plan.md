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

## Trazabilidad del refinamiento SDD

| Historia/criterio | Decisión de `clarify` | Tarea | Prueba o evidencia |
| --- | --- | --- | --- |
| HU-05 / FR-001 | C-001 y C-002: estación real si hay anclaje; excepción durante viaje | T011, T015 | `TestAveriaRechazaEstacionDistinta` y recorrido de reporte |
| HU-05 / FR-002, FR-003, FR-007 | C-003: reintento idempotente | T004, T013, T016 | `TestAveriaBloqueaYEsIdempotente` y comprobación de desbloqueo |
| HU-05 / FR-004, SC-003 | Excluir averiadas del mapa y contadores | T015, T017 | Mapa y contadores antes/después del reporte |
| HU-05 / FR-005, SC-002 | Panel de taller con bicicleta, estación, tipo, descripción y fecha | T008, T016, T017 | Endpoint y pantalla `/incidencias` |
| HU-05 / FR-006 | Confirmación de alta con respuesta 201/200 | T005, T013 | Respuesta de `POST /averias` |
| HU-05 / FR-008 | Refresco cada 10 segundos y actualización manual | T012, T017 | Inspección y prueba manual del panel |
| HU-05 / permisos | Solo usuario reporta y solo taller consulta pendientes | T005, T016, T018 | Respuestas 401/403 |

Todas las filas quedaron contrastadas mediante las pruebas, el recorrido final,
la matriz global y la revisión cruzada de la PR #15.
