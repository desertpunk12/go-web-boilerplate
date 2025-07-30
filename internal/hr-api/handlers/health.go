package handlers

import "github.com/gofiber/fiber/v3"

func Health(ctx fiber.Ctx) error {
	// TODO: check for db connection
	// TODO: check for redis connection
	// TODO: check for other services

	return ctx.Status(fiber.StatusOK).SendString("OK")
}
