package oas3

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Middleware puts the model into the current context
func Middleware(mapper *Mapper) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			op := mapper.ByRoute(route)
			ctx := WithOperation(r.Context(), op)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
