package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                         string
	HTTPAddr                       string
	LoggerLevel                    string
	LoggerToFile                   bool
	LoggerFile                     string
	AntsPoolSize                   int
	RedisAddr                      string
	RedisPassword                  string
	RedisDB                        int
	RedisPoolSize                  int
	PostgresDSN                    string
	PostgresMaxIdleConns           int
	PostgresMaxOpenConns           int
	PostgresConnMaxLifetimeMinutes int
	JWTSecret                      string
	AdminSecret                    string
	AdminSessionTTLMinutes         int
	AdminRoutePrefix               string
	FrontendDist                   string
	PublicBaseURL                  string
}

var App Config

func Load() Config {
	_ = godotenv.Load()
	App = Config{
		AppEnv:                         get("APP_ENV", "development"),
		HTTPAddr:                       get("HTTP_ADDR", ":8080"),
		LoggerLevel:                    get("LOGGER_LEVEL", "info"),
		LoggerToFile:                   getBool("LOGGER_TO_FILE", false),
		LoggerFile:                     get("LOGGER_FILE", "logs/app.log"),
		AntsPoolSize:                   getInt("ANTS_POOL_SIZE", 100),
		RedisAddr:                      get("REDIS_ADDR", "localhost:6379"),
		RedisPassword:                  get("REDIS_PASSWORD", ""),
		RedisDB:                        getInt("REDIS_DB", 0),
		RedisPoolSize:                  getInt("REDIS_POOL_SIZE", 20),
		PostgresDSN:                    get("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=telegram_bot port=5432 sslmode=disable TimeZone=Asia/Shanghai"),
		PostgresMaxIdleConns:           getInt("POSTGRES_MAX_IDLE_CONNS", 10),
		PostgresMaxOpenConns:           getInt("POSTGRES_MAX_OPEN_CONNS", 50),
		PostgresConnMaxLifetimeMinutes: getInt("POSTGRES_CONN_MAX_LIFETIME_MINUTES", 60),
		JWTSecret:                      get("JWT_SECRET", "change-me-jwt-secret"),
		AdminSecret:                    get("ADMIN_SECRET", "change-me-admin-hmac-secret"),
		AdminSessionTTLMinutes:         getInt("ADMIN_SESSION_TTL_MINUTES", 120),
		AdminRoutePrefix:               get("ADMIN_ROUTE_PREFIX", "/api/admin"),
		FrontendDist:                   get("FRONTEND_DIST", "web/dist"),
		PublicBaseURL:                  get("PUBLIC_BASE_URL", ""),
	}
	return App
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func getInt(key string, fallback int) int {
	v, err := strconv.Atoi(get(key, ""))
	if err != nil {
		return fallback
	}
	return v
}
func getBool(key string, fallback bool) bool {
	v, err := strconv.ParseBool(get(key, ""))
	if err != nil {
		return fallback
	}
	return v
}
