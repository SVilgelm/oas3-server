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

func getOperationByMethod(pathItem *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case http.MethodHead:
		return pathItem.Head
	case http.MethodPost:
		return pathItem.Post
	case http.MethodPut:
		return pathItem.Put
	case http.MethodPatch:
		return pathItem.Patch
	case http.MethodDelete:
		return pathItem.Delete
	case http.MethodConnect:
		return pathItem.Connect
	case http.MethodOptions:
		return pathItem.Options
	case http.MethodTrace:
		return pathItem.Trace
	default:
		return pathItem.Get
	}
}

func processOperation(
	mapper *Mapper,
	model *openapi3.Swagger,
	router *mux.Router,
	path string,
	httpMethod string,
) {
	pathOperation := getOperationByMethod(model.Paths[path], httpMethod)
	if pathOperation != nil {
		if pathOperation.OperationID == "" {
			log.Printf("No operationID for path '%s' and method '%s', skiped", path, httpMethod)
			return
		}
		item := mapper.ByID(pathOperation.OperationID)
		if item == nil {
			item = NewItem(pathOperation.OperationID, model)
		}
		var route *mux.Route
		if v, err := getBoolExt("x-wildcard", pathOperation.Extensions); err == nil && v {
			route = router.PathPrefix(path)
		} else {
			route = router.Path(path)
		}
		item.Routes = append(
			item.Routes,
			route.Methods(httpMethod).HandlerFunc(http.NotFound),
		)
		mapper.Add(item)
	}
}

// RegisterOperations creates all routes
func RegisterOperations(model *openapi3.Swagger, router *mux.Router) *Mapper {
	allMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}

	mapper := NewMapper()
	if model != nil {
		for path, meta := range model.Paths {
			if meta == nil {
				log.Printf("Wrong path '%s' definition, skiped", path)
				continue
			}
			for _, httpMethod := range allMethods {
				processOperation(mapper, model, router, path, httpMethod)
			}
		}
	}
	return mapper
}
