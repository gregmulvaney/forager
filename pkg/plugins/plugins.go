package plugins

import (
	"errors"
	"fmt"
	"io/fs"
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
}

type PluginRegister struct {
	plugins []ServicePlugin
}

type Service interface {
	Register(*db.Db, *zap.Logger)
}

func RegisterPlugins(config *Config, db *db.Db, logger *zap.Logger) {
	err := filepath.WalkDir(config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) == ".so" {
			symPlugin, err := plugin.Open(path)
			if err != nil {
				return err
			}

			service, err := lookupSymbol[Service](symPlugin, "Service")
			if err != nil {
				logger.Panic("Failed symbol look up for plugin interface", zap.String("Plugin Path", path), zap.Error(err))
				return nil
			}

			(*service).Register(db, logger)
		}
		return nil
	})

	if err != nil {
		logger.Panic("Failed to load plugin files", zap.Error(err))
	}
}

func lookupSymbol[T any](plugin *plugin.Plugin, name string) (*T, error) {
	symbol, err := plugin.Lookup(name)
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
