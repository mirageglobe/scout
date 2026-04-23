package filesystem

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ThemeIdx int `json:"theme_idx"`
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "scout")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config"), nil
}

func LoadConfig() (Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return Config{ThemeIdx: 0}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{ThemeIdx: 0}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{ThemeIdx: 0}, nil
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
