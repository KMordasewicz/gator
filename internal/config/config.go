package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name,omitempty"`
}

func getConfigFilePath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cofigPath := path + "/" + configFileName
	return cofigPath, nil
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err = json.Unmarshal(content, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func write(config Config) error {
	content, err := json.Marshal(config)
	if err != nil {
		return err
	}
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(path, content, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (config *Config) SetUser(name string) error {
	config.CurrentUserName = name
	return write(*config)
}
