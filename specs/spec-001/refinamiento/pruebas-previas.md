# Pruebas previas al clarify — HU-01 (spec-001)

**Responsable**: Iker Roig · **Revisor**: Alex Morell
**Enunciado**: §5.1, paso 1. Probar la historia en tres situaciones: el
recorrido habitual, una entrada inválida o un caso límite, y una condición de
permisos o de estado.

> **Nota de trazabilidad**: estas pruebas se ejecutan **después** de la sesión
> de clarify del 2026-11-19 y no antes, como pide el enunciado. Se hacen sobre
> la **versión de referencia** (commit `44b5050`, anterior a las respuestas del
> clarify) para dejar constancia de cómo se comportaba la aplicación cuando se
> tomaron las decisiones. En la columna «Esperado» se indica qué criterio lo
> exige: los marcados con ★ salen del clarify del 2026-11-19, así que es
> normal que todavía fallen.

## Entorno

| Dato | Valor |
|------|-------|
| Versión del código | `44b5050` (main + `docs/agrupamiento`), sin cambios de la HU-01 posteriores |
| Spec Kit | 1.1.3 (`.specify/init-options.json`) |
| Arranque | `docker compose up -d` → `backend`: `go run ./cmd/migrate up`, `go run ./cmd/seed`, `go run ./cmd/server` → `frontend`: `npm run dev` (http://localhost:5173) |
| Navegador | Chrome con la extensión Claude in Chrome (pruebas automatizadas, 2026-11-19 11:15–11:21 hora local) |
| Usuario de prueba | `ana@si16.test` / `password` (rol `usuario`, abono vigente) |

Estaciones del seed con casos especiales:

| Estación | Situación |
|----------|-----------|
| Mercat Central | fuera de servicio |
| L'Oceanogràfic | fuera de servicio |
| Torres de Quart | 0 bicicletas disponibles |
| Mercat de Russafa | llena (0 anclajes libres) |
| Jardí Botànic | todos los anclajes ocupados por bicicletas averiadas (0 y 0) |
| Xàtiva - Bailén | un anclaje fuera de servicio sin bicicleta |

---

## P1 — Recorrido habitual

**Precondición**: seed recién cargado; sesión iniciada con `ana@si16.test`.

| # | Paso | Esperado (criterio) | Observado | ¿OK? |
|---|------|---------------------|-----------|------|
| P1.1 | Abrir http://localhost:5173/ | Mapa con las ~60 estaciones; «Actualizado a las HH:MM:SS» visible (FR-001, FR-007, SC-001) | 60 marcadores con `aria-label` descriptivo; «Actualizado a las 11:15:14» | ✅ |
| P1.2 | Hacer clic en una estación normal | Detalle con nombre, bicis disponibles y anclajes libres en < 2 s (FR-002, FR-004, SC-004) | Xàtiva - Bailén: «Operativa · Bicicletas disponibles 5 · Anclajes libres 4» en **58 ms** | ✅ |
| P1.3 | Con el detalle abierto, cambiar la estación desde SQL: `UPDATE bicicletas SET estado='averiada' WHERE id = (SELECT bicicleta_id FROM anclajes WHERE estacion_id = <id> AND bicicleta_id IS NOT NULL LIMIT 1);` y esperar | El contador de disponibles baja 1 en ≤ 30 s (FR-003, SC-002) | Bici 107 a `averiada` (commit 10:16:43.9 UTC): el contador pasa de 5 a 4 en **≈ 16 s**, por el camino de respaldo (barrido + polling) | ✅ |
| P1.4 | Igual que P1.3: ¿el **detalle abierto** cambia solo, sin cerrarlo? | Sí, sin perder la selección (★ FR-004 ampliado) | El detalle abierto se actualizó solo, sin cerrarse ni perder la selección | ✅ |

## P2 — Casos límite y entrada inválida

**Precondición**: la misma que en P1.

| # | Paso | Esperado (criterio) | Observado | ¿OK? |
|---|------|---------------------|-----------|------|
| P2.1 | Mirar en el mapa Torres de Quart, Mercat de Russafa y Jardí Botànic | Cada marcador muestra su estado con icono y texto, no solo con color (FR-008) | Marcadores con icono y texto: «∅ Sin bicis», «▣ Llena», «⊗ Sin bicis · Llena»; `aria-label` p. ej. «Jardí Botànic: Sin bicicletas ni anclajes libres. 0 bicicletas disponibles, 0 anclajes libres» | ✅ |
| P2.2 | Abrir el detalle de Jardí Botànic | «0 · Sin bicicletas» y «0 · Sin anclajes libres»; sin desglose de averías; la suma es menor que la capacidad (FR-006, FR-008) | «0 · Sin bicicletas» y «0 · Sin anclajes libres», sin desglose; en la BD hay 15 anclajes con bici averiada y capacidad 15 (suma 0 < 15) | ✅ |
| P2.3 | Abrir el detalle de Xàtiva - Bailén | El anclaje roto no cuenta como libre (SC-003) | API: 5 disponibles y 4 libres; BD: 5 ocupados (disponible), 4 libres y 1 `fuera_de_servicio` → el anclaje roto no cuenta | ✅ |
| P2.4 | Entrada inválida: `curl -H "Authorization: Bearer <token>" http://localhost:5173/api/v1/estaciones/99999` | 404 con el formato de error `{error:{codigo,mensaje}}` (contrato) | `HTTP 404 {"error":{"codigo":"NO_ENCONTRADO","mensaje":"Estación no encontrada"}}`; también con `/estaciones/abc` | ✅ |
| P2.5 | Parar el backend y **recargar** la página (primera carga sin datos previos) | Mensaje «No se pudo cargar la disponibilidad», botón «Reintentar», reintento automático y sin marcadores (★ FR-010) | «Sin datos todavía · Error 500. Reintentando…»; 0 marcadores; **ni el texto exigido ni el botón «Reintentar»** | ❌ (★) |
| P2.6 | Volver a arrancar el backend sin recargar | El mapa se carga solo y el mensaje desaparece (★ FR-010) | Sin recargar, a los ≤ 25 s aparecen los 60 marcadores y desaparece el error | ✅ |
| P2.7 | Bloquear las teselas del mapa (DevTools → Network → bloquear `tile.openstreetmap.org`) y recargar | Las estaciones se pueden consultar igualmente desde un listado de texto (★ FR-011) | Simulado quitando la capa de teselas (la extensión no puede bloquear peticiones): los 60 marcadores siguen en la página, pero **no hay listado de texto** ni otra forma de consultar | ❌ (★) |

## P3 — Permisos y estado

| # | Paso | Esperado (criterio) | Observado | ¿OK? |
|---|------|---------------------|-----------|------|
| P3.1 | **Estado**: con sesión, abrir el detalle de Mercat Central | «Fuera de servicio», «—» en lugar de los contadores y ninguna acción de recogida ni devolución; el marcador se puede seleccionar (FR-009) | «Fuera de servicio», «—» en ambos contadores, «Esta estación no está disponible para recoger ni devolver bicicletas.» y ningún botón de acción; el marcador se puede seleccionar | ✅ |
| P3.2 | **Estado**: `UPDATE estaciones SET estado='fuera_de_servicio' WHERE nombre='Torres de Quart';` y esperar con el detalle abierto | En ≤ 30 s pasa a «Fuera de servicio», que prevalece sobre «sin bicicletas» (FR-009, SC-007, ★ FR-004). Revertir después | Commit 10:17:18.0 → detalle abierto en «Fuera de servicio» con «—» a las 10:17:29.7 (**11,7 s**); marcador «Torres de Quart: Fuera de servicio». Revertido después | ✅ |
| P3.3 | **Permisos**: `curl http://localhost:5173/api/v1/estaciones` sin token | 401 (★ FR-012, constitución I) | `HTTP 401 {"error":{"codigo":"UNAUTENTICADO",...}}`; con un token inválido también da 401 («Sesión no válida o caducada») | ✅ |
| P3.4 | **Permisos**: en una ventana privada, abrir http://localhost:5173/ directamente | Redirige al login (★ FR-012) | Sin token, `/` redirige a `/login` | ✅ |
| P3.5 | Iniciar sesión desde esa pantalla | Se vuelve al mapa (★ FR-012) | Tras entrar se llega a `/` (el mapa) con 60 marcadores. Funciona porque el mapa es la página por defecto, no porque se recuerde la ruta de origen | ✅ |
| P3.6 | **Sesión caducada**: con el mapa abierto, en DevTools → Application → Local Storage, cambiar `si16.token` por un valor inválido y esperar al siguiente refresco (≤ 15 s) | Se cierra la sesión, redirige al login y, tras entrar, vuelve al mapa (★ FR-012) | Cambiar el token en `localStorage` con la página abierta **no tiene efecto**: el token vive en memoria. Se reprodujo al recargar con el token inválido: en cuanto el servidor responde 401 la aplicación cierra la sesión y va a `/login`; al entrar, vuelve al mapa | ✅ (con matiz) |

---

## Resumen

| Situación | Pruebas | OK | Fallan | Fallos esperados (★) |
|-----------|---------|----|--------|----------------------|
| P1 Habitual | 4 | 4 | 0 | 0 |
| P2 Límite / inválida | 7 | 5 | 2 | 2 (P2.5, P2.7) |
| P3 Permisos / estado | 6 | 6 | 0 | 0 |
| **Total** | **17** | **15** | **2** | **2** |

**Conclusión**:

- Los criterios de la versión de referencia (FR-001…FR-009) **se cumplen en
  todas las pruebas**. La frescura medida (≈ 16 s y 11,7 s por el camino de
  respaldo) queda dentro de los 30 s.
- Los dos fallos son los esperados (★) y corresponden a las decisiones nuevas
  del clarify: **FR-010** (P2.5: falta el texto «No se pudo cargar la
  disponibilidad» y el botón «Reintentar») y **FR-011** (P2.7: no existe el
  listado de texto). Esto confirma los hallazgos G1 y G2 del analyze y sirve
  como evidencia «antes» de las correcciones (§5.3, paso 4).
- **FR-004 ampliado** (P1.4, P3.2) y **FR-012** (P3.3–P3.6) ya se cumplían.
  El clarify los dejó escritos como criterios verificables, pero no hace
  falta cambiar el código. Matiz de P3.5: la vuelta al mapa sale bien porque
  el mapa es la página por defecto; la aplicación no recuerda la ruta de
  origen, cosa que el FR-012 no exige.

## Hallazgo nuevo (no previsto en la lista)

**H1 — Mapa centrado lejos de la red.** Durante P2.5 y P2.6 el navegador tenía
concedido el permiso de ubicación y el equipo estaba fuera del área de
Valencia. Cuando la posición llegó antes que los datos, `MapaBase.tsx` centró
el mapa en el usuario y, al cargar las estaciones, ya no volvió a encuadrar la
red: **los 60 marcadores quedaron fuera de la pantalla** (primer marcador en
x = 3916 y y = −31954 px) y el usuario veía un mapa sin ninguna estación. Es
lo que pide la T030 («si concede permiso, centra en su posición»), pero la
spec no define qué hacer cuando el usuario está **fuera del área de la red**.
Es una ambigüedad funcional: según el §5.2 (paso 4), hay que volver a
**clarify** antes de tocar el plan o el código.

**Resolución**: pregunta 5 del clarify del 2026-11-19 → **FR-013** (si el
usuario está fuera del área de la red, se encuadra toda la red). Ver
[clarify.md](./clarify.md).
