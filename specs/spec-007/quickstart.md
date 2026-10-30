# Quickstart: validación de la HU-07 (monitorización de desequilibrios y alertas de balanceo)

**Branch**: `spec-007` | **Date**: 2026-10-30 | **Plan**: [plan.md](./plan.md) | **Contrato**: [contracts/balanceo.md](./contracts/balanceo.md) | **Modelo**: [data-model.md](./data-model.md)

Guía de validación **end-to-end** del balanceo de flota. Es una guía de ejecución; la implementación vive en `tasks.md` (Phase 2, pendiente de `/speckit.tasks`) y el código. Hereda el setup del [quickstart global](../spec-001/quickstart.md) y de [spec-004](../spec-004/quickstart.md) (infra, migrate, seed, server, frontend) sobre la base ya presente en esta rama.

## 0. Prerrequisitos

- Docker Desktop, Go ≥ 1.22, Node ≥ 20; puertos 5432, 6379, 8080, 5173 libres.
- `.env` presente en la raíz (copiado de `.env.example`).

## 1. Setup (igual que el global)

```powershell
copy .env.example .env        # revisar DATABASE_URL, REDIS_ADDR, JWT_SECRET
docker compose up -d
cd backend
go run ./cmd/migrate up
go run ./cmd/seed             # ahora incluye estaciones en extremos: 5 % y 95 % (research D8)
go run ./cmd/server           # terminal 1 — arranca también el worker alertas (D6)
cd ../frontend && npm install && npm run dev   # terminal 2
```

> El worker `alertas` evalúa la red al arrancar y después cada 60 s: con el seed de extremos, el panel tendrá alertas en **≤ 60 s** tras el arranque del servidor.

## 2. Pruebas automatizadas (gate constitución I — regla T5 + traslados)

```powershell
cd backend
go test ./... -v              # integración del balanceo contra la BD del compose
cd ../frontend
npm test                      # Vitest: página de balanceo
```

| Test | Escenario | Esperado |
|---|---|---|
| T7a | **Umbrales estrictos (gate T5, constitución I)** | Ocupación 9,99 % → alerta `vacia_menos_10`; **10,00 % → sin alerta**; **90,00 % → sin alerta**; 90,01 % → `llena_mas_90` |
| T7b | Unicidad y resolución (FR-003/FR-004) | Evaluar dos veces la misma estación → **1 sola** alerta abierta; volver a `[10, 90]` → `resuelta = true` y fuera del panel; re-cruzar → fila **nueva** abierta; 0 duplicados por (estación, umbral) |
| T7c | Exclusiones (FR-008 + caso límite) | Estación `fuera_de_servicio` → sin alertas; solo bicis `averiada` ancladas → cuenta como ocupación 0 (genera `vacia_menos_10`); bici `disponible` sin anclaje → no cuenta en ninguna estación |
| T7d | Panel (FR-005, SC-002) | `GET /balanceo/alertas` → solo abiertas, orden por urgencia (0 % antes que 9 %; 100 % antes que 91 %), `estacion_nombre` presente, `items: []` cuando la red está equilibrada |
| T7e | Roles (const. I) | Sin token → `401 NO_AUTENTICADO`; rol `usuario`/`gestor`/`operario_taller` → `403 ROL_NO_PERMITIDO` en **ambos** endpoints |
| T7f | Traslado válido (FR-006/FR-007, SC-003) | `POST /balanceo/traslados` con lote factible → `200` con `origen`/`destino` **recalculados en la propia respuesta**; Redis `estacion:{id}:estado` de ambas escrito post-commit; umbrales evaluados inline (p. ej. destino que recibe el lote cruza > 90 % → alerta en la siguiente lectura del panel) |
| T7g | Traslado inconsistente (caso límite) | `unidades > disponibles en origen` o `> anclajes libres en destino` o estación inexistente/fuera de servicio → `409 ESTACION_SIN_CAPACIDAD` **sin alterar nada** (recuentos idénticos antes/después); `unidades < 1`, origen = destino o JSON malformado → `422 ENTRADA_INVALIDA` |
| T7h | Latencia (SC-001) | Con una estación mutada a la fuerza (SQL) fuera de rango, el **worker** genera la alerta en < 60 s (aserción cronometrada en el test del worker: barrido único tras el tick) |

> T7a/T7b cubren la regla crítica T5 del [quickstart global](../spec-001/quickstart.md) (alertas <10 % / >90 %): son **obligatorias antes de fusionar** (gate de calidad de la constitución).

## 3. Escenario manual S7

```powershell
# Login del operario de logística (seed)
$r = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/auth/login `
     -Body '{"email":"logistica@si16.test","password":"password"}' -ContentType 'application/json'
$h = @{ Authorization = "Bearer $($r.token)" }

# Panel de balanceo (esperar ≤ 60 s tras el arranque del server para que el worker evalúe el seed)
Invoke-RestMethod -Uri http://localhost:8080/api/v1/balanceo/alertas -Headers $h | ConvertTo-Json -Depth 5

# Confirmar un traslado: 6 bicis desde la estación LLENA (95 %) hacia la VACÍA (5 %)
# (sustituye 4 y 3 por los ids reales de las estaciones en extremos del seed)
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/balanceo/traslados `
    -Headers $h -Body '{"estacion_origen_id":4,"estacion_destino_id":3,"unidades":6}' `
    -ContentType 'application/json' | ConvertTo-Json -Depth 5
```

**Esperado**: `200` con `origen` y `destino` reflejando los nuevos contadores **de inmediato**. Comprobar además:

1. **Panel ordenado** (SC-002): en `http://localhost:5173/balanceo`, la estación vacía (0–5 %) aparece **antes** que la llena si su extremo es mayor; 0 duplicados por estación/umbral; severidad "alta" y fecha visibles.
2. **Alertas en < 60 s** (SC-001): llevar una estación seed fuera de rango con SQL (`UPDATE anclajes…` o un segundo traslado) y observar aparecer su alerta en el panel en < 1 minuto (worker).
3. **Resolución automática** (SC-004): tras el traslado anterior, si la estación origen vuelve a `[10, 90]`, su alerta **desaparece sola** del panel sin recargar a mano (recarga tras el 200); la tabla `alertas_balanceo` muestra `resuelta = true`.
4. **Traslado rechazado** (caso límite): `unidades: 999` → `409 ESTACION_SIN_CAPACIDAD` y comprobar vía SQL que **nada cambió** en ambas estaciones; `origen == destino` y `unidades: 0` → `422 ENTRADA_INVALIDA`.
5. **Roles** (const. I): login con `ana@si16.test` (rol usuario) → `403 ROL_NO_PERMITIDO` en ambos endpoints.
6. **Estado global** (SC-003): `HGETALL estacion:4:estado` en redis-cli (estación de origen) refleja los nuevos contadores con `actualizado_en` del momento del traslado (≤ 30 s, constitución IV).

## 4. Criterios de aceptación de la HU-07 (SC del spec)

| SC | Cómo se valida |
|---|---|
| SC-001 (100 % de cruces de umbral generan alerta alta en < 1 min) | T7h (worker cronometrado) + inline en T7f + comprobación manual 2 |
| SC-002 (panel ordenado por urgencia, 0 duplicados) | T7d + T7b + comprobación manual 1 |
| SC-003 (traslado → vista inmediata y estado global ≤ 30 s) | T7f + comprobaciones manuales 3 y 6 |
| SC-004 (100 % de alertas resueltas al normalizar, 0 obsoletas) | T7b + comprobación manual 3 |

## 5. Resultados de la validación

_Pendiente de ejecución tras la implementación (Phase 2 / `/speckit.tasks`): se registrarán aquí los resultados contra la rama `spec-007`, siguiendo el patrón de la sección 5 de [spec-004/quickstart.md](../spec-004/quickstart.md)._
