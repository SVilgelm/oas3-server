package server

import (
	"log"
	"net/http"
	"time"

	"github.com/SVilgelm/oas3-server/pkg/oas3"
	"github.com/gorilla/mux"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// LogHTTP is a middleware to log requests
func LogHTTP(mapper *oas3.Mapper) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := statusWriter{ResponseWriter: w}
			defer func() {
				duration := time.Now().Sub(start)
				route := mux.CurrentRoute(r)
				op := mapper.ByRoute(route)
				log.Println(
					"Request",
					op.ID,
					r.Host,
					r.RemoteAddr,
					r.Method,
					r.RequestURI,
					r.Proto,
					sw.status,
					sw.length,
					r.Header.Get("User-Agent"),
					duration,
				)
			}()
			next.ServeHTTP(&sw, r)
		})
	}
}
