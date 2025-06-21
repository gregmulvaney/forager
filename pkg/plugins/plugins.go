package plugins

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"

	"github.com/gregmulvaney/forager/pkg/db"
	"go.uber.org/zap"
)

type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

type ServicePlugin struct {
	plugin *plugin.Plugin
	name   string
	path   string
	hash   string
}

type PluginRegister struct {
	plugins []ServicePlugin
}

type Service interface {
	Register(*sql.DB, *zap.Logger)
}

func hashFile(path string) (string, error) {
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

func RegisterPlugins(config *Config, db *db.Db, logger *zap.Logger) {
	logger.Debug("Loading plugins", zap.String("Directory", config.Directory))

	// Walk plugins dir and register plugins
	err := filepath.WalkDir(config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) == ".so" {
			hash, err := hashFile(path)
			if err != nil {
				logger.Error("Failed to hash plugin file", zap.String("Path", path), zap.Error(err))
				return nil
			}

			symPlugin, err := plugin.Open(path)
			if err != nil {
				return err
			}

			name, err := lookupSymbol[string](symPlugin, "ServiceName")
			if err != nil {
				logger.Error("Failed plugin symbol lookup", zap.String("Path", path), zap.Error(err))
				return nil
			}

			service, err := lookupSymbol[Service](symPlugin, "Service")
			if err != nil {
				logger.Panic("Failed symbol look up for plugin interface", zap.String("Plugin Path", path), zap.Error(err))
				return nil
			}

			(*service).Register(db.Conn, logger)

			// TODO: make this cleaner
			logger.Debug("Successfully loaded plugin",
				zap.String("Name", func() string {
					if name != nil {
						return *name
					}
					return "nil"
				}()),
				zap.String("Path", path),
				zap.String("Hash", hash),
			)
		}
		return nil
	})

	if err != nil {
		logger.Panic("Failed to load plugin files", zap.Error(err))
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
		return nil, errors.New(fmt.Sprintf("Unexpected type from module symbol %T", symbol))
	}
}
