package main

import (
	"os"
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
	})

	app.Get("/style", func(c fiber.Ctx) error {
		c.RequestCtx().SetContentType("text/css")
		// Use relative path instead of embedded assets
		file, err := os.Open("./assets/css/output.css")
		if err != nil {
			return err
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		buf := make([]byte, stat.Size())
		_, err = file.Read(buf)
		if err != nil {
			return err
		}

		return c.Send(buf)
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
