# Research: spec-006

## Decisiones

1. Se reutiliza incidencias_mantenimientos en lugar de crear otra tabla:
   contiene el reporte de usuario y los campos del parte técnico.
2. La identidad del operario se toma del JWT y nunca del body de la petición.
3. La bicicleta se bloquea con FOR UPDATE antes de modificar incidencias o
   estado, evitando dos partes simultáneos inconsistentes.
4. Al marcar una reparación como operativa, la bicicleta queda disponible pero
   sin anclaje. La ubicación posterior es la que la hace aparecer en el mapa.
5. El mantenimiento preventivo se representa con una incidencia resuelta de
   tipo otro; no necesita estación ni usuario reportante.
6. La operación de ubicación bloquea bicicleta y anclaje dentro de la misma
   transacción y recalcula la estación tras el commit.

## Decisiones descartadas

- No se añade un sistema de autenticación nuevo: se reutilizan los roles y el
  middleware común.
- No se modifica 0001_esquema_base.sql; la compatibilidad del mantenimiento
  preventivo se resuelve en 0003_spec006_mantenimiento_preventivo.sql.
- No se muestran bicicletas reparadas en el mapa antes de confirmar un anclaje.
