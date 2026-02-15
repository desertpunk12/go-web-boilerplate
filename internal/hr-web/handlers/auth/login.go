package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"web-boilerplate/internal/hr-web/config"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

var log *zerolog.Logger

func InitLogger(logger *zerolog.Logger) {
	log = logger
}

// Login handles the POST authentication request
func Login(c fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	payload := map[string]string{
		"username": username,
		"password": password,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	resp, err := http.Post(config.API_URL+"/v1/login", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error().Err(err).Str("url", config.API_URL+"/v1/login").Msg("error requesting backend")
		return c.Status(fiber.StatusServiceUnavailable).SendString("Backend Service Unavailable")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid Credentials")
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to parse response")
	}

	// Set token as cookie
	if token, ok := result["token"].(string); ok {
		cookie := new(fiber.Cookie)
		cookie.Name = "auth_token"
		cookie.Value = token
		cookie.Expires = time.Now().Add(24 * time.Hour)
		c.Cookie(cookie)
	}

	return c.Redirect().To("/home")
}
