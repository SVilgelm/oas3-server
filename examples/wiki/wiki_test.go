package main

import (
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	baseURL     string
	baseURLLock sync.Mutex
)

func TestMain(m *testing.M) {
	srv := initServer()
	srv.Config.Address = "127.0.0.1:0"
	srv.Start()
	baseURLLock.Lock()
	baseURL = "http://" + srv.Config.Address + "/"
	baseURLLock.Unlock()
	exitCode := m.Run()
	srv.Shutdown()
	os.Exit(exitCode)
}

func TestModel(t *testing.T) {
	t.Parallel()
	baseURLLock.Lock()
	url := baseURL + "oas3-model"
	baseURLLock.Unlock()
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
	baseURLLock.Lock()
	resp, err := http.Get(baseURL)
	baseURLLock.Unlock()
	t.Log(resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
