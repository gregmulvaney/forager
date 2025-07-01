package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

}
