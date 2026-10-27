// Command seed — datos de prueba SI-16 (versión mínima para probar la HU-03
// de forma independiente; el seed global completo lo aporta spec-001).
//
// Idempotente: vacía y regenera los datos. Ejecutar con el esquema migrado.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"golang.org/x/crypto/bcrypt"

	"si16/backend/internal/config"
	"si16/backend/internal/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
	fmt.Println("Seed completado. Usuarios de prueba (contraseña: 'password'):")
	fmt.Println("  ana@si16.test     usuario, abono vigente, VIAJE ACTIVO (para probar devolución)")
	fmt.Println("  carla@si16.test   usuario, abono vigente, sin viaje")
	fmt.Println("  diego@si16.test   usuario, abono CADUCADO")
	fmt.Println("  taller@si16.test  operario_taller   | logistica@si16.test operario_logistica")
	fmt.Println("  gestor@si16.test  gestor")
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		// Regenerable: vaciar todo (orden por FKs) y reiniciar contadores.
		for _, q := range []string{
			`TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
			                anclajes, bicicletas, estaciones, usuarios
			  RESTART IDENTITY CASCADE`,
		} {
			if _, err := tx.Exec(ctx, q); err != nil {
				return fmt.Errorf("truncate: %w", err)
			}
		}

		// ---- Estaciones (E1 origen de Ana, E2 con anclajes libres) ----
		var e1, e2 int64
		if err := tx.QueryRow(ctx,
			`INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total)
			 VALUES ('Estació Universitat', 39.469900, -0.376300, 8),
			        ('Estació Mercat',      39.479000, -0.373000, 6)
			 RETURNING id`).Scan(&e1); err != nil {
			return fmt.Errorf("estaciones: %w", err)
		}
		// La segunda estación (e2): la id que no es e1 (INSERT múltiple solo
		// devuelve la primera; la recuperamos por nombre).
		if err := tx.QueryRow(ctx,
			`SELECT id FROM estaciones WHERE nombre = 'Estació Mercat'`).Scan(&e2); err != nil {
			return fmt.Errorf("estación 2: %w", err)
		}

		// ---- Anclajes: E1 → AN-00001..AN-00008, E2 → AN-00009..AN-00014 ----
		anclajes := map[int64][2]int64{} // estacion → (primera id, última id)
		for _, r := range []struct {
			estacion int64
			desde, n int
		}{{e1, 1, 8}, {e2, 9, 6}} {
			var first, last int64
			for i := 0; i < r.n; i++ {
				codigo := fmt.Sprintf("AN-%05d", r.desde+i)
				var id int64
				if err := tx.QueryRow(ctx,
					`INSERT INTO anclajes (estacion_id, codigo_anclaje)
					 VALUES ($1, $2) RETURNING id`, r.estacion, codigo).Scan(&id); err != nil {
					return fmt.Errorf("anclaje %s: %w", codigo, err)
				}
				if i == 0 {
					first = id
				}
				last = id
			}
			anclajes[r.estacion] = [2]int64{first, last}
		}

		// ---- Bicicletas ancladas ----
		bicis := map[string]int64{} // codigo → id
		crearBici := func(codigo, estado string, anclajeID *int64) (int64, error) {
			var id int64
			var ancl any
			if anclajeID != nil {
				ancl = *anclajeID
			}
			err := tx.QueryRow(ctx,
				`INSERT INTO bicicletas (codigo, estado) VALUES ($1, $2) RETURNING id`,
				codigo, estado).Scan(&id)
			if err != nil {
				return 0, err
			}
			if ancl != nil {
				if _, err := tx.Exec(ctx,
					`UPDATE anclajes SET estado = 'ocupado', bicicleta_id = $1 WHERE id = $2`,
					id, ancl); err != nil {
					return 0, err
				}
			}
			return id, nil
		}

		a1 := anclajes[e1][0]
		a2 := a1 + 1
		a3 := a1 + 2
		a4 := a1 + 3
		// E2: anclajes 9..14 → primeros libres (para devolver), luego ocupados.
		b9 := anclajes[e2][0] + 1 // AN-00010
		b10 := anclajes[e2][0] + 2

		for _, b := range []struct {
			codigo, estado string
			anclaje        *int64
		}{
			{"B-001", "disponible", &a1},
			{"B-002", "disponible", &a2},
			{"B-003", "disponible", &a3},
			{"B-004", "averiada", &a4}, // anclada pero averiada (no cuenta)
			{"B-005", "disponible", &b9},
			{"B-006", "disponible", &b10},
		} {
			id, err := crearBici(b.codigo, b.estado, b.anclaje)
			if err != nil {
				return fmt.Errorf("bicicleta %s: %w", b.codigo, err)
			}
			bicis[b.codigo] = id
		}

		// ---- Usuarios de todos los roles ----
		hoy := time.Now()
		vigente := hoy.AddDate(1, 0, 0)
		caducada := hoy.AddDate(-1, 0, 0)
		usuarios := []struct {
			email, nombre, rol string
			abono              bool
			venc               *time.Time
		}{
			{"ana@si16.test", "Ana", "usuario", true, &vigente},
			{"carla@si16.test", "Carla", "usuario", true, &vigente},
			{"diego@si16.test", "Diego", "usuario", true, &caducada}, // activo con fecha pasada = caducado
			{"taller@si16.test", "Taller", "operario_taller", false, nil},
			{"logistica@si16.test", "Logística", "operario_logistica", false, nil},
			{"gestor@si16.test", "Gestor", "gestor", false, nil},
		}
		ids := map[string]int64{}
		for _, u := range usuarios {
			var id int64
			if err := tx.QueryRow(ctx,
				`INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
				 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
				u.nombre, u.email, string(hash), u.rol, u.abono, u.venc).Scan(&id); err != nil {
				return fmt.Errorf("usuario %s: %w", u.email, err)
			}
			ids[u.email] = id
		}

		// ---- Viaje ACTIVO de Ana (B-007 en uso, origen E1) para la HU-03 ----
		b7, err := crearBici("B-007", "en_uso", nil)
		if err != nil {
			return fmt.Errorf("bicicleta B-007: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id, fecha_inicio)
			 VALUES ($1, $2, $3, now() - interval '15 minutes')`,
			ids["ana@si16.test"], b7, e1); err != nil {
			return fmt.Errorf("viaje activo de Ana: %w", err)
		}

		return nil
	})
}
