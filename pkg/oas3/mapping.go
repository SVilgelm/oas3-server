package oas3

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/qri-io/jsonschema"

	"github.com/gorilla/mux"
)

type meta struct {
	requestSchema          *jsonschema.RootSchema
	requestParamsNotString map[string]map[string]bool
}

// Item represents a connection between Route and OperationID
type Item struct {
	ID     string
	Model  *openapi3.Swagger
	Routes []*mux.Route
	meta   map[*mux.Route]meta
}

// FindParam returns a parameter from model for given in and name
func (i *Item) FindParam(route *mux.Route, in, name string) *openapi3.Parameter {
	path, err := route.GetPathTemplate()
	if err != nil {
		log.Println("Cannot get a path template:", err.Error())
		return nil
	}
	param := i.Model.Paths[path].Parameters.GetByInAndName(in, name)
	if param == nil {
		methods, err := route.GetMethods()
		if err != nil {
			log.Println("Cannot get methods:", err.Error())
			return nil
		}
		for _, method := range methods {
			op := i.Model.Paths[path].GetOperation(method)
			if op == nil {
				continue
			}
			param = op.Parameters.GetByInAndName(in, name)
			if param != nil {
				break
			}
		}
	}
	return param
}

func getSchemaBuilder(params map[string]map[string]string, required map[string][]string) *strings.Builder {
	size := 32 + len(params) // `{"type":"object","properties":{}}`(33) + commas: len(params) - 1
	for in, pr := range params {
		size += len(in) + 35 + len(pr) // `"` + len(in) + `":{"type":"object","properties":{}}`(35) + commas: len(pr) - 1
		for name, schema := range pr {
			size += 3 + len(name) + len(schema) // `"` + len(name) + `":` + len(schema)
		}
	}
	for _, req := range required {
		size += 13 + len(req) // `"required":[],`(14) + commas: len(req) - 1
		for _, r := range req {
			size += 2 + len(r) // `"` + len(r) + `"`
		}
	}
	builder := strings.Builder{}
	builder.Grow(size)
	return &builder
}

func prepareSchema(params map[string]map[string]string, required map[string][]string) *jsonschema.RootSchema {
	if len(params) == 0 {
		return nil
	}
	schemaData := getSchemaBuilder(params, required)
	schemaData.WriteString(`{"type":"object","properties":{`)
	firstIn := true
	for in, pr := range params {
		if firstIn {
			schemaData.WriteString(`"`)
			firstIn = false
		} else {
			schemaData.WriteString(`,"`)
		}
		schemaData.WriteString(in)
		schemaData.WriteString(`":{"type":"object",`)
		req, ok := required[in]
		if ok && len(req) > 0 {
			schemaData.WriteString(`"required":["`)
			schemaData.WriteString(req[0])
			schemaData.WriteString(`"`)
			for _, name := range req[1:] {
				schemaData.WriteString(`,"`)
				schemaData.WriteString(name)
				schemaData.WriteString(`"`)
			}
			schemaData.WriteString(`],`)
		}
		schemaData.WriteString(`"properties":{`)
		firstPr := true
		for name, schema := range pr {
			if firstPr {
				schemaData.WriteString(`"`)
				firstPr = false
			} else {
				schemaData.WriteString(`,"`)
			}
			schemaData.WriteString(name)
			schemaData.WriteString(`":`)
			schemaData.WriteString(schema)
		}
		schemaData.WriteString(`}}`)
	}
	schemaData.WriteString(`}}`)
	rs := &jsonschema.RootSchema{}
	if err := json.Unmarshal([]byte(schemaData.String()), rs); err != nil {
		log.Printf("Unable to prepare a valid schema: %s", err)
		return nil
	}
	return rs
}

// AddRoute add routes and initializes the Schemas for the operation
func (i *Item) AddRoute(route *mux.Route, pathParameters openapi3.Parameters, operation *openapi3.Operation) {
	i.Routes = append(i.Routes, route)
	params := make(map[string]map[string]string)
	required := make(map[string][]string)
	routeMeta := i.meta[route]
	routeMeta.requestParamsNotString = make(map[string]map[string]bool)

	for _, parameters := range []openapi3.Parameters{pathParameters, operation.Parameters} {
		if len(parameters) > 0 {
			for _, pr := range parameters {
				if pr.Value == nil {
					log.Println("There is invalide parameter. Skipping")
					continue
				}
				if pr.Value.Schema.Value == nil {
					log.Printf("Schema of a parameter '%s/%s' is null. Skipping", pr.Value.In, pr.Value.Name)
					continue
				}
				schema, err := json.Marshal(pr.Value.Schema.Value)
				if err != nil {
					log.Printf("Schema of a parameter '%s/%s' is invalid: %v. Skipping", pr.Value.In, pr.Value.Name, err)
					continue
				}
				in, ok := params[pr.Value.In]
				if !ok {
					in = make(map[string]string)
					params[pr.Value.In] = in
				}
				in[pr.Value.Name] = string(schema)
				if pr.Value.Required {
					required[pr.Value.In] = append(required[pr.Value.In], pr.Value.Name)
				}
				if _, ok := routeMeta.requestParamsNotString[pr.Value.In]; !ok {
					routeMeta.requestParamsNotString[pr.Value.In] = make(map[string]bool)
				}
				routeMeta.requestParamsNotString[pr.Value.In][pr.Value.Name] = pr.Value.Schema.Value.Type != "string"
			}
		}
	}
	routeMeta.requestSchema = prepareSchema(params, required)
	i.meta[route] = routeMeta
}

// NewItem creates new Item with a new Route
func NewItem(operationID string, model *openapi3.Swagger) *Item {
	op := Item{
		ID:    operationID,
		Model: model,
		meta:  make(map[*mux.Route]meta),
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
		route.Methods(httpMethod).HandlerFunc(http.NotFound)
		item.AddRoute(route, model.Paths[path].Parameters, pathOperation)
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
