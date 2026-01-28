package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type LoginParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login(c fiber.Ctx) error {
	var params LoginParams
	err := c.Bind().Body(&params)
	if err != nil {
		h.Log.Error(err, "failed to bind body")
		return err
	}
	h.Log.Info("login attempt", "params", params)

	err = h.processLogin()
	if err != nil {
		h.Log.Error(err, "login process failed")
		return err
	}

	return c.JSON(params)
}

func (h *Handler) processLogin() error {
	// Use h.DB here
	return h.Pool.Ping(context.Background())
}
