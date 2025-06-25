package plugins

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/web/components"
	"go.uber.org/zap"
)

type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

// Server API interface
type ServerApiInterface interface {
	MigrateSchema(ddl string) error
	RegisterRoute(path string, title string, content string)
}

type ServerAPI struct {
	db         *db.Db
	logger     *zap.Logger
	routeGroup fiber.Router
}

func (s *ServerAPI) MigrateSchema(ddl string) error {
	if _, err := s.db.Conn.ExecContext(context.Background(), ddl); err != nil {
		s.logger.Debug("Failed to migrate schema", zap.Error(err))
		return err
	}
	return nil
}

// TODO: Change to allow method selections
func (s *ServerAPI) RegisterRoute(path string, title string, content string) {
	contentComponent := templ.Raw(content)
	layoutComponent := components.Layout(title)

	s.routeGroup.Get(path, func(c fiber.Ctx) error {
		ctx := context.Background()
		ctx = templ.WithChildren(ctx, contentComponent)

		c.Set("Content-Type", "text/html")
		return layoutComponent.Render(ctx, c.Response().BodyWriter())
	})
}

type PluginRegister struct {
	config *Config
	logger *zap.Logger
	db     *db.Db
	router *fiber.App
}

func (p *PluginRegister) RegisterPlugins() {
	p.logger.Debug("Attempting to load plugins in", zap.String("Directory", p.config.Directory))

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
			p.loadPlugin(path, serverAPI)
		}

		return nil
	})

	if err != nil {
		p.logger.Error("Failed to walk plugin directory", zap.String("Directory", p.config.Directory), zap.Error(err))
	}
}

func (p *PluginRegister) loadPlugin(path string, serverAPI *ServerAPI) {
	p.logger.Debug("Attempting to load plugin at", zap.String("Path", path))

	// Hash plugin file
	_, err := hashPlugin(path)
	if err != nil {
		p.logger.Error("Failed to hash plugin file", zap.String("Path", path), zap.Error(err))
		return
	}

	// Open plugin file
	symPlugin, err := plugin.Open(path)
	if err != nil {
		p.logger.Error("Failed to open plugin file", zap.String("Path", path), zap.Error(err))
		return
	}

	symRegister, err := symPlugin.Lookup("Register")
	if err != nil {
		p.logger.Error("Failed to lookup Register symbol", zap.String("Path", path), zap.Error(err))
		return
	}

	registerFunc := symRegister.(func(ServerApiInterface))
	registerFunc(serverAPI)

	_, err = lookupSymbol[string](symPlugin, "ServiceName")
	if err != nil {
		p.logger.Error("Failed to look up plugin name symbol", zap.String("Path", path))
	}
}

func Init(config *Config, db *db.Db, logger *zap.Logger, router *fiber.App) *PluginRegister {
	return &PluginRegister{
		config: config,
		logger: logger,
		db:     db,
		router: router,
	}
}

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
		return &result, nil
	default:
		return nil, fmt.Errorf("Unexpected type from module symbol %T", symbol)
	}
}
