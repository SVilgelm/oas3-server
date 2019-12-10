package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	baseURL string
)

func setupTests() (func(), error) {
	var tearDowns []func()

	tearDown := func() {
		for _, f := range tearDowns {
			f()
		}
	}

	dir, err := ioutil.TempDir("testdata", "data")
	if err != nil {
		return tearDown, err
	}
	tearDowns = append(tearDowns, func() {
		os.RemoveAll(dir)
	})
	dataFolder = dir

	p := Page{
		Title: "testArticle",
		Body:  "This is a test Article",
	}
	err = p.save()
	if err != nil {
		return tearDown, err
	}

	srv, err := initServer()
	if err != nil {
		return tearDown, err
	}
	srv.Config.Address = "127.0.0.1:0"
	err = srv.Start()
	if err != nil {
		return tearDown, err
	}
	baseURL = "http://" + srv.Config.Address + "/"
	tearDowns = append(tearDowns, func() {
		_ = srv.Shutdown()
	})
	return tearDown, nil
}

func TestMain(m *testing.M) {
	tearDown, err := setupTests()
	if err != nil {
		tearDown()
		panic(err)
	}
	exitCode := m.Run()
	tearDown()
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
	t.Logf("Resp %+v", resp)
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
}

func TestList(t *testing.T) {
	t.Parallel()
	resp, err := http.Get(baseURL)
	assert.NoError(t, err)
	t.Logf("Response: %+v", resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("content-type"), "text/html")

	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	req.Header.Set("content-type", "application/json")
	assert.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	t.Logf("Response: %+v", resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("content-type"), "application/json")
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	t.Logf("Body: %s", string(data))
	assert.NoError(t, err)
	var res []string
	err = json.Unmarshal(data, &res)
	assert.NoError(t, err)
	t.Logf("List of articles: %+v", res)
	assert.Contains(t, res, "testArticle")
}
