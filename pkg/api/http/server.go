package http

import (
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
	router *fiber.App
}

func Init(config *Config, logger *zap.Logger) Server {
	return Server{
		config: config,
		logger: logger,
		router: fiber.New(),
	}
}

func (s *Server) regiserMiddleware() {
	s.router.Use(fiberzap.New(fiberzap.Config{
		Logger: s.logger,
	}))
}

func (s *Server) registerRoutes() {
	s.router.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})
}

func (s *Server) Serve() {
	s.regiserMiddleware()
	s.registerRoutes()
	s.logger.Info("Starting HTTP server")
	log.Fatal(s.router.Listen(":3000"))
}
