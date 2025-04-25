package main

import (
	"web-boilerplate/assets"
	"web-boilerplate/ui/pages"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		c.RequestCtx().SetContentType("text/html")
		return pages.Login().Render(c.Context(), c.Response().BodyWriter())
		// return c.SendString("Hello, World!")
	})

	app.Get("/static*", static.New("", static.Config{
		FS:     assets.Assets,
		Browse: true,
	}))

	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
