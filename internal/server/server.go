package server

import (
	"fmt"
	"net/http"

	"github.com/lopezator/filterer/internal/filterer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Handler is an interface for handling gRPC services.
type Handler interface {
	Handle() (string, http.Handler)
}

// Config holds necessary server configuration parameters
type Config struct {
	Addr      string               `yaml:"addr"`
	FieldSets []*filterer.FieldSet `yaml:"field_sets"`
}

// Server is a meta-server composed by a grpc server and an http server
type Server struct {
	addr string
	mux  *http.ServeMux
}

// New creates a new Server.
func New(cfg *Config) (*Server, error) {
	// Create a new mux
	mux := http.NewServeMux()

	// Add services to the mux
	mux.Handle(filterer.NewService(cfg.FieldSets))

	// Return the server
	return &Server{
		addr: cfg.Addr,
		mux:  mux,
	}, nil
}

// Serve serves the grpc + rest server.
func (s *Server) Serve() error {
	fmt.Println("... Listening on", s.addr)
	return http.ListenAndServe(
		s.addr,
		// For gRPC clients, it's convenient to support HTTP/2 without TLS.
		h2c.NewHandler(s.mux, &http2.Server{}),
	)
}
