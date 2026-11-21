-- 0004_spec-007_traslados.sql - auditoría de traslados HU-07.

-- +goose Up
-- +goose StatementBegin

CREATE TABLE traslados_balanceo (
  id                   BIGSERIAL PRIMARY KEY,
  estacion_origen_id   BIGINT NOT NULL REFERENCES estaciones(id),
  estacion_destino_id  BIGINT NOT NULL REFERENCES estaciones(id),
  unidades             INTEGER NOT NULL CHECK (unidades > 0),
  fecha                TIMESTAMPTZ NOT NULL DEFAULT now(),
  operario_id          BIGINT NOT NULL REFERENCES usuarios(id),
  CHECK (estacion_origen_id <> estacion_destino_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS traslados_balanceo;

-- +goose StatementEnd
