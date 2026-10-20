# ESPECIFICACIÓN DE HISTORIAS DE USUARIO - SI-16: BICICLETAS COMPARTIDAS

## Equipo y reparto de especificaciones

| Integrante | Ramas / Specs | Historias de usuario |
| --- | --- | --- |
| Iker Roig | spec-001, spec-002 | HU-01, HU-02 |
| Alex Morell | spec-003, spec-004 | HU-03, HU-04 |
| Josep Peiró | spec-005, spec-006 | HU-05, HU-06 |
| Alberto Copa | spec-007, spec-008 | HU-07, HU-08 |

#### HU-01: Consulta de disponibilidad de estaciones en tiempo real
- Como: Usuario registrado con abono anual activo.
- Quiero: Visualizar un mapa interactivo con todas las estaciones de la red y el estado en vivo de sus anclajes y bicicletas.
- Para: Planificar el punto de origen donde recoger una bicicleta y el punto de destino donde estacionarla.
- Criterios de Aceptación:
  - Dado que un usuario accede a la aplicación móvil o web, cuando se carga la pantalla principal, entonces se muestra la ubicación geolocalizada de las ~60 estaciones.
  - Dado que un usuario selecciona una estación concreta, cuando consulta sus detalles, entonces se indica con precisión el número de bicicletas disponibles y anclajes libres.
  - Dado que ocurren alzas o bajas de bicicletas en cualquier estación, cuando se actualizan los datos en el servidor, entonces la vista del usuario refleja el cambio con un retardo máximo de 30 segundos mediante Redis.

#### HU-02: Desbloqueo de bicicleta mediante código de anclaje
- Como: Usuario con abono anual en regla.
- Quiero: Introducir en la aplicación el código identificador del anclaje de la estación.
- Para: Desbloquear la bicicleta e iniciar mi trayecto.
- Criterios de Aceptación:
  - Dado un usuario con abono activo sin viajes en curso, cuando introduce el código de un anclaje con bicicleta disponible, entonces la bicicleta se desacopla, se registra la hora de inicio y se crea un registro de viaje activo.
  - Dado un usuario con abono inactivo o caducado, cuando solicita un desbloqueo, entonces el sistema rechaza la operación y muestra un mensaje informativo de la causa.
  - Dado un usuario que ya cuenta con un trayecto iniciado, cuando intenta liberar otra bicicleta, entonces el sistema deniega el acceso e impide realizar múltiples viajes simultáneos.

#### HU-03: Devolución de bicicleta y cálculo de tarificación
- Como: Usuario que se encuentra realizando un trayecto.
- Quiero: Fijar la bicicleta en un anclaje vacante de cualquier estación de la red.
- Para: Dar por finalizado el viaje y que se aplique la tarifa según el tiempo transcurrido.
- Criterios de Aceptación:
  - Dado un viaje activo, cuando la bicicleta es introducida en un anclaje libre, entonces el sistema registra la hora de llegada, libera el estado del usuario y bloquea la bicicleta en el anclaje.
  - Dado un viaje cuya duración total sea menor o igual a 30 minutos, cuando el sistema calcula el importe del trayecto, entonces asigna un coste de 0,00 € (incluido en el abono).
  - Dado un viaje cuya duración supere los 30 minutos, cuando se procesa la devolución, entonces se tarifica y cobra el tramo de tiempo que exceda de los 30 minutos iniciales.

#### HU-04: Consulta de historial de viajes y desglose mensual
- Como: Usuario del servicio de bicicletas.
- Quiero: Consultar mi histórico de trayectos realizados y aplicar filtros por mes y año.
- Para: Verificar los importes cobrados, ver mis rutas pasadas y controlar mis gastos.
- Criterios de Aceptación:
  - Dado que un usuario entra en la sección "Mis Viajes", cuando la vista se renderiza, entonces se muestra una lista cronológica con estación origen, estación destino, duración y coste individual.
  - Dado el historial de trayectos, cuando el usuario aplica un filtro de mes y año específicos, entonces la lista exhibe únicamente los viajes comprendidos en ese rango junto al sumatorio del gasto.
  - Dado un viaje que no generó cargos por durar menos de 30 minutos, cuando se muestra en el historial, entonces figura claramente identificado con el concepto "Tramo gratuito (0,00 €)".

#### HU-05: Notificación de averías e inhabilitación preventiva
- Como: Usuario que identifica una deficiencia en una bicicleta.
- Quiero: Notificar el fallo técnico (frenos, cadena, neumático) directamente desde la aplicación.
- Para: Prevenir que otros usuarios alquilen una unidad defectuosa y notificar al equipo técnico.
- Criterios de Aceptación:
  - Dado un usuario que reporta una avería en una estación, cuando envía el formulario con la descripción del problema, entonces la bicicleta cambia su estado a "Averiada".
  - Dado que una bicicleta se marca como "Averiada", cuando se actualiza el mapa global y la lista de anclajes, entonces la bicicleta queda inhabilitada para el alquiler y deja de sumar al conteo de bicicletas operativas.
  - Dado el reporte de una avería, cuando la petición es procesada en Go, entonces se emite un evento que actualiza el panel de incidencias pendientes para los operarios.

#### HU-06: Registro de intervenciones técnicas y reparación en taller
- Como: Operario de mantenimiento de taller.
- Quiero: Registrar los partes de reparación ejecutados en cada bicicleta (avería detectada, piezas sustituidas y datos del técnico).
- Para: Conservar el historial de mantenimiento de cada unidad y reactivar solo bicicletas puestas a punto.
- Criterios de Aceptación:
  - Dado un operario en el taller, cuando consulta la ficha de una bicicleta averiada, entonces visualiza el reporte original introducido por el usuario.
  - Dado que un operario concluye una reparación, cuando cumplimenta el formulario técnico (piezas cambiadas, notas) y marca la unidad como "Operativa", entonces se guarda la intervención firmada por el operario en PostgreSQL.
  - Dado que una bicicleta es reparada y devuelta a un anclaje, cuando se confirma su ubicación, entonces vuelve a figurar como disponible para los usuarios en el mapa.

#### HU-07: Monitorización de desequilibrios y alertas de balanceo de flota
- Como: Operario de transporte y calle.
- Quiero: Recibir alertas de estaciones saturadas (>90% ocupación) o vacías (<10% ocupación) ordenadas por severidad.
- Para: Planificar las rutas con la furgoneta y trasladar bicicletas desde puntos llenos a puntos vacíos.
- Criterios de Aceptación:
  - Dado el estado de ocupación de la red, cuando una estación cae por debajo del 10% o supera el 90% de su capacidad, entonces el sistema genera una alerta automática de prioridad alta.
  - Dado un operario que accede a su panel de logística, cuando abre la lista de balanceo, entonces visualiza las estaciones descompensadas ordenadas por urgencia de intervención.
  - Dado que un operario realiza la carga o descarga de un lote de bicicletas, cuando confirma la acción en su dispositivo, entonces la disponibilidad de ambas estaciones se recalcula de forma inmediata.

#### HU-08: Panel de analítica de uso y exportación de datos a CSV
- Como: Gestor municipal del servicio.
- Quiero: Analizar el uso de la red mediante gráficos por franja horaria y día de la semana, así como exportar los datos a CSV.
- Para: Evaluar el rendimiento del servicio, detectar cuellos de botella y justificar futuras ampliaciones ante el ayuntamiento.
- Criterios de Aceptación:
  - Dado el gestor en el área de administración, cuando establece un rango de fechas, entonces el sistema genera gráficos interactivos con el volumen de viajes distribuidos por hora y por día de la semana.
  - Dado el análisis de la red, cuando el gestor requiere procesar los datos con software externo, entonces la aplicación le permite descargar un archivo en formato CSV con el histórico filtrado.