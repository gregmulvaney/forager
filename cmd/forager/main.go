package main

import (
	"fmt"
	"github.com/gregmulvaney/forager/pkg/api/http"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

func main() {
	// Enable flags
	flagSet := pflag.NewFlagSet("default", pflag.ContinueOnError)
	// Plugins config
	flagSet.String("plugin-dir", "./plugins", "Plugin file directory")
	// HTTP config
	flagSet.String("host", "0.0.0.0", "HTTP service address")
	flagSet.Int("port", 3080, "HTTP plaintext service port")
	flagSet.Int("secure-port", 3443, "HTTP secure service port")
	// TODO: Add log level configuration

	viper.BindPFlags(flagSet)

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	viper.Set("hostname", hostname)
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

	logger, err := initZap()
	defer logger.Sync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
	}
	stdLog := zap.RedirectStdLog(logger)
	defer stdLog()

	// Unmarshal http flags to config struct
	var httpConfig *http.Config
	if err := viper.Unmarshal(&httpConfig); err != nil {
		panic(err)
	}
	httpServer := http.Init(httpConfig, logger)

	httpServer.Serve()
}

func initZap() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
