package gateway

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	JwtSecret string               `yaml:"secret"`
	Basics    map[string]BasicAuth `yaml:"basic"`
	Routes    []Route              `yaml:"routes"`
	Targets   map[string]string    `yaml:"targets"`
}

type Route struct {
	Path   string `yaml:"path"`
	Prefix bool   `yaml:"prefix"`
	Target string `yaml:"target"`
}

type BasicAuth struct {
	Password string `yaml:"password"`
	Tenant   string `yaml:"tenant"`
}

func (cfg *Config) Load(loadPath string) error {
	data, err := ioutil.ReadFile(loadPath)
	if err != nil {
		return fmt.Errorf("error with read config file: %w", err)
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error with unmarshal data: %v", err)
	}

	return nil
}
