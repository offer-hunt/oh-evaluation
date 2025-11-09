package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config хранит конфигурацию приложения, загруженную из переменных окружения.
type Config struct {
	AuthIssuer   string `envconfig:"AUTH_ISSUER" default:"http://localhost:8999"`
	AuthAudience string `envconfig:"AUTH_AUDIENCE" default:"evaluation-service"`
	AuthJwksURL  string `envconfig:"AUTH_JWKS_URL" default:"http://localhost:8999/.well-known/jwks.json"`
	ServerPort   string `envconfig:"HTTP_SERVER_PORT" default:"8080"`
	DBHost       string `envconfig:"DB_HOST" default:"localhost"`
	DBPort       int    `envconfig:"DB_PORT" default:"5432"`
	DBUser       string `envconfig:"DB_USER" default:"evaluation_user"`
	DBPass       string `envconfig:"DB_PASS" default:"evaluation_password"`
	DBName       string `envconfig:"DB_NAME" default:"evaluation_db"`
	DBSchema     string `envconfig:"DB_SCHEMA" default:"evaluation"`
}

// New создает новый экземпляр Config и загружает в него значения.
func New() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}
	return &cfg, nil
}

// DBString возвращает DSN (Data Source Name) для подключения к PostgreSQL.
func (c *Config) DBString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName, c.DBSchema)
}
