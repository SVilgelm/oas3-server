package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/SVilgelm/oas3-server/pkg/oas3"
	"github.com/ghodss/yaml"
)

// Config is a main cofing
type Config struct {
	OAS3     string         `json:"oas3,omitempty"`
	Address  string         `json:"address,omitempty"`
	TLS      TLSConfig      `json:"tls,omitempty"`
	Static   string         `json:"static,omitempty"`
	Validate ValidateConfig `json:"validate,omitempty"`

	Model *openapi3.Swagger `json:"-,omitempty"`
}

// TLSConfig is used for tls settings
type TLSConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Cert    string `json:"cert,omitempty"`
	Key     string `json:"key,omitempty"`
}

// ValidateConfig is used for Validation settings
type ValidateConfig struct {
	Request  bool `json:"request,omitempty"`
	Response bool `json:"response,omitempty"`
}

func (c *Config) init() error {
	if c.TLS.Cert == "" || c.TLS.Key == "" {
		c.TLS.Enabled = false
	}
	if c.OAS3 != "" {
		model, err := oas3.Load(c.OAS3)
		if err != nil {
			return fmt.Errorf("invalid model: %s", err.Error())
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
		log.Println(err)
		cfg = &Config{}
		cfg.init()
	}
	return cfg
}
