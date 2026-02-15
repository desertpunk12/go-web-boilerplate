package middlewares

import (
	"strings"
	"web-boilerplate/internal/hr-api/config"

	"github.com/gofiber/fiber/v3/middleware/cors"
)

func SetupCorsConfig() cors.Config {
	return cors.Config{
		AllowCredentials: true,
		AllowOrigins:     strings.Split(config.ALLOWED_ORIGINS, ","),
		AllowHeaders:     []string{"Origin", " Content-Type", " Accept", " Accept-Language", " Content-Length"},
	}
}
