# Data model: spec-006

HU-06 reutiliza incidencias_mantenimientos, que ya contiene la información
común de averías e intervenciones:

| Campo | Uso en taller |
| --- | --- |
| operario_id | Técnico que firma el parte desde el JWT |
| piezas_sustituidas | Piezas cambiadas |
| notas_taller | Observaciones del técnico |
| fecha_reparacion | Fecha escrita por el backend al cerrar |
| estado | pendiente, en_reparacion o resuelta |

Una reparación concluida actualiza todas las incidencias abiertas de la
bicicleta dentro de una única transacción y cambia la bicicleta a disponible.
También libera cualquier anclaje anterior para evitar que una bicicleta
reparada aparezca disponible antes de confirmar su nueva ubicación.

El mantenimiento preventivo crea una fila resuelta con tipo_fallo = otro y
descripción Mantenimiento preventivo. La migración
0003_spec006_mantenimiento_preventivo.sql permite que ese registro no tenga
estacion_id, porque no procede de un reporte localizado en una estación.

La confirmación de ubicación bloquea la bicicleta y el anclaje con FOR UPDATE,
ocupa el anclaje en la misma transacción y actualiza Redis después del COMMIT.
Una bicicleta disponible sin anclaje sigue fuera de los contadores hasta
completar esta operación.
