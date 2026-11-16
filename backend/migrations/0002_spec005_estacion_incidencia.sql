-- 0002_spec005_estacion_incidencia.sql - datos propios de HU-05.
-- La migracion 0001 es la base compartida y no se modifica despues de creada.

-- +goose Up
-- +goose StatementBegin

ALTER TABLE incidencias_mantenimientos
  ADD COLUMN estacion_id BIGINT REFERENCES estaciones(id);

-- Las incidencias nuevas siempre indican donde se encontro la bicicleta.
-- El proyecto parte de la tabla inicial sin incidencias cargadas, por lo que
-- podemos hacer obligatorio el dato despues de anadir la columna.
ALTER TABLE incidencias_mantenimientos
  ALTER COLUMN estacion_id SET NOT NULL;

-- Solo puede quedar una incidencia abierta por bicicleta. Ademas de expresar
-- la regla de negocio, el indice protege el caso de escrituras concurrentes.
CREATE UNIQUE INDEX ux_incidencia_abierta_bicicleta
  ON incidencias_mantenimientos(bicicleta_id)
  WHERE estado <> 'resuelta';

CREATE INDEX ix_incidencias_estacion
  ON incidencias_mantenimientos(estacion_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS ix_incidencias_estacion;
DROP INDEX IF EXISTS ux_incidencia_abierta_bicicleta;
ALTER TABLE incidencias_mantenimientos DROP COLUMN IF EXISTS estacion_id;

-- +goose StatementEnd
