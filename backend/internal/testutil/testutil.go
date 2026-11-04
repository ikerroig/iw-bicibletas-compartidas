// Package testutil prepara el entorno de los tests de integración y contrato:
// una base de datos PostgreSQL y una base lógica de Redis PROPIAS de cada
// paquete de tests (go test ./... ejecuta paquetes en paralelo), con el
// esquema de backend/migrations aplicado.
//
// Requiere `docker compose up -d`. Variables (ver .env.example):
// TEST_DATABASE_URL (BD base; a su nombre se le añade "_<paquete>") y REDIS_ADDR.
package testutil

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"si16/backend/internal/auth"
	"si16/backend/internal/cache"
	"si16/backend/internal/config"
	"si16/backend/internal/db"
	"si16/backend/internal/hooks"
	"si16/backend/internal/models"
	"si16/backend/internal/routes"
	"si16/backend/internal/services/estacion"
)

// Entorno de un paquete de tests.
type Entorno struct {
	Ctx       context.Context
	Pool      *pgxpool.Pool
	Cache     *cache.Cliente
	Estacion  *estacion.Servicio
	Emisor    *auth.Emisor
	Router    *gin.Engine
	RedisAddr string
	RedisDB   int
	DBURL     string
}

var (
	once    sync.Once
	entorno *Entorno
	errIni  error
)

// Iniciar prepara (una vez por paquete) la BD "<base>_<paquete>" y la base
// lógica de Redis indicada, y limpia los datos antes de cada test.
func Iniciar(t *testing.T, paquete string, redisDB int) *Entorno {
	t.Helper()
	once.Do(func() { entorno, errIni = preparar(paquete, redisDB) })
	if errIni != nil {
		t.Fatalf("entorno de integración no disponible (¿`docker compose up -d`?): %v", errIni)
	}
	entorno.Limpiar(t)
	return entorno
}

func preparar(paquete string, redisDB int) (*Entorno, error) {
	config.CargarDotenv()
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		base = "postgres://si16_user:si16_password@localhost:5432/si16_test?sslmode=disable"
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	nombreBD := strings.TrimPrefix(u.Path, "/") + "_" + paquete
	if err := crearBDSiFalta(ctx, u, nombreBD); err != nil {
		return nil, err
	}
	u.Path = "/" + nombreBD
	dbURL := u.String()

	if err := db.Migrar(ctx, dbURL, "up"); err != nil {
		return nil, fmt.Errorf("migrar %s: %w", nombreBD, err)
	}
	pool, err := db.Conectar(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rc := cache.Nuevo(addr, redisDB)
	if err := rc.Ping(ctx); err != nil {
		return nil, fmt.Errorf("redis %s: %w", addr, err)
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	e := &Entorno{
		Ctx: ctx, Pool: pool, Cache: rc, RedisAddr: addr, RedisDB: redisDB, DBURL: dbURL,
		Estacion: estacion.Nuevo(pool, rc, log),
		Emisor:   auth.NuevoEmisor("secreto-de-tests"),
	}
	e.Router = routes.NuevoRouter(routes.Deps{
		Cfg: config.Config{TarifaMinuto: 0.20}, Pool: pool, Cache: rc,
		Auth: e.Emisor, Estacion: e.Estacion, Log: log,
	})
	return e, nil
}

func crearBDSiFalta(ctx context.Context, u *url.URL, nombre string) error {
	admin := *u
	admin.Path = "/postgres"
	conn, err := pgx.Connect(ctx, admin.String())
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(ctx) }()
	var existe bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, nombre).Scan(&existe); err != nil {
		return err
	}
	if existe {
		return nil
	}
	_, err = conn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{nombre}.Sanitize())
	return err
}

// Limpiar vacía todas las tablas, la base lógica de Redis y los suscriptores.
func (e *Entorno) Limpiar(t *testing.T) {
	t.Helper()
	if _, err := e.Pool.Exec(e.Ctx, `TRUNCATE alertas_balanceo, incidencias_mantenimientos, viajes,
		anclajes, bicicletas, estaciones, usuarios RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	if err := e.Cache.Redis().FlushDB(e.Ctx).Err(); err != nil {
		t.Fatalf("flushdb: %v", err)
	}
	hooks.ReiniciarParaTests()
}

// Anclaje describe un anclaje de fixture.
type Anclaje struct {
	Estado    string // libre | ocupado | fuera_de_servicio
	Bicicleta string // "" (sin bici) | disponible | averiada | en_taller
}

// Atajos de fixture.
var (
	Libre       = Anclaje{Estado: models.AnclajeLibre}
	ConBici     = Anclaje{Estado: models.AnclajeOcupado, Bicicleta: models.BicicletaDisponible}
	ConAveriada = Anclaje{Estado: models.AnclajeOcupado, Bicicleta: models.BicicletaAveriada}
	ConEnTaller = Anclaje{Estado: models.AnclajeOcupado, Bicicleta: models.BicicletaEnTaller}
	Roto        = Anclaje{Estado: models.AnclajeFueraDeServicio}
)

// Repetir devuelve n copias de a.
func Repetir(n int, a Anclaje) []Anclaje {
	out := make([]Anclaje, n)
	for i := range out {
		out[i] = a
	}
	return out
}

var secuencia int

// CrearEstacion inserta una estación con capacidad = len(anclajes).
func (e *Entorno) CrearEstacion(t *testing.T, nombre, estado string, anclajes ...Anclaje) int64 {
	t.Helper()
	secuencia++
	var id int64
	err := e.Pool.QueryRow(e.Ctx, `INSERT INTO estaciones (nombre, latitud, longitud, capacidad_total, estado)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		nombre, 39.47+float64(secuencia)/1000, -0.37, len(anclajes), estado).Scan(&id)
	if err != nil {
		t.Fatalf("crear estación: %v", err)
	}
	for i, a := range anclajes {
		var bici *int64
		if a.Bicicleta != "" {
			var b int64
			if err := e.Pool.QueryRow(e.Ctx, `INSERT INTO bicicletas (codigo, estado) VALUES ($1, $2) RETURNING id`,
				fmt.Sprintf("B-%d-%d", id, i), a.Bicicleta).Scan(&b); err != nil {
				t.Fatalf("crear bici: %v", err)
			}
			bici = &b
		}
		if _, err := e.Pool.Exec(e.Ctx, `INSERT INTO anclajes (estacion_id, codigo_anclaje, estado, bicicleta_id)
			VALUES ($1, $2, $3, $4)`, id, fmt.Sprintf("E%d-A%02d", id, i+1), a.Estado, bici); err != nil {
			t.Fatalf("crear anclaje: %v", err)
		}
	}
	return id
}

// Exec ejecuta SQL directo (simula una mutación sin pasar por la API).
func (e *Entorno) Exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := e.Pool.Exec(e.Ctx, sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

// Token crea un usuario con el rol dado y devuelve su JWT.
func (e *Entorno) Token(t *testing.T, rol models.Rol) string {
	t.Helper()
	secuencia++
	hash, _ := auth.HashPassword("password")
	u := models.Usuario{Nombre: "Test " + string(rol), Rol: rol}
	err := e.Pool.QueryRow(e.Ctx, `INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo)
		VALUES ($1, $2, $3, $4, true) RETURNING id`,
		u.Nombre, fmt.Sprintf("u%d@si16.test", secuencia), hash, rol).Scan(&u.ID)
	if err != nil {
		t.Fatalf("crear usuario: %v", err)
	}
	tok, err := e.Emisor.Emitir(u)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// Usuario crea un usuario de rol `usuario` con el abono indicado (vencimiento
// nil = sin fecha) y devuelve su id y su JWT.
func (e *Entorno) Usuario(t *testing.T, activo bool, vencimiento *time.Time) (int64, string) {
	t.Helper()
	secuencia++
	u := models.Usuario{Nombre: "Abonado", Rol: models.RolUsuario}
	err := e.Pool.QueryRow(e.Ctx, `INSERT INTO usuarios (nombre, email, password_hash, rol, abono_activo, fecha_vencimiento_abono)
		VALUES ($1, $2, 'x', $3, $4, $5) RETURNING id`,
		u.Nombre, fmt.Sprintf("a%d@si16.test", secuencia), u.Rol, activo, vencimiento).Scan(&u.ID)
	if err != nil {
		t.Fatalf("crear usuario: %v", err)
	}
	tok, err := e.Emisor.Emitir(u)
	if err != nil {
		t.Fatal(err)
	}
	return u.ID, tok
}

// Post hace una petición POST JSON autenticada (token vacío = sin cabecera).
func (e *Entorno) Post(t *testing.T, ruta, token, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/v1"+ruta, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.Router.ServeHTTP(rec, req)
	return rec
}

// Get hace una petición GET autenticada (token vacío = sin cabecera).
func (e *Entorno) Get(t *testing.T, ruta, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/v1"+ruta, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.Router.ServeHTTP(rec, req)
	return rec
}

// Esperar reintenta cond hasta que sea cierta o venza el plazo.
func Esperar(t *testing.T, plazo time.Duration, cond func() bool) bool {
	t.Helper()
	limite := time.Now().Add(plazo)
	for time.Now().Before(limite) {
		if cond() {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return cond()
}
