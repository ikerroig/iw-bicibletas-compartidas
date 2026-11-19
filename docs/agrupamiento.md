# Agrupamiento: reparto de historias de usuario

**Sistema asignado**: SI-16 · Servicio de Bicicletas Compartidas
**Repositorio**: <https://github.com/ikerroig/iw-bicibletas-compartidas>

El grupo tiene 4 integrantes y cada uno es responsable de 2 historias de
usuario. Responsable significa recorrer el ciclo completo de Spec Kit en su
spec (specify → clarify → plan → tasks → analyze → implement → converge) y
llevar el peso principal de la implementación, en una rama propia por spec.
El reparto se acordó al inicio y quedó registrado en
[`historia.md`](../historia.md) y en el
[`tasks.md` global](../specs/spec-001/tasks.md#mapa-de-stories--specs--responsables).

## Reparto

| Integrante | Historia de usuario | Spec y rama | Actor principal |
|---|---|---|---|
| **Iker Roig** | HU-01 — Consulta de disponibilidad de estaciones en tiempo real | [`spec-001`](../specs/spec-001/) | Usuario con abono |
| **Iker Roig** | HU-02 — Desbloqueo de bicicleta mediante código de anclaje | [`spec-002`](../specs/spec-002/) | Usuario con abono |
| **Alex Morell** | HU-03 — Devolución de bicicleta y cálculo de tarificación | [`spec-003`](../specs/spec-003/) | Usuario con abono |
| **Alex Morell** | HU-04 — Historial de viajes y desglose mensual | [`spec-004`](../specs/spec-004/) | Usuario con abono |
| **Josep Peiró** | HU-05 — Notificación de averías e inhabilitación preventiva | [`spec-005`](../specs/spec-005/) | Usuario con abono |
| **Josep Peiró** | HU-06 — Registro de intervenciones técnicas y reparación en taller | [`spec-006`](../specs/spec-006/) | Operario de taller |
| **Alberto Copa** | HU-07 — Monitorización de desequilibrios y alertas de balanceo | [`spec-007`](../specs/spec-007/) | Operario de logística |
| **Alberto Copa** | HU-08 — Panel de analítica de uso y exportación a CSV | [`spec-008`](../specs/spec-008/) | Gestor del servicio |

El texto completo de cada historia, con sus criterios de aceptación, está en
[`historia.md`](../historia.md) y en el `spec.md` de cada spec.

### Criterio del reparto

Las historias se agruparon por pares que comparten flujo y datos, para que
cada integrante pudiera razonar su pareja de principio a fin:

- **Iker** (HU-01 y HU-02): la base compartida (modelo de datos, autenticación,
  estado en vivo en Redis) y el inicio del viaje. HU-01 es el MVP del que
  dependen las demás.
- **Alex** (HU-03 y HU-04): el cierre del viaje con su tarifa y la consulta de
  esos mismos viajes en el historial.
- **Josep** (HU-05 y HU-06): el ciclo de vida de una avería, desde el reporte
  del usuario hasta la reparación y reubicación en el taller.
- **Alberto** (HU-07 y HU-08): las vistas operativas y de gestión sobre el
  estado de la red y el uso agregado.

## Identidades en Git y GitHub

Para la trazabilidad por commits, estas son las identidades con las que
aparece cada integrante:

| Integrante | Usuario de GitHub | Nombres de autor en los commits |
|---|---|---|
| Iker Roig | `ikerroig` | `Iker Roig`, `Iker`, `ikerroig` |
| Alex Morell | `amf162-ua` | `amf162-ua`, `Alex Morell Far` |
| Josep Peiró | `jpt35-ua` | `Josep Peiró`, `jpt35-ua` |
| Alberto Copa | `acf91-ua` | `acf91` |

## Trazabilidad

### Ramas y pull requests

| Spec | Rama | Integración en `main` |
|---|---|---|
| spec-001 | `spec-001` | [PR #1](https://github.com/ikerroig/iw-bicibletas-compartidas/pull/1) |
| spec-002 | `spec-002` | [PR #2](https://github.com/ikerroig/iw-bicibletas-compartidas/pull/2) |
| spec-003, spec-004 | `spec-003`, `spec-004` | merge `10ff0bf` (junto con spec-007 y spec-008) |
| spec-005 | `spec-005` | [PR #3](https://github.com/ikerroig/iw-bicibletas-compartidas/pull/3) |
| spec-006 | `spec-006` | [PR #4](https://github.com/ikerroig/iw-bicibletas-compartidas/pull/4) |
| spec-007, spec-008 | `spec-007`, `spec-008` | merge `10ff0bf` (junto con spec-003 y spec-004) |

Trabajo común posterior a la fusión de las 8 specs:

- [PR #5](https://github.com/ikerroig/iw-bicibletas-compartidas/pull/5): corrección de problemas de integración entre specs.
- [PR #6](https://github.com/ikerroig/iw-bicibletas-compartidas/pull/6): validación de punta a punta (T088) y README (T092).

### Commits por spec

Commits (sin merges) que modifican la carpeta de cada spec en `main`, a
fecha de 19-11-2026. El commit de Iker que aparece en todas las specs es la
base común inicial (constitución, 8 specs y plan compartido), creada en su
ordenador según la guía de la práctica.

| Spec | Responsable | Commits del responsable | Otros |
|---|---|---|---|
| spec-001 | Iker Roig | 11 | 1 de Alberto |
| spec-002 | Iker Roig | 4 | — |
| spec-003 | Alex Morell | 4 | 1 de Iker (base común) |
| spec-004 | Alex Morell | 2 | 1 de Iker (base común) |
| spec-005 | Josep Peiró | 1 | 1 de Iker (base común) |
| spec-006 | Josep Peiró | 1 | 1 de Iker (base común) |
| spec-007 | Alberto Copa | 4 | 1 de Iker (base común) |
| spec-008 | Alberto Copa | 3 | 1 de Iker (base común) |

Para comprobarlo:

```bash
git log main --no-merges --format='%an' -- specs/spec-00X | sort | uniq -c
```

## Partes compartidas

Hay piezas que no pertenecen a una sola historia y se acordaron entre todos:

- **Constitución** del proyecto: [`.specify/memory/constitution.md`](../.specify/memory/constitution.md).
- **Plan y modelo de datos comunes** a las 8 specs: [`specs/spec-001/plan.md`](../specs/spec-001/plan.md) y [`specs/spec-001/data-model.md`](../specs/spec-001/data-model.md). Cualquier cambio del esquema se acuerda con el grupo antes de añadir una migración.
- **Contrato OpenAPI** común: [`specs/spec-001/contracts/openapi.yaml`](../specs/spec-001/contracts/openapi.yaml).
- **Validación conjunta** de la fase de pulido (T088–T094 en el [`tasks.md` global](../specs/spec-001/tasks.md)).
