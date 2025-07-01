package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gregmulvaney/forager/pkg/db"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Define config options
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	// Plugins config
	fs.String("plugins-dir", "./plugins", "Directory where plugin files are stored")
	// HTTP config
	fs.String("host", "0.0.0.0", "HTTP service host address")
	fs.Int("port", 3000, "HTTP service plaintext port")
	fs.Int("secure-port", 3443, "HTTP service secure port")
	fs.String("cert-path", "./cert.pem", "HTTP certificate")
	// Logs config
	fs.String("log-level", "debug", "Set log level")
	fs.String("log-mode", "console", "Log mode")

	versionFlag := fs.BoolP("version", "v", false, "Display version number")

	version := "0.0.1"

	// Parse flags
	err := fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		fs.PrintDefaults()
		os.Exit(2)
	case *versionFlag:
		fmt.Println(version)
		os.Exit(0)
	}

	// Bind flags and environment variables
	viper.BindPFlags(fs)

	// Query OS hostname
	hostname, _ := os.Hostname()
	viper.Set("hostname", hostname)
	// Replace all dashes with hyphens for environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Initialize zap logger
	logger, err := initZap(viper.GetString("log-level"), viper.GetString("log-mode"))
	defer logger.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
	}
	stdLog := zap.RedirectStdLog(logger)
	defer stdLog()

	// Initialize DB
	_ = db.Init(logger)
}

func initZap(logLevel string, logMode string) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// Switch log levels based on flags
	switch logLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case "panic":
		level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	}

	zapEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         logMode,
		EncoderConfig:    zapEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()

}
