# Quickstart: validación de la HU-03 (devolución y tarificación)

**Branch**: `spec-003` | **Date**: 2026-10-27 | **Plan**: [plan.md](./plan.md) | **Contrato**: [contracts/devoluciones.md](./contracts/devoluciones.md) | **Modelo**: [data-model.md](./data-model.md)

Guía de validación **end-to-end** de la devolución. Es una guía de ejecución; la implementación vive en `tasks.md` y el código. Hereda el setup del [quickstart global](../spec-001/quickstart.md) (infra, migrate, seed, server, frontend).

## 0. Prerrequisitos

- Docker Desktop, Go ≥ 1.22, Node ≥ 20; puertos 5432, 6379, 8080, 5173 libres.
- `.env` con `TARIFA_MINUTO` definido (ej. `0,20`).

## 1. Setup (igual que el global)

```powershell
copy .env.example .env        # revisar DATABASE_URL, REDIS_ADDR, JWT_SECRET, TARIFA_MINUTO
docker compose up -d
cd backend
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/server           # terminal 1
cd ../frontend && npm install && npm run dev   # terminal 2
```

## 2. Pruebas automatizadas (gate constitución I — prueba T3)

```powershell
cd backend
go test ./internal/services/devolucion/... -v    # unitario de tarifa + casos borde
go test ./...                                    # integración end-to-end contra la BD del compose
```

| Test | Escenario | Esperado |
|---|---|---|
| T3a | Viaje de 25 min → devolución | `200`, `coste_total = 0.00`, `etiqueta = "Tramo gratuito (0,00 €)"`, viaje `finalizado` |
| T3b | Viaje de 45 min, `TARIFA_MINUTO=0,20` → devolución | `200`, `coste_total = 3.00` (15 min excedentes × 0,20) |
| T3c | Viaje de exactamente 30:00 | `200`, `coste_total = 0.00` |
| T3d | Viaje de 30:01 | `200`, `coste_total = 0.20` (1 minuto cobrado) |
| T3e | Sin viaje activo → devolución | `409 SIN_VIAJE_ACTIVO`; el anclaje queda intacto |
| T3f | Doble devolución concurrente del mismo viaje | exactamente 1 × `200`; la otra → `409 VIAJE_YA_CERRADO` |
| T3g | Anclaje ocupado / inexistente / fuera de servicio | `409` con código específico; viaje sigue `en_curso` |
| T3h | Bici reportada averiada durante el viaje (HU-05) → devolución | `200`, viaje `finalizado` y tarificado; la bici queda **anclada y `averiada`** (no suma en `bicis_disponibles`) |
| T3i | Reloj inconsistente (`fecha_fin < fecha_inicio` forzada en BD) | coste ≥ 0; se usa hora validada |
| T3j | Latencia de una devolución válida (SC-003) | completa en < 10 s (asertado en el test de integración y cronometrado en el manual) |
| T3k | Permisos (FR-008, AC4): `POST /devoluciones` sin sesión y con token de rol `operario_taller` | sin sesión → `401 NO_AUTENTICADO`; operario → `403 ROL_NO_PERMITIDO`; en ambos el viaje sigue `en_curso` y el anclaje intacto |
| T3l | Abono caducado (FR-009, AC5): usuario con abono vencido y viaje activo devuelve la bici | `200`, viaje `finalizado` con `coste_total` registrado (el abono no condiciona el cierre) |

## 3. Escenario manual S3 (misma prueba del quickstart global)

```powershell
# Login del usuario seed con abono vigente (el seed imprime las credenciales)
$r = Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/auth/login `
     -Body '{"email":"ana@si16.test","password":"password"}' -ContentType 'application/json'
$h = @{ Authorization = "Bearer $($r.token)" }

# Con un viaje activo (previamente desbloqueado, HU-02), devolver en un anclaje libre:
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/devoluciones `
     -Headers $h -Body '{"codigo_anclaje":"AN-00456"}' -ContentType 'application/json' | ConvertTo-Json
```

**Esperado**: `200` con el viaje `finalizado`, hora de llegada, duración y coste; para viajes ≤ 30 min, `coste_total = 0.00` con la etiqueta "Tramo gratuito (0,00 €)". Comprobar además:

1. **Usuario libre**: un segundo desbloqueo (HU-02) funciona de inmediato (SC-004).
2. **Bici anclada**: `GET /estaciones` → la estación destino muestra la bici y un anclaje libre menos.
3. **Frescura ≤ 30 s** (const. IV): `actualizado_en` del HASH `estacion:{destino}:estado` es nuevo; el mapa refleja el cambio en ≤ 30 s.
4. **Viaje > 30 min** (forzar `fecha_inicio` hacia atrás en BD o esperar): `coste_total` = solo excedente × `TARIFA_MINUTO`.
5. **Rechazos**: repetir con un anclaje ocupado → `409 ANCLAJE_OCUPADO` y el viaje permanece `en_curso` (puede devolverse en otro anclaje).
6. **UI**: en `http://localhost:5173`, la página de devolución (`pages/devolucion/`) muestra el formulario de código de anclaje y, tras enviar, la confirmación con el importe.

## 4. Criterios de aceptación de la HU-03 (SC del spec)

| SC | Cómo se valida |
|---|---|
| SC-001 (100 % viajes ≤ 30 min con coste 0,00 €) | T3a/T3c + inspección SQL: `SELECT count(*) FROM viajes WHERE duracion_minutos <= 30 AND coste_total <> 0` = 0 |
| SC-002 (100 % viajes > 30 min cobrados solo excedente) | T3b/T3d + aserto 46:30 → 17 min (redondeo al minuto empezado) + inspección SQL: coste = minutos excedentes × tarifa unitaria |
| SC-003 (devolución < 10 s) | T3j (aserción en test de integración) + cronómetro manual en el paso 3 |
| SC-004 (usuario libre de inmediato, 0 dobles cierres) | pasos 1 y T3f |
