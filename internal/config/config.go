package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

// Config represents the structure of the entire configuration file
type Config struct {
	ConfigFilePath string
	StoragePath    string
	GitEnabled     bool
	Dotfiles       map[string]string `toml:"dotfiles"`
}

// GetConfigFilePath returns the path to the config file
func (c *Config) GetConfigFilePath() string {
	return c.ConfigFilePath
}

// LoadDotfilesConfig loads the dotfiles from the config file
func LoadDotfilesConfig(configPath string) (map[string]string, error) {
	tree, err := toml.LoadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	dotfilesTree := tree.Get("dotfiles")
	if dotfilesTree == nil {
		return make(map[string]string), nil
	}

	dotfiles := make(map[string]string)
	err = dotfilesTree.(*toml.Tree).Unmarshal(&dotfiles)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling dotfiles: %w", err)
	}

	return dotfiles, nil
}

// AddDotfileToConfig adds a new dotfile to the config file
func AddDotfileToConfig(configPath string, name, path string) error {
	tree, err := toml.LoadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			tree, _ = toml.Load("")
		} else {
			return fmt.Errorf("error loading config file: %w", err)
		}
	}

	dotfilesTree, ok := tree.Get("dotfiles").(*toml.Tree)
	if !ok {
		dotfilesTree, _ = toml.Load("")
		tree.Set("dotfiles", dotfilesTree)
	}

	dotfilesTree.Set(name, path)

	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	err = encoder.Encode(tree)
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	return nil
}

// SaveConfig saves the config to the file
func SaveConfig(configFilePath string, config Config) error {
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	return os.WriteFile(configFilePath, data, 0664)
}
