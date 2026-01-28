package handlers

import (
	"context"
	"time"
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/shared/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
		return fiber.ErrBadRequest
	}

	// 1. Fetch user from DB
	user, err := h.Repo.GetUserByUsername(context.Background(), params.Username)
	if err != nil {
		h.Log.Error(err, "user not found or db error")
		return fiber.ErrUnauthorized
	}

	// 2. Hash provided password for comparison
	hashedPassword, err := helpers.HashText(params.Password)
	if err != nil {
		h.Log.Error(err, "failed to hash password")
		return fiber.ErrInternalServerError
	}

	// 3. Verify password
	if user.Password != hashedPassword {
		h.Log.Info("invalid password attempt")
		return fiber.ErrUnauthorized
	}

	// 4. Generate JWT Token
	// Convert pgtype.UUID to google/uuid for easier string handling if needed,
	// but we can just use the bytes directly for formatting or use a string helper.
	userID := uuid.UUID(user.ID.Bytes).String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(config.TOKEN_TTL).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.SECRET_KEY))
	if err != nil {
		h.Log.Error(err, "failed to sign token")
		return fiber.ErrInternalServerError
	}

	h.Log.Info("login successful", "username", params.Username, "id", userID)

	return c.JSON(fiber.Map{
		"token": tokenString,
		"id":    userID,
	})
}
