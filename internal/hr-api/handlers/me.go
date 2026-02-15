package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) GetMe(c fiber.Ctx) error {
	// Get user claims from the protected middleware
	claims, ok := c.Locals("user").(map[string]interface{})
	if !ok {
		h.Log.Error(nil, "failed to get user claims from context")
		return fiber.ErrUnauthorized
	}

	// Extract user ID from claims
	userIDStr, ok := claims["id"].(string)
	if !ok {
		h.Log.Error(nil, "invalid user id in claims")
		return fiber.ErrUnauthorized
	}

	// Parse the user ID to UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.Log.Error(err, "invalid user id format")
		return fiber.ErrUnauthorized
	}

	// Convert to pgtype.UUID
	pgUUID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	// Query the database for the user
	user, err := h.Repo.GetUser(context.Background(), pgUUID)
	if err != nil {
		h.Log.Error(err, "user not found")
		return fiber.ErrNotFound
	}

	// Return user data without password
	return c.JSON(fiber.Map{
		"id":       uuid.UUID(user.ID.Bytes).String(),
		"name":     user.Name,
		"email":    user.Email,
		"username": user.Username,
	})
}
