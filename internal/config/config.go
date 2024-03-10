package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yogin/go-ec2/internal/utils"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFile    = "gosh.yaml"           // DefaultConfigFile is the default configuration file name
	CurrentConfigVersion = 1                     // CurrentConfigVersion is the current configuration version
	DefaultTimeFormat    = "2006-01-02 15:04:05" // DefaultTimeFormat is the default time format
)

var (
	config *Config // config is the global configuration
)

type Config struct {
	Version       int                `json:"version" yaml:"version"`
	Profiles      map[string]Profile `json:"profiles" yaml:"profiles"`
	ShowUTCTime   bool               `json:"show_utc_time" yaml:"show_utc_time"`     // show UTC time (default: false)
	ShowLocalTime bool               `json:"show_local_time" yaml:"show_local_time"` // show local time (default: false)
	TimeFormat    string             `json:"time_format" yaml:"time_format"`         // time format (default: "2006-01-02 15:04:05")

	configPath string
}

type Profile struct {
	Provider       string `json:"provider" yaml:"provider"`                 // aws, gcp, azure
	Name           string `json:"name" yaml:"name"`                         // profile name
	Region         string `json:"region" yaml:"region"`                     // region (us-west-1, us-east-1, etc)
	PreferPublicIP bool   `json:"prefer_public_ip" yaml:"prefer_public_ip"` // prefer public IP over private IP (default: false)
}

func NewConfig(path *string) *Config {
	if config != nil {
		return config
	}

	config = DefaultConfiguration()

	// config = &Config{
	// 	Version:  CurrentConfigVersion,
	// 	Profiles: make(map[string]Profile),
	// }

	if path != nil {
		config.configPath = *path
	}

	if config.findConfigFile() {
		config.Profiles = make(map[string]Profile) // reset default profiles if configuration file is found

		if err := config.loadConfigFile(); err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
	}

	return config
}

func DefaultConfiguration() *Config {
	profile := Profile{
		Provider: "aws",
		Name:     "default",
	}

	profiles := make(map[string]Profile)
	profiles["default"] = profile

	return &Config{
		Version:       CurrentConfigVersion,
		Profiles:      profiles,
		ShowUTCTime:   true,
		ShowLocalTime: true,
		TimeFormat:    DefaultTimeFormat,
	}
}

func (c *Config) loadYAMLConfigFile() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadJSONConfigFile() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadConfigFile() error {
	// try to load yaml configuration file
	if strings.HasSuffix(c.configPath, ".yaml") || strings.HasSuffix(c.configPath, ".yml") {
		return c.loadYAMLConfigFile()
	}

	// try to load json configuration file
	if strings.HasSuffix(c.configPath, ".json") {
		return c.loadJSONConfigFile()
	}

	return errors.New("unsupported configuration file format")
}

func (c *Config) findConfigFile() bool {
	// if the configuration file is provided and exists, use it
	if len(c.configPath) > 0 && utils.IsFile(c.configPath) {
		return true
	}

	// if not found, check for configuration file in the current directory
	if path := fmt.Sprintf(".%s", DefaultConfigFile); utils.IsFile(path) {
		c.configPath = path
		return true
	}

	// if not found, check for configuration file in the home directory
	if home, err := os.UserHomeDir(); err == nil {
		if path := filepath.Join(home, fmt.Sprintf(".%s", DefaultConfigFile)); utils.IsFile(path) {
			c.configPath = path
			return true
		}
	}

	// if not found, check for global configuration file
	if utils.IsDirectory("/etc") {
		if path := fmt.Sprintf("/etc/%s", DefaultConfigFile); utils.IsFile(path) {
			c.configPath = path
			return true
		}
	}

	return false
}
