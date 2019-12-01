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

func prepareData(r *http.Request, item *Item, route *mux.Route) []byte {
	var value string
	data := strings.Builder{}
	firstData := true
	data.WriteString(`{`)
	for in, pr := range item.meta[route].requestParamsNotString {
		if firstData {
			firstData = false
			data.WriteString(`"`)
		} else {
			data.WriteString(`,"`)
		}
		data.WriteString(in)
		data.WriteString(`":{`)
		firstIn := true
		for name, notString := range pr {
			switch in {
			case openapi3.ParameterInCookie:
				cookie, err := r.Cookie(name)
				if err != nil {
					continue
				}
				value = cookie.Value
			case openapi3.ParameterInQuery:
				values, ok := r.URL.Query()[name]
				if !ok || len(values) == 0 {
					continue
				}
				param := item.FindParam(route, in, name)
				if param == nil {
					log.Printf("Cannot find a parameter /%s/%s", in, name)
					continue
				}
				if param.Schema.Value.Type != "array" && len(values) == 1 {
					value = values[0]
				} else if param.Schema.Value.Type == "array" && param.Schema.Value.Items.Value.Type == "string" {
					value = `[`
					for i, v := range values {
						if i > 0 {
							value += `,"` + v + `"`
						} else {
							value += `"` + v + `"`
						}
					}
					value += `]`
				} else {
					value = "[" + strings.Join(values, ",") + "]"
				}
			case openapi3.ParameterInHeader:
				values, ok := r.Header[textproto.CanonicalMIMEHeaderKey(name)]
				if !ok || len(values) == 0 {
					continue
				}
				param := item.FindParam(route, in, name)
				if param == nil {
					log.Printf("Cannot find a parameter /%s/%s", in, name)
					continue
				}
				if param.Schema.Value.Type != "array" && len(values) == 1 {
					value = values[0]
				} else if param.Schema.Value.Type == "array" && param.Schema.Value.Items.Value.Type == "string" {
					value = `[`
					for i, v := range values {
						if i > 0 {
							value += `,"` + v + `"`
						} else {
							value += `"` + v + `"`
						}
					}
					value += `]`
				} else {
					value = "[" + strings.Join(values, ",") + "]"
				}
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
			if firstIn {
				firstIn = false
				data.WriteString(`"`)
			} else {
				data.WriteString(`,"`)
			}
			data.WriteString(name)
			data.WriteString(`":`)
			if notString {
				data.WriteString(value)
			} else {
				data.WriteString(`"`)
				data.WriteString(value)
				data.WriteString(`"`)
			}
		}
		data.WriteString(`}`)
	}

	data.WriteString(`}`)
	return []byte(data.String())
}

func checkRequest(r *http.Request, item *Item, route *mux.Route) error {
	rs := item.meta[route].requestSchema
	if rs != nil {
		data := prepareData(r, item, route)
		valErr, err := rs.ValidateBytes(data)
		if err != nil {
			return err
		}
		if len(valErr) > 0 {
			return fmt.Errorf("%+v", valErr)
		}
	}
	return nil
}

// Middleware puts the model into the current context and run validations
func Middleware(mapper *Mapper, validateRequest, validateResponse bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			item := mapper.ByRoute(route)
			if item != nil && validateRequest {
				if err := checkRequest(r, item, route); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			ctx := WithOperation(r.Context(), item)
			r = r.WithContext(ctx)
			if item != nil && validateResponse {
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
