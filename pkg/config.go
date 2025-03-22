package pkg

import (
	"io"
	"log/slog"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GraphQLServer Server   `yaml:"graphql_server"`
	CommandServer Server   `yaml:"command_server"`
	QueryServer   Server   `yaml:"query_server"`
	Database      DB       `yaml:"database"`
	Cache         Memcache `yaml:"memcache"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DB struct {
	Username string `yaml:"username"`
	// Token string `yaml:"token"`
	Path string `yaml:"path"`
}

type Memcache struct {
	Host string `yaml:"hostname"`
	Port string `yaml:"port"`
}

func (c *Config) LoadFile(file io.Reader) error {
	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read file", "error", err)
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		slog.Error("Failed to unmarshal data", "error", err)
		return err

	}

	return nil
}
