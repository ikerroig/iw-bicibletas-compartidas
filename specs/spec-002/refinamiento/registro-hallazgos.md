# Registro de hallazgos del analyze — spec-002 (HU-02)

**Enunciado**: §5.2, paso 3. Se ha contrastado cada hallazgo del
[informe inicial](./analyze.md) con el material del proyecto (spec, plan,
tareas, contrato, quickstart, seed, esquema y código) y se ha decidido qué
hacer con él.
**Responsable**: Iker Roig · **Revisor**: Alex Morell · **Fecha**: 2026-11-19

**Estados**: *Corregido* = el documento ya está corregido · *Tareas añadidas* =
la corrección documental está hecha y el código queda para `/speckit-implement`
(§5.3) · *Aceptado* = no requiere cambios, con la justificación indicada.

| ID | Historia / criterio | Gravedad | Evidencia | Decisión | Responsable | Estado |
|----|---------------------|----------|-----------|----------|-------------|--------|
| I2 | HU-02 · `quickstart.md` V2–V5, V9 | ALTA → **MEDIA** | El seed crea el viaje activo de Ana para la demo de HU-03 (`backend/cmd/seed`, `sembrarViajesDemo`); en las pruebas previas, Ana recibe 409 `VIAJE_ACTIVO` (P3.4). Ningún usuario del seed tiene abono vigente y no tiene viaje: hubo que crear `eva` y `fran` por SQL | **Válido.** Se baja a MEDIA porque no afecta a requisitos ni al código, solo a la guía de validación (pero bloquearía T056 y T057). El quickstart indica ahora cómo crear `eva` y `fran` y los usa en V2–V5 y V9. No se toca el seed, que es compartido y lo usa la demo de HU-03 (Alex) | Iker Roig | Corregido |
| I1 | HU-02 · escenarios de aceptación | MEDIA | `spec.md`: el escenario 3 quedó detrás del 7 al insertar los nuevos | **Válido.** Reordenados 1–7 sin cambiar su texto | Iker Roig | Corregido |
| U1 | HU-02 · FR-002, entidad Usuario | MEDIA | La entidad decía «abono anual (activo/inactivo)»; FR-002 y SC-001 usan la fecha para «caducó el DD/MM/AAAA». La columna `fecha_vencimiento_abono` ya existe en `0001_esquema_base.sql` | **Válido.** Se añade la fecha de vencimiento a la entidad; no hay cambio de esquema | Iker Roig | Corregido |
| U2 | HU-02 · FR-001, mensaje de código mal formado | MEDIA | El límite de 64 caracteres solo está en el plan, el contrato y el data-model. P2.3: un código demasiado largo responde «Introduce el código del anclaje». Es el hueco 4 de las pruebas previas, que el clarify dejó para el analyze | **Válido; decisión del grupo, opción A**: se mantiene como validación técnica del plan (el límite protege la base de datos y un código real nunca lo alcanza) y se separa el mensaje: vacío → «Introduce el código del anclaje»; largo o con caracteres de control → «El código no es válido». Se actualizan research D2, el contrato y el plan, y se añaden T059 (test) y T060 (implementación). Se descartan B (llevarlo a la spec, sobredimensiona una regla técnica) y C (dejarlo, mensaje engañoso) | Iker Roig | Tareas añadidas |
| I3 | HU-02 · caso límite de bicicleta no disponible | MEDIA | El caso límite decía «averiada o fuera de servicio»; el esquema solo admite `disponible`, `en_uso`, `averiada` y `en_taller` (`CHECK` de `bicicletas`) | **Válido.** Cambiado a «averiada o en taller», que es lo que hace el código (`ErrBicicletaNoDisponible` si no está `disponible`) | Iker Roig | Corregido |
| I4 | HU-02 · entidad Anclaje, FR-004, FR-012 | MEDIA | La entidad decía «ocupado/libre»; el esquema admite `fuera_de_servicio` y FR-012 lo usa («anclaje en servicio»); P2.7 lo prueba con `E15-A06` | **Válido.** Se añade el estado a la entidad | Iker Roig | Corregido |
| I5 | HU-02 · clarify, pregunta 2 → FR-011 | MEDIA | La respuesta del clarify habla de «la respuesta no llega»; al actualizar el plan, FR-011 y el escenario 5 se ampliaron a los errores 5xx sin dejar constancia | **Válido, fallo de proceso.** No es una decisión funcional nueva (§5.2, paso 4): es la misma respuesta aplicada a un caso equivalente, porque un 5xx tampoco garantiza que el viaje no se haya creado (research D12). Se deja constancia en la spec, junto a la respuesta, y en `clarify.md` | Iker Roig | Corregido |
| A1 | HU-02 · FR-011, escenario 5, caso límite | BAJA | «error interno del servidor» en FR-011 y en el caso límite; «error 5xx» en el escenario 5 | **Válido.** Se usa en los tres sitios «error 5xx del servidor», que coincide con D17 (`status >= 500`) | Iker Roig | Corregido |
| I6 | HU-02 · FR-009 vs. research D9 | BAJA | D9 también redirige si hay token pero no hay usuario guardado; FR-009 habla de «los demás roles» | **Aceptado sin cambio.** Sin usuario guardado no se conoce el rol, y D9 aplica el mismo criterio que el menú (T029): lo seguro es no mostrar el formulario. Es un detalle técnico que ya justifica D9; no cambia lo que la spec exige a los roles | Iker Roig | Aceptado |
| D1 | HU-02 · casos límite de pérdida de conexión | BAJA | Dos casos límite se solapan en parte | **Válido en parte.** Son casos distintos (el antiguo fija que no hay viaje parcial; el nuevo, lo que ve el usuario), así que se mantienen los dos y el antiguo remite a FR-011 | Iker Roig | Corregido |
| C1 | HU-02 · `tasks.md`, estrategia de entrega | BAJA | «Incremental Delivery» terminaba en el paso 5 | **Válido.** Se añade el paso 6 (Phase 6) y T059/T060 a dependencias y checkpoint | Iker Roig | Corregido |
| C2 | HU-02 · `plan.md`, árbol de ficheros | BAJA | Faltaban los tests nuevos de T044, T046, T049 y T050 | **Válido.** Añadidos al árbol | Iker Roig | Corregido |
| K1 | HU-02 · T043, T058; constitución III | BAJA | Los cambios compartidos se comunicaban solo en la PR | **Válido como buena práctica**, no como incumplimiento: el principio III exige acuerdo previo para el esquema y Redis, y aquí no hay ni migraciones ni claves nuevas. T043 dice ahora que hay que avisar al equipo **antes** de empezar | Iker Roig | Corregido (el aviso se hará al empezar T043) |

## Dependencias con historias de otros integrantes

- **spec-003 (Alex Morell)**: `pages/devolucion/index.tsx` y su test usan
  `fijarViajeEnCurso` y `viajeEnCurso` de `stores/viaje.ts`. T053 conserva
  sus firmas y guarda el abono en una señal aparte.
- **Seed (compartido)**: no se toca para no romper la demo de HU-03; los
  usuarios de prueba de la HU-02 se crean por SQL (I2).
- **OpenAPI, `httperr` y `frontend/tests/handlers.ts` (compartidos)**: cambios
  aditivos (T043, T051, T053), avisando al equipo antes (K1).
- Ninguna otra spec consume `GET /alquileres/activo`, `ABONO_VENCIDO` ni
  `models.AbonoVigente` (comprobado con `grep`).

## Resumen

| | Total | Válidos | Válidos en parte | Falsos positivos | Aceptados sin cambio |
|---|---|---|---|---|---|
| Hallazgos del analyze | 13 | 11 (I1–I5, U1, U2, A1, C1, C2, K1) | 1 (D1) | 0 | 1 (I6) |

Una gravedad ajustada (I2, de ALTA a MEDIA) con su motivo. Ningún hallazgo
reveló una decisión funcional pendiente que obligase a volver al clarify: U2
era el hueco que el propio clarify había dejado para esta fase, y se resolvió
como validación técnica.

## Revisión entre compañeros (§5.1 y §5.2) — a rellenar por Alex Morell

> El enunciado pide que el revisor contraste las decisiones del clarify, elija
> **al menos un criterio** de la historia y compruebe que la tarea propuesta
> permitiría verificarlo. Si no hay incidencias reales, se documenta la
> cobertura comprobada, sin inventarse errores.

| Qué se revisó | Criterio / decisión | Cómo se comprobó | Resultado | Fecha |
|---------------|---------------------|------------------|-----------|-------|
| Decisiones del clarify (5 de la sesión 2026-11-19) | Preguntas 1–5 → FR-002/FR-010 (abono caducado vs. sin abono y aviso al entrar), FR-011 (resultado desconocido y comprobación sin reintento de POST), FR-009 (redirección operario/gestor al mapa), FR-012 (orden determinista de validaciones) | Lectura de `clarify.md`, contraste con la narrativa (reducción de reclamaciones por cobros no entendidos) y verificación de los escenarios de aceptación V1, V10, V11, V12a/b, V13, V14 en `quickstart.md` y `verificacion-final.md` | **De acuerdo con las 5 decisiones.** Están sólidamente fundamentadas en la narrativa y la constitución del proyecto. La decisión Q2/FR-011 (ante 5xx o timeout comprobar el estado del viaje antes de reintentar el POST) es especialmente acertada porque un reintento a ciegas podría generar dobles viajes o un 409 confuso. La decisión Q1/FR-010 de avisar en la pantalla antes de que el usuario envíe el formulario evita desplazamientos y frustración innecesaria. | 2026-11-19 |
| Cobertura de un criterio en la matriz | **FR-010** (aviso previo de abono al abrir) y **FR-011** (resultado desconocido del desbloqueo), filas de `docs/matriz-trazabilidad.md` | Trazabilidad completa spec → plan (D16, D17) → tareas (T047, T048, T052, T053 para FR-010; T049, T054, T061–T063 para FR-011) → tests. **Ejecutado**: `cd frontend && npx vitest run tests/alquiler` → **5 archivos, 30/30 tests en verde**, incluyendo `desbloqueo.test.tsx` (10 tests) y `resultadoDesconocido.test.tsx` (6 tests que prueban timeout > 10 s, red caída y error 5xx con un solo POST). Tests de backend de `internal/services/alquiler` (`causa_test.go`, `TestFechaNegocio`) en verde | **Cobertura verificada y comprobada.** Las tareas y los tests reflejan con rigor los criterios de aceptación y los escenarios manuales V10, V11 y V12a/b. No hay lagunas de implementación ni discrepancias entre la especificación y los tests ejecutados. | 2026-11-19 |
| Registro de hallazgos: I2, U2, I6 | I2 → seed y quickstart; U2 → código mal formado; I6 → roles y D9 | Contraste con el código y ficheros compartidos: en **I2**, no tocar el seed compartido fue una decisión acertada para no colisionar con los fixtures de HU-03 (mi historia) y resolverlo vía documentación SQL en el quickstart; en **U2**, separar el mensaje de vacío («Introduce el código...») del de código inválido («El código no es válido») protege el backend sin inflar la spec; en **I6**, la redirección por URL sin usuario autenticado respeta el principio de mínima sorpresa | **Las tres justificaciones se sostienen.** El ajuste de gravedad de I2 (de ALTA a MEDIA) es totalmente procedente al ser un tema de guía de pruebas y no un defecto funcional. Se agradece expresamente haber preservado el seed de HU-03 sin romper dependencias cruzadas. | 2026-11-19 |

**Resultado general de la revisión**: sin incidencias que bloqueen;
se documenta la cobertura comprobada (§5.2, revisión entre compañeros). Los
16 criterios de la HU-02 quedan trazados, cubiertos y con pruebas verificadas
tanto en backend como en frontend.

**Firma del revisor**: Alex Morell — 2026-11-19
