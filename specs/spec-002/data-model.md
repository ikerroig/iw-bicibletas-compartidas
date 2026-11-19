# Data Model: Desbloqueo por código de anclaje (HU-02)

**Branch**: `spec-002` | **Date**: 2026-11-04 | **Plan**: [plan.md](./plan.md)

**Sin cambios de esquema.** Subconjunto del modelo compartido ([specs/spec-001/data-model.md](../spec-001/data-model.md), migración `0001_esquema_base.sql`), que sigue siendo la fuente única.

## Entidades usadas

| Tabla | Campos relevantes | Uso en HU-02 |
|---|---|---|
| `usuarios` | `rol`, `abono_activo`, `fecha_vencimiento_abono` | Lectura en cada solicitud; nunca desde el JWT, salvo el rol |
| `anclajes` | `codigo_anclaje`, `estacion_id`, `estado`, `bicicleta_id` | `FOR UPDATE` y actualización |
| `estaciones` | `nombre`, `estado` | `FOR SHARE`; nombre para el viaje activo |
| `bicicletas` | `codigo`, `estado` | `FOR UPDATE` y actualización |
| `viajes` | Todos | Inserción y lectura del viaje `en_curso` |

## Reglas de validación

| Regla | Origen | Dónde se aplica |
|---|---|---|
| Abono vigente ⇔ `abono_activo` y `fecha_vencimiento_abono` ≥ fecha actual en `Europe/Madrid` | FR-002 | Servicio. Si no es vigente, la causa es «sin abono» cuando `abono_activo = false` o no hay fecha, y «caducado» cuando la fecha ya ha pasado (D15, 2026-11-19) |
| Código: sin espacios alrededor, en mayúsculas, 1–64 caracteres, sin caracteres de control | FR-001 | Backend |
| Sin viaje `en_curso` del usuario | FR-003 | Servicio + `ux_viajes_usuario_activo` |
| Estación `operativa`, anclaje `ocupado`, bicicleta `disponible` | FR-004 | Servicio |
| Bicicleta en un solo viaje activo | FR-007 | `ux_viajes_bicicleta_activa` |
| Solo rol `usuario` | FR-009 | Middleware |

**Invariante de datos**: `anclajes.codigo_anclaje` se guarda siempre en mayúsculas. La BD no lo impone; lo cumplen el seed y cualquier alta futura.

## Transiciones de estado

Un desbloqueo válido aplica, en una transacción:

| Entidad | Antes | Después |
|---|---|---|
| Anclaje | `ocupado`, con bicicleta | `libre`, sin bicicleta |
| Bicicleta | `disponible` | `en_uso` |
| Viaje | (no existe) | `en_curso`, con `fecha_inicio` y `estacion_origen_id` |

Un rechazo no aplica ninguna. «Desacoplar» es solo este cambio de estado: la orden física queda fuera de alcance.

## Vista del viaje activo

No es una tabla: es la forma de la respuesta (FR-008).

| Campo | Origen |
|---|---|
| Campos de `Viaje` | `viajes` |
| `bicicleta_codigo` | `bicicletas.codigo` |
| `estacion_origen_nombre` | `estaciones.nombre` |

## Estado en vivo (Redis)

Sin claves nuevas. Tras el commit, `estacion:{id}:estado` de la estación de origen baja un contador de bicicletas y sube uno de anclajes libres ([contrato global](../spec-001/contracts/redis-state.md)).
