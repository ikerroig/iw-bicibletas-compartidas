-- 0001_esquema_base.sql — Esquema compartido SI-16 (specs/spec-001/data-model.md).
-- Aportado por spec-001 y acordado por el equipo (constitución, principio III).
-- Cualquier cambio posterior va en una migración 000N_spec-00X_tema.sql con
-- ordinal acordado (research R6); no editar esta migración una vez fusionada.

-- +goose Up
-- +goose StatementBegin

CREATE TABLE usuarios (
  id                      BIGSERIAL PRIMARY KEY,
  nombre                  TEXT        NOT NULL,
  email                   TEXT        NOT NULL UNIQUE,
  password_hash           TEXT        NOT NULL,
  rol                     TEXT        NOT NULL DEFAULT 'usuario'
                          CHECK (rol IN ('usuario','operario_taller','operario_logistica','gestor')),
  abono_activo            BOOLEAN     NOT NULL DEFAULT FALSE,
  fecha_vencimiento_abono DATE,
  creado_en               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE estaciones (
  id              BIGSERIAL PRIMARY KEY,
  nombre          TEXT NOT NULL UNIQUE,
  latitud         NUMERIC(9,6) NOT NULL CHECK (latitud  BETWEEN -90  AND 90),
  longitud        NUMERIC(9,6) NOT NULL CHECK (longitud BETWEEN -180 AND 180),
  capacidad_total INTEGER     NOT NULL CHECK (capacidad_total > 0),
  estado          TEXT        NOT NULL DEFAULT 'operativa'
                  CHECK (estado IN ('operativa','fuera_de_servicio'))
);

CREATE TABLE bicicletas (
  id     BIGSERIAL PRIMARY KEY,
  codigo TEXT NOT NULL UNIQUE,
  estado TEXT NOT NULL DEFAULT 'disponible'
         CHECK (estado IN ('disponible','en_uso','averiada','en_taller'))
);

CREATE TABLE anclajes (
  id             BIGSERIAL PRIMARY KEY,
  estacion_id    BIGINT NOT NULL REFERENCES estaciones(id),
  codigo_anclaje TEXT   NOT NULL UNIQUE,
  estado         TEXT   NOT NULL DEFAULT 'libre'
                 CHECK (estado IN ('libre','ocupado','fuera_de_servicio')),
  bicicleta_id   BIGINT UNIQUE REFERENCES bicicletas(id),
  CHECK (estado <> 'ocupado' OR bicicleta_id IS NOT NULL),  -- ocupado ⇒ con bicicleta
  CHECK (bicicleta_id IS NULL OR estado <> 'libre')         -- libre ⇒ sin bicicleta
);
CREATE INDEX ix_anclajes_estacion ON anclajes(estacion_id);

CREATE TABLE viajes (
  id                  BIGSERIAL PRIMARY KEY,
  usuario_id          BIGINT      NOT NULL REFERENCES usuarios(id),
  bicicleta_id        BIGINT      NOT NULL REFERENCES bicicletas(id),
  estacion_origen_id  BIGINT      NOT NULL REFERENCES estaciones(id),
  estacion_destino_id BIGINT      REFERENCES estaciones(id),
  fecha_inicio        TIMESTAMPTZ NOT NULL DEFAULT now(),
  fecha_fin           TIMESTAMPTZ,
  duracion_minutos    INTEGER     CHECK (duracion_minutos IS NULL OR duracion_minutos >= 0),
  coste_total         NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (coste_total >= 0),
  estado              TEXT        NOT NULL DEFAULT 'en_curso'
                      CHECK (estado IN ('en_curso','finalizado')),
  CHECK (estado <> 'finalizado' OR (fecha_fin IS NOT NULL
                                    AND estacion_destino_id IS NOT NULL
                                    AND duracion_minutos IS NOT NULL)),
  CHECK (estado <> 'en_curso' OR (fecha_fin IS NULL AND estacion_destino_id IS NULL))
);
-- Reglas críticas garantizadas a nivel de BD (research R7):
CREATE UNIQUE INDEX ux_viajes_usuario_activo   ON viajes(usuario_id)   WHERE estado = 'en_curso';
CREATE UNIQUE INDEX ux_viajes_bicicleta_activa ON viajes(bicicleta_id) WHERE estado = 'en_curso';
CREATE INDEX ix_viajes_usuario_fecha ON viajes(usuario_id, fecha_inicio DESC);  -- HU-04
CREATE INDEX ix_viajes_fecha         ON viajes(fecha_inicio);                   -- HU-08

CREATE TABLE incidencias_mantenimientos (
  id                  BIGSERIAL PRIMARY KEY,
  bicicleta_id        BIGINT      NOT NULL REFERENCES bicicletas(id),
  usuario_reporta_id  BIGINT      REFERENCES usuarios(id),
  operario_id         BIGINT      REFERENCES usuarios(id),
  tipo_fallo          TEXT        NOT NULL
                      CHECK (tipo_fallo IN ('frenos','cadena','neumatico','otro')),
  descripcion_reporte TEXT        NOT NULL,
  piezas_sustituidas  TEXT,
  notas_taller        TEXT,
  fecha_reporte       TIMESTAMPTZ NOT NULL DEFAULT now(),
  fecha_reparacion    TIMESTAMPTZ,
  estado              TEXT        NOT NULL DEFAULT 'pendiente'
                      CHECK (estado IN ('pendiente','en_reparacion','resuelta')),
  CHECK (estado <> 'resuelta' OR (fecha_reparacion IS NOT NULL
                                  AND operario_id IS NOT NULL
                                  AND (piezas_sustituidas IS NOT NULL OR notas_taller IS NOT NULL)))
);
CREATE INDEX ix_incidencias_bicicleta  ON incidencias_mantenimientos(bicicleta_id);
CREATE INDEX ix_incidencias_pendientes ON incidencias_mantenimientos(fecha_reporte) WHERE estado <> 'resuelta';  -- panel HU-05/HU-06

CREATE TABLE alertas_balanceo (
  id                   BIGSERIAL PRIMARY KEY,
  estacion_id          BIGINT NOT NULL REFERENCES estaciones(id),
  tipo_alerta          TEXT   NOT NULL CHECK (tipo_alerta IN ('vacia_menos_10','llena_mas_90')),
  porcentaje_ocupacion NUMERIC(5,2) NOT NULL CHECK (porcentaje_ocupacion BETWEEN 0 AND 100),
  severidad            TEXT   NOT NULL DEFAULT 'alta',
  fecha_generacion     TIMESTAMPTZ NOT NULL DEFAULT now(),
  resuelta             BOOLEAN NOT NULL DEFAULT FALSE
);
-- Sin duplicados por estación y umbral (HU-07 FR-003):
CREATE UNIQUE INDEX ux_alerta_abierta  ON alertas_balanceo(estacion_id, tipo_alerta) WHERE NOT resuelta;
CREATE INDEX ix_alertas_abiertas ON alertas_balanceo(fecha_generacion DESC) WHERE NOT resuelta;  -- orden por urgencia

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS alertas_balanceo;
DROP TABLE IF EXISTS incidencias_mantenimientos;
DROP TABLE IF EXISTS viajes;
DROP TABLE IF EXISTS anclajes;
DROP TABLE IF EXISTS bicicletas;
DROP TABLE IF EXISTS estaciones;
DROP TABLE IF EXISTS usuarios;
-- +goose StatementEnd
