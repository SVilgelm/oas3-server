package oas3

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/gorilla/mux"
)

type response struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func (w *response) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *response) Write(body []byte) (int, error) {
	return w.buf.Write(body)
}

func (w *response) send() {
	if w.statusCode == 0 {
		w.statusCode = 200
	}
	w.ResponseWriter.WriteHeader(w.statusCode)
	w.ResponseWriter.Write(w.buf.Bytes())
}

func processValues(param *openapi3.Parameter, values []string) string {
	if param.Schema.Value.Type != "array" {
		if len(values) == 1 {
			return values[0]
		}
		return strings.Join(values, " ")
	} else if param.Schema.Value.Items.Value.Type == "string" {
		size := 2 + 2*len(values) + len(values) - 1 // [] + ""*len(values) + commas
		for _, v := range values {
			size += len(v)
		}
		b := strings.Builder{}
		b.Grow(size)
		b.WriteString(`[`)
		b.WriteString(`"`)
		b.WriteString(values[0])
		b.WriteString(`"`)
		for _, v := range values[1:] {
			b.WriteString(`,"`)
			b.WriteString(v)
			b.WriteString(`"`)
		}
		b.WriteString(`]`)
		return b.String()
	}
	return "[" + strings.Join(values, ",") + "]"
}

func getRealParameters(r *http.Request, item *Item, route *mux.Route) map[string]map[string]*string {
	valuesCache := make(map[string]map[string]*string)
	for in, pr := range item.meta[route].requestParamsNotString {
		for name := range pr {
			var value string
			switch in {
			case openapi3.ParameterInCookie:
				cookie, err := r.Cookie(name)
				if err != nil {
					continue
				}
				value = cookie.Value
			case openapi3.ParameterInQuery:
				values := r.URL.Query()[name]
				if len(values) == 0 {
					continue
				}
				param := item.FindParam(route, in, name)
				if param == nil {
					log.Printf("Cannot find a parameter /%s/%s", in, name)
					continue
				}
				value = processValues(param, values)
			case openapi3.ParameterInHeader:
				values := r.Header[textproto.CanonicalMIMEHeaderKey(name)]
				if len(values) == 0 {
					continue
				}
				param := item.FindParam(route, in, name)
				if param == nil {
					log.Printf("Cannot find a parameter /%s/%s", in, name)
					continue
				}
				value = processValues(param, values)
			case openapi3.ParameterInPath:
				vars := mux.Vars(r)
				pathValue, ok := vars[name]
				if !ok {
					continue
				}
				value = pathValue
			default:
				continue
			}
			if _, ok := valuesCache[in]; !ok {
				valuesCache[in] = make(map[string]*string)
			}
			valuesCache[in][name] = &value
		}
	}
	return valuesCache
}

func getParameterDataBuilder(valuesCache map[string]map[string]*string, item *Item, route *mux.Route) *strings.Builder {
	size := 2 // {}
	if len(valuesCache) > 0 {
		size += 5*len(valuesCache) + len(valuesCache) - 1 // len(in) * "":{} commas
		for in, pr := range valuesCache {
			size += len(in)
			size += 3*len(pr) + len(pr) - 1 // len(pr) * "": + commas
			for name, value := range pr {
				size += len(name) + len(*value)
				if !item.meta[route].requestParamsNotString[in][name] {
					size += 2 // ""
				}
			}
		}
	}
	builder := strings.Builder{}
	builder.Grow(size)
	return &builder
}

func prepareParametersData(r *http.Request, item *Item, route *mux.Route) []byte {
	valuesCache := getRealParameters(r, item, route)
	builder := getParameterDataBuilder(valuesCache, item, route)

	firstData := true
	builder.WriteString(`{`)
	for in, pr := range valuesCache {
		if firstData {
			firstData = false
			builder.WriteString(`"`)
		} else {
			builder.WriteString(`,"`)
		}
		builder.WriteString(in)
		builder.WriteString(`":{`)
		firstIn := true
		for name, value := range pr {
			if firstIn {
				firstIn = false
				builder.WriteString(`"`)
			} else {
				builder.WriteString(`,"`)
			}
			builder.WriteString(name)
			builder.WriteString(`":`)
			if item.meta[route].requestParamsNotString[in][name] {
				builder.WriteString(*value)
			} else {
				builder.WriteString(`"`)
				builder.WriteString(*value)
				builder.WriteString(`"`)
			}
		}
		builder.WriteString(`}`)
	}
	builder.WriteString(`}`)
	return []byte(builder.String())
}

func validateRequest(r *http.Request, item *Item, route *mux.Route) error {
	rs := item.meta[route].requestSchema
	if rs != nil {
		data := prepareParametersData(r, item, route)
		valErr, err := rs.ValidateBytes(data)
		if err != nil {
			return err
		}
		if len(valErr) > 0 {
			return fmt.Errorf("validating request parameters: %+v", valErr)
		}
	}
	return nil
}

// Middleware puts the model into the current context and run validations
func Middleware(mapper *Mapper, doRequestValidation, doResponseValidation bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			item := mapper.ByRoute(route)
			if item != nil && doRequestValidation {
				if err := validateRequest(r, item, route); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			ctx := WithOperation(r.Context(), item)
			r = r.WithContext(ctx)
			if item != nil && doResponseValidation {
				log.Println("Validating response")

				rw := response{
					ResponseWriter: w,
					buf:            new(bytes.Buffer),
				}
				next.ServeHTTP(&rw, r)
				rw.send()
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
