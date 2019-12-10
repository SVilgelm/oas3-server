package server

import (
	"net/http"
	"strings"
)

// FileSystem is a static file server
type FileSystem struct {
	fs http.FileSystem
}

// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	path = strings.Trim(path, "/")
	var f http.File
	var err error
	for _, p := range strings.Split(path, "/") {
		f, err = fs.fs.Open(path)
		if err == nil {
			break
		}
		path = path[len(p):]
		path = strings.TrimPrefix(path, "/")
	}
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		path = path + "/index.html"
		return fs.fs.Open(path)
	}

	return f, nil
}
