package handlers

import (
	"context"
	"errors"
	"time"
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/shared/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

	user, err := h.Repo.GetUserByUsername(context.Background(), params.Username)
	if err != nil {
		h.Log.Error(err, "user not found or db error")
		return fiber.ErrUnauthorized
	}

	err = helpers.CompareHashAndPassword(user.Password, params.Password)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			h.Log.Info("invalid password attempt")
			return fiber.ErrUnauthorized
		}
		h.Log.Error(err, "failed to compare password")
		return fiber.ErrInternalServerError
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
