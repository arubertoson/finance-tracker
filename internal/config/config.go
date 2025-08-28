package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// AppName is used as the application identifier for XDG paths
	AppName = "finance"

	// DefaultDBName specifies the default SQLite database filename.
	// This file will be created in the configured database directory
	// if no custom path is provided in the configuration.
	DefaultDBName = "finance.db"

	// DefaultConfigFile specifies the JSON configuration filename.
	// This file stores user preferences and application settings,
	// and will be automatically created with default values if not present.
	DefaultConfigFile = "config.json"

	// DefaultXDGConfigDir specifies the default XDG config location if
	// XDG_CONFIG_HOME is not set (~/.config/finance)
	defaultXDGConfigDir = ".config"

	// DefaultXDGDataDir specifies the default XDG data location if
	// XDG_DATA_HOME is not set (~/.local/share/finance)
	defaultXDGDataDir = ".local/share"
)

// Config holds the application configuration settings
type Config struct {
	DBPath string `json:"db_path"`
}

var defaultConfig = Config{
	DBPath: filepath.Join(GetDataDir(), DefaultDBName),
}

// Load reads the configuration from the XDG config directory.
// If the config file doesn't exist, it returns the default configuration.
//
// Returns:
//   - *Config: A pointer to the loaded configuration
//   - error: An error if the loading process fails
func Load() (*Config, error) {
	configPath := filepath.Join(GetConfigDir(), DefaultConfigFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to the XDG config directory.
// It creates the config directory if it doesn't exist and writes the
// configuration in a JSON format with proper indentation.
//
// Returns:
//   - error: An error if the saving process fails
func (c *Config) Save() error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDir, DefaultConfigFile)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigDir returns the XDG-compliant configuration directory.
// It checks XDG_CONFIG_HOME environment variable first, falling back
// to ~/.config/finance if not set.
func GetConfigDir() string {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		xdgConfig = filepath.Join(homeDir, defaultXDGConfigDir)
	}
	return filepath.Join(xdgConfig, AppName)
}

// GetDataDir returns the XDG-compliant data directory.
// It checks XDG_DATA_HOME environment variable first, falling back
// to ~/.local/share/finance if not set.
func GetDataDir() string {
	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		xdgData = filepath.Join(homeDir, defaultXDGDataDir)
	}
	return filepath.Join(xdgData, AppName)
}
