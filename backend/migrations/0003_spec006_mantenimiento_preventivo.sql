-- 0003_spec006_mantenimiento_preventivo.sql - compatibilidad HU-06.
-- Un mantenimiento preventivo no procede de una estación ni de un reporte
-- de usuario, por lo que estacion_id debe poder quedar vacío.

-- +goose Up
-- +goose StatementBegin

ALTER TABLE incidencias_mantenimientos
  ALTER COLUMN estacion_id DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Las filas preventivas sin estación deben eliminarse antes de volver a hacer
-- obligatorio el campo al deshacer esta migración.
DELETE FROM incidencias_mantenimientos
WHERE estacion_id IS NULL;

ALTER TABLE incidencias_mantenimientos
  ALTER COLUMN estacion_id SET NOT NULL;

-- +goose StatementEnd
