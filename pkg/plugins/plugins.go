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
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gregmulvaney/forager/pkg/api/http"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/pkg/db/queries"
	"github.com/gregmulvaney/forager/web/components"
	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

type ServicePlugin struct {
	path string
	hash string
}

type PluginRegister struct {
	config  *Config
	logger  *zap.Logger
	db      *db.Db
	router  *fiber.App
	plugins []ServicePlugin
}

type layoutRenderer struct{}

func (l *layoutRenderer) RenderWithLayout(title, content string) ([]byte, error) {
	contentComponent := templ.Raw(content)
	layoutComponent := components.Layout(title)

	var buf []byte
	ctx := context.Background()

	// Create a new context with the content as children
	ctx = templ.WithChildren(ctx, contentComponent)

	// Render the layout with the content
	buffer := &strings.Builder{}
	err := layoutComponent.Render(ctx, buffer)
	if err != nil {
		return nil, err
	}

	buf = []byte(buffer.String())
	return buf, nil
}

type Service interface {
	Register(*sql.DB, fiber.Router, *zap.Logger, layoutRenderer)
}

func Init(config *Config, logger *zap.Logger, dbConn *db.Db, http *http.Server) *PluginRegister {
	return &PluginRegister{
		config: config,
		logger: logger,
		db:     dbConn,
		router: http.Router,
	}
}

func (p *PluginRegister) Register() {
	p.logger.Debug("Loading plugins", zap.String("Directory", p.config.Directory))

	serviceRouteGroup := p.router.Group("/service")

	err := filepath.WalkDir(p.config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ".so" {
			p.logger.Debug("Attempting to load plugin located at", zap.String("Path", path))
			// Hash plugin file
			hash, err := hashPlugin(path)
			if err != nil {
				p.logger.Panic("Failed to hash plugin file", zap.Error(err))
			}

			// Open the plugin
			symPlugin, err := plugin.Open(path)
			if err != nil {
				p.logger.Panic("Failed to open plugin", zap.String("Plugin path", path), zap.Error(err))
			}

			// Lookup Service interface
			service, err := lookupSymbol[Service](symPlugin, "Service")
			if err != nil {
				p.logger.Panic("Failed to lookup plugin service interface", zap.Error(err))
			}

			// Lookup plugin name
			name, err := lookupSymbol[string](symPlugin, "ServiceName")
			if err != nil {
				p.logger.Panic("Faild to look up plugin name", zap.Error(err))
			}

			// Lookup default url
			homePath, err := lookupSymbol[string](symPlugin, "ServiceDefaultPath")
			if err != nil {
				p.logger.Panic("Failed to look up plugin default url")
			}

			var renderer layoutRenderer
			(*service).Register(p.db.Conn, serviceRouteGroup, p.logger, renderer)

			ctx := context.Background()

			_, err = p.db.Q.CreatePlugin(ctx, queries.CreatePluginParams{
				Name:     fmt.Sprintf("%s", *name),
				Path:     path,
				Hash:     hash,
				HomePath: fmt.Sprintf("%s", *homePath),
			})

			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					p.logger.Debug("Plugin already exists in database", zap.String("plugin", *name), zap.String("path", path))
				} else {
					p.logger.Debug("Failed to insert plugin data", zap.Error(err))
				}
			}

		}

		return nil
	})

	if err != nil {
		p.logger.Panic("Failed to walk plugin directory", zap.Error(err))
	}

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
