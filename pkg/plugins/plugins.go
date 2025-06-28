package plugins

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
	"regexp"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/pkg/db/queries"
	"github.com/gregmulvaney/forager/web/components"
	"go.uber.org/zap"
)

// Config holds the plugin system configuration.
type Config struct {
	Directory string `mapstructure:"plugins-dir"` // Directory path where plugin files are stored
}

// PluginInterface defines the contract that all plugins must implement.
// Plugins must provide a Register method that accepts the server API.
type PluginInterface interface {
	Register(serverAPI any) // Called when the plugin is loaded to register routes and services
}

// ServerAPI provides the interface that plugins can use to interact with the server.
// It exposes logging, routing, and database functionality to plugins.
type ServerAPI struct {
	logger     *zap.Logger  // Logger instance for plugin logging
	routeGroup fiber.Router // Router group for registering plugin routes
	db         *db.DB       // Database connection for plugin data operations
}

// MigrateSchema executes DDL statements to migrate plugin-specific database schema.
// Returns the database connection for further operations.
func (s *ServerAPI) MigrateSchema(ddl string) *sql.DB {
	if _, err := s.db.Conn.ExecContext(context.Background(), ddl); err != nil {
		s.logger.Error("Error migrating plugin schema", zap.Error(err))
	}

	return s.db.Conn
}

// RegisterViewRoute registers a new HTTP route that renders HTML content.
// The content is wrapped in the application's standard layout with the provided title.
func (s *ServerAPI) RegisterViewRoute(path string, title string, content string) {
	s.routeGroup.Get(path, func(ctx fiber.Ctx) error {
		ctx.Set("Content-Type", "text/html")
		c := templ.WithChildren(ctx.Context(), templ.Raw(content))
		return components.Layout(title).Render(c, ctx.Response().BodyWriter())
	})
}

// ServicePlugin represents a loaded plugin service with URL matching capabilities.
type ServicePlugin struct {
	Name    string           // Human-readable name of the plugin service
	Regex   regexp.Regexp    // Regular expression for matching URLs this service handles
	Handler func(url string) // Function to handle matched URLs
}

// PluginRegister manages the loading and registration of plugins.
// It maintains references to all loaded plugins and their configurations.
type PluginRegister struct {
	config  *Config          // Plugin system configuration
	logger  *zap.Logger      // Logger for plugin operations
	router  *fiber.App       // Main application router
	db      *db.DB           // Database connection
	Plugins []*ServicePlugin // Collection of loaded service plugins
}

// Init creates a new PluginRegister instance with the provided dependencies
func Init(config *Config, logger *zap.Logger, router *fiber.App, db *db.DB) *PluginRegister {
	return &PluginRegister{
		config: config,
		logger: logger,
		router: router,
		db:     db,
	}
}

// RegisterPlugins discovers and loads all plugin files from the configured directory.
// It walks the plugin directory, loads .so files, and registers them with the system.
func (p *PluginRegister) RegisterPlugins() {
	p.logger.Debug("Attempting to load plugins")

	serverAPI := &ServerAPI{
		logger:     p.logger,
		routeGroup: p.router.Group("/services"),
		db:         p.db,
	}

	err := filepath.WalkDir(p.config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) == ".so" {
			p.logger.Debug("Loading plugin from", zap.String("Path", path))
			service, err := p.loadPlugin(path, serverAPI)
			if err != nil {
				p.logger.Error("Failed to load plugin", zap.String("Path", path), zap.Error(err))
			}
			p.Plugins = append(p.Plugins, service)
		}

		return nil
	})

	if err != nil {
		p.logger.Error("Error walking plugin directory", zap.String("Directory", p.config.Directory), zap.Error(err))
	}
}

// loadPlugin loads a single plugin file and registers it with the provided server API.
// It opens the .so file, looks up required symbols, and calls the plugin's Register method.
func (p *PluginRegister) loadPlugin(path string, serverAPI *ServerAPI) (*ServicePlugin, error) {
	symPlugin, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	hash, err := hashPlugin(path)

	symService, err := symPlugin.Lookup("Service")
	if err != nil {
		return nil, err
	}

	var service PluginInterface
	service = symService.(PluginInterface)

	(service).Register(serverAPI)

	name, err := lookupSymbol[string](symPlugin, "ServiceName")

	p.db.Q.CreatePlugin(context.Background(), queries.CreatePluginParams{
		Name: fmt.Sprintf("%s", *name),
		Path: path,
		Hash: hash,
	})

	return &ServicePlugin{}, nil
}

// hashPlugin calculates the SHA256 hash of a plugin file for integrity checking
func hashPlugin(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// lookupSymbol performs type-safe symbol lookup from a loaded plugin
func lookupSymbol[T any](plugin *plugin.Plugin, symbolName string) (*T, error) {
	symbol, err := plugin.Lookup(symbolName)
	if err != nil {
		return nil, err
	}

	switch symbol.(type) {
	case *T:
		return symbol.(*T), nil
	case T:
		result := symbol.(T)
		return &result, err
	default:
		return nil, fmt.Errorf("Failed to lookup symbol %T", symbol)
	}
}
