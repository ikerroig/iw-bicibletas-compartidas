# Feature Specification: Panel de analítica de uso y exportación de datos a CSV

**Feature Branch**: `spec-008`

**Created**: 2026-10-20

**Status**: Draft

**Input**: User description: "HU-08 — Panel de analítica de uso y exportación de datos a CSV (historia.md, SI-16)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Analítica de la red con exportación (Priority: P1)

Un gestor municipal del servicio necesita evaluar el rendimiento de la red,
detectar cuellos de botella y justificar futuras ampliaciones ante el
ayuntamiento. Desde el área de administración establece un rango de fechas y
el sistema genera gráficos interactivos con el volumen de viajes distribuidos
por hora del día y por día de la semana. Cuando necesita procesar los datos
con software externo, descarga un archivo CSV con el histórico filtrado de ese
mismo rango.

**Why this priority**: Es la herramienta de decisión y rendición de cuentas del
servicio público; sin métricas exportables el gestor no puede justificar
inversiones ni detectar problemas de uso.

**Independent Test**: Con datos de viajes repartidos en varias semanas,
seleccionar un rango, verificar los gráficos por hora y día de la semana, y
descargar el CSV comprobando que contiene exactamente los viajes del rango
con los mismos totales que los gráficos.

**Acceptance Scenarios**:

1. **Given** el gestor en el área de administración, **When** establece un
   rango de fechas, **Then** el sistema genera gráficos interactivos con el
   volumen de viajes distribuidos por hora y por día de la semana.
2. **Given** el análisis de la red, **When** el gestor requiere procesar los
   datos con software externo, **Then** la aplicación le permite descargar un
   archivo en formato CSV con el histórico filtrado.

### Edge Cases

- El rango seleccionado no contiene viajes: los gráficos se muestran vacíos
  con aviso de "sin datos en el periodo" y el CSV se descarga con solo la
  cabecera.
- Rango futuro o parcialmente futuro: solo se computan los datos existentes
  hasta la fecha actual, sin error.
- Rango muy amplio (varios años): la consulta se limita al histórico
  disponible y los gráficos permanecen consultables; el CSV refleja el rango
  completo solicitado con los datos existentes.
- El CSV contiene nombres de estaciones con tildes, ñ o caracteres
  especiales: se exportan sin corrupción de caracteres, abriéndose
  correctamente en software externo.
- Los totales de los gráficos y los del CSV del mismo rango difieren: no debe
  ocurrir — ambos provienen del mismo filtro; cualquier discrepancia se
  considera defecto.
- Un usuario sin permiso de gestor intenta acceder al panel: se deniega el
  acceso al área de administración.
- Doble clic en la exportación: se descarga un único archivo consistente por
  acción.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST permitir al gestor filtrar el análisis por un
  rango de fechas inicial y final.
- **FR-002**: El sistema MUST generar gráficos interactivos con el volumen de
  viajes del rango seleccionado, distribuidos por hora del día y por día de
  la semana.
- **FR-003**: El sistema MUST permitir descargar un archivo CSV con el
  histórico de viajes filtrado por el mismo rango seleccionado.
- **FR-004**: El archivo CSV MUST contener los campos de cada viaje
  (identificador, fecha, hora, estación de origen, estación de destino,
  duración y coste) con codificación compatible con software externo de hojas
  de cálculo.
- **FR-005**: El área de administración y su exportación MUST estar restringida
  a usuarios con rol de gestor; el sistema MUST denegar el acceso al resto de
  usuarios.
- **FR-006**: Los gráficos y el CSV de un mismo rango MUST reflejar
  exactamente el mismo conjunto de viajes y totales.
- **FR-007**: El sistema MUST basarse en viajes finalizados del servicio
  (histórico completo de la red, no solo de un usuario).

### Key Entities

- **Viaje**: Registro finalizado de la red; fecha, hora, estación de origen,
  estación de destino, duración y coste; unidad de medida de la analítica.
- **Usuario (Gestor)**: Cuenta con permiso de administración que accede al
  panel de analítica y a la exportación.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de los rangos con datos generan gráficos por hora del
  día y día de la semana en menos de 5 segundos para rangos de hasta un año.
- **SC-002**: El CSV descargado contiene el 100 % de los viajes del rango
  filtrado y sus totales coinciden con los de los gráficos en el 100 % de los
  casos.
- **SC-003**: El 100 % de los intentos de acceso de usuarios sin rol de gestor
  al panel de administración son denegados.
- **SC-004**: El 100 % de los CSV descargados se abren correctamente en
  software externo de hojas de cálculo sin pérdida de caracteres.

## Assumptions

- La analítica se calcula sobre viajes finalizados de toda la red; los viajes
  aún activos no se cuentan hasta su cierre (HU-03).
- Los gráficos son interactivos en el sentido de permitir consultar valores
  al explorarlos; no se definen tipos de gráfico concretos más allá de los
  ejes hora/día de la semana.
- El CSV usa formato estándar de intercambio (separador y codificación
  compatibles con hojas de cálculo, sin pérdida de tildes).
- El rol de gestor municipal es preexistente en el sistema; su alta queda
  fuera del alcance de esta spec (entre las 8 HU validadas no se incluye).
- Los detalles de implementación (pila tecnológica, esquema de datos) se rigen
  por la constitución del proyecto y quedan fuera de esta especificación.
- El alcance es exactamente la HU-08 de `historia.md`, validada por el
  profesor; no se añaden historias nuevas.
