package http

import (
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

func Init(config *Config, logger *zap.Logger) (Server, error) {
	return Server{
		config: config,
		logger: logger,
		Router: fiber.New(),
	}, nil
}

func (s *Server) registerMiddlewares() {
	// todo
}

func (s *Server) registerRoutes() {
	// todo
}

func (s *Server) Serve() {
	s.registerMiddlewares()
	s.registerRoutes()
}
