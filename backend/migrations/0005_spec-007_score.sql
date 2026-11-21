-- 0005_spec-007_score.sql - orden de urgencia del panel HU-07.

-- +goose Up
-- +goose StatementBegin

CREATE INDEX ix_alertas_balanceo_urgencia
  ON alertas_balanceo (
    (CASE WHEN tipo_alerta = 'vacia_menos_10'
          THEN 10 - porcentaje_ocupacion
          ELSE porcentaje_ocupacion - 90 END) DESC,
    fecha_generacion ASC
  )
  WHERE NOT resuelta;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS ix_alertas_balanceo_urgencia;

-- +goose StatementEnd
