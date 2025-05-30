package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/contrib/fiberzap/v2"
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
		c.Set("Content-type", "text/html")
		return views.Index().Render(c.Context(), c.Response().BodyWriter())
	})

	s.Router.Use("/static/*", adaptor.HTTPHandler(http.FileServer(http.FS(web.Static))))
}

func (s *Server) Serve() {
	s.registerMiddleware()
	s.registerRoutes()
	s.logger.Info("Starting HTTP server at", zap.String("host", s.config.Host), zap.Int("port", s.config.Port))
	log.Fatal(s.Router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)))
}
