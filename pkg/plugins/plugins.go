package plugins

import (
	"errors"
	"os"
	"path/filepath"
	"plugin"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Config struct {
	Directory string `mapstructure:"plugins-dir"`
}

type ServicePlugin interface {
	RegisterRoutes(*fiber.App)
}

func Register(config *Config, app *fiber.App, logger *zap.Logger) {

	// Walk plugins directory for plugin files
	err := filepath.WalkDir(config.Directory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".so" {
			// Open plugin
			servicePlugin, err := plugin.Open(path)
			if err != nil {
				return err
			}

			// Load the plugin interface
			symPlugin, err := lookupSymbol[ServicePlugin](servicePlugin, "ServicePlugin")
			if err != nil {
				return err
			}

			logger.Debug("PLUGIN: loaded 'example' plugin")

			(*symPlugin).RegisterRoutes(app)

		}
		return nil
	})

	if err != nil {
		panic(err)
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
		return nil, errors.New("Unexpected type")
	}
}
