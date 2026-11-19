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
- **Mensajes (añadido 2026-11-19, hallazgo U2 del analyze)**: el 422 `VALIDACION` distingue dos mensajes. Código vacío o cuerpo inválido → «Introduce el código del anclaje»; código de más de 64 caracteres o con caracteres de control → «El código no es válido». Antes ambos casos decían «Introduce el código del anclaje», que es engañoso si el usuario sí ha escrito algo (P2.3 de las pruebas previas). Se mantiene como validación técnica en el plan y no en la spec (decisión del grupo, opción A): el límite de 64 caracteres protege la base de datos, no es una regla del negocio, y el usuario que teclea un código real nunca lo alcanza.

## D3. Datos del viaje activo

- **Decisión**: las respuestas añaden `bicicleta_codigo` y `estacion_origen_nombre`. Es un tipo propio de `services/alquiler` que embebe `models.Viaje`, sin tocar el struct compartido. Los campos se añaden como opcionales al esquema `Viaje` del OpenAPI.
- **Racional**: FR-008. Una sola respuesta evita la petición extra al detalle de estación. `estacion_origen_nombre` es el nombre que ya usa spec-004 y `bicicleta_codigo` el de `EstacionDetalle`.
- **Alternativas consideradas**: un endpoint adicional: rechazada, sobra. Campos nuevos en `models.Viaje`: rechazada porque el struct es de todas las specs y aumentaría los conflictos.

## D4. Orden de validación y causas

- **Decisión**: bloqueo del anclaje → abono → viaje activo → existencia del anclaje → estación → estado del anclaje → bicicleta. Mapeo: abono 403 `SIN_ABONO` o `ABONO_VENCIDO` (D15); viaje 409 `VIAJE_ACTIVO`; inexistente 404 `ANCLAJE_NO_EXISTE`; estación o anclaje no aptos 409 `ANCLAJE_NO_DISPONIBLE`; bicicleta 409 `BICICLETA_NO_DISPONIBLE`.
- **Racional**: las reglas del usuario van primero porque quien no puede alquilar debe saberlo aunque también se equivoque de código. *2026-11-19*: el clarify de refinamiento convierte este orden en requisito (FR-012); el código ya lo cumple y solo falta un test que lo fije.
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
- **Acceso directo por URL**: ~~no se añade guardia de ruta~~. *Sustituido el 2026-11-19 (FR-009 ampliado en el clarify de refinamiento)*: `Desbloqueo.tsx` redirige a `/` con `<Navigate>` si el rol de la sesión no es `usuario`, también cuando no hay usuario guardado (mismo criterio que el menú, T029). La guardia va en la propia página y no en `routes.tsx`, que es compartido (const. V). El 403 del backend se mantiene como garantía (const. I).

## D10. Redis no disponible

- **Decisión**: el desbloqueo no depende de Redis. El fallo posterior al commit se registra y lo repara el barrido (R16).
- **Racional**: PostgreSQL es la fuente de verdad del viaje; un rechazo con el viaje creado sería peor que unos segundos de contador antiguo.

## D15. Causa concreta del rechazo por abono (FR-002, SC-001) — *añadido 2026-11-19*

- **Decisión**: el servicio distingue dos causas. `ErrSinAbono` cuando `abono_activo = false`, cuando no hay `fecha_vencimiento_abono` o cuando el usuario ya no existe; `ErrAbonoCaducado` cuando el abono está activo pero su fecha es anterior a hoy (fecha de `Europe/Madrid`, D1). El handler responde 403 `SIN_ABONO` «No tienes abono anual» y 403 `ABONO_VENCIDO` «Tu abono anual caducó el DD/MM/AAAA». La fecha sale de la columna `DATE`, así que se formatea sin conversión de zona. El texto se construye en una única función de `services/alquiler` (`causaAbono`), que también usa D16, para que el aviso previo y el rechazo digan exactamente lo mismo. `models.AbonoVigente` no cambia.
- **Racional**: FR-002 pide «indicar la causa», y las pruebas previas (P3.3) mostraron el mismo mensaje para Carla (sin abono) que para Bruno (caducado). Se conserva `ABONO_VENCIDO` para el caso caducado, que es lo que ese código ya significaba, y se añade `SIN_ABONO`: el cambio de contrato es aditivo.
- **Alternativas consideradas**: mantener un único código y variar solo el mensaje (el frontend y los tests no podrían distinguir las causas sin analizar texto); devolver la fecha en un campo aparte del envelope de error (el envelope `{codigo, mensaje}` es común a todas las specs y no se toca).

## D16. Aviso del abono al abrir la pantalla (FR-010) — *añadido 2026-11-19*

- **Decisión**: `GET /alquileres/activo` añade a su respuesta un objeto `abono`: `{"vigente": bool, "codigo": null | "SIN_ABONO" | "ABONO_VENCIDO", "mensaje": null | string}`, calculado con `causaAbono` (D15) en el momento de la consulta. `stores/viaje.ts` lo guarda junto al viaje. Si `vigente` es `false`, `Desbloqueo.tsx` muestra `mensaje` con `role="alert"` encima del formulario y desactiva el campo y el botón. El rechazo del `POST` se mantiene: si el abono caduca con la pantalla abierta, el servidor lo detecta igualmente (FR-010).
- **Racional**: la página ya llama a `/alquileres/activo` al abrirse, así que no hace falta otra petición. El dato lo calcula el backend (const. I), con la misma función que decide el rechazo. Solo la HU-02 consume este endpoint (`AvisoViaje` y `Desbloqueo`), así que añadir un campo no afecta a otras historias.
- **Alternativas consideradas**: exponer `fecha_vencimiento_abono` en el `Usuario` del login (cambia un modelo compartido, const. III; además, el usuario del login se guarda para toda la sesión y no reflejaría una renovación); un endpoint nuevo `/alquileres/estado` (una petición más para el mismo momento); calcular la vigencia en el frontend (duplicaría la regla fuera del backend, const. I).

## D17. Resultado desconocido de un desbloqueo (FR-011) — *añadido 2026-11-19*

- **Decisión**: `desbloquear()` en `services/api/alquiler.ts` usa un `AbortController` con un plazo de 10 s. `Desbloqueo.tsx` trata como **resultado desconocido** tres casos: `ApiError` con `status = 0` (`SIN_CONEXION`), la cancelación por plazo y cualquier `status >= 500`. En esos casos pasa al estado «comprobando», muestra «No sabemos si se ha desbloqueado; comprobando…» y llama una sola vez a `cargarViajeEnCurso()`: si hay viaje, lo muestra (FR-008); si no, muestra «No se ha desbloqueado, inténtalo de nuevo» con el formulario activo; si la consulta también falla, muestra el mismo aviso de comprobación con un botón «Volver a comprobar», que repite solo la consulta. El `POST` nunca se reenvía de forma automática. Los 4xx no cambian: son rechazos con causa y no hay duda sobre el resultado.
- **Racional**: con la red cortada o un 5xx no se puede saber si el commit llegó a hacerse (D12 confirma aunque la petición se cancele). Consultar el viaje activo da el dato real del servidor. Reenviar el `POST` sería peor: si el primero entró, respondería `VIAJE_ACTIVO` y confundiría al usuario (clarify, pregunta 2).
- **Alternativas consideradas**: reintentos automáticos del `POST` (descartados en el clarify, opción C); una clave de idempotencia en el `POST` (resolvería el reintento, pero cambia el contrato y la tabla `viajes`, const. III, para un caso que la consulta del viaje ya cubre).

---

**Resumen de resolución (2026-11-19)**: D15–D17 añadidas y D9 corregida tras el clarify de refinamiento; D4 pasa a ser requisito (FR-012) sin cambios. 0 NEEDS CLARIFICATION. Cambios en piezas compartidas, todos aditivos y a comunicar al equipo (const. III): constante `SinAbono` en `backend/internal/httperr`, código `SIN_ABONO` y objeto `abono` en `specs/spec-001/contracts/openapi.yaml`, y el handler por defecto de `/alquileres/activo` en `frontend/tests/handlers.ts`. Sin migraciones ni claves Redis nuevas.
