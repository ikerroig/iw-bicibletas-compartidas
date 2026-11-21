# Convergencia 1 — spec-008 / HU-08

**Fecha**: 2026-11-20 · **Rama**: `refinamiento/spec-008`

## Resultado

**Convergido respecto al alcance funcional**: la implementacion inspeccionada satisface FR-001…FR-007, SC-001…SC-004, la historia US1 y las decisiones vigentes del plan. No se identificaron tareas nuevas que anexar a `tasks.md`; sus tareas T001-T020 ya estan marcadas completas y se dejan intactas.

## Comprobaciones de implementacion

| Referencia | Evidencia inspeccionada | Resultado |
|---|---|---|
| FR-001 | `backend/internal/services/analitica/service.go`: formato estricto `YYYY-MM-DD`, rango inclusivo y rechazo de rango invertido; handlers devuelven 422 antes de generar la descarga. | Cumple |
| FR-002 | Servicio crea buckets de 24/7 y agrupa la fecha de inicio en Europe/Madrid; `frontend/src/pages/analitica/` muestra los dos graficos requeridos. | Cumple |
| FR-003/FR-004 | `backend/internal/services/analitica/csv.go`: mismo filtro, BOM UTF-8, columnas pactadas, CRLF, CSV RFC 4180 y emision fila a fila. | Cumple |
| FR-005 | `backend/internal/routes/spec_008_analitica.go`: ambos endpoints exigen autenticacion y rol `gestor`; pruebas T8g cubren 401/403/200. | Cumple |
| FR-006 | `construirFiltro` es compartido por agregados y CSV; T8a compara numero de filas y coste con los agregados. | Cumple |
| FR-007 | Las consultas filtran `estado = 'finalizado'`; T8d verifica exclusion de viajes en curso. | Cumple |
| SC-001 | T8i comprueba el tiempo de respuesta; el quickstart registra una ejecucion previa por debajo de 5 s. | Evidencia previa disponible |
| SC-002/SC-003 | T8a comprueba consistencia; T8g comprueba denegacion a roles no gestor. | Cumple |
| SC-004 | T8h verifica BOM, cabecera, CRLF y caracteres; S8 registra apertura del CSV con tildes y ñ. | Evidencia previa disponible |
| Decision de alcance 2026-11-20 | La UI conserva exactamente los graficos por hora y dia; no hay ranking ni campo `por_estacion` en el modelo compartido. | Cumple |
| Constitucion III | La HU-08 no incorpora migraciones ni claves Redis. | Cumple |

## Validacion y limites

- `go test ./internal/services/analitica`, `go vet ./...` y `npm run build` pasan en esta verificacion.
- La suite frontend completa pasa: **16 archivos, 97 pruebas**. En Node 26 hay que proporcionar `NODE_OPTIONS=--localstorage-file=<temporal>` para que Vitest tenga `localStorage`; con esa opcion las 6 pruebas de HU-08 tambien pasan.
- Docker y los servicios de `docker compose` estan activos. En la verificacion cruzada del 2026-11-21, la suite `go test -p 1 ./...`, `go vet ./...`, las 97 pruebas del frontend y `npm run build` pasaron usando PostgreSQL publicado temporalmente en `127.0.0.1:5433`; no se modifico `.env`.
- La revision manual de los endpoints de HU-08 comprobo `200` para gestor, `422` para rango invertido, `403` para usuario normal, `401` sin token y respuesta vacia correcta para un rango futuro. El intento de levantar Vite desde este entorno concreto fallo por permisos de `esbuild` al resolver la ruta del repositorio; no afecta al build ni a las pruebas del frontend.
- La revision cruzada queda registrada en `refinamiento/hallazgos.md`; no se encontraron incumplimientos ni trabajo funcional pendiente.

## Tareas nuevas

**Ninguna**. No se anadio una fase de convergencia a `tasks.md` porque no se encontro trabajo funcional restante.
