package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gregmulvaney/forager/pkg/api/http"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Define config options
	fs := pflag.NewFlagSet("default", pflag.ContinueOnError)
	// Plugins config
	fs.String("plugins-dir", "./plugins", "Directory where plugins are stored")
	// HTTP config
	fs.String("host", "0.0.0.0", "HTTP service host address")
	fs.Int("port", 3000, "HTTP service plaintext port")
	fs.Int("secure-port", 3443, "HTTP service secure port")
	// fs.String("certificate", "yeet", "HTTP service certificate")
	// Logs config
	fs.String("log-level", "debug", "Log level: debug, info, warn, error, fatal, or panic")

	viper.BindPFlags(fs)

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	viper.Set("hostname", hostname)
	// TODO: Centralize the version number
	viper.Set("version", "0.0.1")
	// Replace all dashes with underscores for environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	// Load environment variables
	viper.AutomaticEnv()

	versionFlag := fs.BoolP("version", "v", false, "Print service version")
	err = fs.Parse(os.Args[1:])
	switch {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "Error: %s \n", err)
	case *versionFlag:
		fmt.Println(viper.GetString("version"))
		os.Exit(0)
	}

	// Initialize logger
	logger, err := initZap(viper.GetString("log-level"))
	defer logger.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
	}
	stdLog := zap.RedirectStdLog(logger)
	defer stdLog()

	// Unmarshall http config
	var httpConfig *http.Config
	if err := viper.Unmarshal(&httpConfig); err != nil {
		logger.Panic("Failed to unmarshall HTTP config", zap.Error(err))
	}

	httpServer := http.Init(httpConfig, logger)

	httpServer.Serve()
}

func initZap(logLevel string) (*zap.Logger, error) {
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
		Encoding:         "console",
		EncoderConfig:    zapEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()

}
