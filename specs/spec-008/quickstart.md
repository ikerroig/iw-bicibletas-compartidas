# Quickstart: validación de la HU-08 (panel de analítica de uso y exportación CSV)

**Branch**: `spec-008` | **Date**: 2026-11-01 | **Plan**: [plan.md](./plan.md) | **Contrato**: [contracts/analitica.md](./contracts/analitica.md) | **Modelo**: [data-model.md](./data-model.md)

Guía de validación **end-to-end** de la analítica y la exportación. Es una guía de ejecución; la implementación vive en `tasks.md` (Phase 2, pendiente de `/speckit.tasks`) y el código. Hereda el setup del [quickstart global](../spec-001/quickstart.md) y de [spec-004](../spec-004/quickstart.md) (infra, migrate, seed, server, frontend) sobre la base ya presente en esta rama.

## 0. Prerrequisitos

- Docker Desktop, Go ≥ 1.22, Node ≥ 20; puertos 5432, 6379, 8080, 5173 libres.
- `.env` presente en la raíz (copiado de `.env.example`).

## 1. Setup (igual que el global)

```powershell
copy .env.example .env        # revisar DATABASE_URL, REDIS_ADDR, JWT_SECRET
docker compose up -d
cd backend
go run ./cmd/migrate up
go run ./cmd/seed             # ahora incluye el lote de viajes de demo HU-08 (research D7)
go run ./cmd/server           # terminal 1
cd ../frontend && npm install && npm run dev   # terminal 2 (instala chart.js — research D6)
```

> El seed añade ~40 viajes finalizados de `ana@si16.test` repartidos en las últimas 4 semanas, con horas 7–22 y todos los días de la semana, para que gráficos y CSV tengan forma reconocible sin SQL manual.

## 2. Pruebas automatizadas

```powershell
cd backend
go test ./... -v              # integración de la analítica contra la BD del compose
cd ../frontend
npm test                      # Vitest: página de analítica
```

| Test | Escenario | Esperado |
|---|---|---|
| T8a | **Consistencia gráficos ↔ CSV (FR-006, SC-002)** | Con fixtures de viajes en varias semanas: `GET /analitica/viajes` y `GET /analitica/export` del mismo rango → `total_viajes` == nº de filas de datos del CSV; `total_coste` == suma de `coste_eur`; `por_hora[i]` == filas del CSV en la hora `i` |
| T8b | Filtro del rango (FR-001, clarificación 1) | Viaje del día 15 a las 23:30 hora española → cuenta con `hasta = 15` (inclusivo) y **no** cuenta con `hasta = 14`; viaje del 1 a las 00:30 → cuenta con `desde = 1` |
| T8c | Agrupación hora/día (FR-002, clarificación 2) | Fixtures con horas conocidas → `por_hora` de 24 posiciones con las horas locales correctas; domingo → índice **6**, lunes → **0**, sábado → **5** de `por_dia_semana` (7 posiciones) |
| T8d | Viajes activos excluidos (FR-007) | Viaje `en_curso` en el rango → no suma en barras ni en CSV; al finalizarlo (HU-03) sí aparece |
| T8e | Rangos sin datos / futuro (casos límite) | Rango sin viajes y rango completamente futuro → `200` con arrays a 0, `total_viajes: 0` y CSV de **solo cabecera**; sin error |
| T8f | Entrada inválida (clarificación 1) | `desde`/`hasta` ausentes, mal formateados o `desde > hasta` → `422 ENTRADA_INVALIDA` en **ambos** endpoints, sin efectos |
| T8g | Roles (FR-005, SC-003) | Sin token → `401`; rol `usuario`/`operario_taller`/`operario_logistica` → `403 ROL_NO_PERMITIDO` en ambos endpoints; rol `gestor` → `200` |
| T8h | Formato CSV (FR-004, R11, SC-004) | Descarga → BOM UTF-8 (`EF BB BF`), cabecera `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur`, CRLF, coma; nombres con tildes/ñ intactos; `coste_eur` con punto decimal (`3.00`); orden por `fecha_inicio` ASC; nombre `viajes_AAAAMMDD_AAAAMMDD.csv` |
| T8i | Latencia (SC-001) | `GET /analitica/viajes` con rango de hasta un año cronometrado **< 5 s** contra la BD del compose |

## 3. Escenario manual S8

```powershell
# Login del gestor (seed)
$r = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/auth/login `
     -Body '{"email":"gestor@si16.test","password":"password"}' -ContentType 'application/json'
$h = @{ Authorization = "Bearer $($r.token)" }

# Rango de los últimos 30 días (o el que cubra el lote de seed)
$desde = (Get-Date).AddDays(-30).ToString('yyyy-MM-dd')
$hasta = (Get-Date).ToString('yyyy-MM-dd')

# Agregados para los gráficos
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/analitica/viajes?desde=$desde&hasta=$hasta" `
    -Headers $h | ConvertTo-Json -Depth 3

# Descarga del CSV del MISMO rango
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/analitica/export?desde=$desde&hasta=$hasta" `
    -Headers $h -OutFile viajes.csv
```

**Esperado**: `200` con `por_hora` de 24 posiciones y `por_dia_semana` de 7 (lunes→domingo) con las horas/días del seed; el CSV descargado contiene exactamente los mismos viajes. Comprobar además en `http://localhost:5173/analitica`:

1. **Gráficos** (SC-001): las dos barras renderizan los datos; el tooltip muestra el valor al explorar; el rango de los 30 días se pinta en < 5 s.
2. **Consistencia** (SC-002): el `total_viajes` del JSON coincide con el número de filas de datos del CSV y con el total mostrado bajo los gráficos.
3. **CSV correcto** (SC-004): abrir `viajes.csv` en Excel/LibreOffice → sin caracteres corruptos (mojibake) en "Estació Universitat"/"Estació Mercat"; columnas en el orden del contrato; `coste_eur` con punto decimal.
4. **Sin datos** (caso límite): elegir un rango futuro (p. ej. el próximo año) → gráficos vacíos con el aviso **"sin datos en el periodo"** y CSV de solo cabecera.
5. **Rango invertido** (clarificación 1): `desde > hasta` (o manipular el input) → mensaje de rango inválido (422), sin descarga.
6. **Roles** (SC-003): login con `ana@si16.test` (rol usuario) → la API responde `403 ROL_NO_PERMITIDO` en ambos endpoints (probar con `Invoke-RestMethod` y ver el mensaje de permiso en la página).

## 4. Criterios de aceptación de la HU-08 (SC del spec)

| SC | Cómo se valida |
|---|---|
| SC-001 (gráficos < 5 s para rangos ≤ 1 año) | T8i + comprobación manual 1 |
| SC-002 (CSV = 100 % de los viajes del rango; totales idénticos a los gráficos) | T8a + comprobación manual 2 |
| SC-003 (100 % de accesos no gestor denegados) | T8g + comprobación manual 6 |
| SC-004 (100 % de CSV abren sin pérdida de caracteres) | T8h + comprobación manual 3 |

## 5. Resultados de la validación

**Fecha**: 2026-11-01 | **Rama**: `spec-008` | **Commit base**: rama con T001–T018 implementadas (T019/T020 bloqueadas, pendientes de acuerdo del equipo)

### 5.1 Pruebas automatizadas

| Suite | Resultado |
|---|---|
| `go test ./...` (backend) | ✅ en verde — `services/analitica` (unit servicio + CSV) y `tests` (T8a–T8i, 9/9) incluidos |
| `go vet ./...` | ✅ limpio |
| `gofmt -l` | ✅ sin avisos reales (el listado del working tree es CRLF preexistente de Windows; ficheros nuevos sin avisos) |
| `npm test` (frontend) | ✅ 19/19 — página de analítica (6 tests: datos, vacío, descarga, doble clic, 403, 422), cliente API, viajes y devolución |
| `npm run build` | ⚠️ falla **solo** con `TS2307 Cannot find module 'chart.js'` en `GraficoBarras.tsx` — comportamiento esperado hasta T020 (`"chart.js": "^4"` pendiente de acuerdo del equipo); sin ningún otro error de tipos |

### 5.2 Escenario S8 en vivo (API real, server :8080 + seed T003)

Ejecutado el 2026-11-01 con `gestor@si16.test` / `password`:

| Paso | Resultado |
|---|---|
| Agregados últimos 30 días | ✅ `200` — `total_viajes: 41`, `total_coste: 101.4`, `por_hora` 24 pos., `por_dia_semana` 7 pos., latencia 5 ms (< 5 s, SC-001) |
| Export del mismo rango | ✅ `200` — 41 filas de datos (consistencia SC-002: filas == `total_viajes`; suma `coste_eur` == 101.4 == `total_coste`), BOM UTF-8, CRLF, cabecera de 6 columnas, `Content-Disposition: attachment; filename=viajes_AAAAMMDD_AAAAMMDD.csv`, nombres con ñ/tildes intactos (SC-004) |
| Rango con bordes (`2026-10-01`…`2026-10-15`) | ✅ `200` — 12 viajes; export consistente |
| Rango futuro (2099) | ✅ `200` — arrays a 0, `total_viajes: 0`, CSV de solo cabecera |
| Rango anual (SC-001) | ✅ `200` — 42 viajes, respuesta instantánea |
| Rango invertido (`desde > hasta`) | ✅ `422 ENTRADA_INVALIDA` en ambos endpoints (agregados y export), sin cabeceras de descarga |
| Rol `usuario` (carla@si16.test) | ✅ `403 ROL_NO_PERMITIDO` en ambos endpoints (SC-003) |
| Sin token | ✅ `401 NO_AUTENTICADO` |

### 5.3 Página (`localhost:5173/analitica`)

Los comportamientos de UI (totales, aviso "sin datos en el periodo", descarga autenticada con Blob/`<a download>`, botón deshabilitado durante la descarga, mensajes `role="alert"` para 403/422) están verificados por los 6 tests de `frontend/src/pages/analitica/index.test.tsx` (wrapper del gráfico mockeado). El renderizado **visual de las barras con chart.js en navegador** queda pendiente de la aprobación de T020 (`npm install` de `chart.js` + `npm run build`); la ruta `/analitica` está dada de alta en `routes.tsx` y el nav.
