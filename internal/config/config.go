package cfg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func WriteConfig(cfg Config) error {
	usr_homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("homedir not found")
	}

	config_json, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	config_path := filepath.Join(usr_homedir, configFileName)
	return os.WriteFile(config_path, config_json, 0o600)
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username

	return WriteConfig(*cfg)
}

func ReadConfig() (Config, error) {
	usr_homedir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("homedir not found")
	}
	config_path := filepath.Join(usr_homedir, configFileName)

	config_file, err := os.ReadFile(config_path)
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = json.Unmarshal(config_file, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
