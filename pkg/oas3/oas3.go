package oas3

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
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

// JSON returns the OAS3 model
func JSON(w http.ResponseWriter, r *http.Request) {
	item := OperationFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	coder := json.NewEncoder(w)
	coder.Encode(item.Model)
}

// Console returns the OAS3 Developer console
func Console(w http.ResponseWriter, r *http.Request) {
	item := OperationFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	coder := json.NewEncoder(w)
	coder.Encode(item.Model)
}
