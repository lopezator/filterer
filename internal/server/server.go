package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"connectrpc.com/vanguard/vanguardgrpc"
	"github.com/lopezator/filterer/internal/filterer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

	handler, err := vanguardgrpc.NewTranscoder(srv)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	listener2, err := net.Listen("tcp", "127.0.0.1:18181")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	go func() {
		// We use the h2c package in order to support HTTP/2 without TLS,
		// so we can handle gRPC requests, which requires HTTP/2, in
		// addition to Connect and gRPC-Web (which work with HTTP 1.1).
		err = http.Serve(listener2, h2c.NewHandler(handler, &http2.Server{}))
		if !errors.Is(err, http.ErrServerClosed) {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	log.Printf("server: server listening at %v", listener.Addr())
	if err := srv.Serve(listener); err != nil {
		return fmt.Errorf("server: failed to serve: %w", err)
	}

	return nil
}
