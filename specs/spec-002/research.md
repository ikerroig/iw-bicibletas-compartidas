# Research: Desbloqueo por código de anclaje (HU-02)

**Branch**: `spec-002` | **Date**: 2026-11-04 | **Plan**: [plan.md](./plan.md)

Hereda del plan global R5 (autenticación y roles), R7 (integridad del alquiler) y R16 (frescura) en [specs/spec-001/research.md](../spec-001/research.md). Aquí solo figuran las decisiones propias de la HU-02. La numeración coincide con los Deltas de [plan.md](./plan.md); las decisiones sin Delta (D4, D10) ya están en `e090f41`.

## D1. Fecha de vigencia del abono

- **Decisión**: la fecha actual se calcula en `Europe/Madrid` y se pasa a `models.AbonoVigente`, que no cambia. La zona se embebe con `import _ "time/tzdata"`, para no depender de la base de zonas del sistema (Windows o imágenes mínimas).
- **Racional**: FR-002 fija el fin de vigencia en las 23:59:59 hora de España. El cambio de hora queda cubierto por la propia zona.
- **Alternativas consideradas**: `CURRENT_DATE` de PostgreSQL: rechazada porque depende de la zona de la BD (UTC por defecto en la imagen). `SET TIME ZONE` en la sesión: rechazada porque afectaría a otras consultas.
- **Alcance**: solo el abono. La zona de negocio de historial y analítica (spec-004, spec-008) no se decide aquí y se comunica al equipo. `specs/spec-001/data-model.md` dice `CURRENT_DATE`; FR-002 lo precisa y hay que actualizarlo.

## D2. Normalización y validación del código

- **Decisión**: el backend aplica `TrimSpace` y `ToUpper` y consulta por igualdad exacta. Devuelve 422 `VALIDACION` si queda vacío, supera 64 caracteres (igual que `POST /devoluciones`) o contiene caracteres de control, incluido NUL.
- **Racional**: FR-001. La igualdad usa el índice único de `codigo_anclaje`. Los códigos se guardan en mayúsculas (`E01-A01`), invariante que se documenta en [data-model.md](./data-model.md). PostgreSQL rechaza NUL en texto, y sin validar acabaría en 500.
- **Alternativas consideradas**: `upper(codigo_anclaje) = $1`: rechazada porque no usa el índice y exigiría uno nuevo, es decir, cambio de esquema. Normalizar en el frontend: rechazada por la constitución I.

## D3. Datos del viaje activo

- **Decisión**: las respuestas añaden `bicicleta_codigo` y `estacion_origen_nombre`. Es un tipo propio de `services/alquiler` que embebe `models.Viaje`, sin tocar el struct compartido. Los campos se añaden como opcionales al esquema `Viaje` del OpenAPI.
- **Racional**: FR-008. Una sola respuesta evita la petición extra al detalle de estación. `estacion_origen_nombre` es el nombre que ya usa spec-004 y `bicicleta_codigo` el de `EstacionDetalle`.
- **Alternativas consideradas**: un endpoint adicional: rechazada, sobra. Campos nuevos en `models.Viaje`: rechazada porque el struct es de todas las specs y aumentaría los conflictos.

## D4. Orden de validación y causas

- **Decisión**: bloqueo del anclaje → abono → viaje activo → existencia del anclaje → estación → estado del anclaje → bicicleta. Mapeo: abono 403 `ABONO_VENCIDO`; viaje 409 `VIAJE_ACTIVO`; inexistente 404 `ANCLAJE_NO_EXISTE`; estación o anclaje no aptos 409 `ANCLAJE_NO_DISPONIBLE`; bicicleta 409 `BICICLETA_NO_DISPONIBLE`.
- **Racional**: las reglas del usuario van primero porque quien no puede alquilar debe saberlo aunque también se equivoque de código.
- **Bloqueos**: el `FOR UPDATE` del anclaje va en una sentencia propia para que las lecturas siguientes vean lo confirmado durante la espera. Con un JOIN en la misma consulta, PostgreSQL revaluaría solo parte de las filas y podría leer datos antiguos.

## D5. Lectura de la estación

- **Decisión**: `SELECT estado FROM estaciones … FOR SHARE`.
- **Racional**: FR-004. Un cambio concurrente a `fuera_de_servicio` espera al commit o se ve antes de crear el viaje. `FOR SHARE` permite desbloqueos simultáneos en la misma estación.
- **Alternativas consideradas**: `FOR UPDATE`: rechazada porque serializaría todos los desbloqueos de la estación.

## D6. Tiempo tras el commit

- **Decisión**: el recálculo tiene un presupuesto de 2 s, sin depender de la cancelación de la petición. `RecalcularEstacion` notifica el hook con un contexto propio de 3 s, independiente del anterior, sin cambiar su firma.
- **Racional**: si el cliente se desconecta, el viaje existe y el mapa debe reflejarlo. `cache.Nuevo` activa `ContextTimeoutEnabled`: sin él, go-redis ignora el plazo del contexto (5,6 s medidos frente a 2,3 s con Redis mudo). Con Redis lento, los suscriptores no deben recibir un contexto vencido. El peor caso añade unos 5 s, dentro de los 10 s de SC-003.
- **Alternativas consideradas**: recálculo en segundo plano: rechazada porque el comportamiento síncrono es el contrato de R16 y los tests lo comprueban.

## D11. Plazo de la transacción

- **Decisión**: la transacción de desbloqueo tiene un plazo de 4 s, derivado del contexto de la petición. Si se agota, responde error interno sin efectos.
- **Racional**: SC-003. Sin plazo, un bloqueo retenido dejaría la petición colgada y ocuparía conexiones del pool. Con D6, el peor caso es 4 + 2 + 3 = 9 s.
- **Alternativas consideradas**: `SET LOCAL lock_timeout`: rechazada porque no acota el resto de la transacción.

## D12. Commit tras una desconexión

- **Decisión**: `db.EnTransaccion` confirma con `context.WithoutCancel` y un plazo propio de 2 s.
- **Racional**: si el contexto se cancela durante el COMMIT, PostgreSQL puede confirmar mientras el cliente recibe error: el viaje existiría sin recálculo ni hook. Cuando `fn` termina bien, la decisión de confirmar ya está tomada.

## D13. Lectura del abono

- **Decisión**: `SELECT … FROM usuarios … FOR SHARE`.
- **Racional**: FR-007, igual que D5. Una revocación simultánea espera o se ve antes de crear el viaje. `FOR SHARE` no bloquea desbloqueos paralelos.

## D14. Aviso de viaje en curso

- **Decisión**: `Layout` muestra a rol `usuario` un aviso con hora de inicio, código de bicicleta y enlace a `/alquiler`, en todas las páginas menos esa, que ya muestra el viaje. El estado vive en `stores/viaje.ts`, compartido con la página.
- **Racional**: FR-008 pide mostrar el viaje al volver a abrir la aplicación, y tras el login se llega al mapa. Se consulta al entrar, sin polling: el viaje solo cambia por acciones del propio usuario.
- **Dependencia con spec-003**: la devolución debe llamar a `fijarViajeEnCurso(null)` para que desaparezca el aviso.

## D7. Contrato

- **Decisión**: declarar `401` en los dos endpoints y usar `E01-A01` como ejemplo.
- **Racional**: alinea el contrato con `RequireAuth` y permite probar el 401 con el validador.

## D8. Limpieza

- **Decisión**: quitar las ramas `!ok` tras `SesionDe` (inalcanzables tras `RequireAuth`) y el parámetro `signal` sin uso de `viajeActivo`. `Entorno.Post` y `Entorno.Get` siguen separados porque `testutil` es compartido.

## D9. Permisos

- **Decisión**: `RequireRole(usuario)` en el desbloqueo. `GET /alquileres/activo` admite cualquier sesión porque solo devuelve el viaje propio. La entrada del menú solo aparece al rol `usuario`.
- **Acceso directo por URL**: no se añade guardia de ruta. FR-009 pide ocultar la opción en el menú, y el backend ya rechaza con 403 y la página muestra la causa.

## D10. Redis no disponible

- **Decisión**: el desbloqueo no depende de Redis. El fallo posterior al commit se registra y lo repara el barrido (R16).
- **Racional**: PostgreSQL es la fuente de verdad del viaje; un rechazo con el viaje creado sería peor que unos segundos de contador antiguo.
