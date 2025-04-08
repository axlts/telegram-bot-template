package config

import (
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Bot `yaml:"bot"`
}

func Parse(filepath string) (cfg *Config, err error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &cfg)
	return
}

func (c *Config) Validate() error {
	return c.Bot.validate() //nolint:staticcheck
}
