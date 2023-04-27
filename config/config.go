package config

import (
	"github.com/pelletier/go-toml"
	"os"
)

type Database struct {
	RootUser     string `toml:"root_user"`
	RootPassword string `toml:"root_password"`
	Host         string `toml:"host"`
	Port         string `toml:"port"`
}

type Config struct {
	Database    Database      `toml:"database"`
	Connection  Connection    `toml:"connection"`
	Credentials []Credentials `toml:"credentials"`
}

type Connection struct {
	Port string `toml:"port"`
	Host string `toml:"host"`
}

type Credentials struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// NewConfigFromTOML reads from contents from the given fileName, and unmarshalls it to
// Config struct.
func NewConfigFromTOML(fileName string) (*Config, error) {
	configFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var config Config
	err = toml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
