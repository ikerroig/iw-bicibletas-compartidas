# Contrato HTTP de HU-05

## POST `/api/v1/averias`

Requiere `Authorization: Bearer <token>` con rol `usuario`.

```json
{
  "bicicleta_codigo": "BICI-0001",
  "estacion_id": 1,
  "tipo_fallo": "frenos",
  "descripcion": "La maneta derecha no frena"
}
```

La respuesta inicial es `201`; un reintento de una incidencia abierta devuelve
`200`. En ambos casos el cuerpo tiene `incidencia`, `mensaje` y `repetido`.

## GET `/api/v1/incidencias`

Requiere el rol `operario_taller` y devuelve:

```json
{ "items": [
  {
    "id": 1,
    "bicicleta_id": 1,
    "bicicleta_codigo": "BICI-0001",
    "estacion_id": 1,
    "estacion_nombre": "Plaça de l'Ajuntament",
    "tipo_fallo": "frenos",
    "descripcion_reporte": "La maneta derecha no frena",
    "fecha_reporte": "2026-11-10T10:00:00Z",
    "estado": "pendiente"
  }
] }
```

Los reportes con descripción ausente o inválida reciben `422` con el envelope
de error común. La falta de sesión recibe `401` y el rol insuficiente `403`.

Cuando la bicicleta está anclada en otra estación, el reporte recibe `409` con
el código `BICICLETA_EN_OTRA_ESTACION`. En ese caso no se modifica la bicicleta
ni se crea la incidencia.
