<!-- Sync Impact Report
Version change: 1.0.0 → 1.0.1
Bump type: PATCH (nombres reales de integrantes; sin cambios de alcance ni
  redefinición de principios)
Modified principles: none (solo contenido de sección, sin tocar principios)
Added sections: none
Removed sections: none
Modified content:
  - Sección "Organización del Equipo y Reparto de Historias": etiquetas
    "Integrante A–D" → nombres reales: Iker Roig (spec-001/002),
    Alex Morell (spec-003/004), Josep Peiró (spec-005/006),
    Alberto Copa (spec-007/008).
Follow-up TODOs: ninguno. Los TODOs de la ratificación inicial quedan
  resueltos: nombres de integrantes aplicados; typos de HU-05 en historia.md
  corregidos ("Petición" → "petición", "y inhabilitación" → "e inhabilitación");
  alcance aclarado por el usuario (abono anual = atributo preexistente del
  usuario; "rutas pasadas" = estaciones origen/destino sin GPS; sin HU nuevas).
-->
# SI-16 Constitution
<!-- Sistema de Información SI-16 – Servicio Público de Bicicletas Compartidas -->

## Core Principles

### I. Autoridad del Backend en Reglas de Negocio (NON-NEGOTIABLE)

Todas las reglas de negocio críticas del sistema MUST validarse obligatoriamente en
el backend (Go/Gin); las comprobaciones del frontend son solo complemento de usabilidad
y nunca son fuente de verdad:

- Abono anual activo: solo un usuario con abono vigente MUST poder desbloquear
  bicicletas; un abono caducado MUST ser rechazado con mensaje de causa.
- Prohibición de viajes simultáneos: un usuario con un viaje activo MUST ser denegado
  en cualquier segundo desbloqueo.
- Primeros 30 minutos gratuitos: viajes de duración ≤ 30 min tarifican 0,00 €;
  solo el excedente de los 30 minutos iniciales MUST cobrarse.
- Bloqueo inmediato por avería: una bicicleta reportada como averiada MUST quedar
  inhabilitada para el alquiler de inmediato y fuera del conteo de unidades operativas.
- Alertas de ocupación: el sistema MUST generar una alerta de prioridad alta cuando
  una estación queda por debajo del 10 % o por encima del 90 % de su capacidad.

Rationale: el cliente es manipulable; solo la validación server-side garantiza
integridad económica y seguridad, y es la única versión testeable de cada regla.

### II. Pila Tecnológica Acordada

- Frontend: SolidJS + Tailwind CSS.
- Backend: Go con framework Gin.
- Base de datos principal: PostgreSQL.
- Estado en tiempo real: Redis, con refresco máximo de 30 segundos para la
  disponibilidad de estaciones y anclajes.

Cualquier sustitución o incorporación de tecnología MUST ser acordada por todo el
equipo mediante enmienda a esta constitución ANTES de su uso en una spec.

Rationale: un stack uniforme permite revisión cruzada entre las 8 especificaciones
e elimina fricciones de integración entre ramas.

### III. Arquitectura Separada con Modelo de Datos Compartido

- El código se organiza en `/backend` (Go/Gin) y `/frontend` (SolidJS); ninguna
  funcionalidad vive fuera de estas dos raíces.
- El modelo de datos en PostgreSQL y la estructura de claves de Redis son únicos y
  compartidos por las 8 especificaciones.
- Cualquier cambio en el esquema compartido (migraciones PostgreSQL, nuevas claves o
  cambios de formato en Redis) MUST contar con acuerdo del equipo ANTES de aplicarse.

Rationale: las 8 historias operan sobre entidades compartidas (estaciones, anclajes,
bicicletas, viajes); esquemas divergentes romperían la integración entre specs.

### IV. Disponibilidad en Tiempo Real con Frescura ≤ 30 s

La disponibilidad de estaciones y anclajes mostrada al usuario MUST reflejar los
cambios del servidor con un retardo máximo de 30 segundos, sirviendo el estado
desde Redis. El frontend MUST consumir ese estado en vivo; queda prohibido mostrar
disponibilidad proveniente de datos estáticos sin expiración.

Rationale: la planificación de origen y destino (HU-01) solo es útil si el mapa es
fiable dentro de la ventana acordada; la expiración corta de Redis la garantiza.

### V. Independencia por Especificación

El equipo está formado por 4 integrantes; cada integrante es responsable de
2 especificaciones (8 historias de usuario en total, definidas en `historia.md`).
Cada especificación MUST poder desarrollarse, probarse y fusionarse sin esperar a
las demás, con responsabilidad clara de su integrante.

Rationale: el reparto 4 × 2 con una rama por especificación habilita trabajo
paralelo independiente y evita cuellos de botella entre personas.

## Organización del Equipo y Reparto de Historias

`historia.md` es la fuente única de verdad de las 8 historias de usuario
(HU-01 … HU-08). Reparto acordado; cada integrante cubre un dominio cohesivo con
sus 2 especificaciones:

| Integrante | Ramas / Specs | Historias de usuario (dominio) |
| --- | --- | --- |
| Iker Roig | spec-001, spec-002 | HU-01 Consulta de disponibilidad en tiempo real; HU-02 Desbloqueo por código de anclaje (ciclo de alquiler) |
| Alex Morell | spec-003, spec-004 | HU-03 Devolución y tarificación; HU-04 Historial de viajes y desglose mensual (cierre y facturación) |
| Josep Peiró | spec-005, spec-006 | HU-05 Notificación de averías e inhabilitación; HU-06 Intervenciones técnicas y taller (ciclo de vida de la bicicleta) |
| Alberto Copa | spec-007, spec-008 | HU-07 Alertas de balanceo de flota; HU-08 Analítica de uso y exportación CSV (operación y reporting) |

- El emparejamiento minimiza dependencias: A inicia viajes, B los cierra y factura,
  C gestiona el estado físico de la flota, D consume datos agregados.
- Toda HU nueva o cambio de alcance MUST reasignarse por acuerdo del equipo y
  reflejarse en `historia.md` antes de abrir su rama.

## Acuerdos de Colaboración y Calidad

- Rama por especificación: cada spec trabaja en su propia rama `spec-001` …
  `spec-008`, creada desde `main`. Está prohibido integrar directamente en `main`.
- Todo cambio se fusiona mediante Pull Request revisado por, al menos, otro
  integrante del equipo distinto del autor.
- Gate de calidad: antes de fusionar una spec, las reglas de negocio críticas que
  cubre (principio I) MUST estar implementadas en el backend acompañadas de pruebas
  automatizadas que demuestren su cumplimiento y el rechazo de los casos inválidos.
- Esquema compartido: todo cambio en PostgreSQL/Redis MUST comunicarse al equipo y
  contar con su acuerdo antes de aplicar la migración; las specs que compartan
  entidades MUST coordinar sus migraciones para no romper ramas ajenas.
- Deriva de alcance: `historia.md` y esta constitución se revisan en cada fusión de
  spec; cualquier discrepancia se resuelve por acuerdo del equipo antes de continuar.

## Governance

- Esta constitución es el documento rector del SI-16 y prevalece sobre prácticas
  locales de cada integrante en todos los ámbitos que cubre.
- Enmiendas: cualquier cambio MUST proponerse al equipo, obtener su aprobación por
  consenso, aplicarse en este archivo con la versión y fecha ISO correspondientes y
  registrarse su motivo en el reporte de sincronización de la enmienda.
- Política de versionado (SemVer): MAJOR = eliminación o redefinición de principios
  existentes; MINOR = principio o sección nueva, o expansión material de alcance;
  PATCH = clarificaciones, redacción y correcciones no semánticas.
- Cumplimiento: cada PR y su revisión MUST verificarse contra los principios antes
  de fusionarse; cualquier violación bloquea la fusión. El reparto de historias, el
  estado de las ramas `spec-*` y el acuerdo de cambios de esquema se revisan en cada
  integración como revisión de cumplimiento obligatoria.
- Los cambios de esquema compartido requieren acuerdo del equipo (principio III);
  si alteran el contrato de datos entre specs, constituyen una enmienda MINOR.

**Version**: 1.0.1 | **Ratified**: 2026-10-20 | **Last Amended**: 2026-10-20
