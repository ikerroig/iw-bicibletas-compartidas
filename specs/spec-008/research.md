# Research: Panel de analítica de uso y exportación CSV HU-08 (Phase 0)

**Branch**: `spec-008` | **Date**: 2026-11-01 | **Plan**: [plan.md](./plan.md)

Investigación propia de la HU-08. **No repite** las decisiones globales ya resueltas en [specs/spec-001/research.md](../spec-001/research.md) (R1 pila, R2 esquema, R5 auth/roles, R6 migraciones, R11 exportación CSV, R14 analítica y gráficos): esta spec las **adopta tal cual** y solo detalla los puntos de diseño de su `spec.md` (clarificaciones del 2026-11-01 incluidas). La rama `spec-008` contiene la base de spec-003/spec-004/spec-007 (auth JWT, `apierr`, historial con filtro Europe/Madrid, patrón de servicios/handlers/rutas por spec); se verifica qué existe antes de diseñar. Resultado: **0 NEEDS CLARIFICATION**.

---

## D1. Filtro del rango: validación de entrada y semántica inclusiva en Europe/Madrid

- **Decisión**: ambos endpoints (`GET /analitica/viajes` y `GET /analitica/export`) exigen `desde` y `hasta` con formato estricto `YYYY-MM-DD` y aplican el **mismo** predicado (FR-001, clarificación 1):
  1. **Validación previa** (→ `422 ENTRADA_INVALIDA`, patrón ya consolidado por spec-003/spec-004/spec-007): cualquiera de los dos parámetros ausente, con formato distinto de `YYYY-MM-DD` estricto (p. ej. día inexistente o dígito suelto) o `desde > hasta`. La petición no llega a tocar la BD.
  2. **Filtro temporal** (idéntico al del historial de HU-04, `services/viajes`): `(fecha_inicio AT TIME ZONE 'Europe/Madrid')::date BETWEEN $desde AND $hasta` sobre viajes con `estado = 'finalizado'`. **Ambos extremos inclusivos**: un viaje iniciado el día `hasta` a las 23:59 hora española entra en el rango.
  3. Rangos futuros o parcialmente futuros **no son error**: el predicado sencillamente no encuentra viajes posteriores a "ahora" y devuelve contadores a cero / CSV de solo cabecera (caso límite del spec — sin recorte especial: no puede existir un viaje finalizado con `fecha_inicio` en el futuro, salvo corrupción de datos o reloj).
- **Racional**: reutiliza el filtro ya probado y timezone-correcto de la rama (test T4g: viaje del día 31 a las 23:30 hora española cae en su mes real); la validación previa evita sorpresas SQL y respeta la separación 422/200 ya usada en el proyecto. La clarificación del spec fija Europe/Madrid y la inclusividad, así que no hay nada abierto.
- **Alternativas consideradas**: rangos `datetime` con hora (rechazado: el OpenAPI global define `format: date` y el spec habla de "rango de fechas"); filtrar en UTC (rechazado: contradice la clarificación y rompería la consistencia con HU-04); recortar explícitamente `hasta` a hoy (rechazado: innecesario — el predicado ya no encuentra viajes futuros; añadiría lógica sin efecto observable).

## D2. Agregación por hora (00–23) y día de la semana (lunes–domingo)

- **Decisión**: la agrupación usa **siempre la fecha de inicio** del viaje reinterpretada en Europe/Madrid (FR-002, clarificación 2), con dos consultas indexadas sobre el mismo filtro de D1 (`ix_viajes_fecha`):
  - **`por_hora[24]`**: `EXTRACT(hour FROM (fecha_inicio AT TIME ZONE 'Europe/Madrid'))` → bucket `0…23` (hora local española del desbloqueo).
  - **`por_dia_semana[7]`**: `((EXTRACT(dow FROM (fecha_inicio AT TIME ZONE 'Europe/Madrid'))::int + 6) % 7)` → bucket `0…6` en orden **lunes → domingo** (PostgreSQL devuelve `dow` 0=domingo; la fórmula desplaza a lunes=0, domingo=6 — mapeo verificado en los tests).
  - Los buckets sin viajes se rellenan con `0` **en el servicio Go** (arrays fijos de 24/7 inicializados a cero y incrementados con los filas agrupadas): sin `generate_series` ni `LEFT JOIN`, 24+7 posiciones siempre presentes en la respuesta (schema global: `minItems/maxItems` 24 y 7).
- **Racional**: el spec y la clarificación fijan huso, origen (inicio) y orden de días; `date_part`/`extract` sobre la columna `TIMESTAMPTZ` con `AT TIME ZONE` es exacto sea cual sea el huso del servidor (const. I: la regla vive en backend/SQL). El relleno en Go es trivial, testeable unitariamente y evita un tercero `generate_series` cuyo plan de consulta no aporta nada a esta escala (~10²–10⁴ viajes).
- **Nota de coherencia**: el historial HU-04 filtra con la misma conversión, así que un viaje del día 31 a las 23:30 hora española cuenta en la hora 23 y el día de semana locales, no en los del huso del servidor.
- **Alternativas consideradas**: `generate_series(0,23) LEFT JOIN` en SQL (rechazada: mismos resultados, más ruido en la consulta y en los tests); agrupar por `fecha_fin` (rechazado: contradice la clarificación — el volumen se analiza por momento de origen); domingo=0 (rechazado: la clarificación fija lunes→domingo).

## D3. Un único filtro compartido: consistencia gráficos ↔ CSV (FR-006) y totales

- **Decisión**: `services/analitica` construye el predicado y los args del rango **una sola vez por petición** en una función interna (`construirFiltro`) y lo reutiliza en todas las consultas:
  - `GET /analitica/viajes`: Q1 `por_hora`, Q2 `por_dia_semana` (D2) y Q3 `SELECT count(*), COALESCE(sum(coste_total), 0)` — los tres con el mismo WHERE/args. `total_viajes` = suma de `por_hora` (se calcula al rellenar los buckets, imposible que difiera de las barras) y `total_coste` = Q3.
  - `GET /analitica/export`: única consulta en streaming con el **mismo** WHERE/args, `ORDER BY fecha_inicio ASC`, JOIN a `estaciones` para los nombres.
  - Ambos endpoints comparten la misma función de filtro → el CSV de un rango contiene exactamente los viajes que suman los gráficos del mismo rango (SC-002, "por construcción" — patrón R11).
- **Racional**: el spec declara defecto cualquier discrepancia entre gráficos y CSV; una única fuente de predicado elimina la clase entera de bugs de doble implementación. `total_viajes` derivado de las barras hace imposible un total desalineado de sus propias barras.
- **Alternativas consideradas**: vista materializada o pre-agregación nocturna (rechazada: R14 la descarta a esta escala); recalcular totales en el frontend (rechazado: const. I — el backend es la única autoridad).

## D4. Exportación CSV: streaming, BOM, RFC 4180 y mapeo de columnas (R11 + FR-004)

- **Decisión**: `GET /analitica/export` genera el fichero en **streaming** (filas de la BD → `encoding/csv.Writer` sobre el `ResponseWriter`), según R11 y la clarificación 3:
  - **Cabecera HTTP**: `Content-Type: text/csv; charset=utf-8`, `Content-Disposition: attachment; filename=viajes_AAAAMMDD_AAAAMMDD.csv` (fechas del rango sin guiones).
  - **BOM UTF-8** (`EF BB BF`) escrito **a mano** antes del writer (Go no lo emite solo; el BOM es lo que hace que Excel/LibreOffice reconozcan UTF-8 → SC-004, tildes/ñ).
  - **Separador coma**, comillas RFC 4180 (comportamiento por defecto de `encoding/csv`) y `UseCRLF = true` (finales de línea CRLF).
  - **Columnas** (en este orden): `id,fecha_hora,estacion_origen,estacion_destino,duracion_minutos,coste_eur`.
  - **`fecha_hora`**: `fecha_inicio` formateada en Europe/Madrid como `AAAA-MM-DD HH:MM` (`2006-01-02 15:04` con `time.LoadLocation("Europe/Madrid")`).
  - **`coste_eur`**: decimal con **punto** y 2 decimales (`fmt.Sprintf("%.2f")`, p. ej. `2.00`) — con separador de columna coma, el separador decimal español (coma) rompería el parseo; el punto es el único compatible con RFC 4180 + coma. **`duracion_minutos`**: entero (ya `ceil` registrado por HU-03).
  - **Nombres de estación**: literales de la BD (con tildes/ñ); el writer los comilla si contienen coma o comillas.
  - Rango vacío → **solo la fila de cabecera** (caso límite del spec).
- **Racional**: R11 del plan global ya eligió BOM+coma+RFC 4180+CRLF y el nombre del fichero; la decisión sobre el decimal es la única que el YAML global no fija y se resuelve a favor de la interoperabilidad (mismo criterio que R11 rechazó `;`). El streaming evita cargar en memoria rangos amplios (caso límite "varios años") y emite el primer byte en cuanto la BD devuelve la primera fila.
- **Alternativas consideradas**: separador `;` + decimal coma (rechazado explícitamente en R11: no es RFC 4180); XLSX (rechazado en R11: HU-08 pide CSV); escribir todo en memoria y responder de golpe (rechazado: innecesario, rompe el caso de rangos muy amplios); columna `fecha` y `hora` separadas (rechazado: contradice el contrato global y la clarificación, que unifican en `fecha_hora`).

## D5. Autorización: rol `gestor` exclusivo en ambos endpoints

- **Decisión**: `routes/spec_008_analitica.go` encadena `auth.RequireAuth(d.JWTSecret)` + `auth.RequireRole("gestor")` en `GET /analitica/viajes` y `GET /analitica/export` (middlewares ya existentes en la rama, patrón de spec-007). Sin token → `401 NO_AUTENTICADO`; token de cualquier otro rol (`usuario`, `operario_taller`, `operario_logistica`) → `403 ROL_NO_PERMITIDO`. El frontend puede ocultar el enlace del panel, pero la API es la autoridad (SC-003, FR-005, const. I).
- **Racional**: el modelo global asigna el rol `gestor` a "analítica y exportación CSV (HU-08)" y el OpenAPI marca ambos endpoints como "rol gestor"; el spec es tajante: denegación del 100 % de los intentos no gestor.
- **Alternativas consideradas**: incluir también a algún operario (rechazado: el spec es exclusivo del gestor municipal); comprobar el rol solo en el frontend (rechazado: const. I — el cliente no es autoridad); ocultar 404 en lugar de 403 (rechazado: el OpenAPI define 403 y el quickstart global espera 403).

## D6. Frontend: panel con Chart.js (R14), descarga autenticada y estados

- **Decisión**: página `frontend/src/pages/analitica/index.tsx` + cliente `frontend/src/services/api/analitica.ts` (patrón de `api/viajes.ts` sobre `api.request`):
  - **Selector de rango**: dos inputs `date` (`desde`/`hasta`, valor por defecto = **últimos 30 días** incluyendo hoy — decisión UX, ajustable sin tocar contrato) con validación cliente `desde ≤ hasta` (el backend manda: su 422 se muestra igualmente).
  - **Gráficos**: dos barras con **Chart.js** (dependencia nombrada en el plan global R14; alta en `frontend/package.json` y lockfile autorizada y aplicada en T020): "Viajes por hora del día" (eje X 00–23) y "Viajes por día de la semana" (eje X lunes…domingo), pintadas **tal cual** los arrays del backend (const. I). Interactividad = tooltip Chart.js al explorar (Assumption del spec).
  - **Estado "sin datos"**: cuando `total_viajes == 0` se muestra el aviso **"sin datos en el periodo"** sobre los gráficos (el backend no necesita flag: el total de la respuesta lo determina).
  - **Descarga**: el JWT viaja en cabecera `Authorization`, así que **no** se usa `<a href>` crudo; se hace `fetch` autenticado de `/analitica/export`, se crea un `Blob` y se dispara la descarga con un `<a download>` temporal (mismo patrón que cualquier descarga protegida por Bearer).
  - **Errores**: 403 → mensaje de permiso; 422 → mensaje de rango inválido; ambos con `role="alert"`.
  - **Ruta/nav**: una entrada en `frontend/src/routes.tsx` y un enlace "Analítica" en el nav de `frontend/src/index.tsx` (ficheros compartidos, un elemento cada uno — conflictos de merge triviales).
- **Racional**: sigue el patrón de páginas ya consolidado en la rama (señales + Tailwind + `data-testid` + Vitest con stub de fetch); Chart.js es la decisión global R14 (maduro, sin ataduras de framework, wrapper aislado en la página).
- **Alternativas consideradas**: SVG/canvas propio (rechazado: reinventar tooltips y ejes sin necesidad); librerías Solid específicas (rechazadas en R14: menos maduras); descarga vía query token en URL (rechazada: el JWT no debe viajar en query strings ni en historial del navegador).

## D7. Seed: viajes finalizados repartidos en semanas, horas y días (demo S8)

- **Decisión**: ampliar `backend/cmd/seed/main.go` **añadiendo** (sin tocar los bloques existentes HU-03/HU-04/HU-07) un lote de **viajes finalizados de demostración** para la analítica:
  - ~40 viajes de la usuaria seed `ana@si16.test` con la bicicleta nueva `B-200` (`disponible`, sin anclaje), repartidos deterministamente en las **últimas 4 semanas** (28 días): patrón de horas locales `{7, 8, 9, 12, 17, 18, 20, 22}` y distintos días de la semana, para que las 24 barras/7 días tengan forma reconocible; duraciones y costes coherentes con la tarifa HU-03 (≤ 30 min → `0.00`; si no, excedente × 0,20).
  - Cumplen el CHECK de `finalizado` (destino, fin y duración NOT NULL); los viajes `finalizado` no chocan con `ux_viajes_bicicleta_activa` (índice parcial sobre `en_curso`), así que una sola bici sirve para todo el lote. Insertados **después** de los viajes HU-04 existentes: no alteran sus ids ni sus fixtures.
- **Racional**: sin este lote, el panel y S8 solo se podrían demostrar con SQL manual; con él, el escenario del quickstart es reproducible en dos comandos (mismo criterio que la ampliación de seed de spec-007). El patrón determinista permite asertar valores esperados en la validación manual.
- **Alternativas consideradas**: fixtures solo en los tests (insuficiente para la demo manual S8); script de demo separado (rechazado: el seed es el mecanismo establecido); repartir viajes entre varios usuarios (rechazado: la analítica es de red, un único usuario de demo simplifica y no cambia nada observable).

## D8. Rendimiento y observabilidad (SC-001)

- **Decisión**: 3 consultas pequeñas para el panel (Q1/Q2/Q3 de D3) + 1 streaming para el CSV, todas apoyadas en `ix_viajes_fecha` (ya en `0001_esquema_base.sql`) y con `WHERE estado = 'finalizado'`. Presupuesto SC-001: **< 5 s** para rangos de hasta un año, verificado con un test de integración cronometrado sobre el volumen del seed (rango de un año con ~40–100 viajes → milisegundos reales; el test asienta el presupuesto, no el valor observado). El backend registra la duración de cada agregado con `log` (patrón del worker de spec-007) para que la regresión sea visible en producción.
- **Racional**: a la escala definida en el plan global (~10²–10⁴ viajes) ninguna agregación necesita caché ni materialización (R14); el índice ya existe — la HU-08 **no requiere migración alguna**. El streaming del CSV cubre el caso "varios años" sin memoria proporcional al resultado.
- **Alternativas consideradas**: caché Redis de agregados (rechazada: frescura/invalidación sin beneficio medible a esta escala, y añadiría claves nuevas que exigen acuerdo — const. III); materialized view (rechazada en R14).

## D9. Pruebas: patrón de la rama (integración directa, sin kin-openapi)

- **Decisión**: tests de integración en `backend/tests/analitica_test.go` con **fixtures propias** (TRUNCATE+RESTART IDENTITY CASCADE de `viajes` + inserts de viajes finalizados con horas/días conocidos, patrón de `historial_test.go`/`balanceo_test.go`) montando `routes.Register` con `httptest`; aserciones directas del JSON y del CSV contra el contrato documentado en [contracts/analitica.md](./contracts/analitica.md). Vitest + stub de fetch para la página. Escenarios T8a–T8i en [quickstart.md](./quickstart.md).
- **Racional**: el plan global (T081) menciona `kin-openapi`, pero la rama ya consolidada (spec-003/spec-004/spec-007) valida contratos con aserciones directas y **sin** esa dependencia; añadirla solo para esta spec rompería la homogeneidad de `go.mod` sin cubrir ningún caso nuevo (los endpoints son dos y sus payloads están fijados por el schema global). Se mantiene el patrón de la rama.
- **Alternativas consideradas**: incorporar `kin-openapi` (rechazada por ahora: dependencia nueva en un `go.mod` compartido, sin beneficio incremental — se deja anotado como mejora global del equipo si se desea validar todo el OpenAPI de una vez).

---

**Resumen**: 9 decisiones (D1–D9), 0 NEEDS CLARIFICATION. D1/D2 materializan las clarificaciones del spec (rango inclusivo Europe/Madrid; agrupación por inicio, 00–23 y lunes–domingo); D3/D4 detallan la consistencia FR-006 y el CSV R11 ya fijado; D5/D6 completan autorización y frontend; D7/D8/D9 cierran seed, rendimiento y pruebas. **La HU-08 no requiere ningún cambio de esquema PostgreSQL ni claves Redis nuevas**: usa `viajes`, `estaciones` y `ix_viajes_fecha` tal y como están en `0001_esquema_base.sql`. Los cambios al OpenAPI global (`422`, T019) y a `frontend/package.json` (`chart.js`, T020) fueron autorizados por el equipo y aplicados.
