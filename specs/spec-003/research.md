# Research: Devolución y tarificación HU-03 (Phase 0)

**Branch**: `spec-003` | **Date**: 2026-10-27 | **Plan**: [plan.md](./plan.md)

Investigación propia de la HU-03. **No repite** las decisiones globales ya resueltas en [specs/spec-001/research.md](../spec-001/research.md) (R1 pila, R3 frescura Redis, R6 migraciones, R7 concurrencia, R8 tarifa, R9 disponibilidad): esta spec las **adopta tal cual** y solo investiga los puntos que quedaban abiertos en el `spec.md` de spec-003. Resultado: **0 NEEDS CLARIFICATION**.

---

## D1. Fórmula exacta de tarificación (adopta R8, sin desviaciones)

- **Decisión**: cálculo en el cierre, dentro de la transacción de devolución:
  - `segundos = fecha_fin − fecha_inicio` (ambas TIMESTAMPTZ; `fecha_fin` = hora de fijado en el anclaje);
  - si `segundos < 0` (reloj inconsistente, edge case del spec): se descarta `fecha_fin` propuesta y se usa `now()` validado; el coste **nunca** resulta negativo (piso 0,00 €);
  - `duracion_minutos = ceil(segundos / 60)`;
  - `coste_total = 0,00 €` si `segundos ≤ 1800`; si no, `ceil((segundos − 1800) / 60) × TARIFA_MINUTO`, con piso 0,00 €.
  - `TARIFA_MINUTO` es configuración operativa (env, ej. 0,20) — el valor numérico está fuera de alcance por asunción del spec.
- **Racional**: idéntico a R8 global → sin sorpresas al fusionar con HU-04 (historial muestra el mismo coste) y HU-08 (analítica suma el mismo campo). Casos borde del spec cubiertos: 30:00 exacto → `segundos = 1800` → 0,00 €; 30:01 → 1 minuto cobrado.
- **Alternativas consideradas**: fracción exacta de segundo (rechazada: HU-04 habla de minutos y la facturación es en céntimos); tabla `tarifas` persistida (rechazada: alcance nuevo, requeriría enmienda).

## D2. Transacción de cierre: orden de operaciones y bloqueo

- **Decisión**: `POST /devoluciones` ejecuta **una única transacción** con este orden:
  1. `SELECT … FOR UPDATE` del **anclaje** por `codigo_anclaje` (bloquea la fila físicamente);
  2. si el anclaje no está `libre` → rollback + 409 `ANCLAJE_OCUPADO` (o `ANCLAJE_NO_ENCONTRADO`/`ANCLAJE_FUERA_DE_SERVICIO` si aplica), sin ningún efecto (FR-007, edge case "todos ocupados");
  3. `SELECT … FOR UPDATE` del viaje `en_curso` del usuario autenticado; si no existe → rollback + 409 `SIN_VIAJE_ACTIVO` (FR-007);
  4. validación de reloj (D1) y cálculo de `fecha_fin`, `duracion_minutos`, `coste_total`;
  5. `UPDATE viajes SET fecha_fin, duracion_minutos, coste_total, estacion_destino_id, estado='finalizado' WHERE id = … AND estado='en_curso'` — la cláusula `AND estado='en_curso'` hace el cierre **idempotente de una sola vez** (FR-006): si `rows affected = 0` (doble confirmación casi simultánea) → rollback + 409 `VIAJE_YA_CERRADO`;
  6. `UPDATE anclajes SET estado='ocupado', bicicleta_id = … WHERE id = … AND estado='libre'`;
  7. `UPDATE bicicletas SET estado = CASE WHEN estado = 'averiada' THEN 'averiada' ELSE 'disponible' END WHERE id = …` — si la bici fue reportada averiada **durante** el viaje (HU-05), la devolución la deja anclada pero **sigue averiada** (edge case del spec: "la avería no altera la tarificación del tramo ya recorrido" ni el estado de flota);
  8. COMMIT.
  Tras el commit: reescritura de Redis (D4).
- **Racional**: adopta R7 global (`FOR UPDATE` + índices únicos parciales `ux_viajes_usuario_activo` y `ux_viajes_bicicleta_activa` como red de seguridad). El orden anclaje→viaje evita deadlock con el desbloqueo (HU-02) si el equipo establece un orden canónico anclaje-primero. La doble confirmación revienta por la condición `estado='en_curso'` del UPDATE (no por índice), y el resultado es "solo la primera tiene efecto" (edge case del spec).
- **Alternativas consideradas**: cerrar viaje sin bloquear anclaje (rechazado: dos devoluciones en el mismo anclaje librarían un `UPDATE` sobre fila ya ocupada); endpoint idempotente devolviendo 200 en la segunda llamada (rechazado: el spec exige que solo la primera tenga efecto y el openapi global fija 409).

## D3. Códigos de error del endpoint

- **Decisión**: respuesta siempre con el envelope global `{ error: { codigo, mensaje } }`:

  | HTTP | `codigo` | Condición (FR/edge case) |
  |---|---|---|
  | 401 | `NO_AUTENTICADO` | Sin JWT o token inválido |
  | 409 | `SIN_VIAJE_ACTIVO` | FR-007: no hay viaje `en_curso` del usuario |
  | 409 | `ANCLAJE_NO_ENCONTRADO` | `codigo_anclaje` inexistente |
  | 409 | `ANCLAJE_OCUPADO` | Anclaje con bicicleta / edge case "todos ocupados" |
  | 409 | `ANCLAJE_FUERA_DE_SERVICIO` | Anclaje retirado; no admite devolución |
  | 409 | `VIAJE_YA_CERRADO` | Doble confirmación casi simultánea |
  | 422 | `ENTRADA_INVALIDA` | `codigo_anclaje` vacío o malformado |

- **Racional**: alineado con el `Error` schema de `specs/spec-001/contracts/openapi.yaml` y con los códigos ya comentados en el tag spec-003 (`SIN_VIAJE_ACTIVO | ANCLAJE_OCUPADO`); el resto son extensiones naturales documentadas aquí para que tasks.md y los tests no improvisen.
- **Alternativas consideradas**: 404 para anclaje inexistente (rechazado: el contrato global de spec-003 solo declara 200/409; agrupar todos los rechazos de estado en 409 mantiene el contrato abierto sin ampliarlo).

## D4. Invalidación del estado en vivo (adopta R3, alcance exacto)

- **Decisión**: inmediatamente después del COMMIT, el servicio reescribe la clave `estacion:{id}:estado` de la **estación destino** (la del anclaje usado): `anclajes_libres −1`, `bicis_disponibles +1` (la bici pasa a `disponible` — o no suma si sigue `averiada`), `ocupacion_pct` recalculado, `actualizado_en = now()`, TTL 120 s. La estación **origen** ya fue invalidada por el desbloqueo (HU-02), pero si el cierre detecta que origen = destino se reescribe una sola vez. La actualización es *post-commit best-effort*: si Redis falla, el viaje ya está cerrado (PostgreSQL es la autoridad) y el barrido de 15 s del worker global repara la clave antes de la cota de 30 s.
- **Racional**: constitución IV + R3 global; la devolución es una de las mutaciones listadas como *write-path* de actualización síncrona. Recalcular desde la fórmula R9 (nunca contadores incrementales ciegos) evita deriva.
- **Alternativas consideradas**: delegar solo en el barrido de 15 s (rechazado: HU-07 exige recálculo inmediato para paneles de operarios); pub/sub (ya descartado en R3 global para el MVP).

## D5. Payload de respuesta y etiqueta "Tramo gratuito"

- **Decisión**: `200` con el schema `Viaje` completo del contrato global (incluye `fecha_fin`, `duracion_minutos`, `coste_total`, `estado='finalizado'`, `estacion_destino_id`). El campo opcional `etiqueta` se rellena con `"Tramo gratuito (0,00 €)"` cuando `coste_total = 0` y `estado = 'finalizado'` (HU-04 lo reutiliza en el historial; se define aquí el origen del dato).
- **Racional**: HU-03 SC/AC exige dejar claro el concepto de tramo gratuito; producirlo en el cierre evita que cada consumidor (historial HU-04, analítica HU-08) reimplemente la regla.
- **Alternativas consideradas**: que el frontend calcule la etiqueta (rechazado: const. I — la regla vive en el backend).

## D6. Validación de entrada

- **Decisión**: `codigo_anclaje` es obligatorio, string no vacío tras trim, con longitud máxima razonable (64). No se valida el formato exacto del código (prefijo `AN-`, etc.): la existencia se resuelve contra la BD y un código desconocido produce `ANCLAJE_NO_ENCONTRADO`. Entrada vacía/malformada → `422 ENTRADA_INVALIDA` antes de tocar la BD.
- **Racional**: el spec no fija formato de código; la unicidad de `codigo_anclaje` (modelo compartido) hace que la búsqueda sea un `SELECT` indexado.
- **Alternativas consideradas**: regex estricta de formato (rechazada: fricción innecesaria y no requerida; el seed define la forma real).

## D7. Rol autorizado

- **Decisión**: el endpoint exige JWT con rol `usuario` (middleware `RequireAuth` + `RequireRole("usuario")` del auth global R5). Operarios y gestores reciben 403 — la devolución es una acción del ciclo de alquiler del usuario, no del personal.
- **Racional**: `data-model.md` global ya asigna la devolución (HU-03) al rol `usuario`; el spec describe "usuario que se encuentra realizando un trayecto".
- **Alternativas consideradas**: cualquier rol autenticado (rechazado: abriría devoluciones por personal sin viaje activo, que caerían en `SIN_VIAJE_ACTIVO`, pero viola el principio de mínimo privilegio).

---

**Resumen**: 7 decisiones (D1–D7), 0 NEEDS CLARIFICATION. D1/D2/D4 adoptan R8/R7/R3 del plan global sin desviaciones; D3/D5/D6/D7 son especificaciones propias necesarias para que tasks.md e implementación no improvisen. Sin cambios de esquema ni de stack → sin puntos que someter a acuerdo de equipo (principio III).
