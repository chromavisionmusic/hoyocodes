package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot BotConfig `yaml:"bot"`
}

type User struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type Links struct {
	Invite  string `yaml:"invite"`
	Website string `yaml:"website"`
	Server  string `yaml:"server"`
	Github  string `yaml:"github"`
}

type BotConfig struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version"`
	Owner        User     `yaml:"owner"`
	Contributors []User   `yaml:"contributors,omitempty"`
	Links        Links    `yaml:"links"`
	Presence     Presence `yaml:"presence"`
}

type Presence struct {
	Status       string `yaml:"status"`
	ActivityType string `yaml:"activity_type"`
	Activity     string `yaml:"activity"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if cfg.Bot.Name == "" {
		return nil, fmt.Errorf("bot name is required")
	}

	if cfg.Bot.Version == "" {
		return nil, fmt.Errorf("bot version is required")
	}

	return &cfg, nil
}
