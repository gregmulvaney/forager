package http

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/web"
	"go.uber.org/zap"
)

type Config struct {
	AppEnv     string `mapstructure:"app-env"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	SecurePort int    `mapstructure:"secure-port"`
	CertPath   string `mapstructure:"cert-path"`
}

type Server struct {
	config *Config
	logger *zap.Logger
	db     *db.DB
	Router *fiber.App
}

func (s *Server) registerMiddleware() {
	if s.config.AppEnv == "production" {
		s.Router.Use(compress.New(compress.ConfigDefault))
	}
}

func (s *Server) registerRoutes() {
	s.Router.Use("/static", static.New("/", static.Config{
		FS: web.Static,
	}))
}

func (s *Server) ListenAndServe() {
	s.registerMiddleware()
	s.registerRoutes()

	s.logger.Info("starting plaintext HTTP service on", zap.String("Host", s.config.Host), zap.Int("Port", s.config.Port))

	if err := s.Router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)); err != nil {
		s.logger.Panic("Failed to initialize plaintext HTTP service", zap.Error(err))
	}
}

func Init(config *Config, logger *zap.Logger, db *db.DB) *Server {
	router := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	return &Server{
		config: config,
		logger: logger,
		db:     db,
		Router: router,
	}
}

func Render(ctx fiber.Ctx, component templ.Component) error {
	// Set the appropriate content type for HTML responses
	ctx.Set("Content-Type", "text/html")
	// Render the component to the response body
	return component.Render(ctx.Context(), ctx.Response().BodyWriter())
}
