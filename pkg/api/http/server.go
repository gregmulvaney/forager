package http

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/pkg/plugins"
	"github.com/gregmulvaney/forager/web"
	"github.com/gregmulvaney/forager/web/pages"
	"go.uber.org/zap"
)

// Config holds the HTTP server configuration settings
type Config struct {
	Host       string `mapstructure:"host"`        // Host address to bind the server to
	Port       int    `mapstructure:"port"`        // Port number for HTTP connections
	SecurePort int    `mapstructure:"secure-port"` // Port number for HTTPS connections
	CertPath   string `mapstructure:"cert-path"`   // Path to SSL certificate file
}

// Server represents the HTTP server instance with all its dependencies
type Server struct {
	config  *Config                 // Server configuration
	logger  *zap.Logger             // Structured logger instance
	db      *db.DB                  // Database connection
	Router  *fiber.App              // Fiber web framework router
	Plugins *plugins.PluginRegister // Plugin registry for extensibility
}

// registerRoutes sets up all HTTP routes for the server
func (s *Server) registerRoutes() {
	// Configure static file serving for web assets
	s.Router.Use("/static", static.New("/", static.Config{
		FS: web.Static,
	}))

	s.Router.Get("/", func(ctx fiber.Ctx) error {
		return Render(ctx, pages.Index())
	})
}

// registerMiddleware configures HTTP middleware for the server
func (s *Server) registerMiddleware() {
	// Enable gzip compression for responses
	s.Router.Use(compress.New(compress.ConfigDefault))
}

// ListenAndServe starts the HTTP server and begins listening for requests
func (s *Server) ListenAndServe() {
	// Set up middleware and routes before starting
	s.registerMiddleware()
	s.registerRoutes()

	// Log server startup information
	s.logger.Info("Starting HTTP service", zap.String("Host", s.config.Host), zap.Int("Port", s.config.Port))

	// Start the server and handle any startup errors
	if err := s.Router.Listen(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)); err != nil {
		s.logger.Panic("Failed to start HTTP service", zap.Error(err))
	}
}

// Init creates and initializes a new HTTP server instance
func Init(config *Config, logger *zap.Logger, db *db.DB) *Server {
	return &Server{
		config: config,      // Store server configuration
		logger: logger,      // Store logger instance
		db:     db,          // Store database connection
		Router: fiber.New(), // Initialize new Fiber app
	}
}

// Render is a helper function to render templ components as HTTP responses
func Render(ctx fiber.Ctx, component templ.Component) error {
	// Set the appropriate content type for HTML responses
	ctx.Set("Content-Type", "text/html")
	// Render the component to the response body
	return component.Render(ctx.Context(), ctx.Response().BodyWriter())
}
