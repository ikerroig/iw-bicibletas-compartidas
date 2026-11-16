# Plan: spec-006

## Alcance técnico

### Backend

- Reutilizar la tabla de incidencias para mostrar el historial de la bicicleta.
- Implementar el servicio transaccional de partes de reparación y mantenimiento
  preventivo.
- Resolver incidencias abiertas, firmarlas con el operario autenticado y
  separar la reactivación de la ubicación física.
- Implementar la ubicación en un anclaje libre con bloqueo de concurrencia.
- Registrar las rutas de ficha, intervención y ubicación con autorización
  exclusiva para operario_taller.

### Frontend

- Añadir el cliente API de taller.
- Crear la pantalla /taller con consulta por ID de bicicleta.
- Mostrar el reporte original, el historial y el formulario del parte.
- Añadir la confirmación de ubicación cuando la bicicleta ya esté operativa.

## Flujo principal

    operario_taller -> GET /taller/bicicletas/{id}
                     -> visualizar reporte e historial
                     -> POST /intervenciones
                     -> cerrar incidencias y bicicleta = disponible sin anclaje
                     -> POST /ubicacion
                     -> anclaje = ocupado y Redis actualizado
                     -> bicicleta visible en el mapa

## Trazabilidad

| Requisito | Código | Verificación |
| --- | --- | --- |
| FR-001, FR-004 | servicio de ficha e historial | reporte original y mantenimiento visible |
| FR-002, FR-003, FR-006 | servicio de intervención | firma, fecha, cierre y estado |
| FR-005, FR-007 | servicio de ubicación | anclaje libre, Redis y contadores |
| Seguridad | rutas y middleware | 403 para roles distintos |

## Decisiones de cierre

- La operación de reparación y la de ubicación son independientes para
  respetar FR-007.
- Las operaciones críticas se ejecutan con transacciones y bloqueos de filas.
- Redis se actualiza después del COMMIT; PostgreSQL sigue siendo la fuente de
  verdad si Redis no está disponible.
