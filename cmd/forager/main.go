package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gregmulvaney/forager/pkg/api/http"
	"github.com/gregmulvaney/forager/pkg/plugins"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Enable config flags
	flagSet := pflag.NewFlagSet("default", pflag.ContinueOnError)
	// Plugins config
	flagSet.String("plugins-dir", "./plugins", "Storage directory for plugins")
	// HTTP config
	flagSet.String("host", "0.0.0.0", "HTTP service address")
	flagSet.Int("port", 3000, "HTTP plaintext port")
	flagSet.Int("secure-port", 3443, "HTTP secure port")
	// Log options
	flagSet.String("log-level", "debug", "Log level: debug, info, warn, error, fatal, or panic")

	viper.BindPFlags(flagSet)

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	viper.Set("hostname", hostname)
	// TODO: Pull the version number from a centralized location
	viper.Set("version", "0.0.1")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	versionFlag := flagSet.BoolP("version", "v", false, "Print service version")

	err = flagSet.Parse(os.Args[1:])
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

	var httpConfig *http.Config
	if err := viper.Unmarshal(&httpConfig); err != nil {
		panic(err)
	}

	// Initialize HTTP server struct
	httpSrv, _ := http.Init(httpConfig, logger)

	// Unmarshal plugin config
	var pluginConfig *plugins.Config
	if err := viper.Unmarshal(&pluginConfig); err != nil {
		panic(err)
	}

	// Initialize plugins
	err = plugins.Init(pluginConfig, &httpSrv, logger)
	if err != nil {
		panic(err)
	}

	// Start HTTP server
	httpSrv.Serve()
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
