package plugins

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
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

// Config holds plugin system configuration
type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

// ServicePluginInterface defines the contract that all service plugins must implement
type ServicePluginInterface interface {
	Register(ServerApiInterface)
}

// ServerApiInterface provides the API that plugins can use to interact with the server
type ServerApiInterface interface {
	MigrateSchema(ddl string) (*sql.DB, error)
	RegisterRoute(path string, title string, content string)
	GetQueries() *queries.Queries
}

// ServerAPI implements ServerApiInterface and provides plugins access to server functionality
type ServerAPI struct {
	db         *db.Db
	logger     *zap.Logger
	routeGroup fiber.Router
}

// MigrateSchema executes DDL statements to migrate the database schema
func (s *ServerAPI) MigrateSchema(ddl string) (*sql.DB, error) {
	if _, err := s.db.Conn.ExecContext(context.Background(), ddl); err != nil {
		return nil, err
	}
	return s.db.Conn, nil
}

// RegisterRoute registers a new HTTP route with the given path, title, and HTML content
func (s *ServerAPI) RegisterRoute(path string, title string, content string) {
	s.routeGroup.Get(path, func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		ctx := templ.WithChildren(c.Context(), templ.Raw(content))
		return components.Layout(title).Render(ctx, c.Response().BodyWriter())
	})
}

// GetQueries returns the database queries instance for plugin use
func (s *ServerAPI) GetQueries() *queries.Queries {
	return s.db.Q
}

// ServicePlugin represents a loaded plugin with its metadata
type ServicePlugin struct {
	Name        *string
	DomainRegex *regexp.Regexp
	Handler     func()
}

// PluginRegister manages the loading and registration of plugins
type PluginRegister struct {
	config  *Config
	db      *db.Db
	logger  *zap.Logger
	router  *fiber.App
	Plugins []*ServicePlugin
}

// Init creates a new PluginRegister instance with the provided dependencies
func Init(config *Config, db *db.Db, logger *zap.Logger, router *fiber.App) *PluginRegister {
	return &PluginRegister{
		config: config,
		db:     db,
		logger: logger,
		router: router,
	}
}

// RegisterPlugins scans the plugin directory and loads all .so files as plugins
func (p *PluginRegister) RegisterPlugins() {
	p.logger.Debug("Attempting to load plugins")

	serverAPI := &ServerAPI{
		db:         p.db,
		logger:     p.logger,
		routeGroup: p.router.Group("/services"),
	}

	err := filepath.WalkDir(p.config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) == ".so" {
			p.logger.Debug("Attempting to load plugin at", zap.String("Path", path))
			service, err := p.loadPlugin(path, serverAPI)
			if err != nil {
				p.logger.Error("Failed to load plugin", zap.String("Path", path), zap.Error(err))
			}
			p.Plugins = append(p.Plugins, service)
		}

		return nil
	})

	if err != nil {
		p.logger.Error("Failed to walk plugin directory", zap.String("Directory", p.config.Directory), zap.Error(err))
	}
}

// loadPlugin loads a single plugin from the specified path and registers it with the server
func (p *PluginRegister) loadPlugin(path string, serverAPI *ServerAPI) (*ServicePlugin, error) {
	symPlugin, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	hash, err := hashPlugin(path)
	if err != nil {
		return nil, err
	}

	name, err := lookupSymbol[string](symPlugin, "ServiceName")
	if err != nil {
		return nil, err
	}

	p.db.Q.CreatePlugin(context.Background(), queries.CreatePluginParams{
		Name: fmt.Sprintf("%s", *name),
		Path: path,
		Hash: hash,
	})

	service, err := lookupSymbol[ServicePluginInterface](symPlugin, "Service")
	if err != nil {
		return nil, err
	}

	regex, err := lookupSymbol[string](symPlugin, "DomainRegex")
	if err != nil {
		return nil, err
	}

	(*service).Register(serverAPI)

	return &ServicePlugin{
		Name:        name,
		DomainRegex: regexp.MustCompile(*regex),
		Handler: func() {
			p.logger.Debug("Domain matched", zap.String("Plugin", *name))
		},
	}, nil
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
		return nil, errors.New("Failed to lookup symbol")
	}
}
