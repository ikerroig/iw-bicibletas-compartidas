# Contrato HTTP de HU-06

Todos los endpoints requieren un JWT con rol operario_taller.

## GET /api/v1/taller/bicicletas/{id}

Devuelve la ficha de la bicicleta y sus incidencias/intervenciones ordenadas
por fecha descendente:

    {
      "bicicleta": { "id": 10, "codigo": "BICI-0010", "estado": "averiada" },
      "incidencias": [
        {
          "id": 4,
          "bicicleta_id": 10,
          "descripcion_reporte": "La maneta no frena",
          "tipo_fallo": "frenos",
          "estado": "pendiente",
          "fecha_reporte": "2026-11-16T10:00:00Z"
        }
      ]
    }

Una bicicleta inexistente responde 404 BICICLETA_NO_ENCONTRADA.

## POST /api/v1/taller/bicicletas/{id}/intervenciones

Body:

    {
      "piezas_sustituidas": "Maneta derecha",
      "notas_taller": "Frenos ajustados",
      "marcar_operativa": true
    }

Debe incluir al menos una pieza o una nota. La API guarda el operario tomado
del JWT y la fecha del servidor. Si marcar_operativa es true, cierra las
incidencias abiertas, cambia la bicicleta a disponible y la deja sin anclaje
hasta la confirmación de ubicación. Responde 201 con la incidencia actualizada.

Sin incidencia abierta se crea un mantenimiento preventivo de tipo otro, con
descripción Mantenimiento preventivo.

Los datos incompletos responden 422 ENTRADA_INVALIDA. Una bicicleta en viaje
responde 409 BICICLETA_EN_USO.

## POST /api/v1/taller/bicicletas/{id}/ubicacion

Body:

    { "codigo_anclaje": "T06-A02" }

Solo acepta una bicicleta disponible sin anclaje y un anclaje libre de una
estación operativa. Responde 200 con el estado recalculado de la estación.
Los errores de anclaje ocupado, fuera de servicio o bicicleta no reparada
responden 409; un anclaje inexistente responde 404.
