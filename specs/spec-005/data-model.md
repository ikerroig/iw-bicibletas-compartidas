# Data model: spec-005

La entidad existente `incidencias_mantenimientos` se amplía de forma aditiva:

| Campo | Tipo | Regla |
| --- | --- | --- |
| `estacion_id` | `BIGINT NOT NULL` | FK a `estaciones(id)`; estación donde se encontró la bicicleta |

También se crea `ux_incidencia_abierta_bicicleta`, un índice único parcial
sobre `bicicleta_id` mientras `estado <> 'resuelta'`. Una bicicleta puede tener
muchas incidencias históricas resueltas, pero solo una incidencia pendiente o en
reparación a la vez.

La migración se encuentra en
`backend/migrations/0002_spec005_estacion_incidencia.sql` para conservar la
migración base compartida sin cambios.

El resto de la semántica es la del modelo común: una bicicleta `averiada` no
cuenta en `bicis_disponibles`, no puede iniciar un alquiler y puede terminar un
viaje activo sin volver a estado `disponible`.
