package oas3

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SVilgelm/oas3-server/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

// Load parses the YAML/JSON-encoded file with OpenApi 3 Specification
func Load(fileName string) (*openapi3.Swagger, error) {
	model, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(fileName)
	if err != nil {
		return nil, err
	}
	log.Println("Loaded OpenAPI 3 Specification file:", fileName)
	return model, nil
}

// Model returns the OAS3 model in JSON/YAML format
func Model(w http.ResponseWriter, r *http.Request) {
	item := OperationFromContext(r.Context())
	types := utils.GetContentTypes(r)
	var data []byte
	var err error
	var contentType string
	if utils.Contains(types, "application/json") {
		contentType = "application/json"
		data, err = json.Marshal(item.Model)
	} else if utils.Contains(types, "application/yaml") {
		contentType = "application/yaml"
		data, err = yaml.Marshal(item.Model)
	} else {
		http.Error(
			w,
			"Supported media type not found. Use 'application/json' or 'application/yaml'",
			http.StatusUnsupportedMediaType,
		)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Console returns the OAS3 Developer console
func Console(w http.ResponseWriter, r *http.Request) {
	item := OperationFromContext(r.Context())

	data, err := yaml.Marshal(item.Model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
