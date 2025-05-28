package main

import "github.com/gofiber/fiber/v2"

type service struct{}

func (s *service) RegisterRoutes(app *fiber.App) {
	app.Get("/plugin", func(c *fiber.Ctx) error {
		return c.SendString("plugin works")
	})
}

var ServicePlugin service
