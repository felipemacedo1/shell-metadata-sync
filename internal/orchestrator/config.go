package orchestrator

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string   `yaml:"version"`
	Name    string   `yaml:"name"`
	Sources Sources  `yaml:"sources"`
	Storage Storage  `yaml:"storage"`
	Pipelines map[string]Pipeline `yaml:"pipelines"`
	Logging Logging  `yaml:"logging"`
	Monitoring Monitoring `yaml:"monitoring"`
}

type Sources struct {
	Github GithubSource `yaml:"github"`
}

type GithubSource struct {
	Users      []string  `yaml:"users"`
	RateLimit  RateLimit `yaml:"rate_limit"`
}

type RateLimit struct {
	Enabled        bool `yaml:"enabled"`
	MaxRetries     int  `yaml:"max_retries"`
	BackoffSeconds int  `yaml:"backoff_seconds"`
}

type Storage struct {
	MongoDB MongoDBStorage `yaml:"mongodb"`
	JSON    JSONStorage    `yaml:"json"`
}

type MongoDBStorage struct {
	Enabled     bool              `yaml:"enabled"`
	URIEnv      string            `yaml:"uri_env"`
	Database    string            `yaml:"database"`
	Collections map[string]string `yaml:"collections"`
}

type JSONStorage struct {
	Enabled   bool     `yaml:"enabled"`
	OutputDir string   `yaml:"output_dir"`
	Files     []string `yaml:"files"`
}

type Pipeline struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Schedule    string  `yaml:"schedule"`
	Enabled     bool    `yaml:"enabled"`
	Stages      []Stage `yaml:"stages"`
}

type Stage struct {
	Name     string `yaml:"name"`
	Parallel bool   `yaml:"parallel"`
	Tasks    []Task `yaml:"tasks"`
}

type Task struct {
	Type   string                 `yaml:"type"`
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
	Retry  int                    `yaml:"retry"`
}

type Logging struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type Monitoring struct {
	Enabled bool     `yaml:"enabled"`
	Metrics []string `yaml:"metrics"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand environment variables
	if config.Storage.MongoDB.URIEnv != "" {
		if uri := os.Getenv(config.Storage.MongoDB.URIEnv); uri != "" {
			// Store URI in a private field if needed
		}
	}

	return &config, nil
}

func (c *Config) GetUsers() []string {
	return c.Sources.Github.Users
}

func (c *Config) GetPipeline(name string) (Pipeline, bool) {
	p, ok := c.Pipelines[name]
	return p, ok
}
