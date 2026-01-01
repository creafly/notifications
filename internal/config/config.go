package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/creafly/vault"
	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Kafka     KafkaConfig
	Identity  IdentityConfig
	Outbox    OutboxConfig
	CORS      CORSConfig
	I18n      I18nConfig
	Tracing   TracingConfig
	RateLimit RateLimitConfig
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond float64
	BurstSize         int
}

type IdentityConfig struct {
	ServiceURL string
}

type TracingConfig struct {
	Enabled        bool
	OTLPEndpoint   string
	ServiceName    string
	ServiceVersion string
	Environment    string
}

type ServerConfig struct {
	Host    string
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	URL         string
	AutoMigrate bool
}

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Enabled bool
}

type OutboxConfig struct {
	Enabled               bool
	PollInterval          time.Duration
	BatchSize             int
	CleanupInterval       time.Duration
	RetentionPeriod       time.Duration
	FailedRetentionPeriod time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type I18nConfig struct {
	DefaultLocale string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	secrets := vault.NewSecretLoaderFromEnv("notifications")

	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
	ginMode := getEnv("GIN_MODE", "debug")

	databaseURL := buildDatabaseURL(secrets)

	return &Config{
		Server: ServerConfig{
			Host:    getEnv("SERVER_HOST", "0.0.0.0"),
			Port:    getEnv("SERVER_PORT", "8081"),
			GinMode: ginMode,
		},
		Database: DatabaseConfig{
			URL:         databaseURL,
			AutoMigrate: getEnvBool("AUTO_MIGRATE", true),
		},
		Kafka: KafkaConfig{
			Brokers: splitNonEmpty(getEnv("KAFKA_BROKERS", ""), ","),
			GroupID: getEnv("KAFKA_GROUP_ID", "notifications-service"),
			Enabled: getEnvBool("KAFKA_ENABLED", true),
		},
		Identity: IdentityConfig{
			ServiceURL: getEnv("IDENTITY_SERVICE_URL", "http://localhost:8080"),
		},
		Outbox: OutboxConfig{
			Enabled:               getEnvBool("OUTBOX_ENABLED", true),
			PollInterval:          getEnvDuration("OUTBOX_POLL_INTERVAL", time.Second),
			BatchSize:             getEnvInt("OUTBOX_BATCH_SIZE", 100),
			CleanupInterval:       getEnvDuration("OUTBOX_CLEANUP_INTERVAL", time.Hour),
			RetentionPeriod:       getEnvDuration("OUTBOX_RETENTION_PERIOD", 24*time.Hour),
			FailedRetentionPeriod: getEnvDuration("OUTBOX_FAILED_RETENTION_PERIOD", 7*24*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins:   splitNonEmpty(corsOrigins, ","),
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "Accept-Language"},
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           86400,
		},
		I18n: I18nConfig{
			DefaultLocale: getEnv("DEFAULT_LOCALE", "en-US"),
		},
		Tracing: TracingConfig{
			Enabled:        getEnvBool("TRACING_ENABLED", false),
			OTLPEndpoint:   getEnv("OTLP_ENDPOINT", "localhost:4317"),
			ServiceName:    "notifications",
			ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
			Environment:    getEnv("ENVIRONMENT", "development"),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: getEnvFloat("RATE_LIMIT_RPS", 100),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 200),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return i
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		d, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return d
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return defaultValue
		}
		return f
	}
	return defaultValue
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func buildDatabaseURL(secrets *vault.SecretLoader) string {
	host := getEnv("DATABASE_HOST", "localhost")
	port := getEnv("DATABASE_PORT", "5432")
	name := getEnv("DATABASE_NAME", "notifications")
	user := getEnv("DATABASE_USER", "postgres")
	sslMode := getEnv("DATABASE_SSL_MODE", "disable")

	password := secrets.GetSecret("database_password", "DATABASE_PASSWORD", "")
	if password == "" {
		log.Fatal("DATABASE_PASSWORD is required (from Vault or environment)")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		url.QueryEscape(password),
		host,
		port,
		name,
		sslMode,
	)
}
