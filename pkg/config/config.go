package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/SVilgelm/oas3-server/pkg/oas3"
	"gopkg.in/yaml.v2"
)

// Config is a main cofing
type Config struct {
	OAS3    string    `yaml:"oas3,omitempty" json:"oas3,omitempty"`
	Address string    `yaml:"address,omitempty" json:"address,omitempty"`
	TLS     TLSConfig `yaml:"tls,omitempty" json:"tls,omitempty"`
	Static  string    `yaml:"static,omitempty" json:"static,omitempty"`

	Model *openapi3.Swagger `yaml:"-,omitempty" json:"-,omitempty"`
}

// TLSConfig is used to cofigure tls settings
type TLSConfig struct {
	Enabled bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Cert    string `yaml:"cert,omitempty" json:"cert,omitempty"`
	Key     string `yaml:"key,omitempty" json:"key,omitempty"`
}

func (c *Config) init() error {
	if c.TLS.Cert == "" || c.TLS.Key == "" {
		c.TLS.Enabled = false
	}
	if c.OAS3 != "" {
		model, err := oas3.Load(c.OAS3)
		if err != nil {
			return err
		}
		c.Model = model
	}
	if c.Address == "" {
		c.Address = "0.0.0.0:8000"
	}
	return nil
}

// Load loads a config file
func Load(fileName string) (*Config, error) {
	var cfg Config
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal the file '%s'. %s", fileName, err)
	}
	log.Println("Loaded config file:", fileName)
	err = cfg.init()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SafeLoad loads a config file and omit any errors by setting the default values
func SafeLoad(fileName string) *Config {
	cfg, err := Load(fileName)
	if err != nil {
		cfg = &Config{}
		cfg.init()
	}
	return cfg
}
