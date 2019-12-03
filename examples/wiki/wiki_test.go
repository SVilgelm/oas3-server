package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	baseURL string
)

func TestMain(m *testing.M) {
	srv := initServer()
	srv.Config.Address = "127.0.0.1:0"
	err := srv.Start()
	if err != nil {
		panic(err)
	}
	baseURL = "http://" + srv.Config.Address + "/"
	exitCode := m.Run()
	err = srv.Shutdown()
	if err != nil {
		panic(err)
	}
	os.Exit(exitCode)
}

func TestModel(t *testing.T) {
	t.Parallel()
	url := baseURL + "oas3-model"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("content-type", "application/json")
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("content-type"), "application/json")

	req, err = http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("content-type", "application/yaml")
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("content-type"), "application/yaml")

	req, err = http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("content-type", "application/pdf")
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
}

func TestList(t *testing.T) {
	t.Parallel()
	resp, err := http.Get(baseURL)
	t.Log(resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
