package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DBUrl			string	`json:"db_url"`
	CurrentUsername	string	`json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, configFileName), nil
}

func Read() (Config, error) {
	//I will change it from home directory
	path, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	fileData, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(fileData, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func write(cfg Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	fileData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, fileData, 0o777)
}

func (c *Config) SetUser(username string) (error) {
	c.CurrentUsername = username
	return write(*c)
}

