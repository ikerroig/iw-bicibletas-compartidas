# Research: Plan técnico global SI-16 (Phase 0)

**Branch**: `spec-001` | **Date**: 2026-10-20 | **Plan**: [plan.md](./plan.md)

Investigación asociada al Technical Context de `plan.md`. Cada entrada resuelve un punto abierto con el formato **Decisión / Racional / Alternativas consideradas**. Resultado final: **0 NEEDS CLARIFICATION pendientes**.

---

## R1. Pila tecnológica

- **Decisión**: Frontend SolidJS 1.9+ + Tailwind CSS (Vite 6, TypeScript 5); Backend Go 1.22+ con Gin; PostgreSQL 16; Redis 7. Migraciones con `goose`, acceso con `pgx/v5`, JWT con `golang-jwt/v5`, contraseñas con `bcrypt`.
- **Racional**: Fijado por la constitución SI-16 (principio II); stack uniforme permite revisión cruzada entre las 8 specs y elimina fricciones de integración entre ramas. Versiones son las actuales LTS/estables y compatibles entre sí.
- **Alternativas consideradas**: Node/Express o Spring (prohibidas por constitución); ORM pesados (GORM, Prisma): se rechazan porque ocultan SQL y dificultan los índices únicos parciales y transacciones `FOR UPDATE` que exigen las reglas críticas — se usa SQL explícito con pgx.

## R2. Estructura de carpetas y prevención de colisiones entre los 4 integrantes

- **Decisión**: Dos raíces de código (`/backend`, `/frontend`). Dentro de cada raíz, **una partición por spec**: `backend/internal/{services,handlers}/<dominio-spec>` + `backend/internal/routes/spec_00N_*.go` + `frontend/src/pages/<area-spec>` + `frontend/src/services/api/<spec>.ts`. Solo dos ficheros compartidos por rama: `routes/register.go` y `frontend/src/routes.tsx` (cada spec añade una línea/registro trivial).
- **Racional**: Principio V (independencia por spec). Dos specs de integrantes distintos no tocan nunca el mismo fichero, salvo el registro compartido cuyo conflicto de fusión se resuelve aceptando ambos bloques. `docker-compose.yml`, `.env.example` y migraciones son los únicos puntos de coordinación, ya regulados por el principio III (acuerdo de esquema).
- **Alternativas consideradas**: módulos por capa (`handlers/todos`, `pages/todos`): rechazado porque concentra conflictos en ficheros monolíticos; un repo por spec: rechazado porque rompe el modelo compartido y la revisión cruzada.

## R3. Estrategia de frescura ≤ 30 s en tiempo real

- **Decisión**: Redis es la fuente de lectura del estado en vivo. Doble mecanismo: (a) **actualización síncrona post-commit**: cada operación que muta estaciones/bicicletas/anclajes (desbloqueo, devolución, avería, traslado, ubicación de taller) reescribe las claves `estacion:{id}:estado` afectadas justo después de confirmar la transacción SQL; (b) **barrido de seguridad** cada 15 s en un worker Go que recalcula las ~60 estaciones con una única consulta agregada, garantía de cota ≤ 30 s aunque un camino de escritura se olvide de invalidar. El frontend hace *polling* de `GET /estaciones` cada 15 s y muestra `actualizado_en` (edge case de desconexión de HU-01).
- **Racional**: Satisface constitución IV y SC-002/SC-003 de HU-01 con margen (15 s < 30 s). El barrido de 60 claves cada 15 s es despreciable de coste y convierte el TTL/olvido en fallo degradado, no en datos estáticos (FR-005: prohibido mostrar datos sin expiración → cada clave lleva `actualizado_en` y TTL de seguridad de 120 s; si expira, el endpoint recalcula desde PostgreSQL).
- **Alternativas consideradas**: SSE o WebSockets: rechazados para el MVP (más complejidad de conexión/Reparto; el presupuesto de 30 s no lo exige) — el contrato REST deja la puerta abierta a añadirlos sin romper nada; solo polling sin actualización síncrona: rechazado porque el desbloqueo/devolución tardaría hasta 30 s en reflejarse en paneles de operarios (HU-07 exige recálculo inmediato).

## R4. Modelo de claves Redis

- **Decisión**: `estacion:{id}:estado` → **HASH** con campos `bicis_disponibles`, `anclajes_libres`, `ocupacion_pct`, `estado` (`operativa|fuera_de_servicio`), `actualizado_en` (RFC3339 UTC). TTL 120 s (refrescado en cada escritura). Lectura masiva del mapa con `SCAN MATCH estacion:*:estado` + `HGETALL` en pipeline; canal pub/sub opcional `estacion:estado:cambios` (no vinculante, para futuras mejoras push). Formato completo en [contracts/redis-state.md](./contracts/redis-state.md).
- **Racional**: El usuario/constitución fija el patrón `estacion:{id}:estado`; el HASH permite leer solo los campos necesarios y actualizarlos de forma atómica; `ocupacion_pct` precalculado evita recálculos en cada lectura del panel de HU-07.
- **Alternativas consideradas**: claves string JSON: rechazadas (todo el objeto se reescribe por un solo cambio, peor atomicidad); un único hash global `red:estado`: rechazado (contenido grande, escrituras contendedas entre specs).

## R5. Autenticación, sesiones y roles

- **Decisión**: Login `POST /auth/login` (email + contraseña → JWT HS256, exp 24 h, claims `sub`, `rol`, `nombre`). Contraseña guardada como `bcrypt` en `usuarios.password_hash` (**adición al modelo base, requiere acuerdo de equipo** — ver data-model). Middleware Gin `RequireAuth` + `RequireRole(...)`. Roles: `usuario`, `operario_taller`, `operario_logistica`, `gestor`. El frontend guarda el token en `localStorage` y lo envía en `Authorization: Bearer`.
- **Racional**: Las 8 specs asumen "sesión iniciada" y diferencian operario de taller (HU-06), operario de calle/transporte (HU-07) y gestor municipal (HU-08); dos roles de operario separados porque HU-06 y HU-07 exigen privilegios excluyentes sobre fichas vs. panel de logística. JWT stateless sin tabla de sesiones: menos superficie de coordinación entre specs.
- **Alternativas consideradas**: sesiones opacas en Redis: viable pero añade estado que ninguna HU pide; un único rol `operario`: rechazado porque HU-06 ("solo operario de taller") y HU-07 ("panel exclusivo del operario de transporte") exigen separación de privilegios; OAuth externo: fuera de alcance (sistema universitario, usuarios preexistentes).

## R6. Acceso a datos, migraciones y coordinación de esquema

- **Decisión**: `pgx/v5` con SQL explícito (sin ORM). Migraciones `goose` en `backend/migrations/` con numerales secuenciales globales: `0001_esquema_base.sql` (esquema compartido, aportado por spec-001) y, para cada spec posterior, `000N_spec-00X_tema.sql` **tras acordar el número con el equipo** (principio III). Transacciones siempre que una operación toque más de una tabla.
- **Racional**: SQL explícito controla los `SELECT … FOR UPDATE` y los índices únicos parciales que garantizan "un viaje activo por usuario/bicicleta". La numeración acordada evita que dos ramas reclamen el mismo ordinal al fusionarse.
- **Alternativas consideradas**: ORM (ver R1); migraciones por espec (embed en cada paquete): rechazado (orden de aplicación no determinista entre ramas); migraciones con timestamp: viable, pero añade poca claridad sobre el orden ya que el equipo coordina de todos modos.

## R7. Integridad del ciclo de alquiler (concurrencia)

- **Decisión**: Desbloqueo y devolución ejecutan **una única transacción** con: `SELECT … FOR UPDATE` sobre el anclaje (y la fila de viaje en devolución), validación de las reglas críticas dentro de la transacción, escrituras, y **índices únicos parciales como red de seguridad**:
  - `ux_viajes_usuario_activo`: `UNIQUE (usuario_id) WHERE estado='en_curso'` → imposible tener 2 viajes activos (HU-02 FR-007/FR-003);
  - `ux_viajes_bicicleta_activa`: `UNIQUE (bicicleta_id) WHERE estado='en_curso'` → una bicicleta solo en un viaje;
  - `ux_anclaje_bicicleta`: `UNIQUE (bicicleta_id)` en anclajes → una bicicleta en un solo anclaje (HU-03 FR-006).
  Un duplicado simultáneo revienta por índice → la API devuelve 409 sin efectos parciales (todo-o-nada de HU-02/HU-05).
- **Racional**: Constitución I (backend autoridad) + SC-002 de HU-01 y edge cases "dos desbloqueos casi simultáneos". El índice es garantía a nivel de base de datos, no solo de código.
- **Alternativas consideradas**: bloqueo optimista con columna `version`: rechazado (los dos viajes simultáneos se detectan tarde y el usuario ve errores de retry); advisory locks de PostgreSQL: innecesarios con `FOR UPDATE` + índices.

## R8. Tarificación (30 min gratuitos)

- **Decisión**: `tarifa_eur_minuto` es **configuración operativa** (variable de entorno `TARIFA_MINUTO`, valor de ejemplo 0,20) — el valor numérico está fuera de alcance por asunción de HU-03. Cálculo en el cierre del viaje, dentro de su transacción:
  - `segundos = fecha_fin − fecha_inicio` (fecha_fin = hora de fijado en el anclaje; reloj validado: si `segundos < 0` se usa `now()` y se registra incidencia de reloj);
  - `duracion_minutos = ceil(segundos / 60)` (columna del modelo);
  - `coste_total = 0,00 €` si `segundos ≤ 1800`; si no, `ceil((segundos − 1800)/60) × TARIFA_MINUTO`, con piso en 0,00.
  Con `TARIFA_MINUTO` a 0 el coste del excedente es 0 (acepta configuración "todo gratuito").
- **Racional**: Regla crítica de constitución I; con enteros coherentes: 30:00 → 0 €; 30:01 → 1 minuto cobrado. La duración mostrada (ceil) nunca contradice el cobro.
- **Alternativas consideradas**: cobro por fracción exacta de segundo: rechazado (facturación en céntimos y HU-04 habla de minutos); tabla `tarifas` persistida: rechazado (ninguna HU gestiona tarifas; sería alcance nuevo que requeriría enmienda).

## R9. Cálculo de disponibilidad y ocupación (fuente única)

- **Decisión**: todo cómputo deriva de `anclajes` + `bicicletas` (nunca contadores almacenados en SQL):
  - `bicis_disponibles(estación)` = anclajes de la estación cuya bicicleta está en estado `disponible`;
  - `anclajes_libres(estación)` = anclajes en estado `libre`;
  - `ocupación (%) = 100 × bicis_disponibles / capacidad_total` (solo si la estación está `operativa` y `capacidad_total > 0`);
  - averiadas y en_taller **no cuentan** como disponibles (HU-01 FR-006, HU-05 FR-004, HU-07 FR-008) pero, si ocupan un anclaje, **reducen `anclajes_libres`** (físicamente ocupado).
  Bicicleta `disponible` **sin anclaje** (recién reparada, HU-06) no pertenece a ninguna estación → no aparece en ningún mapa/contador hasta confirmar ubicación (FR-007 de HU-06) sin necesidad de un estado extra.
- **Racional**: Una sola fórmula satisface HU-01, HU-05, HU-06 y HU-07 de forma consistente (SC-003 de HU-01, SC-003 de HU-06); los valores en Redis son solo su caché con marca de tiempo.
- **Alternativas consideradas**: contadores incrementales `bicis_disponibles` en la tabla estaciones: rechazado (deriva entre specs que los mutan — riesgo de conflicto de integración justo en el modelo compartido); triggers de PostgreSQL: viable pero oculta la lógica del equipo que debe revisarla en código Go testeable.

## R10. Generación de alertas de balanceo (HU-07)

- **Decisión**: doble mecanismo espejo de R3: (a) tras cada mutación de estación, el servicio correspondiente evalúa los umbrales de las estaciones afectadas y crea/resuelve alertas; (b) worker `alertas` cada 60 s recalcula toda la red (garantía SC-001 < 1 min). Umbrales **estrictos**: alerta si `ocupación < 10` o `> 90` exactamente (10,00 y 90,00 no alertan). Índice `UNIQUE (estacion_id, tipo_alerta) WHERE NOT resuelta` impide duplicados por umbral (FR-003); al volver al rango normal se pone `resuelta = true` (FR-004) y puede reabrirse una nueva.
- **Racional**: constitución I (alertas = regla crítica server-side); cobertura inmediata + red de seguridad temporal cubre ambos SC sin dependencia de que todas las specs llamen al hook.
- **Alternativas consideradas**: solo triggers post-commit (sin barrido): un write-path olvidado deja alertas obsoletas; solo cron cada minuto: HU-07 exige recálculo inmediato tras confirmar traslado (SC-003).

## R11. Exportación CSV (HU-08)

- **Decisión**: `GET /analitica/export` genera CSV en streaming con BOM UTF-8 (`EF BB BF`), separador coma, comillas RFC 4180 (cabeceras en español incluidas: `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur`), finales de línea CRLF y `Content-Disposition: attachment; filename=viajes_YYYYMMDD_YYYYMMDD.csv`. Mismo filtro de fechas (misma consulta/servicio) que `GET /analitica/viajes` → totales idénticos por construcción (FR-006).
- **Racional**: BOM + UTF-8 elimina el edge case de tildes/ñ en Excel/LibreOffice (SC-004); un único servicio de consulta elimina la discrepancia gráficos↔CSV.
- **Alternativas consideradas**: CSV con `;` (locale español): rechazado (no es RFC 4180; BOM+coma es interoperable); XLSX: fuera de alcance (HU-08 pide CSV).

## R12. docker-compose y entorno

- **Decisión**: `docker-compose.yml` en la raíz con `postgres:16-alpine` (puerto 5432, volumen `pgdata`, healthcheck `pg_isready`, usuario/db/contraseña `si16` vía `.env`) y `redis:7-alpine` (puerto 6379, `--appendonly yes`, healthcheck `redis-cli ping`). El backend se conecta con `DATABASE_URL` y `REDIS_ADDR` del `.env` (ver `/.env.example`). Levantamiento: `docker compose up -d`.
- **Racional**: requisito explícito del usuario; volumen persistente conserva datos entre reinicios; healthchecks permiten `depends_on: condition: service_healthy` cuando se añada el contenedor de la app. Raíz (no dentro de `/backend`) para que las 8 specs y el frontend apunten al mismo entorno — es infraestructura, no funcionalidad, compatible con principio III.
- **Alternativas consideradas**: Dockerfiles de app incluidos desde el inicio: se aplazan (no son necesarios para desarrollar y probar localmente); DB embebida (SQLite): prohibido por constitución (PostgreSQL).

## R13. Estrategia de pruebas

- **Decisión**: Backend: `go test ./…` — tests unitarios de servicios con repositorios mock (cálculo de tarifa, umbrales, validaciones) + tests de integración contra una BD real del compose (esquema y transacciones). **Obligatorias antes de fusionar** (gate constitución I): (1) abono vencido → 403 con causa; (2) segundo viaje activo → 409; (3) viaje ≤ 30 min → 0,00 € y > 30 min → solo excedente; (4) bicicleta averiada → desbloqueo rechazado y excluida de contadores; (5) alertas <10 % y >90 % generadas, sin duplicados, resueltas al normalizar. Frontend: Vitest + @solidjs/testing-library + MSW (contratos de `openapi.yaml`). Contratos: validación de respuestas contra `contracts/openapi.yaml` en los tests de integración (biblioteca `kin-openapi`).
- **Racional**: "Gate de calidad" de la constitución; cada HU tiene además su prueba independiente descrita en su spec (se enlazan en [quickstart.md](./quickstart.md)).
- **Alternativas consideradas**: tests solo de frontend (insuficientes: las reglas son del backend); E2E con Playwright desde el día 1: se deja como ampliación posterior, el quickstart cubre el flujo manual.

## R14. Analítica y gráficos (HU-08)

- **Decisión**: agregación SQL en backend (`date_part('hour', fecha_inicio)` y `extract(dow)`) sobre `viajes` finalizados con índice `ix_viajes_fecha`; respuesta `GET /analitica/viajes` = `por_hora[24]`, `por_dia_semana[7]`, totales. Frontend: **Chart.js** (barras) con renderizado bajo demanda en SolidJS.
- **Racional**: agregación en SQL es exacta y testeable (FR-006); Chart.js es maduro, ligero y sin ataduras de framework (el wrapper se aísla en `pages/analitica/`).
- **Alternativas consideradas**: librerías Solid específicas (`solid-charts`): menos maduras; pre-agregación nocturna/materialized views: innecesario a escala de ~60 estaciones.

## R15. Adiciones al modelo base acordado por el usuario

- **Decisión**: el esquema del usuario se adopta **literalmente** como base; se proponen al equipo 3 adiciones mínimas imprescindibles (detalladas y marcadas en [data-model.md](./data-model.md)):
  1. `usuarios.password_hash` (sin ella no existe "sesión iniciada", asumida por las 8 specs);
  2. `estaciones.estado` (`operativa|fuera_de_servicio`): HU-01 exige mostrar estaciones fuera de servicio y HU-07 excluirlas del cálculo de ocupación;
  3. índices/constraints derivados (únicos, CHECKs, FK) — son consecuencia directa de los FR, no cambian ningún campo pedido.
- **Racional**: cumplir el requisito sin romper el modelo compartido: las adiciones se someten a acuerdo de equipo antes de migrar (principio III).
- **Alternativas consideradas**: inventar estados o tablas extra (rechazado: fuera de lo pedido); no añadir nada (imposible: HU-01/HU-07 y la autenticación no serían implementables).

---

**Resumen de resolución**: 15 decisiones, 0 NEEDS CLARIFICATION. Pendiente de acuerdo de equipo (principio III): adiciones R15 y numeración de migraciones R6.
