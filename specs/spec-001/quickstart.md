# Quickstart: levantar y validar el sistema SI-16

**Branch**: `spec-001` | **Date**: 2026-10-20 | **Plan**: [plan.md](./plan.md) | **Contratos**: [contracts/openapi.yaml](./contracts/openapi.yaml) | **Modelo**: [data-model.md](./data-model.md)

Guía de validación **end-to-end** del sistema completo (8 specs). Es una guía de ejecución: los detalles de implementación viven en `tasks.md` y el código.

## 0. Prerrequisitos

- Docker Desktop (con `docker compose` v2)
- Go ≥ 1.22
- Node.js ≥ 20 (npm ≥ 10)
- Puerto libres: 5432 (PostgreSQL), 6379 (Redis), 8080 (backend), 5173 (frontend)

## 1. Setup

```powershell
# 1) Infraestructura compartida (raíz del repo)
copy .env.example .env        # revisar DATABASE_URL, REDIS_ADDR, JWT_SECRET, TARIFA_MINUTO (ej. 0,20)
docker compose up -d          # postgres:16-alpine + redis:7-alpine (con healthchecks)

# 2) Esquema y datos de prueba
cd backend
go run ./cmd/migrate up       # aplica 0001_esquema_base.sql (ver data-model.md)
go run ./cmd/seed             # ~60 estaciones, flota, anclajes, usuarios de TODOS los roles

# 3) Backend (http://localhost:8080/api/v1) — en una terminal
go run ./cmd/server

# 4) Frontend (http://localhost:5173) — en otra terminal
cd ../frontend
npm install
npm run dev
```

**Resultado esperado**: `docker compose ps` muestra ambos servicios `healthy`; el seed imprime los usuarios de prueba (usuario con abono vigente, usuario con abono caducado, operario_taller, operario_logistica, gestor).

## 2. Pruebas automatizadas (gate de calidad de la constitución)

```powershell
cd backend
go test ./...                                  # unit + integración (usa la BD del compose)
go test ./internal/services/... -run TestReglasCriticas   # T1–T5 (ver abajo)
cd ../frontend
npm test                                       # Vitest + testing-library + MSW
```

**T1–T5 — reglas críticas (constitución I), deben existir y pasar antes de fusionar cualquier spec:**

| Test | Regla | Esperado |
|---|---|---|
| T1 | Abono vigente | Desbloqueo con abono caducado → 403 `ABONO_VENCIDO`; con abono vigente → 201 |
| T2 | Un viaje activo por usuario | Segundo desbloqueo → 409 `VIAJE_ACTIVO`; 2 desbloqueos concurrentes → exactamente 1 viaje creado |
| T3 | 30 min gratuitos | Viaje ≤ 1800 s → `coste_total = 0.00`; > 1800 s → solo excedente × TARIFA_MINUTO |
| T4 | Bloqueo por avería | Bicicleta averiada: desbloqueo → 409 y excluida de `bicis_disponibles` |
| T5 | Alertas de ocupación | Ocupación 9,99 % → alerta `vacia_menos_10`; 10,00 % → sin alerta; 90,00 % → sin alerta; 90,01 % → `llena_mas_90`; duplicados imposibles; vuelta a rango → resuelta |

## 3. Escenarios de validación manual (por spec)

Usar `POST /auth/login` con los usuarios seed (o la UI). `curl` de ejemplo para un endpoint crítico:

```powershell
$r = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/auth/login `
     -Body '{"email":"ana@si16.test","password":"password"}' -ContentType 'application/json'
$h = @{ Authorization = "Bearer $($r.token)" }
Invoke-RestMethod -Uri http://localhost:8080/api/v1/estaciones -Headers $h | ConvertTo-Json -Depth 5
```

| # | Spec | Escenario | Esperado |
|---|---|---|---|
| S1 | HU-01 (spec-001) | Abrir el mapa / `GET /estaciones` | ~60 estaciones geolocalizadas con contadores; seleccionar una → detalle < 2 s. Provocar un cambio (una devolución, S3) y refrescar: el contador cambia en **≤ 30 s** (medir con `actualizado_en`) |
| S1a | HU-01 (spec-001) | Cortar el backend (o la red en DevTools) con el mapa abierto; esperar > 30 s; volver a levantarlo | Los contadores siguen visibles con «Actualizado a las HH:MM:SS»; a los > 30 s aparece el aviso «Datos desactualizados»; desaparece con la primera respuesta correcta (FR-007, SC-005) |
| S1b | HU-01 (spec-001) | Con SQL, dejar una estación con 0 bicis disponibles, otra con 0 anclajes libres y otra con todos los anclajes ocupados por bicis averiadas | Cada marcador muestra su estado con icono y texto (no solo color); el detalle muestra «0 · Sin bicicletas» / «0 · Sin anclajes libres», o ambos; al devolver disponibilidad vuelve a normal en ≤ 30 s; la suma de contadores puede ser menor que la capacidad (FR-006, FR-008, SC-006) |
| S1c | HU-01 (spec-001) | `UPDATE estaciones SET estado='fuera_de_servicio'` en una estación con 0 bicis; después revertir | En ≤ 30 s el marcador pasa a «Fuera de servicio» (prevalece sobre «sin bicicletas»), sigue siendo seleccionable y su detalle muestra «—» sin contadores ni acciones; al revertir vuelve a operativa en ≤ 30 s (FR-009, SC-007) |
| S1d | HU-01 (spec-001) | Parar las escrituras síncronas (p. ej. un `UPDATE` directo en `anclajes`, sin pasar por la API) y medir | El contador cambia en pantalla en ≤ 30 s solo gracias al barrido de 10 s (camino de respaldo, R16) |
| S2 | HU-02 (spec-002) | Desbloquear con código de anclaje válido | 201 con viaje activo; usuario caducado → 403 con causa; usuario con viaje activo → 409; código inexistente/estación fuera de servicio/anclaje sin bici → 404/409 **sin** modificar nada |
| S3 | HU-03 (spec-003) | Devolver en anclaje libre; repetir con viaje forzado > 30 min (ajustar `fecha_inicio` en BD o esperar) | Viaje finalizado con hora de llegada, usuario liberado (puede desbloquear de nuevo), bici anclada; ≤ 30 min → `coste_total = 0.00` con etiqueta "Tramo gratuito (0,00 €)"; > 30 min → solo excedente |
| S4 | HU-04 (spec-004) | Sección "Mis Viajes" con filtro mes/año | Lista cronológica (origen, destino, duración, coste), sumatorio correcto, viajes de 0,00 € etiquetados, viaje activo ausente del historial |
| S5 | HU-05 (spec-005) | Reportar avería con descripción obligatoria | Bicicleta pasa a `averiada` al instante; desbloqueo posterior → 409; desaparece de `bicis_disponibles` (S1); incidencia en panel `GET /incidencias` (operario_taller) en < 30 s; reporte repetido → sin duplicar bloqueo |
| S6 | HU-06 (spec-006) | Operario abre ficha, registra parte y marca "Operativa"; confirma ubicación | Ficha muestra el reporte original; parte firmado (operario + fecha) persistido; incidencia cerrada; hasta ubicar, la bici no aparece en ningún mapa; tras `POST /taller/bicicletas/{id}/ubicacion` → visible en su estación en **≤ 30 s** (S1) |
| S7 | HU-07 (spec-007) | Llevar una estación seed al 95 % y otra al 5 % (traslados o SQL); abrir panel; normalizar; confirmar un traslado | Alertas de prioridad alta generadas en < 60 s, ordenadas por urgencia, **sin duplicados** por estación/umbral; al volver a 10–90 % → resueltas solas; traslado → contadores de ambas estaciones recalculados de inmediato |
| S8 | HU-08 (spec-008) | Gestor: rango de fechas con datos y vacío; descargar CSV; probar con usuario normal | Gráficos por hora (24) y día de la semana (7); CSV = mismo rango y **mismos totales** que los gráficos, abre con tildes/ñ correctos; rango vacío → gráficos vacíos y CSV solo cabecera; usuario sin rol gestor → 403 |

## 4. Criterios de aceptación de nivel sistema

- Frescura: ningún cambio de disponibilidad tarda > 30 s en el mapa (medido en S1 y S6) — constitución IV.
- Consistencia: 0 usuarios con > 1 viaje activo y 0 viajes sin duración/coste tras cierre (`SELECT` de control en BD).
- Exclusión de averiadas consistente: HU-01, HU-05 y HU-07 muestran el mismo número de disponibles para una misma estación.
- Autorización: cada endpoint responde 403 al rol equivocado (recorrido de los tags de `contracts/openapi.yaml`).

## 5. Referencias

- Endpoints, payloads y códigos de error: [contracts/openapi.yaml](./contracts/openapi.yaml)
- Reglas de cálculo (tarifa, disponibilidad, umbrales): [data-model.md](../data-model.md) y [research.md](../research.md)
- Estado en vivo y claves Redis: [contracts/redis-state.md](./contracts/redis-state.md)
