package main

import (
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/internal/hr-api/middlewares"
	"web-boilerplate/internal/hr-api/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {

	app := fiber.New()

	err := config.LoadEnvFile()
	if err != nil {
		panic(err)
	}
	err = config.LoadAllConfig()
	if err != nil {
		panic(err)
	}

	// Disable cache control middleware in development and add dynamic route for style
	if !config.IS_PROD {
		app.Use(func(c fiber.Ctx) error {
			c.Set("Cache-Control", "no-store")
			return c.Next()
		})
	}

	//TODO: Setup middlewares
	middlewares.SetupMiddlewares(app)

	//TODO: Setup routes
	routes.SetupRoutes(app)

	//Start Server
	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}

}
