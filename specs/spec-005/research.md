# Research: spec-005

## Decisiones

1. Se reutiliza `incidencias_mantenimientos`, ya presente en el modelo común.
   La estación del reporte se añade con la migración `0002_spec005_estacion_incidencia.sql`;
   no se modifica la migración inicial compartida.
2. El reporte usa una transacción PostgreSQL y bloquea la fila de la bicicleta
   con `FOR UPDATE`. Así el INSERT de la incidencia y el estado `averiada` se
   confirman juntos o se deshacen juntos.
3. Solo existe una incidencia abierta por bicicleta mediante una consulta bajo
   bloqueo y un índice único parcial. Un reintento devuelve la incidencia
   existente y no repite la transición.
4. El usuario reporta con el código de bicicleta y la estación seleccionada.
   El servicio de taller consulta las incidencias abiertas con bicicleta,
   estación, descripción, tipo y fecha.
5. Después del commit se llama a `RecalcularEstacion`. Si Redis no responde,
   el cambio de PostgreSQL se conserva y los mecanismos de refresco/fallback
   existentes recuperan el estado.

## Decisiones descartadas

- No se edita `0001_esquema_base.sql`, porque es la base acordada por el equipo.
- No se permite reportar una bicicleta `en_taller`: el operario ya controla
  esa unidad y el reporte de usuario no debe alterar el flujo de reparación.

## Decisiones añadidas durante la implementación

6. La estación enviada por el usuario se contrasta con el anclaje actual de la
   bicicleta dentro de la transacción. Esto evita asociar una avería a una
   estación diferente por un código mal introducido o por datos obsoletos.
7. El panel del taller se refresca cada 10 segundos, por debajo del límite de
   30 segundos indicado para la HU-05.
