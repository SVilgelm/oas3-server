package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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
	item := s.mapper.ByID(operationID)
	if item == nil {
		return fmt.Errorf("The operation '%s' not found", operationID)
	}
	log.Printf("Linking new handler for the operation '%s'", operationID)
	for _, route := range item.Routes {
		route.HandlerFunc(handler)
	}
	return nil
}

// Handle links the handler with the operation
func (s *Server) Handle(operationID string, handler http.Handler) error {
	item := s.mapper.ByID(operationID)
	if item == nil {
		return fmt.Errorf("The operation '%s' not found", operationID)
	}
	log.Printf("Linking new handler for the operation '%s'", operationID)
	for _, route := range item.Routes {
		route.Handler(handler)
	}
	return nil
}

// Shutdown gracefully shutdowns the server
func (s *Server) Shutdown() error {
	return s.HTTPServer.Shutdown(context.Background())
}

// Start runs the server
func (s *Server) Start() error {
	addr := s.Config.Address
	if addr == "" {
		if s.Config.TLS.Enabled {
			addr = ":https"
		} else {
			addr = ":http"
		}
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.Config.Address = ln.Addr().(*net.TCPAddr).String()

	errors := make(chan error, 1)

	go func(listener net.Listener, errors chan error) {
		if s.Config.TLS.Enabled {
			errors <- s.HTTPServer.ServeTLS(listener, s.Config.TLS.Cert, s.Config.TLS.Key)
		} else {
			errors <- s.HTTPServer.Serve(listener)
		}
		close(errors)
	}(ln, errors)
	time.Sleep(1 * time.Microsecond)

	if len(errors) > 0 {
		return <-errors
	}
	var u string
	if s.Config.TLS.Enabled {
		u = "https://"
	} else {
		u = "http://"
	}
	u += s.Config.Address + "/"
	log.Println("Service is listening on", u)
	return nil
}

// Serve starts Server
func (s *Server) Serve() error {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	err := s.Start()
	if err != nil {
		close(gracefulStop)
		return err
	}
	log.Println("Please press Ctrl+C to stop service")
	<-gracefulStop
	log.Println("Gracefully stopping service")

	return s.Shutdown()
}

// NewServer creates new server
func NewServer(cfg *config.Config) *Server {
	srv := Server{
		HTTPServer: &http.Server{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      mux.NewRouter(),
		},
		Config: cfg,
	}
	srv.R = srv.HTTPServer.Handler.(*mux.Router)
	srv.mapper = oas3.RegisterOperations(srv.Config.Model, srv.R)
	srv.R.Use(LogHTTP(srv.mapper), oas3.Middleware(
		srv.mapper,
		srv.Config.Validate.Request,
		srv.Config.Validate.Response,
	))
	srv.HandleFunc("oas3.model", oas3.Model)
	srv.HandleFunc("oas3.console", oas3.Console)
	if _, err := os.Stat(cfg.Static); !os.IsNotExist(err) {
		fileServer := http.FileServer(FileSystem{http.Dir(cfg.Static)})
		srv.Handle("static", fileServer)
	}

	return &srv
}
