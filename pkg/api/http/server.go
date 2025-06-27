package http

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gregmulvaney/forager/pkg/plugins"
	"github.com/gregmulvaney/forager/web"
	"github.com/gregmulvaney/forager/web/pages"
	"go.uber.org/zap"
)

type Config struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	SecurePort int    `mapstructure:"secure-port"`
	CertPath   string `mapstructure:"cert-path"`
}

type Server struct {
	config  *Config
	logger  *zap.Logger
	Router  *fiber.App
	Plugins *plugins.PluginRegister
}

func Init(config *Config, logger *zap.Logger) *Server {
	return &Server{
		config: config,
		logger: logger,
		Router: fiber.New(fiber.Config{}),
	}
}

func (s *Server) registerRoutes() {
	// Route to all static files
	s.Router.Use("/static", static.New("/", static.Config{
		FS: web.Static,
	}))

	s.Router.Get("/", func(ctx fiber.Ctx) error {
		return Render(ctx, pages.Index())
	})

	s.Router.Get("/settings", func(ctx fiber.Ctx) error {
		return Render(ctx, pages.Settings())
	})

	s.Router.Get("/services", func(ctx fiber.Ctx) error {
		return Render(ctx, pages.Services())
	})
}

func (s *Server) registerMiddleware() {
	// TODO
	s.Router.Use(compress.New(compress.ConfigDefault))
}

func (s *Server) ListenAndServe() {
	s.registerMiddleware()
	s.registerRoutes()
	s.logger.Info("Starting HTTP service", zap.String("Host", s.config.Host), zap.Int("Port", s.config.Port))
	if err := s.Router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)); err != nil {
		s.logger.Panic("Failed to start HTTP service", zap.Error(err))
	}
}

func Render(ctx fiber.Ctx, component templ.Component) error {
	ctx.Set("Content-Type", "text/html")
	return component.Render(ctx.Context(), ctx.Response().BodyWriter())
}
