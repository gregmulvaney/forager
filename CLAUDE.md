# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
- `make run` - Run the forager application
- `go run cmd/forager/main.go` - Direct Go execution of the main application

### Building
- `make build-plugin` - Build a plugin example as a shared object (.so file)
- `go build -buildmode=plugin -o tmp/plugin/example.so ./plugin-example/main.go` - Build plugins manually

### Database Operations
- `make sqlc` - Generate Go code from SQL queries using sqlc
- `sqlc generate -f ./sqlc/sqlc.yaml` - Direct sqlc generation

### Frontend Assets
- `make tailwind` - Build CSS using Tailwind
- `tailwindcss -o ./web/static/style.css -i ./tailwind.css --minify` - Direct Tailwind compilation

## Architecture Overview

This is a Go-based HTTP service with a plugin architecture, built around:

### Core Components
- **Main Application**: `cmd/forager/main.go` - Entry point with configuration management using Viper and pflag
- **HTTP Server**: `pkg/api/http/server.go` - Fiber-based web server with middleware support
- **Database Layer**: `pkg/db/` - SQLite database with sqlc-generated queries
- **Plugin System**: `pkg/plugins/plugins.go` - Dynamic plugin loading system using Go's plugin package

### Plugin Architecture
- Plugins are built as shared objects (.so files) and loaded dynamically at runtime
- Plugin interface defined in `pkg/plugins/plugins.go` with `Service` interface
- Example plugin in `plugin-example/` directory shows the pattern
- Plugins register themselves and receive database and logger instances

### Database Setup
- Uses SQLite with schema defined in `sqlc/schema.sql`
- Queries defined in `sqlc/queries.sql` and generated to `pkg/db/queries/`
- Database initialization happens in `pkg/db/db.go` with embedded schema
- sqlc configuration in `sqlc/sqlc.yaml`

### Configuration
- Uses Viper for configuration management with pflag for CLI flags
- Environment variables automatically loaded (dashes converted to underscores)
- Default config: host=0.0.0.0, port=3000, secure-port=3443, plugins-dir=./plugins
- Structured logging with Zap logger

### Web Assets
- Static files served from `web/static/`
- CSS built with Tailwind CSS
- Assets embedded using Go's embed directive

## Key Patterns

### Plugin Development
- Implement the `Service` interface with a `Register(*db.Db, *zap.Logger)` method
- Export a global `Service` variable for plugin loading
- Use embedded SQL files for plugin-specific database operations

### Database Operations
- All database operations go through sqlc-generated queries
- Schema changes require running `make sqlc` to regenerate Go code
- Plugins can register their own SQL during initialization