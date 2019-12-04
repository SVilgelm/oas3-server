package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContentTypes(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "", nil)
	assert.NoError(t, err)

	assert.Empty(t, req.Header.Get("content-type"))
	assert.Equal(t, []string{"application/octet-stream"}, GetContentTypes(req))

	req.Header.Set("content-type", "application/json")
	assert.Equal(t, []string{"application/json"}, GetContentTypes(req))
	req.Header.Set("content-type", "application/json, text/plain")
	assert.Equal(t, []string{"application/json", "text/plain"}, GetContentTypes(req))
	req.Header.Set("content-type", "application/json, , text/plain")
	assert.Equal(t, []string{"application/json", "text/plain"}, GetContentTypes(req))
}
