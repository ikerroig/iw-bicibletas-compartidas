// Package config carga la configuración del backend desde variables de
// entorno y, opcionalmente, un fichero .env en la raíz del repo (si existe).
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL  string
	RedisAddr    string
	JWTSecret    string
	ServerAddr   string
	TarifaMinuto float64
}

// Load lee .env (si existe, sin pisar variables ya definidas) y valida
// las variables obligatorias.
func Load() (*Config, error) {
	loadDotEnvWalkingUp()

	cfg := &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisAddr:    os.Getenv("REDIS_ADDR"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		ServerAddr:   os.Getenv("SERVER_ADDR"),
		TarifaMinuto: 0.20,
	}
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = ":8080"
	}
	if v := os.Getenv("TARIFA_MINUTO"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("TARIFA_MINUTO inválido (%q): %w", v, err)
		}
		cfg.TarifaMinuto = f
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL es obligatoria")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR es obligatoria")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET es obligatorio")
	}
	return cfg, nil
}

// loadDotEnvWalkingUp busca .env desde el CWD subiendo hasta 5 directorios
// (backend/cmd/server → backend → raíz con .env).
func loadDotEnvWalkingUp() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 5; i++ {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			loadDotEnv(path)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

// loadDotEnv parsea un .env plano (KEY=VALUE) sin pisar el entorno existente.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists && k != "" {
			os.Setenv(k, v)
		}
	}
}
