package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds the application configuration.
type Config struct {
	Dict string `toml:"dict"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Dict: "dictcc-en-de.db",
	}
}

// Dir returns the platform-specific config directory for dict.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "dict"), nil
}

// DirWithOverride returns the config directory, using override if provided.
func DirWithOverride(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return Dir()
}

// DictsDir returns the path to the dicts subdirectory.
func DictsDir(configDir string) string {
	return filepath.Join(configDir, "dicts")
}

// Path returns the path to the config file.
func Path(configDir string) string {
	return filepath.Join(configDir, "config.toml")
}

// Load reads the config file from the default directory.
func Load() (Config, error) {
	dir, err := Dir()
	if err != nil {
		return DefaultConfig(), err
	}
	return LoadFrom(dir)
}

// LoadFrom reads the config file from the specified directory.
func LoadFrom(configDir string) (Config, error) {
	cfg := DefaultConfig()

	path := Path(configDir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil // Use defaults
	}
	if err != nil {
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save writes the config to the specified directory.
func (c Config) Save(configDir string) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(Path(configDir), data, 0644)
}

// DictPath returns the full path to the configured dictionary.
func (c Config) DictPath(configDir string) string {
	return filepath.Join(DictsDir(configDir), c.Dict)
}
