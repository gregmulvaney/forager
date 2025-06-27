package plugins

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"plugin"
	"regexp"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

type PluginInterface interface {
	Register(serverAPI any)
}

type ServerAPI struct {
	logger     *zap.Logger
	routeGroup fiber.Router
}

func (s *ServerAPI) MigrateSchema(ddl string) {
	// TODO:
}

func (s *ServerAPI) RegisterRoute(path string, method string) {
	s.logger.Debug("Yay")
}

type ServicePlugin struct {
	Name    string
	Regex   regexp.Regexp
	Handler func(url string)
}

type PluginRegister struct {
	config  *Config
	logger  *zap.Logger
	router  *fiber.App
	Plugins []*ServicePlugin
}

// Init creates a new PluginRegister instance with the provided dependencies
func Init(config *Config, logger *zap.Logger, router *fiber.App) *PluginRegister {
	return &PluginRegister{
		config: config,
		logger: logger,
		router: router,
	}
}

func (p *PluginRegister) RegisterPlugins() {
	p.logger.Debug("Attempting to load plugins")

	serverAPI := &ServerAPI{
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

func (p *PluginRegister) loadPlugin(path string, serverAPI *ServerAPI) (*ServicePlugin, error) {
	symPlugin, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	symService, err := symPlugin.Lookup("Service")
	if err != nil {
		return nil, err
	}

	var service PluginInterface
	service = symService.(PluginInterface)

	(service).Register(serverAPI)

	return &ServicePlugin{}, nil
}

func hashPlugin() {
	// TODO:
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
