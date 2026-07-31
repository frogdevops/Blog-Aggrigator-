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

func Read() (Config, error) {
	config, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	data, err := os.ReadFile(config)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil

}

func SetUser(c Config) {
	c.CurrentUserName = fmt.Sprintf("")
}
