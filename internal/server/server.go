package server

import (
	"fmt"
	"log"
	"net"

	"github.com/lopezator/filterer/internal/filterer"
	"google.golang.org/grpc"
)

// Registerer is an interface for registering gRPC services.
type Registerer interface {
	Register(*grpc.Server) error
}

// Config holds necessary server configuration parameters
type Config struct {
	GRPCAddr string
}

// Server is a meta-server composed by a grpc server and a http server
type Server struct {
	GRPCAddr    string
	Registerers []Registerer
}

// New creates a new Server.
func New(cfg *Config) (*Server, error) {
	filtererServer, err := filterer.NewServer()
	if err != nil {
		return nil, err
	}
	return &Server{
		GRPCAddr: cfg.GRPCAddr,
		Registerers: []Registerer{
			filtererServer,
		},
	}, nil
}

// Serve creates a new gRPC server.
func (s *Server) Serve() error {
	// Initialize & register gRPC services.
	srv := grpc.NewServer()
	for _, registerer := range s.Registerers {
		err := registerer.Register(srv)
		if err != nil {
			return err
		}
	}

	// Listen on the specified address & serve.
	listener, err := net.Listen("tcp", s.GRPCAddr)
	if err != nil {
		return fmt.Errorf("server: failed to listen: %w", err)
	}
	log.Printf("server: server listening at %v", listener.Addr())
	if err := srv.Serve(listener); err != nil {
		return fmt.Errorf("server: failed to serve: %w", err)
	}

	return nil
}
