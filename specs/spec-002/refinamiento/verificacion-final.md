# Verificación final del refinamiento — HU-02 (spec-002)

**Responsable**: Iker Roig · **Revisor**: Alex Morell · **Fecha**: 2026-11-19
**Tareas**: T056 (validación manual antes/después, enunciado §5.3 paso 4) y
T057 (regresión, §5.4).
**Versión probada**: commit `058b3bd` (implementación de T043–T055, T059 y
T060), con el backend reiniciado sobre ese código y el seed recién cargado.
Usuarios `eva` y `fran` creados como indica el
[quickstart](../quickstart.md) (hallazgo I2).

## Tests automáticos

| | Antes (`0d6f478` + docs) | Fase roja (tests nuevos sin código) | Después (`058b3bd`) |
|---|---|---|---|
| Backend `go test -count=1 ./...` | todo `ok` ([tests-antes-backend.txt](./evidencias/tests-antes-backend.txt)) | no compila: faltan `ErrSinAbono` y `httperr.SinAbono` ([tests-rojo-backend.txt](./evidencias/tests-rojo-backend.txt)) | todo `ok` ([tests-despues-backend.txt](./evidencias/tests-despues-backend.txt)) |
| Frontend `npx vitest run` | 81/81 ([tests-antes-frontend.txt](./evidencias/tests-antes-frontend.txt)) | HU-02: 10 de 29 fallan, justo los de FR-009, FR-010 y FR-011 ([tests-rojo-frontend.txt](./evidencias/tests-rojo-frontend.txt)) | **96/96** (81 + 15 nuevos) ([tests-despues-frontend.txt](./evidencias/tests-despues-frontend.txt)) |
| `go vet ./...`, `npx tsc --noEmit` | sin errores | — | sin errores |
| `npm run lint` | 0 errores | — | 0 errores, 6 avisos en ficheros de otras specs (analítica, balanceo), ya existentes |

Tests nuevos o ampliados: `causa_test.go` (T044), `spec002_abono_test.go` y
`spec002_estacion_test.go` (T045), `spec002_orden_test.go` (T046),
`spec002_test.go` de contrato (T047), `spec002_codigo_test.go` (T059),
`desbloqueo.test.tsx` (T048), `resultadoDesconocido.test.tsx` (T049) y
`guardia.test.tsx` (T050).

## T056 — Validación manual: antes y después

«Antes» es lo observado en las [pruebas previas](./pruebas-previas.md) sobre
`0d6f478`. «Después» se ejecutó el 2026-11-19 entre las 13:15 y las 13:30 por
API (`curl`) y en Chrome con la extensión Claude in Chrome.

| Escenario | Entrada | Esperado | Antes | Después | ¿OK? |
|-----------|---------|----------|-------|---------|------|
| **V1** (FR-002, SC-001) | Bruno (caducado el 09/11/2026) y Carla (sin abono) desbloquean `E03-A01` | Causas distintas, con la fecha en el caso de Bruno; sin efectos | P3.1–P3.3: los dos, 403 `ABONO_VENCIDO` «Tu abono anual no está vigente» | Bruno: 403 `ABONO_VENCIDO` «Tu abono anual caducó el 09/11/2026». Carla: 403 `SIN_ABONO` «No tienes abono anual». `E03-A01` sigue ocupado con su bici disponible | ✅ |
| **V7** (FR-001, D2) | Código vacío, de 65 caracteres y con `\u0000` | 422; «Introduce el código del anclaje» si está vacío, «El código no es válido» en los otros dos | P2.2–P2.3: 422 con «Introduce el código del anclaje» en todos | Vacío: «Introduce el código del anclaje». 65 caracteres y NUL: «El código no es válido». Siempre 422 `VALIDACION` | ✅ |
| **V10** (FR-010) | Bruno y Carla abren `/alquiler` | Ven su motivo antes de escribir; campo y botón desactivados | P3.2: formulario activo; el motivo solo aparece al enviar | Bruno: aviso «Tu abono anual caducó el 09/11/2026». Carla: «No tienes abono anual». En los dos, campo y botón desactivados. `GET /alquileres/activo` devuelve el mismo texto que el 403 | ✅ |
| **V11** (FR-010) | Fran abre `/alquiler` con el abono vigente; sin recargar, se le caduca por SQL y envía `E03-A02` | El formulario estaba activo, pero el servidor rechaza | — (no existía el aviso) | Formulario activo y sin aviso al abrir. Tras el `UPDATE`, al enviar: «Tu abono anual caducó el 18/11/2026» y el código se conserva en el campo | ✅ |
| **V12a** (FR-011) | Eva envía `E03-A02` con la red caída para el desbloqueo y para la consulta; después vuelve la red y pulsa «Volver a comprobar» | Aviso de que no se sabe el resultado y botón para comprobar; al volver la red, el resultado real; un solo intento de desbloqueo | No probado en las pruebas previas (hueco 7); el código mostraba «No se pudo conectar con el servidor» | Con la red caída: «No sabemos si se ha desbloqueado: no hemos podido comprobarlo…», botón «Volver a comprobar» y «Desbloquear» desactivado. Al pulsar con la red de vuelta: «No se ha desbloqueado, inténtalo de nuevo» y formulario activo. 1 intento de `POST`, 0 viajes | ✅ |
| **V12b** (FR-011) | Eva envía `E03-A02`; el `POST` llega al servidor (201) pero la respuesta se pierde | La página comprueba y muestra el viaje; un solo `POST` | — | Tras la comprobación: «Viaje en curso · BICI-0015 · Plaça de la Reina». 1 `POST` (201 en el servidor), 1 consulta; en la BD, **1** viaje de Eva | ✅ |
| **V13** (FR-009) | El gestor y el operario de taller abren `/alquiler` escribiendo la dirección | Acaban en el mapa sin ver el formulario | P3.9: el gestor veía el formulario y solo al enviar recibía «No tienes permiso…» | Gestor: redirigido a `/`, sin formulario; menú «Mapa» y «Analítica». Operario de taller: redirigido a `/`, sin formulario | ✅ |
| **V14** (FR-012) | Bruno envía `NO-EXISTE` | 403 `ABONO_VENCIDO`, no 404 | P3.5: 403 `ABONO_VENCIDO` (el orden ya existía, pero no estaba especificado) | 403 `ABONO_VENCIDO` «Tu abono anual caducó el 09/11/2026» | ✅ |

**Cómo se simuló la red en V12.** La extensión no puede activar el modo
«Offline» de DevTools, así que se sustituyó `window.fetch` en la página: en
V12a, las peticiones a `/alquileres/desbloqueo` y `/alquileres/activo` fallan
con `TypeError` sin salir del navegador hasta que se «restablece la red»; en
V12b, la petición del desbloqueo llega al servidor y es su respuesta la que se
descarta. Así se reproducen los dos casos que la app no puede distinguir. El
plazo de 10 s lo cubre el test automático T049 (caso c).

## T057 — Regresión de HU-02

| Escenario | Resultado después del refinamiento | ¿OK? |
|-----------|------------------------------------|------|
| V2 (FR-001, FR-005, FR-008) | Fran con `"  e03-a06 "` → 201 con BICI-0019 en Plaça de la Reina. Eva desde la interfaz (V12b) → vista del viaje con hora, bici y estación | ✅ |
| V3 (FR-006) | Eva recarga `/alquiler` → sigue viendo «Viaje en curso · BICI-0015», sin formulario | ✅ |
| V4 (FR-003) | Eva, con viaje, pide `E03-A05` → 409 `VIAJE_ACTIVO` | ✅ |
| V5 (FR-004) | Fran: código inexistente → 404 `ANCLAJE_NO_EXISTE`; bici averiada → 409 `BICICLETA_NO_DISPONIBLE`; estación fuera de servicio → 409 «La estación de este anclaje está fuera de servicio» | ✅ |
| V6 (FR-009) | El gestor no ve «Desbloquear» en el menú; por API → 403 `ROL_INSUFICIENTE` | ✅ |
| V7 (FR-001) | Ver T056 | ✅ |
| V8 (const. IV) | Plaça de la Reina: de 17 bicis y 8 libres a 15 y 10 tras los dos desbloqueos (capacidad 25) | ✅ |
| V9 (FR-008) | Eva inicia sesión y llega al mapa con el aviso del viaje (BICI-0015) y el enlace «Ver» a `/alquiler` | ✅ |
| SC-002 | Usuarios con más de un viaje en curso: **0** | ✅ |

## Conclusión

Los 6 criterios que la matriz tenía en 🛠 (FR-002, SC-001, FR-009, FR-010,
FR-011 y FR-012) y la mejora técnica de D2 están implementados, probados
automáticamente y comprobados en el navegador. La regresión de V2–V9 no
muestra cambios en lo que ya funcionaba. Queda pasar `/speckit-converge`
(§5.3) para confirmar que no falta nada respecto a la spec.
