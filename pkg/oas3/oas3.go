package oas3

import (
	"encoding/json"
	"log"
	"net/http"

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

// JSON returns the OAS3 model in JSON format
func JSON(w http.ResponseWriter, r *http.Request) {
	item := OperationFromContext(r.Context())

	data, err := json.Marshal(item.Model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// YAML returns the OAS3 model in YAML format
func YAML(w http.ResponseWriter, r *http.Request) {
	item := OperationFromContext(r.Context())

	data, err := yaml.Marshal(item.Model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
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
