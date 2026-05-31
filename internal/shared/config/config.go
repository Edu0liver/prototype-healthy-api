// Package config loads typed application configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config is the fully-resolved application configuration.
type Config struct {
	App       App
	Database  Database
	Redis     Redis
	JWT       JWT
	OpenAI    OpenAI
	Evolution Evolution
	Crypto    Crypto
	Worker    Worker
	Storage   Storage
	Email     Email
	Security  Security
}

type Security struct {
	// RateLimitPerMinute caps authenticated requests per tenant per minute.
	RateLimitPerMinute int
}

type App struct {
	Env     string // development | production
	Port    string
	Version string
	// PublicBaseURL is the externally reachable base URL (used to build webhook URLs).
	PublicBaseURL string
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	URL      string // built from above fields
}

type Redis struct {
	URL string
}

type JWT struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type OpenAI struct {
	APIKey         string
	BaseURL        string
	EmbeddingModel string
	WhisperModel   string
	Timeout        time.Duration
}

type Evolution struct {
	BaseURL      string
	GlobalAPIKey string
	WebhookURL   string // where Evolution posts webhooks
	WebhookToken string // shared secret validated on inbound webhooks
	Timeout      time.Duration
}

type Crypto struct {
	// EncryptionKey is a 32-byte key (hex-encoded, 64 chars) for AES-256-GCM.
	EncryptionKey string
}

type Worker struct {
	Concurrency     int
	StreamName      string
	ConsumerGroup   string
	DebounceSeconds int
}

type Storage struct {
	Endpoint  string // MinIO/S3 host:port (no scheme)
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

type Email struct {
	ResendAPIKey string
	FromAddress  string
}

// Load reads configuration from the environment, optionally seeding from a .env file.
func Load() (*Config, error) {
	_ = godotenv.Load() // best effort; real env wins

	cfg := &Config{
		App: App{
			Env:           env("APP_ENV", "development"),
			Port:          env("PORT", "8080"),
			Version:       env("APP_VERSION", "dev"),
			PublicBaseURL: env("PUBLIC_BASE_URL", "http://localhost:8080"),
		},
		Database: Database{
			Host:     env("PG_HOST", "localhost"),
			Port:     env("PG_PORT", "5432"),
			User:     env("PG_USER", "app_user"),
			Password: env("PG_PASSWORD", "app_pw"),
			Name:     env("PG_DB", "lumia"),
			SSLMode:  env("PG_SSLMODE", "disable"),
		},
		Redis:    Redis{URL: env("REDIS_URL", "redis://localhost:6379")},
		JWT: JWT{
			Secret:     env("JWT_SECRET", ""),
			AccessTTL:  durationEnv("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: durationEnv("JWT_REFRESH_TTL", 720*time.Hour),
		},
		OpenAI: OpenAI{
			APIKey:         env("OPENAI_API_KEY", ""),
			BaseURL:        env("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			EmbeddingModel: env("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
			WhisperModel:   env("OPENAI_WHISPER_MODEL", "whisper-1"),
			Timeout:        durationEnv("OPENAI_TIMEOUT", 60*time.Second),
		},
		Evolution: Evolution{
			BaseURL:      env("EVOLUTION_BASE_URL", "http://localhost:8081"),
			GlobalAPIKey: env("EVOLUTION_API_KEY", ""),
			WebhookURL:   env("EVOLUTION_WEBHOOK_URL", ""),
			WebhookToken: env("EVOLUTION_WEBHOOK_TOKEN", ""),
			Timeout:      durationEnv("EVOLUTION_TIMEOUT", 30*time.Second),
		},
		Crypto: Crypto{EncryptionKey: env("ENCRYPTION_KEY", "")},
		Worker: Worker{
			Concurrency:     intEnv("WORKER_CONCURRENCY", 8),
			StreamName:      env("WORKER_STREAM", "stream:inbound"),
			ConsumerGroup:   env("WORKER_GROUP", "orchestrators"),
			DebounceSeconds: intEnv("DEBOUNCE_SECONDS", 8),
		},
		Storage: Storage{
			Endpoint:  env("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKey: env("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey: env("STORAGE_SECRET_KEY", "minioadmin"),
			Bucket:    env("STORAGE_BUCKET", "lumia"),
			UseSSL:    boolEnv("STORAGE_USE_SSL", false),
			Region:    env("STORAGE_REGION", "us-east-1"),
		},
		Security: Security{RateLimitPerMinute: intEnv("RATE_LIMIT_PER_MINUTE", 600)},
		Email: Email{
			ResendAPIKey: env("RESEND_API_KEY", ""),
			FromAddress:  env("EMAIL_FROM", "no-reply@example.com"),
		},
	}

	cfg.Database.URL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.SSLMode,
	)
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.App.Env == "production" }

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func durationEnv(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func boolEnv(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
