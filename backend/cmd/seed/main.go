// Comando seed: datos de prueba para los escenarios S1–S8 del quickstart.
//
//	go run ./cmd/seed
//
// BORRA y recrea todos los datos (no el esquema). Es determinista: misma
// salida en cada ejecución. Contraseña de todos los usuarios: "password".
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"

	"si16/backend/internal/auth"
	"si16/backend/internal/cache"
	"si16/backend/internal/config"
	"si16/backend/internal/db"
	"si16/backend/internal/services/estacion"
)

// ~60 puntos reales de València (coordenadas aproximadas).
var lugares = []struct {
	nombre   string
	lat, lng float64
}{
	{"Plaça de l'Ajuntament", 39.4699, -0.3763}, {"Estació del Nord", 39.4664, -0.3772},
	{"Plaça de la Reina", 39.4744, -0.3754}, {"Mercat Central", 39.4735, -0.3790},
	{"Torres de Serrans", 39.4793, -0.3760}, {"Torres de Quart", 39.4757, -0.3834},
	{"Plaça de l'Àngel", 39.4777, -0.3787}, {"Colón", 39.4697, -0.3700},
	{"Porta de la Mar", 39.4718, -0.3683}, {"Plaça d'Amèrica", 39.4661, -0.3632},
	{"Gran Via Marqués del Túria", 39.4640, -0.3700}, {"Russafa", 39.4622, -0.3736},
	{"Mercat de Russafa", 39.4607, -0.3742}, {"Plaça de Canovas", 39.4646, -0.3638},
	{"Xàtiva - Bailén", 39.4668, -0.3806}, {"Pl. Espanya", 39.4661, -0.3866},
	{"Àngel Guimerà", 39.4683, -0.3895}, {"Plaça de Sant Agustí", 39.4677, -0.3790},
	{"Jardí Botànic", 39.4757, -0.3866}, {"Hospital General", 39.4683, -0.3943},
	{"Nou d'Octubre", 39.4682, -0.4009}, {"Campanar", 39.4826, -0.3986},
	{"Nou Campanar", 39.4859, -0.4053}, {"Bioparc", 39.4798, -0.4108},
	{"Beniferri", 39.4932, -0.4027}, {"Benicalap", 39.4945, -0.3906},
	{"Torrefiel", 39.4936, -0.3796}, {"Orriols", 39.4899, -0.3713},
	{"Pl. Alfons el Magnànim", 39.4728, -0.3725}, {"Museu de Belles Arts", 39.4787, -0.3707},
	{"Jardins del Real", 39.4806, -0.3662}, {"Pl. Sant Bertomeu", 39.4805, -0.3790},
	{"Avinguda de Blasco Ibáñez", 39.4762, -0.3570}, {"Facultats", 39.4772, -0.3613},
	{"UPV - Camí de Vera", 39.4810, -0.3410}, {"Benimaclet", 39.4867, -0.3592},
	{"Primat Reig", 39.4840, -0.3640}, {"Avinguda d'Aragó", 39.4700, -0.3600},
	{"Mestalla", 39.4746, -0.3583}, {"Ciutat de les Arts i les Ciències", 39.4549, -0.3507},
	{"Museu de les Ciències", 39.4580, -0.3530}, {"L'Oceanogràfic", 39.4528, -0.3486},
	{"Montolivet", 39.4566, -0.3620}, {"En Corts", 39.4565, -0.3720},
	{"Malilla", 39.4482, -0.3735}, {"Patraix", 39.4592, -0.3933},
	{"Safranar", 39.4549, -0.3902}, {"Jesús", 39.4599, -0.3857},
	{"Pl. Xúquer", 39.4741, -0.3557}, {"Ayora", 39.4683, -0.3478},
	{"El Grau", 39.4597, -0.3292}, {"Port de València", 39.4570, -0.3270},
	{"Platja de la Malva-rosa", 39.4776, -0.3247}, {"Platja de les Arenes", 39.4660, -0.3240},
	{"Cabanyal", 39.4700, -0.3310}, {"Natzaret", 39.4480, -0.3340},
	{"Tarongers", 39.4766, -0.3428}, {"Algirós", 39.4708, -0.3531},
	{"Sant Pau", 39.4710, -0.4045}, {"Pl. de Zaragoza", 39.4840, -0.3686},
}

type perfil int

const (
	normal perfil = iota
	fueraServicio
	sinBicis       // 0 disponibles (S1b)
	llena          // 0 anclajes libres (S1b)
	todasAveriadas // 0 disponibles y 0 libres: anclajes ocupados por averiadas (S1b)
	casiVacia      // ~5 % de ocupación (S7)
	casiLlena      // ~95 % de ocupación (S7)
	conAnclajeRoto // un anclaje fuera de servicio sin bici (SC-003)
)

// Perfiles forzados por índice de estación (el resto, normal).
var perfiles = map[int]perfil{
	3: fueraServicio, 41: fueraServicio,
	5: sinBicis, 12: llena, 18: todasAveriadas,
	7: casiVacia, 25: casiVacia, 9: casiLlena, 33: casiLlena,
	14: conAnclajeRoto,
}

func main() {
	cfg, err := config.Cargar()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx := context.Background()
	pool, err := db.Conectar(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	hash, err := auth.HashPassword("password")
	if err != nil {
		log.Fatal(err)
	}
	rnd := rand.New(rand.NewPCG(16, 2026))
	hoy := time.Now().UTC()

	err = db.EnTransaccion(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
			anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`); err != nil {
			return err
		}
		if err := sembrarUsuarios(ctx, tx, hash, hoy); err != nil {
			return err
		}
		return sembrarRed(ctx, tx, rnd)
	})
	if err != nil {
		log.Fatalf("seed: %v", err)
	}

	// Poblar Redis para que el mapa esté listo sin esperar al worker.
	rc := cache.Nuevo(cfg.RedisAddr, 0)
	defer func() { _ = rc.Cerrar() }()
	if n, err := estacion.Nuevo(pool, rc, nil).RecalcularTodas(ctx); err != nil {
		log.Printf("aviso: Redis no poblado (%v); el worker lo hará al arrancar el servidor", err)
	} else {
		log.Printf("Redis: %d estaciones publicadas", n)
	}

	fmt.Println("Seed completado. Usuarios (contraseña: password):")
	for _, u := range usuarios {
		fmt.Printf("  %-26s %-19s %s\n", u.email, u.rol, u.nota)
	}
}

var usuarios = []struct {
	nombre, email, rol, nota string
	abono                    bool
	diasVencimiento          int // relativo a hoy; 0 = sin fecha
}{
	{"Ana García", "ana@si16.test", "usuario", "abono vigente", true, 200},
	{"Bruno Caducado", "bruno@si16.test", "usuario", "abono caducado", true, -10},
	{"Carla Sin Abono", "carla@si16.test", "usuario", "sin abono", false, 0},
	{"Tomás Taller", "taller@si16.test", "operario_taller", "", true, 365},
	{"Laura Logística", "logistica@si16.test", "operario_logistica", "", true, 365},
	{"Gema Gestora", "gestor@si16.test", "gestor", "", true, 365},
}

func sembrarUsuarios(ctx context.Context, tx pgx.Tx, hash string, hoy time.Time) error {
	for _, u := range usuarios {
		var venc *time.Time
		if u.diasVencimiento != 0 {
			v := hoy.AddDate(0, 0, u.diasVencimiento)
			venc = &v
		}
		if _, err := tx.Exec(ctx, `INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
			VALUES ($1,$2,$3,$4,$5,$6)`, u.nombre, u.email, hash, u.rol, u.abono, venc); err != nil {
			return fmt.Errorf("usuario %s: %w", u.email, err)
		}
	}
	return nil
}

func sembrarRed(ctx context.Context, tx pgx.Tx, rnd *rand.Rand) error {
	numBici := 0
	nuevaBici := func(estado string) (int64, error) {
		numBici++
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO bicicletas (codigo, estado) VALUES ($1,$2) RETURNING id`,
			fmt.Sprintf("BICI-%04d", numBici), estado).Scan(&id)
		return id, err
	}

	for i, l := range lugares {
		p := perfiles[i]
		capacidad := 10 + rnd.IntN(4)*5 // 10, 15, 20 o 25
		estadoEst := "operativa"
		if p == fueraServicio {
			estadoEst = "fuera_de_servicio"
		}
		var estID int64
		if err := tx.QueryRow(ctx, `INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total, estado)
			VALUES ($1,$2,$3,$4,$5) RETURNING id`, l.nombre, l.lat, l.lng, capacidad, estadoEst).Scan(&estID); err != nil {
			return fmt.Errorf("estación %s: %w", l.nombre, err)
		}

		// Cuántas bicis operativas y averiadas lleva según el perfil.
		var operativas, averiadas, rotos int
		switch p {
		case sinBicis:
			averiadas = 1
		case llena:
			operativas, averiadas = capacidad-1, 1
		case todasAveriadas:
			averiadas = capacidad
		case casiVacia:
			operativas = capacidad / 20 // 0–1 bicis → < 10 %
		case casiLlena:
			operativas = capacidad - 1 // 95–96 % con capacidad 20/25
			if 100*operativas <= 90*capacidad {
				operativas = capacidad // con capacidad 10/15, solo 100 % supera el 90 %
			}
		case conAnclajeRoto:
			operativas, rotos = capacidad/2, 1
		case fueraServicio:
			operativas = capacidad / 3
		default:
			operativas = capacidad/4 + rnd.IntN(capacidad/2+1)
			if rnd.IntN(4) == 0 {
				averiadas = 1
			}
		}

		for n := 1; n <= capacidad; n++ {
			codigo := fmt.Sprintf("E%02d-A%02d", i+1, n)
			var estadoAnc string
			var bici *int64
			switch {
			case n <= operativas:
				id, err := nuevaBici("disponible")
				if err != nil {
					return err
				}
				estadoAnc, bici = "ocupado", &id
			case n <= operativas+averiadas:
				id, err := nuevaBici("averiada")
				if err != nil {
					return err
				}
				estadoAnc, bici = "ocupado", &id
			case n <= operativas+averiadas+rotos:
				estadoAnc = "fuera_de_servicio"
			default:
				estadoAnc = "libre"
			}
			if _, err := tx.Exec(ctx, `INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
				VALUES ($1,$2,$3,$4)`, estID, codigo, estadoAnc, bici); err != nil {
				return fmt.Errorf("anclaje %s: %w", codigo, err)
			}
		}
	}
	// Flota fuera de estaciones: reparadas pendientes de ubicar (HU-06) y en taller.
	for _, e := range []string{"disponible", "disponible", "en_taller", "en_taller"} {
		if _, err := nuevaBici(e); err != nil {
			return err
		}
	}
	log.Printf("Red: %d estaciones, %d bicicletas", len(lugares), numBici)
	return nil
}
