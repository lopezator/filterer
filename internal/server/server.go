package server

import (
	"fmt"
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/lopezator/filterer/internal/filterer"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Handler is an interface for handling gRPC services.
type Handler interface {
	Handle() (string, http.Handler)
}

// Config holds necessary server configuration parameters
type Config struct {
	Addr string
}

// Server is a meta-server composed by a grpc server and a http server
type Server struct {
	addr string
	mux  *http.ServeMux
}

// New creates a new Server.
func New(cfg *Config) (*Server, error) {
	// Create a new mux
	mux := http.NewServeMux()

	// Add services to the mux
	mux.Handle(filterer.NewService())

	// Return the server
	return &Server{
		addr: cfg.Addr,
		mux:  mux,
	}, nil
}

// Serve serves the grpc + rest server.
func (s *Server) Serve() error {
	fmt.Println("... Listening on", s.addr)
	// CORS policies.
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200,
	})
	return http.ListenAndServe(
		s.addr,
		// For gRPC clients, it's convenient to support HTTP/2 without TLS.
		h2c.NewHandler(c.Handler(s.mux), &http2.Server{}),
	)
}
