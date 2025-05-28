package http

import (
	"fmt"
	"log"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Config struct {
	Host       string
	Port       int
	SecurePort int
}

type Server struct {
	config *Config
	logger *zap.Logger
	Router *fiber.App
}

func Init(config *Config, logger *zap.Logger) Server {
	return Server{
		config: config,
		logger: logger,
		Router: fiber.New(),
	}
}

func (s *Server) registerMiddleware() {
	s.Router.Use(fiberzap.New(fiberzap.Config{
		Logger: s.logger,
	}))
}

func (s *Server) registerRoutes() {
	s.Router.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})
}

func (s *Server) Serve() {
	s.registerMiddleware()
	s.registerRoutes()
	s.logger.Info("Starting HTTP server at", zap.String("host", s.config.Host), zap.Int("port", s.config.Port))
	log.Fatal(s.Router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)))
}
