package oas3

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/gorilla/mux"
)

// Item represents a connection between Route and OperationID
type Item struct {
	ID     string
	Routes []*mux.Route
	Model  *openapi3.Swagger
}

// NewItem creates new Item with a new Route
func NewItem(operationID string, model *openapi3.Swagger) *Item {
	op := Item{
		ID:    operationID,
		Model: model,
	}
	return &op
}

// Mapper stores all Items
type Mapper struct {
	ids    map[string]*Item
	routes map[*mux.Route]*Item
}

// Add adds new Item
func (o *Mapper) Add(item *Item) {
	o.ids[item.ID] = item
	for _, route := range item.Routes {
		o.routes[route] = item
	}
}

// ByID finds an Item by its OperationID
func (o *Mapper) ByID(operationID string) *Item {
	return o.ids[operationID]
}

// ByRoute finds an Item by a Route
func (o *Mapper) ByRoute(route *mux.Route) *Item {
	return o.routes[route]
}

// NewMapper creates a Mapper
func NewMapper() *Mapper {
	ops := Mapper{
		ids:    make(map[string]*Item),
		routes: make(map[*mux.Route]*Item),
	}
	return &ops
}

func getBoolExt(name string, extensions map[string]interface{}) (bool, error) {
	var res bool = false
	if raw, ok := extensions[name]; ok && raw != nil {
		err := json.Unmarshal(raw.(json.RawMessage), &res)
		if err != nil {
			return false, err
		}
	}
	return res, nil
}

func createOperation(
	model *openapi3.Swagger,
	router *mux.Router,
	path string,
	pathMethod *openapi3.Operation,
	httpMethod string,
	ops *Mapper,
) {
	if pathMethod != nil {
		if pathMethod.OperationID == "" {
			log.Printf("No operationID for path '%s' and method '%s', skiped", path, httpMethod)
			return
		}
		op := ops.ByID(pathMethod.OperationID)
		if op == nil {
			op = NewItem(pathMethod.OperationID, model)
		}
		var route *mux.Route
		if v, err := getBoolExt("x-wildcard", pathMethod.Extensions); err == nil && v {
			route = router.PathPrefix(path)
		} else {
			route = router.Path(path)
		}
		op.Routes = append(
			op.Routes,
			route.Methods(httpMethod).HandlerFunc(http.NotFound),
		)
		ops.Add(op)
	}
}

// RegisterOperations creates all routes
func RegisterOperations(model *openapi3.Swagger, router *mux.Router) *Mapper {
	ops := NewMapper()
	for path, meta := range model.Paths {
		if meta == nil {
			log.Printf("Wrong path '%s' definition, skiped", path)
			continue
		}
		createOperation(model, router, path, meta.Get, http.MethodGet, ops)
		createOperation(model, router, path, meta.Put, http.MethodPut, ops)
		createOperation(model, router, path, meta.Post, http.MethodPost, ops)
		createOperation(model, router, path, meta.Patch, http.MethodPatch, ops)
		createOperation(model, router, path, meta.Delete, http.MethodDelete, ops)
	}
	return ops
}
