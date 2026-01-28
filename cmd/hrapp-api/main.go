package main

import (
	"context"
	"os"
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/middlewares"
	loggerpkg "web-boilerplate/internal/hr-api/pkg/logger"
	"web-boilerplate/internal/hr-api/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {

	err := config.LoadEnvFile()
	if err != nil {
		panic(err)
	}
	err = config.LoadAllConfig()
	if err != nil {
		panic(err)
	}

	// Initialize Dependencies
	logInst := loggerpkg.New(os.Getenv("LOG_LEVEL"))

	dbInst, err := db.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		logInst.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer dbInst.Close()

	app := fiber.New()

	// Disable cache control middleware in development and add dynamic route for style
	if !config.IS_PROD {
		app.Use(func(c fiber.Ctx) error {
			c.Set("Cache-Control", "no-store")
			return c.Next()
		})
	}

	// Setup middlewares
	middlewares.SetupLogger(app, logInst)
	middlewares.SetupMiddlewares(app)

	// Setup routes
	routes.SetupRoutes(app, logInst, dbInst)

	//Start Server
	err = app.Listen(":" + config.PORT)
	if err != nil {
		logInst.Fatal().Err(err).Msg("server failed to start")
	}
}
