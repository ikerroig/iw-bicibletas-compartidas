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

> El worker `alertas` evalúa la red al arrancar y después cada 55 s: con el seed de extremos, el panel tendrá alertas en **< 60 s** tras el arranque del servidor.

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
| T7h | Latencia (SC-001) | (i) Barrido completo sobre el volumen de estaciones definido para el proyecto, cronometrado **< 5 s** (presupuesto verificable de SC-001); (ii) con el periodo de test inyectado, mutación de una estación fuera de rango (SQL) → alerta en `alertas_balanceo` con latencia **< periodo + 5 s**. La garantía de SC-001 (< 60 s = 55 s + barrido) depende de cumplir ese límite medido |

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

**Fecha**: 2026-10-31 · **Rama**: `spec-007` · Backend `:8080`, frontend dev `:5173`, Postgres 16 + Redis 7 vía compose.

### 5.1 Pruebas automatizadas (gate constitución I)

| Comando | Resultado |
|---|---|
| `go test ./...` (backend) | ✅ en verde (~8,4 s): `services/balanceo`, `services/devolucion`, `services/viajes` y el paquete de integración `tests` (T7a–T7h + HU-03/HU-04) |
| `go vet ./...` | ✅ limpio |
| `gofmt -l` | ✅ sin avisos reales (solo fin de línea CRLF del working tree Windows; los ficheros nuevos de spec-007 no aparecen) |
| `npm test` (frontend) | ✅ 13/13 en 3 ficheros (incl. `pages/balanceo/index.test.tsx`) |
| `npm run build` | ✅ `tsc -b && vite build` sin errores (typecheck TS) |

> Corrección aplicada durante el gate: el fixture `diaMes` de `backend/tests/historial_test.go` anclaba al día de hoy antes de restar meses; con hoy = día 31, `AddDate(0, -1, 0)` normalizaba (31-sept → 1-oct) y colapsaba un viaje del mes anterior dentro del actual (T4a/T4b). Anclado al día 1 antes del offset — fix de robustez temporal, sin cambio de comportamiento del producto.

### 5.2 Escenario manual S7 (API en vivo, sin reiniciar el server)

1. **Login** `POST /auth/login` `logistica@si16.test` → `200` con token ✅
2. **Alertas en < 60 s (SC-001)**: tras `go run ./cmd/seed` (idempotente), el panel pasó de `items: []` a **2 alertas en ≤ 12 s** (barrido del worker): Estació Nord id3 `vacia_menos_10` 5 % y Estació Port id4 `llena_mas_90` 95 %, ambas severidad `alta` ✅
3. **Panel ordenado (SC-002)**: `GET /api/v1/balanceo/alertas` con token → `200`, solo abiertas, orden por urgencia (empate de score → `fecha_generacion` ASC) ✅
4. **Traslado válido (SC-003)**: `POST /api/v1/balanceo/traslados` `{"origen":4,"destino":3,"unidades":10}` → `200` con `origen` `{bicis 9, libres 11, 45 %}` y `destino` `{bicis 11, libres 9, 55 %}` recalculados en la respuesta ✅
5. **Redis post-commit**: `HGETALL estacion:3:estado` → `bicis_disponibles 11 / anclajes_libres 9 / ocupacion_pct 55.00`; `HGETALL estacion:4:estado` → `9 / 11 / 45.00`, con `actualizado_en` del momento del traslado ✅
6. **Resolución automática (SC-004)**: tras volver ambas a `[10, 90]`, el panel devolvió `{"items":[]}` sin intervención manual (alertas inline resueltas) ✅
7. **Traslado rechazado (caso límite)**: `unidades: 999` → `409 ESTACION_SIN_CAPACIDAD` con mensaje que distingue la causa (`disponibles 9, unidades pedidas 999`); sin cambios en ambas estaciones ✅
8. **Roles (const. I)**: sin token → `401 NO_AUTENTICADO`; login `ana@si16.test` (rol `usuario`) → `403 ROL_NO_PERMITIDO` ✅

La validación visual del panel en `http://localhost:5173/balanceo` (render, formulario de traslado, tarjetas `EstacionEstado`) queda cubierta por `frontend/src/pages/balanceo/index.test.tsx` (Vitest, 13/13) y el dev server está operativo en `:5173`.
