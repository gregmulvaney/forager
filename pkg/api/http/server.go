package http

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	//"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gregmulvaney/forager/web"
	"github.com/gregmulvaney/forager/web/pages"
	"go.uber.org/zap"
)

type Config struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	SecurePort int    `mapstructure:"secure-port"`
}

type Server struct {
	config *Config
	logger *zap.Logger
	router *fiber.App
}

func Init(config *Config, logger *zap.Logger) Server {
	return Server{
		config: config,
		logger: logger,
		router: fiber.New(),
	}
}

func (s *Server) registerMiddleware() {
	// s.router.Use(compress.New())
}

func (s *Server) registerRoutes() {
	s.router.Use("/static", static.New("/", static.Config{
		FS: web.Static,
	}))

	s.router.Get(healthcheck.DefaultLivenessEndpoint, healthcheck.NewHealthChecker(healthcheck.Config{}))
	s.router.Get(healthcheck.DefaultReadinessEndpoint, healthcheck.NewHealthChecker(healthcheck.Config{}))
	s.router.Get(healthcheck.DefaultStartupEndpoint, healthcheck.NewHealthChecker(healthcheck.Config{}))

	s.router.Get("/", func(ctx fiber.Ctx) error {
		return Render(ctx, pages.Index())
	})

	s.router.Get("/settings", func(ctx fiber.Ctx) error {
		return Render(ctx, pages.Settings())
	})
}

func (s *Server) Serve() {
	s.registerMiddleware()
	s.registerRoutes()
	s.logger.Info("Starting HTTP service on", zap.String("Host", s.config.Host), zap.Int("Port", s.config.Port))
	if err := s.router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)); err != nil {
		s.logger.Panic("Failed to start HTTP service", zap.Error(err))
	}
}

func Render(ctx fiber.Ctx, component templ.Component) error {
	ctx.Set("Content-Type", "text/html")
	return component.Render(ctx.Context(), ctx.Response().BodyWriter())
}
