package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigInitEmpty(t *testing.T) {
	t.Parallel()
	cfg := Config{}
	err := cfg.init()
	assert.NoError(t, err)

	assert.Empty(t, cfg.OAS3)
	assert.Nil(t, cfg.Model)

	assert.Equal(t, "0.0.0.0:8000", cfg.Address)

	assert.Empty(t, cfg.TLS.Key)
	assert.Empty(t, cfg.TLS.Cert)
	assert.False(t, cfg.TLS.Enabled)

	assert.False(t, cfg.Validate.Request)
	assert.False(t, cfg.Validate.Response)
}

func TestConfigInitTLSEnabled(t *testing.T) {
	t.Parallel()
	cfg := Config{
		TLS: TLS{
			Key:     "key",
			Cert:    "cert",
			Enabled: true,
		},
	}
	err := cfg.init()
	assert.NoError(t, err)
	assert.Equal(t, "key", cfg.TLS.Key)
	assert.Equal(t, "cert", cfg.TLS.Cert)
	assert.True(t, cfg.TLS.Enabled)

	cfg.TLS.Enabled = false
	err = cfg.init()
	assert.NoError(t, err)
	assert.False(t, cfg.TLS.Enabled)

	cfg.TLS.Key = ""
	cfg.TLS.Cert = ""
	cfg.TLS.Enabled = true
	err = cfg.init()
	assert.NoError(t, err)
	assert.False(t, cfg.TLS.Enabled)
}

func TestConfigInitAddress(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Address: "localhost",
	}
	err := cfg.init()
	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Address)

	cfg.Address = ""
	err = cfg.init()
	assert.NoError(t, err)
	assert.Equal(t, "0.0.0.0:8000", cfg.Address)
}

func TestConfigInitOAS3(t *testing.T) {
	t.Parallel()
	cfg := Config{
		OAS3: "",
	}
	err := cfg.init()
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.OAS3)
	assert.Nil(t, cfg.Model)

	cfg.OAS3 = "fake-file"
	err = cfg.init()
	t.Log(err)
	assert.Error(t, err)
	assert.Nil(t, cfg.Model)

	cfg.OAS3 = "testdata/model.yaml"
	err = cfg.init()
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Model)
	assert.Equal(t, "9.9.9", cfg.Model.Info.Version)
}

func TestLoad(t *testing.T) {
	t.Parallel()
	cfg, err := Load("fake-file")
	t.Log(err)
	assert.Nil(t, cfg)
	assert.EqualError(t, err, "open fake-file: no such file or directory")

	cfg, err = Load("testdata/bad config.yaml")
	t.Log(err)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to unmarshal the file 'testdata/bad config.yaml'")

	cfg, err = Load("testdata/config with incorrect model.yaml")
	t.Log(err)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid model")

	cfg, err = Load("testdata/config.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Model)
	assert.Equal(t, "9.9.9", cfg.Model.Info.Version)
}
