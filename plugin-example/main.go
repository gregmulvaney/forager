package main

import (
	"context"
	"database/sql"
	"plugin-example/sqlc"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type LayoutRenderer interface {
	RenderWithLayout(title string, content string) ([]byte, error)
}

type service struct{}

func (s *service) Register(db *sql.DB, router fiber.Router, logger *zap.Logger, renderer LayoutRenderer) {
	logger.Debug("Attempting to migrate plugin database schema", zap.String("Plugin", "ServiceName"))
	if _, err := db.ExecContext(context.Background(), sqlc.DDL); err != nil {
		logger.Error("Failed to migrate database schema", zap.String("Plugin", ServiceName), zap.Error(err))
	}

	router.Get("/example", func(ctx fiber.Ctx) error {
		content := `<div class="p-6">
			<h1 class="text-2xl font-bold mb-4">Example Plugin</h1>
			<p class="text-gray-600">This is an example plugin rendered with the layout component.</p>
		</div>`

		html, err := renderer.(interface {
			RenderWithLayout(string, string) ([]byte, error)
		}).RenderWithLayout("Example Plugin", content)
		if err != nil {
			logger.Error("Failed to render with layout", zap.Error(err))
			return ctx.Status(500).SendString("Internal Server Error")
		}

		ctx.Set("Content-Type", "text/html")
		return ctx.Send(html)
	})

}

var Service service
var ServiceName = "Example"
var ServiceDefaultPath = "/example"
