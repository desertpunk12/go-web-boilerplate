package main

import (
	"fmt"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// Test using fiberzerolog
	app.Use(fiberzerolog.New())

	fmt.Println("fiberzerolog import test successful")
}
