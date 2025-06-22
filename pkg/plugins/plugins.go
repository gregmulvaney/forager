package plugins

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"

	"github.com/gofiber/fiber/v3"
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
	plugins []ServicePlugin
}

type Service interface {
	Register()
}

func Init(config *Config, logger *zap.Logger) *PluginRegister {
	return &PluginRegister{
		config: config,
		logger: logger,
	}
}

func (p *PluginRegister) Register() {
	p.logger.Debug("Loading plugins", zap.String("Directory", p.config.Directory))

	err := filepath.WalkDir(p.config.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ".so" {
			// Hash plugin file
			_, err := hashPlugin(path)
			if err != nil {
				p.logger.Panic("Failed to hash plugin file", zap.Error(err))
			}

			// Open the plugin
			_, err = plugin.Open(path)
			if err != nil {
				p.logger.Panic("Failed to open plugin", zap.String("Plugin path", path), zap.Error(err))
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
