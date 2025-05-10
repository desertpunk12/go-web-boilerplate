package middlewares

import "github.com/gofiber/fiber/v3/middleware/cors"

func SetupCorsConfig() cors.Config {
	return cors.Config{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin", " Content-Type", " Accept", " Accept-Language", " Content-Length"},
	}
}
