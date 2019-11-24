package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/SVilgelm/oas3-server/pkg/config"
	"github.com/SVilgelm/oas3-server/pkg/oas3"
	"github.com/gorilla/mux"
)

// Server is a OpenAPI 3 Specification Web Server
type Server struct {
	HTTPServer *http.Server
	Config     *config.Config
	R          *mux.Router
	mapper     *oas3.Mapper
}

// HandleFunc links the handler with the operation
func (s *Server) HandleFunc(operationID string, handler http.HandlerFunc) error {
	op := s.mapper.ByID(operationID)
	if op == nil {
		return fmt.Errorf("The operation '%s' not found", operationID)
	}
	log.Printf("Linking new handler for the operation '%s'", operationID)
	for _, route := range op.Routes {
		route.HandlerFunc(handler)
	}
	return nil
}

// Handle links the handler with the operation
func (s *Server) Handle(operationID string, handler http.Handler) error {
	op := s.mapper.ByID(operationID)
	if op == nil {
		return fmt.Errorf("The operation '%s' not found", operationID)
	}
	log.Printf("Linking new handler for the operation '%s'", operationID)
	for _, route := range op.Routes {
		route.Handler(handler)
	}
	return nil
}

// Shutdown gracefully shutdowns the server
func (s *Server) Shutdown() {
	if err := s.HTTPServer.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}

// Start runs the server
func (s *Server) Start() {
	s.R.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		msg := "Route"
		op := s.mapper.ByRoute(route)
		if op != nil {
			msg += " '" + op.ID + "'"
		}
		msg += ": "
		methods, err := route.GetMethods()
		if err == nil {
			msg += "[" + strings.Join(methods, "|") + "] "
		}
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			msg += pathTemplate + " "
			pathRegexp, err := route.GetPathRegexp()
			if err == nil {
				msg += "(" + pathRegexp + ") "
			}

		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil && len(queriesTemplates) > 0 {
			msg += "? " + strings.Join(queriesTemplates, ",") + " "
			queriesRegexps, err := route.GetQueriesRegexp()
			if err == nil && len(queriesRegexps) > 0 {
				msg += "(" + strings.Join(queriesRegexps, ",") + ") "
			}
		}
		log.Println(msg)
		return nil
	})

	go func() {
		var err error
		if s.Config.TLS.Enabled {
			err = s.HTTPServer.ListenAndServeTLS(s.Config.TLS.Cert, s.Config.TLS.Key)
		} else {
			err = s.HTTPServer.ListenAndServe()
		}
		if err != nil {
			log.Fatalf("Failed to serve: %+v", err)
		}
	}()
	var u string
	if s.Config.TLS.Enabled {
		u = "https://"
	} else {
		u = "http://"
	}
	u += s.Config.Address + "/"
	log.Println("Service is listening on", u)
}

// Serve starts Server
func (s *Server) Serve() {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	s.Start()
	log.Println("Please press Ctrl+C to stop service")
	<-gracefulStop
	log.Println("Gracefully stopping service")

	s.Shutdown()
}

// NewServer creates new server
func NewServer(cfg *config.Config) *Server {
	srv := Server{
		HTTPServer: &http.Server{
			Addr:         cfg.Address,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      mux.NewRouter(),
		},
		Config: cfg,
	}
	srv.R = srv.HTTPServer.Handler.(*mux.Router)
	srv.mapper = oas3.RegisterOperations(srv.Config.Model, srv.R)
	srv.R.Use(LogHTTP(srv.mapper), oas3.Middleware(srv.mapper))
	srv.HandleFunc("oas3.model", oas3.Model)
	srv.HandleFunc("oas3.console", oas3.Console)
	if _, err := os.Stat(cfg.Static); !os.IsNotExist(err) {
		fileServer := http.FileServer(FileSystem{http.Dir(cfg.Static)})
		srv.Handle("static", fileServer)
	}

	return &srv
}
