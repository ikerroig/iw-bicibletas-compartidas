# Verificación final de HU-01 — T109 (antes/después) y T110 (regresión)

**Responsable**: Iker Roig · **Revisor**: _(compañero)_
**Fecha**: 2026-11-19, 11:52–12:01 (hora local)
**Versión del código**: commit `8c4e4f2` (rama `refinamiento/spec-001`; tras
T101–T114 y converge 2 *converged*)
**Entorno**: Docker (PostgreSQL 16 + Redis 7), seed recién cargado,
`go run ./cmd/server`, `npm run dev` (http://localhost:5173), Chrome con la
extensión Claude in Chrome. Usuario de prueba `ana@si16.test` (rol `usuario`).
**Enunciado**: §5.3 paso 4 («para cada corrección, conservar una prueba
reproducible del fallo previo y del resultado posterior») y §5.4 «Regresión».

## T109 — Antes y después de cada corrección (escenarios S1f–S1i)

«Antes» = [pruebas-previas.md](./pruebas-previas.md) (commit `44b5050`). «Después» = esta verificación.

| Escenario | Entrada | Esperado | Antes (observado) | Después (observado) |
|-----------|---------|----------|-------------------|---------------------|
| **S1f** — primera carga fallida (FR-010) | Backend parado y recarga del mapa con sesión iniciada | «No se pudo cargar la disponibilidad», botón «Reintentar» y ningún marcador | ❌ P2.5: «Sin datos todavía · Error 500. Reintentando…», sin botón | ✅ Mensaje «No se pudo cargar la disponibilidad · Lo volveremos a intentar automáticamente», botón «Reintentar», 0 marcadores y 0 contadores |
| S1f — sin datos durante más de 30 s (converge F2) | Igual, esperando 33 s | Solo el mensaje de FR-010 | ❌ (detectado en converge 1): también salía «Datos desactualizados · Mostramos los últimos datos conocidos» | ✅ A los 33 s: mensaje de FR-010, **sin** «Datos desactualizados» ni «últimos datos conocidos» |
| S1f — recuperación | Backend arriba y pulsar «Reintentar» | Carga inmediata y el mensaje desaparece | — (no había botón) | ✅ 60 marcadores **602 ms** después del clic; mensaje retirado; «Actualizado a las 11:55:01» |
| **S1g** — listado con el mapa de fondo caído (FR-011, SC-008) | Capa de teselas eliminada de la página (la extensión no puede bloquear peticiones) → pestaña «Lista» | Todas las estaciones, con el mismo estado y contadores que el mapa | ❌ P2.7: no existía el listado | ✅ 60 filas; sus `aria-label` son **idénticos** a los de los 60 marcadores (Jardí Botànic «Sin bicicletas ni anclajes libres», Mercat Central «Fuera de servicio», etc.) |
| S1g — frescura y foco (FR-003, converge F1) | Fila «Algirós» enfocada; `UPDATE bicicletas SET estado='averiada'` (bici 499, commit 10:55:18.6 UTC) | La fila se actualiza en ≤ 30 s sin perder el foco | ❌ (converge 1): cada refresco volvía a crear las filas y se perdía el foco | ✅ «8 → 7 bicicletas» a las 10:55:34.4 (**15,8 s**); **la misma fila** sigue en la página y **conserva el foco**. Bici revertida después |
| **S1h** — sesión no válida (FR-012) | `curl` sin token; con el mapa abierto, backend reiniciado con otro `JWT_SECRET` | 401; la aplicación lleva al login y, al entrar, vuelve al mapa | ✅ P3.3–P3.6 ya cumplían | ✅ `HTTP 401 UNAUTENTICADO`; en ≤ 20 s la aplicación pasa sola a `/login`; tras entrar, `/` con 60 marcadores |
| **S1i** — encuadre fuera del área (FR-013) | Geolocalización simulada en Alicante (38,385 / −0,513) **antes** de los datos | Encuadre en toda la red, estaciones visibles | ❌ H1: con la ubicación concedida, los 60 marcadores fuera de pantalla | ✅ `data-encuadre="red"`, **60/60** marcadores en pantalla |
| S1i — posición **después** de los datos | Igual, con la posición 2 s más tarde | Se mantiene la red | ❌ (lógica anterior: centraba en el usuario) | ✅ `red`, 60/60 en pantalla |
| S1i — usuario **dentro** del área | Posición en València (39,4699 / −0,3763) | Centrado en el usuario | ✅ | ✅ `data-encuadre="usuario"`, 37 estaciones cercanas en pantalla (zoom de calle) |

## T110 — Regresión

### Automática ([evidencias/regresion-T110.txt](./evidencias/regresion-T110.txt))

| Suite | Resultado |
|-------|-----------|
| Backend `go test -count=1 ./...` (sin caché) | 9 paquetes `ok`, 0 `FAIL` |
| Frontend `vitest run` | 14 ficheros, **81/81** tests |
| Frontend `tsc --noEmit` | Sin errores de tipos |
| Lint de los ficheros de HU-01 | Sin avisos (los 6 avisos del proyecto son anteriores y de spec-007/008) |

### Manual (S1–S1e, con el código final)

| Escenario | Resultado |
|-----------|-----------|
| S1 — mapa y detalle | ✅ 60 marcadores; detalle de Xàtiva - Bailén en **8 ms** |
| S1a — datos desactualizados con datos previos (FR-007) | ✅ Backend parado a las 11:59:29: a los 40 s, «Datos desactualizados» con la hora (11:59:25), 60 marcadores y el detalle (5 / 4) visibles, **sin** el mensaje de FR-010. Al volver el backend, el aviso desaparece («Actualizado a las 12:00:25») |
| S1b — estados visuales | ✅ Torres de Quart «Sin bicicletas», Mercat de Russafa «Sin anclajes libres», Jardí Botànic «Sin bicicletas ni anclajes libres» |
| S1c — fuera de servicio | ✅ Mercat Central y L'Oceanogràfic «Fuera de servicio» (el paso por SQL en ≤ 30 s ya se midió en P3.2: 11,7 s; el detalle no ha cambiado) |
| S1d — camino de respaldo | ✅ Todas las modificaciones de esta verificación se hicieron con `UPDATE` directo, sin la API: 15,8 s (S1g) |
| S1e — rendimiento (SC-004) | ✅ Red local **sin limitar**: marcadores pintados en ≤ 263 ms, respuesta de `/estaciones` en 171 ms (13,8 kB) y detalle en 8 ms. ⚠️ **Pendiente de repetir a mano con el perfil «Fast 4G»** de DevTools, que la extensión no puede activar |

## Incidencias de la sesión de pruebas (no son defectos de la aplicación)

- Un clic con una referencia equivocada abrió «Reportar avería» (solo navegación, sin enviar nada); se repitió la prueba.
- En la regresión, el texto escrito por la extensión no llegaba a los campos del login (la ventana de Chrome no tenía el foco del teclado). Se rellenó el formulario desde la página. El login con teclado había funcionado en esta misma sesión (11:16 y 11:57) y lo cubre `sesionCaducada.test.tsx`.

## Conclusión

Todos los criterios de HU-01 quedan verificados sobre la versión final. El
único punto abierto es repetir **S1e con «Fast 4G»** a mano para cerrar del
todo el SC-004.
