package main

import (
	"os"
	"web-boilerplate/assets"
	"web-boilerplate/internal/hr/config"
	"web-boilerplate/ui/pages"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	app := fiber.New()

	err := config.LoadEnvFile()
	if err != nil {
		panic(err)
	}
	err = config.LoadAllConfig()
	if err != nil {
		panic(err)
	}

	// Disable cache control middleware in development and add dynamic route for style
	if !config.IS_PROD {
		app.Use(func(c fiber.Ctx) error {
			c.Set("Cache-Control", "no-store")
			return c.Next()
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
	}

	app.Get("/", func(c fiber.Ctx) error {
		c.RequestCtx().SetContentType("text/html")
		return pages.Login().Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/home", func(c fiber.Ctx) error {
		c.RequestCtx().SetContentType("text/html")
		var users []pages.User
		users = append(users, pages.User{
			ID:       "1",
			Fullname: "admin",
			Email:    "admin@example.com",
			IsActive: true,
		},
			pages.User{
				ID:       "2",
				Fullname: "notadmin",
				Email:    "notadmin@example.com",
				IsActive: false,
			})
		return pages.Home(users).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/static*", static.New("", static.Config{
		FS:     assets.Assets,
		Browse: true,
	}))

	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
