package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type LOG_LEVEL_TYPE int8

const (
	LOG_LEVEL_NOTFOUND LOG_LEVEL_TYPE = iota - 1
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_DEBUG
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL
)

var (
	IS_PROD         = false
	LOG_LEVEL       = LOG_LEVEL_DEBUG
	SECRET_KEY      = "qweasd123"
	ALLOWED_ORIGINS = ""
	REDIS_KEYS_TTL  = time.Hour * 24 * 7 // 7 days
	TOKEN_TTL       = time.Hour * 5      // 5 hours
	BASE_URL        = "localhost:4000"
	API_URL         = "localhost:3000"
)

func LoadAllConfig() error {
	// Load env file
	err := LoadEnvFile()
	if err != nil {
		return err
	}

	LOG_LEVEL, err = determineLogLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return err
	}

	BASE_URL = os.Getenv("BASE_URL")
	API_URL = os.Getenv("API_URL")
	IS_PROD = os.Getenv("IS_PROD") == "true"

	SECRET_KEY = os.Getenv("SECRET_KEY")

	expireRaw := os.Getenv("TOKEN_EXPIRE_TIME")
	if expireRaw != "" {
		TOKEN_TTL, err = time.ParseDuration(expireRaw)
		if err != nil {
			return fmt.Errorf("error parsing token expire time duration, %w", err)
		}
	} else {
		TOKEN_TTL = time.Hour * 5
	}

	ALLOWED_ORIGINS = os.Getenv("ALLOWED_ORIGINS")
	if len(ALLOWED_ORIGINS) == 0 {
		ALLOWED_ORIGINS = "*"
	}

	REDIS_KEYS_TTL, err = time.ParseDuration(os.Getenv("REDIS_KEYS_TTL"))
	if err != nil {
		REDIS_KEYS_TTL = time.Hour * 24 * 7
	}

	return nil
}
func LoadEnvFile() error {
	paths := []string{".env", "cmd/hrapp-web/.env"}
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			err := godotenv.Load(path)
			if err != nil {
				return err
			}
			fmt.Printf("Loaded .env from: %s\n", path)
			return nil
		}
	}
	return nil
}

// func SetupCorsConfig() cors.Config {
// 	return cors.Config{
// 		AllowCredentials: true,
// 		AllowOrigins:     []string{"*"},
// 		AllowHeaders:     []string{"Origin", " Content-Type", " Accept", " Accept-Language", " Content-Length"},
// 	}
// }

func determineLogLevel(logLevel string) (LOG_LEVEL_TYPE, error) {
	switch logLevel {
	case "":
		return LOG_LEVEL_DEBUG, nil
	case "info":
		return LOG_LEVEL_INFO, nil
	case "warn":
		return LOG_LEVEL_WARN, nil
	case "debug":
		return LOG_LEVEL_DEBUG, nil
	case "error":
		return LOG_LEVEL_ERROR, nil
	case "fatal":
		return LOG_LEVEL_FATAL, nil
	default:
		return LOG_LEVEL_NOTFOUND, fmt.Errorf("invalid log level")
	}
}
