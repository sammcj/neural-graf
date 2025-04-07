package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	API      APIConfig      `mapstructure:"api"`
	Dgraph   DgraphConfig   `mapstructure:"dgraph"`
	MCP      MCPConfig      `mapstructure:"mcp"`
	Shutdown ShutdownConfig `mapstructure:"shutdown"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

// APIConfig contains API server settings
type APIConfig struct {
	Port int `mapstructure:"port"`
}

// DgraphConfig contains Dgraph connection settings
type DgraphConfig struct {
	Address string `mapstructure:"address"`
}

// MCPConfig contains MCP server settings
type MCPConfig struct {
	UseSSE  bool   `mapstructure:"useSSE"`
	Address string `mapstructure:"address"`
}

// ShutdownConfig contains graceful shutdown settings
type ShutdownConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
}

// LoadConfig loads the configuration from a file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read config file
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			// It's okay if config file doesn't exist
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Read from environment variables
	v.SetEnvPrefix("MCPGRAPH")
	v.AutomaticEnv()

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "MCP-Graph")
	v.SetDefault("app.version", "0.1.0")

	// API defaults
	v.SetDefault("api.port", 8080)

	// Dgraph defaults
	v.SetDefault("dgraph.address", "localhost:9080")

	// MCP defaults
	v.SetDefault("mcp.useSSE", true)
	v.SetDefault("mcp.address", ":3000")

	// Shutdown defaults
	v.SetDefault("shutdown.timeout", 5*time.Second)
}

// SaveConfigExample saves an example configuration file
func SaveConfigExample(path string) error {
	v := viper.New()
	setDefaults(v)

	return v.WriteConfigAs(path)
}
