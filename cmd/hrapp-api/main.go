package main

import (
	"context"
	"fmt"
	"os"
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/middlewares"
	loggerpkg "web-boilerplate/internal/hr-api/pkg/logger"
	"web-boilerplate/internal/hr-api/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {
	err := config.LoadAllConfig()
	if err != nil {
		fmt.Printf("Failed to load configs from environment, err: %v", err)
		panic(err)
	}

	// Initialize Dependencies
	logLvl := os.Getenv("LOG_LEVEL")
	if logLvl == "" {
		logLvl = "info"
	}
	logInst := loggerpkg.New(logLvl)

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

	// Setup middlewares (order matters - recover should be first)
	logAdapter := loggerpkg.NewZerologAdapter(logInst)
	middlewares.SetupMiddlewareRecover(app, logAdapter)
	middlewares.SetupLogger(app, logInst)
	middlewares.SetupMiddlewares(app)

	// Setup routes
	routes.SetupRoutes(app, logInst, dbInst)

	if config.BASE_URL == "" {
		config.BASE_URL = "localhost:3000"
	}
	fmt.Printf("baseurl:%s\n", config.BASE_URL)
	//Start Server
	err = app.Listen(config.BASE_URL)
	if err != nil {
		logInst.Fatal().Err(err).Msg("server failed to start")
	}
}
