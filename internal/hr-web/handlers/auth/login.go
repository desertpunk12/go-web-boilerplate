package auth

import (
	"web-boilerplate/internal/hr-web/config"
	"web-boilerplate/internal/hr-web/ui/pages"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

func LoginHandler(c fiber.Ctx) error { // TODO: change this according to library spec
	return pages.Login(config.BASE_URL).Render(c.Context(), c.Response().BodyWriter())
}

func Login(username, password string) (any, error) {
	//TODO:
	cc := client.New()
	resp, err := cc.Get(config.BASE_URL + "/v1/login")
	if err != nil {
		return nil, err
	}

	return string(resp.Body()), nil
}
