package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
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
	s.router.Use(healthcheck.New())
	s.router.Use(compress.New())

}

func (s *Server) registerRoutes() {

}

func (s *Server) Serve() {
	s.registerMiddleware()
	s.registerRoutes()
	if err := s.router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)); err != nil {
		s.logger.Panic("Failed to start HTTP service", zap.Error(err))
	}
}
