package oas3

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/SVilgelm/oas3-server/pkg/utils"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/qri-io/jsonschema"

	"github.com/gorilla/mux"
)

type meta struct {
	requestSchema          *jsonschema.RootSchema
	requestParamsNotString utils.DoubleMapBool
}

// Item represents a connection between Route and OperationID
type Item struct {
	ID     string
	Model  *openapi3.Swagger
	Routes []*mux.Route
	meta   map[*mux.Route]meta
}

// FindParam returns a parameter from model for given in and name
func (i *Item) FindParam(in, name string, route *mux.Route) *openapi3.Parameter {
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

func getSchemaBuilder(params *utils.DoubleMapString, required map[string][]string) *strings.Builder {
	size := 32 + len(*params) // `{"type":"object","properties":{}}`(33) + commas: len(params) - 1
	for in, pr := range *params {
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

func prepareSchema(params *utils.DoubleMapString, required map[string][]string) (*jsonschema.RootSchema, error) {
	if len(*params) == 0 {
		return nil, nil
	}
	schemaData := getSchemaBuilder(params, required)
	schemaData.WriteString(`{"type":"object","properties":{`)
	quoteIn := `"`
	for in, pr := range *params {
		schemaData.WriteString(quoteIn)
		quoteIn = `,"`
		schemaData.WriteString(in)
		schemaData.WriteString(`":{"type":"object",`)
		if req := required[in]; len(req) > 0 {
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
		quotePr := `"`
		for name, schema := range pr {
			schemaData.WriteString(quotePr)
			quotePr = `,"`
			schemaData.WriteString(name)
			schemaData.WriteString(`":`)
			schemaData.WriteString(schema)
		}
		schemaData.WriteString(`}}`)
	}
	schemaData.WriteString(`}}`)
	rs := &jsonschema.RootSchema{}
	if err := json.Unmarshal([]byte(schemaData.String()), rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func getSchema(pr *openapi3.ParameterRef) (string, error) {
	if pr.Value == nil {
		return "", errors.New("invalid parameter")
	}
	if pr.Value.Schema.Value == nil {
		return "", fmt.Errorf("null schema of a parameter '%s/%s'", pr.Value.In, pr.Value.Name)
	}
	schema, err := json.Marshal(pr.Value.Schema.Value)
	if err != nil {
		return "", fmt.Errorf("invalid schema of a parameter '%s/%s': %v", pr.Value.In, pr.Value.Name, err)
	}
	return string(schema), nil
}

// AddRoute add routes and initializes the Schemas for the operation
func (i *Item) AddRoute(route *mux.Route, pathParameters openapi3.Parameters, operation *openapi3.Operation) error {
	i.Routes = append(i.Routes, route)
	params := make(utils.DoubleMapString)
	required := make(map[string][]string)
	routeMeta := i.meta[route]
	routeMeta.requestParamsNotString = make(utils.DoubleMapBool)

	for _, parameters := range []openapi3.Parameters{pathParameters, operation.Parameters} {
		for _, pr := range parameters {
			schema, err := getSchema(pr)
			if err != nil {
				return err
			}
			params.Set(pr.Value.In, pr.Value.Name, schema)
			if pr.Value.Required {
				required[pr.Value.In] = append(required[pr.Value.In], pr.Value.Name)
			}
			routeMeta.requestParamsNotString.Set(pr.Value.In, pr.Value.Name, pr.Value.Schema.Value.Type != "string")
		}
	}
	rs, err := prepareSchema(&params, required)
	if err != nil {
		return err
	}
	routeMeta.requestSchema = rs
	i.meta[route] = routeMeta
	return nil
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
) error {
	pathOperation := getOperationByMethod(model.Paths[path], httpMethod)
	if pathOperation == nil {
		return nil
	}
	if pathOperation.OperationID == "" {
		log.Printf("No operationID for path '%s' and method '%s', skipped", path, httpMethod)
		return nil
	}
	item := mapper.ByID(pathOperation.OperationID)
	if item == nil {
		item = NewItem(pathOperation.OperationID, model)
	}
	wildcard, err := getBoolExt("x-wildcard", pathOperation.Extensions)
	if err != nil {
		return err
	}
	var route *mux.Route
	switch wildcard {
	case true:
		route = router.PathPrefix(path)
	case false:
		route = router.Path(path)
	}
	route.Methods(httpMethod).HandlerFunc(http.NotFound)
	err = item.AddRoute(route, model.Paths[path].Parameters, pathOperation)
	if err != nil {
		return err
	}
	mapper.Add(item)

	return nil
}

// RegisterOperations creates all routes
func RegisterOperations(model *openapi3.Swagger, router *mux.Router) (*Mapper, error) {
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
	if model == nil {
		return mapper, nil
	}
	for path, meta := range model.Paths {
		if meta == nil {
			log.Printf("Wrong path '%s' definition, skipped", path)
			continue
		}
		for _, httpMethod := range allMethods {
			err := processOperation(mapper, model, router, path, httpMethod)
			if err != nil {
				return nil, err
			}
		}
	}
	return mapper, nil
}
