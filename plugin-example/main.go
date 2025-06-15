package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

var Name = "Example"

type service struct{}

func (s *service) Init() {
	fmt.Printf("Plugin")
}

func (s *service) RegisterRoutes(router fiber.Router) {
	router.Get("/example", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Example")
	})
}

var Service service
