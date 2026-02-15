package routes

import (
	"web-boilerplate/internal/hr-web/config"
	"web-boilerplate/internal/hr-web/handlers/auth"
	"web-boilerplate/internal/hr-web/ui/pages"
	gpages "web-boilerplate/ui/pages"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func SetupRoutes(app *fiber.App, log *zerolog.Logger) {
	app.Get("/", func(c fiber.Ctx) error {
		c.RequestCtx().SetContentType("text/html")
		return pages.Login(config.BASE_URL).Render(c, c.Response().BodyWriter())
	})

	auth.InitLogger(log)
	app.Post("/v1/login", auth.Login)

	app.Get("/home", func(c fiber.Ctx) error {
		c.RequestCtx().SetContentType("text/html")
		var users []gpages.User
		users = append(users, gpages.User{
			ID:       "1",
			Fullname: "admin",
			Email:    "admin@example.com",
			IsActive: true,
		},
			gpages.User{
				ID:       "2",
				Fullname: "notadmin",
				Email:    "notadmin@example.com",
				IsActive: false,
			})
		return gpages.Home(users).Render(c, c.Response().BodyWriter())
	})
}
