# Quickstart: validación de la HU-04 (historial de viajes y desglose mensual)

**Branch**: `spec-004` | **Date**: 2026-10-28 | **Plan**: [plan.md](./plan.md) | **Contrato**: [contracts/viajes.md](./contracts/viajes.md) | **Modelo**: [data-model.md](./data-model.md)

Guía de validación **end-to-end** del historial. Es una guía de ejecución; la implementación vive en `tasks.md` y el código. Hereda el setup del [quickstart global](../spec-001/quickstart.md) y de [spec-003](../spec-003/quickstart.md) (infra, migrate, seed, server, frontend).

## 0. Prerrequisitos

- Docker Desktop, Go ≥ 1.22, Node ≥ 20; puertos 5432, 6379, 8080, 5173 libres.
- `.env` presente en la raíz (copiado de `.env.example`).

## 1. Setup (igual que el global)

```powershell
copy .env.example .env        # revisar DATABASE_URL, REDIS_ADDR, JWT_SECRET
docker compose up -d
cd backend
go run ./cmd/migrate up
go run ./cmd/seed             # ahora incluye viajes finalizados de prueba (research D8)
go run ./cmd/server           # terminal 1
cd ../frontend && npm install && npm run dev   # terminal 2
```

## 2. Pruebas automatizadas (gate constitución I — prueba T4)

```powershell
cd backend
go test ./... -v              # integración del historial contra la BD del compose
cd ../frontend
npm test                      # Vitest: página "Mis Viajes"
```

| Test | Escenario | Esperado |
|---|---|---|
| T4a | Historial sin filtro | `200`, solo viajes `finalizado` del usuario, orden `fecha_inicio DESC`, filas con fecha/hora, origen, destino (nombres), duración y coste |
| T4b | Filtro `desde`/`hasta` de un mes | `200`, solo viajes de ese mes (Europe/Madrid), `suma_coste` = suma exacta de ese mes |
| T4c | Viaje con `coste_total = 0.00` | `etiqueta = "Tramo gratuito (0,00 €)"`; computa `0` en `suma_coste` |
| T4d | Aislamiento (FR-005) | Usuario B no ve viajes de A; sin token → `401 NO_AUTENTICADO` |
| T4e | Paginación con `page_size=1` | página 1 y 2 distintas; `total` y `suma_coste` **idénticos** en ambas (sumatorio del periodo) |
| T4f | Página fuera de rango (`page=999`) | `200`, `items: []`, `total` y `suma_coste` del periodo completo |
| T4g | Zona horaria (FR-002) | Viaje del día 31 a las 23:30 hora española cae en el mes esperado al filtrar |
| T4h | Periodo sin viajes | `total: 0`, `suma_coste: 0`, `items: []` sin error |
| T4i | Entrada inválida | `page=0`, `page_size=500`, `desde>hasta` o fecha malformada → `422 ENTRADA_INVALIDA` |
| T4j | Latencia (SC-004) | Respuesta del historial en < 3 s con un año completo de viajes (aserción cronometrada en el test de integración) |
| T4k | AC4 (clarify 2026-11-19): cualquier rol autenticado | token `operario_taller` o `gestor` → `200` con su **propio** historial; el propietario se resuelve desde el JWT (un parámetro de usuario ajeno se ignora) |

## 3. Escenario manual S4

```powershell
# Login del usuario seed con viajes finalizados
$r = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/auth/login `
     -Body '{"email":"ana@si16.test","password":"password"}' -ContentType 'application/json'
$h = @{ Authorization = "Bearer $($r.token)" }

# Historial completo (primera página)
Invoke-RestMethod -Uri http://localhost:8080/api/v1/viajes -Headers $h | ConvertTo-Json -Depth 5

# Filtro de mes (mes/año de la UI = desde/hasta inclusivos, Europe/Madrid)
# Sustituye 2026-10 por el mes actual si estás validando en otra fecha:
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/viajes?desde=2026-10-01&hasta=2026-10-31" -Headers $h | ConvertTo-Json -Depth 5
```

**Esperado**: `200` con `total`, `suma_coste` y `items` (fecha/hora, estaciones por **nombre**, duración, coste; etiqueta en los gratuitos). Comprobar además:

1. **UI**: en `http://localhost:5173`, abrir "Mis viajes" → lista cronológica (más recientes primero), filtros de mes y año, sumatorio visible y etiqueta "Tramo gratuito (0,00 €)" en los viajes de coste 0 (SC-003).
2. **Filtro mes/año**: elegir un mes con viajes → solo esos viajes y sumatorio de ese mes (SC-002); elegir un mes futuro → "sin viajes en el periodo" con suma `0,00 €`, sin error.
3. **Paginación**: con `page_size=1` (o muchos viajes en el seed), cambiar de página → la lista cambia pero el sumatorio **no** varía (FR-008).
4. **Viaje activo**: el viaje `en_curso` de ana (seed) **no** aparece en el historial.
5. **Aislamiento**: login con `carla@si16.test` → no ve ningún viaje de ana; estado vacío inicial ("aún no tienes viajes").
6. **Sin recálculo**: el coste y la duración de cada fila coinciden con los de `SELECT * FROM viajes WHERE id = …` (FR-007).

## 4. Criterios de aceptación de la HU-04 (SC del spec)

| SC | Cómo se valida |
|---|---|
| SC-001 (100 % de viajes finalizados con campos correctos) | T4a + inspección SQL cruzando `viajes` con el response |
| SC-002 (filtros exactos y sumatorio = suma del periodo) | T4b/T4e + comprobación manual 2 y 3 |
| SC-003 (viajes sin coste con la etiqueta) | T4c + comprobación manual 1 |
| SC-004 (historial y filtros < 3 s con un año de uso) | T4j (aserción cronometrada) |
| AC4 (cualquier rol autenticado ve su propio historial; 401 sin sesión) | T4d (aislamiento + 401) + T4k (roles `operario_taller`/`gestor`) |
| AC5 (viaje activo no aparece ni computa) | T4a (total = solo finalizados; `estado = finalizado` por fila) |
| AC6 (primera carga sin filtro: todo, orden DESC, sumatorio total) | T4a (orden `fecha_inicio DESC`) + T4e/T4f (`suma_coste` del historial completo sin filtro) |

## 5. Resultados de la validación (2026-10-28, rama spec-004)

Ejecutado con el seed de T003 y el backend en `:8080`:

| Comprobación | Resultado |
|---|---|
| `go test ./...` (T4a–T4j) | ✅ verde (unit + integración) |
| `npm test` (6 tests de la página + 3 de HU-03) | ✅ 9/9 |
| `npm run build` (typecheck TS) | ✅ |
| `go vet ./...` / `gofmt` | ✅ (avisos solo CRLF del working tree, no reales) |
| S4 API historial ana | ✅ `total=5`, `suma=14.00`, orden DESC, nombres por JOIN, etiqueta en el viaje de 0,00 € |
| Viaje `en_curso` de ana | ✅ **no** aparece en el historial |
| Filtro `2026-10-01..2026-10-31` | ✅ `total=3`, `suma=5.00` (0,00 + 2,00 + 3,00) |
| Paginación `page_size=1` pág. 1 vs 2 | ✅ items distintos (id 4 / id 3); `total=5` y `suma=14.00` idénticos |
| Página fuera de rango (`page=999`) | ✅ `items=[]`, conserva `total=5` y `suma=14.00` |
| 401 sin token / 422 `page=0` | ✅ `NO_AUTENTICADO` / `ENTRADA_INVALIDA` |
| Aislamiento (`carla@si16.test`) | ✅ `total=0`, `suma=0` |
| UI `http://localhost:5173/viajes` | ✅ SPA responde 200; interacciones (filtros, paginación, vacíos, error) cubiertas por los 6 tests de componente |

> Nota: las fechas de los tests de zona horaria usan horas UTC que caen en
> el día local esperado **tanto con CEST como con CET** (el 31-oct-2026
> Madrid ya está en CET, DST terminó el 25-oct).

> Nota (2026-10-29): reproceso del seed con costes coherentes con la tarifa
> HU-03 (0,20 €/min tras 30 min gratis): 12 min → 0,00 · 40 min → 2,00 ·
> 45 min → 3,00 · 50 min → 4,00 · 55 min → 5,00. Antes había viajes de
> 25 min por 2,50 € y de 40 min por 1,25 € (costes fijados a mano), que
> contradecían la regla tarifaria al validar a ojo en la UI.
