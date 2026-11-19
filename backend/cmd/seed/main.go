// Comando seed: datos de prueba para los escenarios S1–S8 del quickstart y
// las demostraciones de las HU-03 (viaje activo), HU-04 (historial), HU-07
// (estaciones en extremos) y HU-08 (lote de analítica).
//
//	go run ./cmd/seed
//
// BORRA y recrea todos los datos (no el esquema). Contraseña de todos los
// usuarios: "password".
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
	casiVacia      // ~5 % de ocupación (S7 / HU-07)
	casiLlena      // ~95 % de ocupación (S7 / HU-07)
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
		if err := sembrarRed(ctx, tx, rnd); err != nil {
			return err
		}
		return sembrarViajesDemo(ctx, tx, hoy)
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
	fmt.Println("Demo: viaje activo de Ana (HU-03), 5 viajes finalizados de Ana (HU-04),")
	fmt.Println("      40 viajes demo de Carla (HU-08), estaciones al 5 % y 95 % (HU-07).")
}

var usuarios = []struct {
	nombre, email, rol, nota string
	abono                    bool
	diasVencimiento          int // relativo a hoy; 0 = sin fecha
}{
	{"Ana García", "ana@si16.test", "usuario", "abono vigente, viaje activo (HU-03)", true, 200},
	{"Bruno Caducado", "bruno@si16.test", "usuario", "abono caducado", true, -10},
	{"Carla Sin Abono", "carla@si16.test", "usuario", "sin abono (viajes demo HU-08)", false, 0},
	{"Diego Expirado", "diego@si16.test", "usuario", "abono caducado (activo con fecha pasada)", true, -10},
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

// sembrarViajesDemo añade los datos de demostración de las HU-03/HU-04/HU-08.
// Estaciones: tras TRUNCATE RESTART IDENTITY, los ids son 1..60 en el orden de
// `lugares` (1 = Plaça de l'Ajuntament, 2 = Estació del Nord, 3 = Plaça de la
// Reina; ninguna de las tres tiene perfil forzado).
func sembrarViajesDemo(ctx context.Context, tx pgx.Tx, hoy time.Time) error {
	ids := map[string]int64{}
	for _, email := range []string{"ana@si16.test", "carla@si16.test"} {
		var id int64
		if err := tx.QueryRow(ctx, `SELECT id FROM usuarios WHERE email = $1`, email).Scan(&id); err != nil {
			return fmt.Errorf("usuario %s: %w", email, err)
		}
		ids[email] = id
	}
	ana, carla := ids["ana@si16.test"], ids["carla@si16.test"]
	est1, est2, est3 := int64(1), int64(2), int64(3)

	crearBici := func(codigo, estado string) (int64, error) {
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO bicicletas (codigo, estado) VALUES ($1,$2) RETURNING id`,
			codigo, estado).Scan(&id)
		return id, err
	}

	// ---- HU-03: viaje ACTIVO de Ana (para probar la devolución) ----
	bActiva, err := crearBici("BICI-9001", "en_uso")
	if err != nil {
		return fmt.Errorf("bicicleta BICI-9001: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id, fecha_inicio)
		 VALUES ($1,$2,$3, now() - interval '15 minutes')`, ana, bActiva, est1); err != nil {
		return fmt.Errorf("viaje activo de Ana: %w", err)
	}

	// ---- HU-04: 5 viajes finalizados de Ana (research D8) ----
	// Repartidos en ≥ 2 meses (actual, anterior y anteactual): uno con coste
	// 0,00 (etiqueta de tramo gratuito), otros con coste > 0. Los costes siguen
	// la tarifa HU-03 (0,20 €/min tras los 30 primeros minutos gratis) para que
	// duración y precio no se contradigan. Día 15 mediodía UTC = tarde en
	// Europe/Madrid (evita bordes de zona horaria y de fin de mes). Cumple el
	// CHECK de 'finalizado': fecha_fin, estacion_destino_id y duracion_minutos
	// NOT NULL. Suma total 14,00 € (5,00 € en el mes en curso).
	bHist, err := crearBici("BICI-9002", "disponible")
	if err != nil {
		return fmt.Errorf("bicicleta BICI-9002: %w", err)
	}
	filas := []struct {
		mes     int // 0 = mes actual, -1 = anterior, -2 = anteactual
		dia     int
		minutos int
		coste   string
	}{
		{0, 5, 12, "0.00"},   // ≤ 30 min → tramo gratuito → etiqueta
		{0, 10, 40, "2.00"},  // 10 min excedente × 0,20
		{0, 20, 45, "3.00"},  // 15 min excedente × 0,20
		{-1, 15, 50, "4.00"}, // mes anterior: 20 min excedente × 0,20
		{-2, 15, 55, "5.00"}, // dos meses atrás: 25 min excedente × 0,20
	}
	for i, f := range filas {
		// Ningún viaje finalizado puede quedar en el futuro: falsearía el
		// historial y la analítica, y la devolución lo tomaría por un cierre
		// recién hecho (VIAJE_YA_CERRADO). En el mes en curso el día se limita
		// a hoy y, si aún no ha llegado la hora, el viaje se adelanta a ahora.
		dia := f.dia
		if f.mes == 0 && dia > hoy.Day() {
			dia = hoy.Day()
		}
		duracion := time.Duration(f.minutos) * time.Minute
		inicio := time.Date(hoy.Year(), hoy.Month()+time.Month(f.mes), dia,
			12, 0, 0, 0, time.UTC)
		if limite := hoy.Add(-time.Duration(len(filas)-i) * time.Minute); inicio.Add(duracion).After(limite) {
			inicio = limite.Add(-duracion)
		}
		fin := inicio.Add(duracion)
		destino := est2
		if i%2 == 1 {
			destino = est3
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id,
			                     estacion_destino_id, fecha_inicio, fecha_fin,
			                     duracion_minutos, coste_total, estado)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8::numeric, 'finalizado')`,
			ana, bHist, est1, destino, inicio, fin, f.minutos, f.coste); err != nil {
			return fmt.Errorf("viaje finalizado %d de Ana: %w", i+1, err)
		}
	}

	// ---- HU-08: 40 viajes demo para la analítica ----
	// research D7: repartidos deterministamente en las últimas 4 semanas (28
	// días), patrón de horas locales {7,8,9,12,17,18,20,22} y distintos días
	// de la semana, para que las 24 barras y los 7 días tengan forma
	// reconocible. Asignados a Carla para no alterar las cifras del historial
	// de Ana (HU-04). Duración y coste siguen la tarifa HU-03.
	bDemo, err := crearBici("BICI-9003", "disponible")
	if err != nil {
		return fmt.Errorf("bicicleta BICI-9003: %w", err)
	}
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		return fmt.Errorf("carga Europe/Madrid: %w", err)
	}
	horasDemo := []int{7, 8, 9, 12, 17, 18, 20, 22}
	for i := 0; i < 40; i++ {
		diasAtras := 1 + (i % 27)           // 1..27 días atrás (4 semanas)
		hora := horasDemo[i%len(horasDemo)] // horas locales repartidas
		dia := time.Now().In(madrid).AddDate(0, 0, -diasAtras)
		inicio := time.Date(dia.Year(), dia.Month(), dia.Day(), hora,
			(i*7)%60, 0, 0, madrid)
		minutos := 10 + (i*7)%60 // 10..69 min
		fin := inicio.Add(time.Duration(minutos) * time.Minute)
		coste := 0.0
		if minutos > 30 {
			coste = float64(minutos-30) * 0.20
		}
		destino := est2
		if i%2 == 1 {
			destino = est3
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO viajes (usuario_id, bicicleta_id, estacion_origen_id,
			                     estacion_destino_id, fecha_inicio, fecha_fin,
			                     duracion_minutos, coste_total, estado)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8::numeric, 'finalizado')`,
			carla, bDemo, est1, destino, inicio, fin, minutos,
			fmt.Sprintf("%.2f", coste)); err != nil {
			return fmt.Errorf("viaje demo HU-08 %d: %w", i+1, err)
		}
	}
	return nil
}
