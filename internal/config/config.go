package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	BURL            string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigPath() (string, error) {
	config, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	path := config + "/.gatorconfig.json"

	return path, nil
}

func Read() (Config, error) {

	path, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil

}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	path, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("get config path: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
