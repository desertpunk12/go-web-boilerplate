package handlers

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

type LoginParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(c fiber.Ctx) error {
	var params LoginParams
	err := c.Bind().Body(params)
	if err != nil {
		return err
	}
	log.Print(params)

	err = login()
	if err != nil {
		return err
	}

	return c.JSON(params)
}

func login() error {

	return nil
}
