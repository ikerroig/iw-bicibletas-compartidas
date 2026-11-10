// Exportación CSV de la HU-08 (T010): streaming con BOM UTF-8, coma, comillas
// RFC 4180, CRLF y las columnas del contrato R11. Reutiliza construirFiltro
// del servicio → el CSV del rango contiene exactamente los viajes que suman
// los gráficos (FR-006, research D3/D4).
package analitica

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// FilaCSV es una fila del fichero (también usada por los unitarios).
type FilaCSV struct {
	ID        int64
	FechaHora string // ya formateada en Europe/Madrid: AAAA-MM-DD HH:MM
	Origen    string
	Destino   string
	Duracion  int
	Coste     float64
}

// cabeceraCSV es la fila de cabecera fijada por la clarificación 3 (R11).
var cabeceraCSV = []string{
	"id", "fecha_hora", "estacion_origen", "estacion_destino",
	"duracion_minutos", "coste_eur",
}

// nombreCabecera devuelve la fila CSV ya serializada (coma, RFC 4180, CRLF).
func filaCampos(f FilaCSV) []string {
	return []string{
		strconv.FormatInt(f.ID, 10),
		f.FechaHora,
		f.Origen,
		f.Destino,
		strconv.Itoa(f.Duracion),
		fmt.Sprintf("%.2f", f.Coste), // decimal con PUNTO (R11/clarificación 3)
	}
}

// escribirCSV emite BOM + cabecera + filas a w (unit testable sin BD).
func escribirCSV(w io.Writer, filas []FilaCSV) error {
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.UseCRLF = true // finales de línea CRLF (clarificación 3)
	if err := cw.Write(cabeceraCSV); err != nil {
		return err
	}
	for _, f := range filas {
		if err := cw.Write(filaCampos(f)); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// NombreFichero → viajes_AAAAMMDD_AAAAMMDD.csv (fechas sin guiones).
func NombreFichero(desde, hasta string) string {
	return "viajes_" + strings.ReplaceAll(desde, "-", "") + "_" +
		strings.ReplaceAll(hasta, "-", "") + ".csv"
}

// ExportarCSV consulta en streaming las filas del rango (mismo
// construirFiltro que los agregados) y las emite fila a fila: el consumo de
// memoria no crece con el tamaño del rango (caso "varios años").
func (s *Service) ExportarCSV(ctx context.Context, p Parametros, w io.Writer) error {
	filtro, args := construirFiltro(p)

	rows, err := s.pool.Query(ctx,
		`SELECT v.id,
		        to_char(v.fecha_inicio AT TIME ZONE '`+zonaHoraria+`', 'YYYY-MM-DD HH24:MI') AS fecha_hora,
		        eo.nombre AS estacion_origen,
		        COALESCE(ed.nombre, '') AS estacion_destino,
		        v.duracion_minutos,
		        v.coste_total::float8
		   FROM viajes v
		   JOIN estaciones eo ON eo.id = v.estacion_origen_id
		   LEFT JOIN estaciones ed ON ed.id = v.estacion_destino_id
		  WHERE v.estado = 'finalizado'`+filtro+`
		  ORDER BY v.fecha_inicio ASC, v.id ASC`, args...)
	if err != nil {
		log.Printf("analitica: consulta CSV: %v", err)
		return err
	}
	defer rows.Close()

	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.UseCRLF = true
	if err := cw.Write(cabeceraCSV); err != nil {
		return err
	}
	for rows.Next() {
		var f FilaCSV
		if err := rows.Scan(&f.ID, &f.FechaHora, &f.Origen, &f.Destino,
			&f.Duracion, &f.Coste); err != nil {
			log.Printf("analitica: scan CSV: %v", err)
			return err
		}
		if err := cw.Write(filaCampos(f)); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("analitica: rows CSV: %v", err)
		return err
	}
	cw.Flush()
	return cw.Error()
}
