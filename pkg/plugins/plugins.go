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

	"github.com/gofiber/fiber/v3"
	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/gregmulvaney/forager/pkg/db/queries"
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
	router  *fiber.App
	db      *db.Db
	plugins []ServicePlugin
}

type Service interface {
	Register(*sql.DB)
}

func Init(config *Config, logger *zap.Logger, dbConn *db.Db) *PluginRegister {
	return &PluginRegister{
		config: config,
		logger: logger,
		db:     dbConn,
	}
}

func (p *PluginRegister) Register() {
	p.logger.Debug("Loading plugins", zap.String("Directory", p.config.Directory))

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

			(*service).Register(p.db.Conn)

			ctx := context.Background()

			p.db.Q.CreatePlugin(ctx, queries.CreatePluginParams{
				Name: fmt.Sprintf("%s", *name),
				Hash: hash,
			})

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
