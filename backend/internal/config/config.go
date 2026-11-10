// Package config lee la configuración del backend desde variables de entorno,
// cargando antes el fichero .env de la raíz del repo si existe.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config es la configuración del proceso (ver /.env.example).
type Config struct {
	DatabaseURL  string
	RedisAddr    string
	JWTSecret    string
	TarifaMinuto float64 // € por minuto de excedente (R8)
	Puerto       string
}

// Cargar busca un .env subiendo desde el directorio actual (para que funcione
// tanto desde la raíz como desde backend/) y valida los campos obligatorios.
func Cargar() (Config, error) {
	CargarDotenv()

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   valorODefecto("REDIS_ADDR", "localhost:6379"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Puerto:      valorODefecto("PUERTO", "8080"),
	}
	tarifa := valorODefecto("TARIFA_MINUTO", "0.20")
	t, err := strconv.ParseFloat(strings.Replace(tarifa, ",", ".", 1), 64)
	if err != nil || t < 0 {
		return Config{}, fmt.Errorf("TARIFA_MINUTO inválida %q: debe ser un número ≥ 0", tarifa)
	}
	cfg.TarifaMinuto = t

	var faltan []string
	if cfg.DatabaseURL == "" {
		faltan = append(faltan, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		faltan = append(faltan, "JWT_SECRET")
	}
	if len(faltan) > 0 {
		return Config{}, errors.New("faltan variables de entorno: " + strings.Join(faltan, ", ") + " (copia .env.example a .env)")
	}
	return cfg, nil
}

// Load es el alias histórico que usan los tests de spec-00X (devuelven *Config).
func Load() (*Config, error) {
	cfg, err := Cargar()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// CargarDotenv carga el primer .env encontrado subiendo directorios. No pisa
// variables ya definidas en el entorno.
func CargarDotenv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 5; i++ {
		ruta := filepath.Join(dir, ".env")
		if _, err := os.Stat(ruta); err == nil {
			_ = godotenv.Load(ruta)
			return
		}
		padre := filepath.Dir(dir)
		if padre == dir {
			return
		}
		dir = padre
	}
}

func valorODefecto(clave, defecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return defecto
}
