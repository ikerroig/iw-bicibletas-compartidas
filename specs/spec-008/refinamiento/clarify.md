# Registro de clarificaciones — spec-008 / HU-08

**Rama**: `refinamiento/spec-008` · **Commit base**: `f8197e91a218288ad2facddf079513801f7967d4` · **Spec Kit**: `1.1.3` (`.specify/init-options.json`)

## Sesión 2026-11-01

| Pregunta | Respuesta | Justificación | HU |
|---|---|---|---|
| ¿Cómo se interpreta el rango de fechas y qué ocurre si está invertido? | `desde` y `hasta` inclusivos sobre la fecha de inicio del viaje en Europe/Madrid; si `desde > hasta`, error de validación sin gráficos ni CSV. | Fija los bordes de consulta y hace comprobable el rechazo de entradas incoherentes. | HU-08 |
| ¿Qué fecha y criterio agrupan los viajes por hora y día de semana? | Fecha de inicio en Europe/Madrid; 24 franjas (00–23) y 7 días, lunes→domingo. | Evita diferencias de zona horaria y permite verificar los buckets sin ambigüedad. | HU-08 |
| ¿Qué formato exacto debe tener el CSV para abrirse sin pérdida de caracteres? | UTF-8 con BOM, coma, RFC 4180, CRLF; columnas y nombre de fichero fijados en FR-004. | Garantiza interoperabilidad con hojas de cálculo y preserva tildes y ñ. | HU-08 |

## Sesión 2026-11-20

| Pregunta | Respuesta | Justificación | HU |
|---|---|---|---|
| ¿Qué alcance de desglose por estación debe tener HU-08 si el equipo no aprueba ampliar el contrato de datos compartido? | No añadir ranking por estación; conservar los gráficos por hora y día de la semana. | El equipo eligió la alternativa más simple cuando no era posible consultar el cambio; coincide con `historia.md` y evita modificar el OpenAPI compartido. | HU-08 |
| ¿Qué ocurre si alguien sin rol gestor intenta consultar la analítica o exportar el CSV? | Se mantiene el requisito existente: sin sesión, `401`; con otro rol, `403`; solo `gestor` accede. | Confirmación del escenario de permisos ya especificado en FR-005 y cubierto por T8g; no añade ni cambia alcance. | HU-08 |

La decisión anterior sustituye las propuestas provisionales de ranking discutidas durante esta sesión. El contrato `AnaliticaViajes` permanece sin cambios.
