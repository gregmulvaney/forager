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

	"github.com/gofiber/fiber/v3"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/pkg/db/queries"
	"go.uber.org/zap"
)

type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

type PluginInterface interface {
	Register(serverAPI any) error
	ParseURL(url string) error
}

type ServerAPI struct {
	logger     *zap.Logger
	routeGroup fiber.Router
	db         *db.DB
}

func (s *ServerAPI) MigrateSchema(ddl string) (*sql.DB, error) {
	if _, err := s.db.Conn.ExecContext(context.Background(), ddl); err != nil {
		return nil, err
	}

	return s.db.Conn, nil
}

func (s *ServerAPI) RegisterViewRoute(path string, title string, content string) {
	// TODO:
}

type ServicePlugin struct {
	Name    *string
	Regex   regexp.Regexp
	Handler func(url string) error
}

type PluginRegister struct {
	Config  *Config
	logger  *zap.Logger
	router  *fiber.App
	db      *db.DB
	Plugins []*ServicePlugin
}

func (p *PluginRegister) RegisterPlugins() {
	p.logger.Debug("Loading plugins...")

	serverAPI := &ServerAPI{
		logger:     p.logger,
		routeGroup: p.router.Group("/services"),
		db:         p.db,
	}

	err := filepath.WalkDir(p.Config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) == ".so" {
			p.logger.Debug("Loading plugin file", zap.String("Path", path))

			service, err := p.loadPlugin(path, serverAPI)
			if err != nil {
				p.logger.Debug("Error loading plugin file", zap.String("Path", path))
			}

			p.Plugins = append(p.Plugins, service)
		}

		return nil
	})

	if err != nil {
		p.logger.Panic("Error walking plugin directory", zap.String("Path", p.Config.Directory))
	}

}

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
	if err != nil {
		return nil, err
	}

	version, err := lookupSymbol[string](symPlugin, "Version")
	if err != nil {
		return nil, err
	}

	regex, err := lookupSymbol[string](symPlugin, "Regex")
	if err != nil {
		return nil, err
	}

	p.db.Q.CreatePlugin(context.Background(), queries.CreatePluginParams{
		Name:    fmt.Sprintf("%s", *name),
		Path:    path,
		Hash:    hash,
		Version: fmt.Sprintf("%s", *version),
	})

	return &ServicePlugin{
		Name:    name,
		Regex:   *regexp.MustCompile(*regex),
		Handler: (service).ParseURL,
	}, nil
}

// Init creates a new PluginRegister instance with the provided dependencies
func Init(config *Config, logger *zap.Logger, router *fiber.App, db *db.DB) *PluginRegister {
	return &PluginRegister{
		Config: config,
		logger: logger,
		router: router,
		db:     db,
	}
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
