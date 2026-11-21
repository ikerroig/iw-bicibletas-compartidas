# Trazabilidad — spec-008 / HU-08

**Fecha**: 2026-11-20 · **Rama**: `refinamiento/spec-008`

| Criterio | Decision del plan / alcance | Tareas | Prueba o evidencia |
|---|---|---|---|
| FR-001 — rango inclusivo Europe/Madrid y rechazo de rango invertido | Validar `desde`/`hasta` en backend; filtrar por fecha local de inicio. | T005, T007, T009, T014 | `service_test.go`; T8b y T8f; validacion del selector |
| FR-002 — 24 horas y 7 dias (lunes-domingo), sobre inicio del viaje | El servidor calcula los buckets y el frontend pinta los arrays. Sin ranking por estacion por decision del equipo del 2026-11-20. | T004, T007-T009, T014, T018, T020 | T8c; tests de pagina para 24/7; `GraficoBarras.tsx`; modelo sin `por_estacion` |
| FR-003 — exportar viajes del mismo rango | CSV generado en backend y descargado con autenticacion desde el cliente. | T006-T007, T010-T013 | T8a, T8e y T8h; tests de descarga del frontend |
| FR-004 — formato del CSV | BOM UTF-8, coma, RFC 4180, CRLF, columnas y nombre de fichero fijados por FR-004/R11. | T006-T007, T010-T011, T017 | Tests unitarios CSV; T8h; escenario S8 del quickstart |
| FR-005 — acceso exclusivo de gestor | `RequireAuth` y `RequireRole("gestor")` en ambas rutas. | T007, T011-T012, T014 | T8g y prueba de mensajes 403 en la pagina |
| FR-006 — graficos y CSV contienen los mismos viajes y totales | Un unico `construirFiltro` alimenta agregados y exportacion. | T007, T009-T011, T016 | T8a compara filas, coste y buckets |
| FR-007 — solo viajes finalizados de toda la red | Consultas sin filtro por usuario y con `estado = 'finalizado'`. | T007, T009-T010 | T8d excluye viajes en curso |
| SC-001 — graficos en menos de 5 s hasta un año | Consultas indexadas y prueba cronometrada. | T007, T016-T017 | T8i; quickstart registra medicion previa inferior a 5 s |
| SC-002 — CSV completo y consistente | Predicado compartido y comparacion automatizada entre agregados y filas. | T007, T010-T011, T016-T017 | T8a; S8 registra igualdad de totales |
| SC-003 — denegar a quien no es gestor | Autorizacion server-side en ambos endpoints. | T007, T012, T016-T017 | T8g; S8 registra 401/403 |
| SC-004 — CSV se abre sin perdida de caracteres | BOM UTF-8, serializacion RFC 4180 y nombres Unicode preservados. | T006-T007, T017 | T8h y comprobacion manual del CSV en S8 |
| Plan/Constitucion III — no ampliar esquema compartido | HU-08 es de solo lectura; no crear migraciones ni claves Redis. | T001-T004, T018-T020 | Modelo de datos y migraciones existentes; inspeccion de servicios |

## Estado de validacion

Actualizacion de la verificacion cruzada (2026-11-21): la migracion y el seed se ejecutaron correctamente contra PostgreSQL publicado temporalmente en `127.0.0.1:5433` (el puerto `55432` definido en el entorno local no estaba disponible). La suite backend completa (`go test -p 1 ./...`) paso, junto con `go vet ./...`; la suite frontend paso con 16 archivos y 97 pruebas, y `npm run build` tambien paso. No se modifico el fichero `.env`.

La API se comprobo manualmente con el seed: el gestor obtuvo `200` en agregados y export CSV; un rango futuro devolvio arrays y totales a cero; un rango invertido devolvio `422 ENTRADA_INVALIDA`; un usuario normal devolvio `403 ROL_NO_PERMITIDO`; y una peticion sin token devolvio `401`.

La revision cruzada queda registrada en `refinamiento/hallazgos.md`. No se detectaron incumplimientos funcionales ni tareas nuevas.
