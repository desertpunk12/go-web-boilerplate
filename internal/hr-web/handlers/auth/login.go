package auth

import (
	"log"
	"web-boilerplate/internal/hr-web/config"
	"web-boilerplate/internal/hr-web/ui/pages"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

func LoginHandler(c fiber.Ctx) error { // TODO: change this according to library spec
	return pages.Login(config.API_URL).Render(c, c.Response().BodyWriter())
}

func Login(username, password string) (any, error) {
	cc := client.New()
	log.Printf("requesting login: %s", config.API_URL+"/v1/login")
	resp, err := cc.Get(config.API_URL + "/v1/login")
	if err != nil {
		return nil, err
	}
	log.Printf("Login response: %s", string(resp.Body()))

	return string(resp.Body()), nil
}
