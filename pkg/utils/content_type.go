package utils

import (
	"mime"
	"net/http"
	"strings"
)

// GetContentTypes returns a list of content-types from a Request
func GetContentTypes(r *http.Request) []string {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return []string{"application/octet-stream"}
	}
	var res []string
	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			continue
		}
		res = append(res, t)
	}
	return res
}
