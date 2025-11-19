package utils

import (
	"os"
	"strconv"
	"time"
)

// Config contém configurações globais
type Config struct {
	HTTPTimeout       time.Duration
	MaxRetries        int
	RateLimitDelay    time.Duration
	MaxConcurrency    int
	EnableStructured  bool
	MongoURI          string
	GitHubToken       string
}

// LoadConfig carrega configurações de variáveis de ambiente
func LoadConfig() *Config {
	return &Config{
		HTTPTimeout:      getEnvDuration("HTTP_TIMEOUT", 30*time.Second),
		MaxRetries:       getEnvInt("MAX_RETRIES", 3),
		RateLimitDelay:   getEnvDuration("RATE_LIMIT_DELAY", 1*time.Second),
		MaxConcurrency:   getEnvInt("MAX_CONCURRENCY", 10),
		EnableStructured: getEnvBool("ENABLE_STRUCTURED_LOGS", false),
		MongoURI:         os.Getenv("MONGO_URI"),
		GitHubToken:      getGitHubToken(),
	}
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getGitHubToken() string {
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

// GetMongoURI obtém URI do MongoDB
func GetMongoURI() string {
	return os.Getenv("MONGO_URI")
}
