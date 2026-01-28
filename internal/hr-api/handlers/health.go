package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func (h *Handler) Health(ctx fiber.Ctx) error {
	// check for db connection
	// check for db connection
	if err := h.Pool.Ping(ctx.Context()); err != nil {
		h.Log.Error(err, "database ping failed")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Database Unreachable")
	}

	// TODO: check for redis connection
	// TODO: check for other services

	h.Log.Info("health check passed")
	return ctx.Status(fiber.StatusOK).SendString("OK")
}
