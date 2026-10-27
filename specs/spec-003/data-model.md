# Data Model: Declaración y cierre de viaje (HU-03, subconjunto del modelo compartido)

**Branch**: `spec-003` | **Date**: 2026-10-27 | **Plan**: [plan.md](./plan.md) | **Decisiones**: [research.md](./research.md)

**Este documento NO define esquema nuevo.** La fuente única de verdad es el modelo compartido en [specs/spec-001/data-model.md](../spec-001/data-model.md); la HU-03 **lee y escribe** un subconjunto de sus tablas ya existentes en `0001_esquema_base.sql`. Cualquier divergencia de este documento respecto al modelo compartido se resuelve a favor del global y requiere acuerdo del equipo (constitución III).

## Entidades tocadas por la devolución

| Tabla | Lectura (con bloqueo) | Escritura | Campos afectados |
|---|---|---|---|
| `viajes` | viaje `en_curso` del usuario (`SELECT … FOR UPDATE`) | cierre | `fecha_fin`, `duracion_minutos`, `coste_total`, `estacion_destino_id`, `estado → 'finalizado'` |
| `anclajes` | anclaje por `codigo_anclaje` (`SELECT … FOR UPDATE`) | ocupación | `estado → 'ocupado'`, `bicicleta_id` |
| `bicicletas` | id de la bici del viaje | disponibilidad | `estado → 'disponible'` (excepto si ya era `'averiada'`: se conserva) |
| `estaciones` | id de la estación del anclaje (para Redis) | — | — |

Entidades **no** tocadas: `usuarios` (solo se lee el id del JWT; la "liberación" del usuario es implícita: no tendrá viaje `en_curso`), `incidencias_mantenimientos`, `alertas_balanceo`.

## Máquina de estados relevante

```text
viaje:      en_curso --cierre (HU-03)--> finalizado        (única vez; FR-006)
bicicleta:  en_uso  --devolución------> disponible          (caso normal)
            en_uso  --devolución------> averiada            (si fue reportada HU-05 durante el viaje: queda anclada y averiada)
anclaje:    libre   --devolución------> ocupado (+ bicicleta_id)
```

Transiciones garantizadas:
- `viajes.estado = 'finalizado'` exige `fecha_fin`, `estacion_destino_id` y `duracion_minutos` NOT NULL (CHECK del modelo compartido) — el cierre incompleto es imposible a nivel de BD (edge case "cálculo falla": o se cierra completo o se hace rollback y se reintenta; nunca se libera al usuario sin registro de cierre).
- `en_curso → finalizado` solo puede ocurrir una vez: el UPDATE filtrado por `estado='en_curso'` + índice `ux_viajes_usuario_activo` hacen imposible el doble cierre (doble confirmación simultánea → solo la primera tiene efecto).
- Una bicicleta en más de un anclaje: imposible por `UNIQUE (bicicleta_id)` en `anclajes` (FR-006).

## Reglas de cálculo aplicadas al cierre (detalle operativo)

Dentro de la transacción de devolución (orden completo en [research.md D2](./research.md)):

1. `segundos = fecha_fin − fecha_inicio` con `fecha_fin` validada (nunca negativa; si el reloj inconsistente da `< 0`, se usa `now()`).
2. `duracion_minutos = ceil(segundos / 60)` → columna `viajes.duracion_minutos` (CHECK ≥ 0).
3. Coste:
   - `segundos ≤ 1800` → `coste_total = 0.00` (FR-003); respuesta incluye `etiqueta = "Tramo gratuito (0,00 €)"` (HU-04).
   - `segundos > 1800` → `coste_total = ceil((segundos − 1800) / 60) × TARIFA_MINUTO`, piso `0.00` (FR-004). `TARIFA_MINUTO` = variable de entorno (configuración operativa, valor de ejemplo 0,20).
4. `estacion_destino_id` = estación del anclaje usado (cualquier estación de la red, no la de origen — asunción del spec).

El coste es `NUMERIC(10,2) CHECK (coste_total >= 0)` — la no-negatividad está reforzada en BD.

## Efecto en el estado en vivo (Redis)

Tras el COMMIT, la estación destino ve afectados sus contadores (fórmula R9 del modelo global, recalculados — nunca incrementos ciegos):

```text
anclajes_libres(destino)  −1
bicis_disponibles(destino) +1   (solo si la bici pasó a 'disponible'; si sigue 'averiada', no suma)
ocupacion_pct(destino)    recalculado
actualizado_en            = now()   (RFC3339 UTC), TTL 120 s
```

Clave: `estacion:{id}:estado` (HASH), formato completo en [specs/spec-001/contracts/redis-state.md](../spec-001/contracts/redis-state.md). El barrido de 15 s del worker global actúa como red de seguridad si la escritura post-commit falla (constitución IV).

## Validaciones de dominio (server-side, const. I)

| # | Validación | Fuente | Comportamiento |
|---|---|---|---|
| V1 | El usuario autenticado tiene exactamente un viaje `en_curso` | FR-007, spec | Sin viaje → 409 `SIN_VIAJE_ACTIVO`; anclaje intacto |
| V2 | El anclaje existe y está `libre` | FR-001, edge "todos ocupados" | No encontrado/ocupado/fuera de servicio → 409 con código específico; viaje sigue activo |
| V3 | Cierre único del viaje | FR-006 | UPDATE condicionado; doble confirmación → 409 `VIAJE_YA_CERRADO` |
| V4 | Duración ≥ 0 y coste ≥ 0 | edge "reloj inconsistente" | Reloj invalidado → `now()`; CHECKs de BD como red final |
| V5 | Bici averiada durante el viaje se devuelve anclada y averiada | edge HU-05 | `estado` de la bici no se fuerza a `disponible` si ya era `averiada` |

## Migraciones

**Ninguna.** La HU-03 no añade tablas, columnas, índices ni restricciones. Se usa `0001_esquema_base.sql` tal cual. No hay nada que someter a acuerdo de equipo bajo el principio III.
