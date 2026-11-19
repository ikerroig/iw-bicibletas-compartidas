# Pruebas previas al clarify — HU-02 (spec-002)

**Responsable**: Iker Roig · **Revisor**: Alex Morell
**Enunciado**: §5.1, paso 1. Probar la historia en tres situaciones: el
recorrido habitual, una entrada inválida o un caso límite, y una condición de
permisos o de estado.

> A diferencia de la spec-001, estas pruebas se hacen **antes** del clarify,
> en el orden que pide el enunciado. Se ejecutan sobre la versión de
> referencia de `main` (commit `0d6f478`), sin ningún cambio del refinamiento.
> La columna «Esperado» cita el requisito de la spec que lo exige.

## Entorno

| Dato | Valor |
|------|-------|
| Versión del código | `0d6f478` (main tras fusionar las 8 specs y el fix de integración) |
| Spec Kit | 1.1.3 (`.specify/init-options.json`) |
| Arranque | `docker compose up -d` → `backend`: `go run ./cmd/migrate up`, `go run ./cmd/seed`, `go run ./cmd/server` → `frontend`: `npm run dev` (http://localhost:5173) |
| Herramientas | `curl` contra la API (pruebas de reglas y concurrencia) y Chrome con la extensión Claude in Chrome (interfaz), 2026-11-19 12:43–12:55 hora local |

Usuarios de prueba. Los del seed, más seis creados por SQL **solo en la base de
datos local** (mismo hash de contraseña que `ana`), porque el seed no tiene
ningún usuario con abono vigente y sin viaje:

| Usuario | Rol | Situación |
|---------|-----|-----------|
| `ana@si16.test` | usuario | abono vigente, **viaje activo** |
| `bruno@si16.test` | usuario | abono marcado activo pero vencido hace 10 días |
| `carla@si16.test` | usuario | sin abono |
| `gestor@si16.test`, `taller@si16.test` | gestor / operario_taller | abono vigente |
| `eva`, `fran`, `kiko`, `lola` `@si16.test` *(creados)* | usuario | abono vigente, sin viaje |
| `hugo@si16.test` *(creado)* | usuario | abono que **vence hoy** |
| `ines@si16.test` *(creado)* | usuario | abono que **venció ayer** |

Anclajes utilizados:

| Anclaje | Situación |
|---------|-----------|
| `E03-A01`…`E03-A17` (Plaça de la Reina) | bici disponible |
| `E15-A01` (Xàtiva - Bailén) | bici disponible |
| `E06-A02` (Torres de Quart) | anclaje vacío |
| `E19-A01` (Jardí Botànic) | bici averiada |
| `E04-A01` (Mercat Central) | estación fuera de servicio |
| `E15-A06` (Xàtiva - Bailén) | anclaje fuera de servicio |

---

## P1 — Recorrido habitual

**Precondición**: sesión de `eva@si16.test` en Chrome.

| # | Paso | Esperado (criterio) | Observado | ¿OK? |
|---|------|---------------------|-----------|------|
| P1.1 | Menú → «Desbloquear» | Formulario con campo «Código de anclaje» (FR-001, FR-009) | Formulario «Desbloquear bicicleta» con el campo y el ejemplo `E01-A01` | ✅ |
| P1.2 | Escribir `E03-A08` y pulsar Intro | Se crea el viaje y se muestran hora de inicio, código de bici y estación de origen (FR-005, FR-008) | «Viaje en curso · Inicio: 19/11/26, 12:46 · Bicicleta: BICI-0021 · Estación de origen: Plaça de la Reina · No puedes desbloquear otra bicicleta hasta devolver esta.» | ✅ |
| P1.3 | Tiempo de respuesta del desbloqueo | < 10 s (SC-003) | Por API, 24–32 ms en los desbloqueos válidos; en la interfaz, inmediato | ✅ |
| P1.4 | Recargar `/alquiler` | Se sigue viendo el viaje activo con los mismos datos (FR-006, FR-008) | Mismos datos tras recargar; el formulario no aparece | ✅ |
| P1.5 | Ir al mapa | El viaje en curso se ve también al volver a la app (FR-008); el contador de la estación de origen baja | El mapa muestra el viaje en curso. Plaça de la Reina: de 17 a **14** bicis disponibles y 11 anclajes libres tras los 3 desbloqueos de la sesión (suma = capacidad 25) | ✅ |
| P1.6 | En la BD, revisar el anclaje, la bici y el viaje | Anclaje libre, bici en uso y viaje `en_curso` con hora de inicio (FR-005, FR-007) | `E03-A02/A03` libres sin bici; BICI-0015/0016 `en_uso`; viajes `en_curso` con `fecha_inicio` | ✅ |

## P2 — Entradas inválidas y casos límite

Por API (`POST /api/v1/alquileres/desbloqueo`) con el token de `eva`, salvo
donde se indica otro usuario. Después de cada rechazo se comprobó en la BD que
el anclaje y la bici no habían cambiado.

| # | Paso | Esperado (criterio) | Observado | ¿OK? |
|---|------|---------------------|-----------|------|
| P2.1 | Código inexistente `E99-A99` (API y formulario) | Rechazo con causa y sin efectos (FR-004, *edge case* 1) | `404 ANCLAJE_NO_EXISTE` «No existe ningún anclaje con ese código»; en la interfaz aparece el mismo texto en rojo bajo el campo | ✅ |
| P2.2 | Código vacío (`"   "`) o sin el campo | Rechazo | `422 VALIDACION` «Introduce el código del anclaje» | ✅ |
| P2.3 | Código de 65 caracteres | Rechazo | `422 VALIDACION` «Introduce el código del anclaje»: el mensaje no explica que el código es demasiado largo | ✅ (con matiz) |
| P2.4 | Anclaje vacío `E06-A02` | Rechazo con causa (FR-004) | `409 ANCLAJE_NO_DISPONIBLE` «Este anclaje no tiene ninguna bicicleta» | ✅ |
| P2.5 | Bici averiada `E19-A01` | Rechazo con causa (FR-004) | `409 BICICLETA_NO_DISPONIBLE` «La bicicleta de este anclaje no está disponible» | ✅ |
| P2.6 | Estación fuera de servicio `E04-A01` | Rechazo con causa (FR-004) | `409 ANCLAJE_NO_DISPONIBLE` «La estación de este anclaje está fuera de servicio» | ✅ |
| P2.7 | Anclaje fuera de servicio `E15-A06` | Rechazo con causa (FR-004) | `409 ANCLAJE_NO_DISPONIBLE` «Este anclaje está fuera de servicio» | ✅ |
| P2.8 | `hugo` (abono vence **hoy**) con `"  e03-a02 "` | Se acepta: el abono vale hasta las 23:59:59 de hoy, y el código se acepta en minúsculas y con espacios (FR-001, FR-002) | `201`, viaje con BICI-0015 en Plaça de la Reina | ✅ |
| P2.9 | `fran`: 5 desbloqueos **simultáneos** en anclajes distintos | Como mucho uno crea viaje (FR-007, SC-002) | 1 × `201` y 4 × `409 VIAJE_ACTIVO`; en la BD, 1 solo viaje en curso de `fran` | ✅ |
| P2.10 | `kiko` y `lola` desbloquean a la vez el **mismo** anclaje `E15-A01` | Solo uno se lleva la bici; el otro recibe un rechazo sin efectos (FR-007) | `kiko`: `201` con BICI-0107. `lola`: `409` «Este anclaje no tiene ninguna bicicleta»; `lola` sin viaje | ✅ |

## P3 — Permisos y estado

| # | Paso | Esperado (criterio) | Observado | ¿OK? |
|---|------|---------------------|-----------|------|
| P3.1 | `bruno` (abono vencido) por API | Rechazo con causa (FR-002, SC-001) | `403 ABONO_VENCIDO` «Tu abono anual no está vigente» | ✅ |
| P3.2 | `bruno` desde el formulario | El mismo mensaje en pantalla | «Tu abono anual no está vigente» bajo el campo ⁽¹⁾ | ✅ |
| P3.3 | `carla` (sin abono) | Rechazo **indicando la causa** (FR-002) | `403 ABONO_VENCIDO` «Tu abono anual no está vigente»: el mismo mensaje que con el abono caducado, aunque Carla nunca ha tenido abono | ✅ (con matiz) |
| P3.4 | `ines` (abono vencido **ayer**) | Rechazo (FR-002) | `403 ABONO_VENCIDO` | ✅ |
| P3.5 | `bruno` con un código inexistente | La spec no fija qué causa va primero | `403 ABONO_VENCIDO`: el abono se comprueba antes que el anclaje | — (no definido) |
| P3.6 | `ana` (con viaje activo) | Denegación (FR-003, SC-002) | `409 VIAJE_ACTIVO` «Ya tienes un viaje en curso» | ✅ |
| P3.7 | `gestor` y `taller` por API | Rechazo aunque tengan abono (FR-009) | `403 ROL_INSUFICIENTE` «No tienes permiso para esta operación» en ambos | ✅ |
| P3.8 | `gestor`: menú | Sin la opción «Desbloquear» (FR-009) | Menú: solo «Mapa» y «Analítica» | ✅ |
| P3.9 | `gestor`: abrir `/alquiler` escribiendo la URL | La spec solo dice que la opción no aparece en el menú | Se muestra el formulario. Al enviar `E03-A09`: «No tienes permiso para esta operación» | ✅ (con matiz) |
| P3.10 | Sin token | 401 (constitución I) | `401 UNAUTENTICADO` «Inicia sesión para continuar» | ✅ |

⁽¹⁾ Los dos primeros intentos desde Chrome no llegaron a enviar el formulario:
el campo no cambió de valor y no salió ninguna petición. Es un fallo de la
automatización, no de la aplicación. Con un clic sobre el botón ejecutado en la
página se ve la petición (`403`) y el mensaje.

---

## Resumen

| Situación | Pruebas | OK | OK con matiz | Fallan | No definido |
|-----------|---------|----|--------------|--------|-------------|
| P1 — Recorrido habitual | 6 | 6 | 0 | 0 | 0 |
| P2 — Entradas inválidas y límites | 10 | 9 | 1 | 0 | 0 |
| P3 — Permisos y estado | 10 | 7 | 2 | 0 | 1 |
| **Total** | **26** | **22** | **3** | **0** | **1** |

Ninguna prueba falla contra la spec actual: la implementación cumple lo que
está escrito. Los matices salen de huecos de la propia spec, no de errores del
código.

## Huecos detectados para el clarify

1. **Causa del rechazo por abono** (P3.3): FR-002 pide «indicar la causa», pero
   «sin abono» y «abono caducado» dan el mismo mensaje. ¿Hay que distinguirlos?
   ¿Hay que mostrar la fecha de vencimiento?
2. **Orden de las causas** (P3.5): si fallan varias reglas a la vez (abono,
   viaje activo, anclaje), ¿cuál se muestra? Ahora va primero el usuario y
   después el anclaje, pero la spec no lo dice.
3. **Acceso a la pantalla por URL** (P3.9): FR-009 solo oculta la opción del
   menú. ¿Un gestor que escribe `/alquiler` debe ver el formulario, ser
   redirigido o ver un aviso?
4. **Mensaje de código mal formado** (P2.3): un código demasiado largo da
   «Introduce el código del anclaje». ¿Se define un formato o una longitud
   máxima y un mensaje propio?
5. **Intentos repetidos con códigos erróneos**: no hay límite. ¿Hace falta
   alguno para evitar que se prueben códigos al azar?
6. **Aviso anticipado del abono**: Bruno ve el formulario y solo se entera de
   que su abono ha caducado al enviarlo. ¿Debe avisarse antes?
7. **Pérdida de conexión** (*edge case* «pierde la conexión»): no se ha podido
   probar en esta tanda; la spec no dice qué ve el usuario si la respuesta no
   llega ni cómo sabe si el viaje se creó.
