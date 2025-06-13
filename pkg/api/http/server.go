package http

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gregmulvaney/forager/web"
	"github.com/gregmulvaney/forager/web/views"
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
	s.Router.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-type", "text/html")
		return views.Index().Render(c.Context(), c.Response().BodyWriter())
	})

	s.Router.Use("/static/*", adaptor.HTTPHandler(http.FileServer(http.FS(web.Static))))
}

func (s *Server) registerRoutes() {
	// todo
}

func (s *Server) Serve() {
	s.registerMiddlewares()
	s.registerRoutes()
	err := s.Router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
	if err != nil {
		s.logger.Error("Failed to start fiber", zap.Error(err))
	}
}
