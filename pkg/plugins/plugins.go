package plugins

import (
	"github.com/gregmulvaney/pkg/api/http"

	"github.com/gofiber/fiber/v2"
)

type Plugin interface {
	RegisterRoutes(*fiber.App)
}

func Register(httpServer *http.Server)
