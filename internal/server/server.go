package server

import (
	"fmt"
	"net/http"

	"connectrpc.com/vanguard"
	"github.com/lopezator/filterer/internal/filterer"
)

// Handler is an interface for handling gRPC services.
type Handler interface {
	Handle() (string, http.Handler)
}

// Config holds necessary server configuration parameters
type Config struct {
	GRPCAddr string
}

// Server is a meta-server composed by a grpc server and a http server
type Server struct {
	GRPCAddr   string
	Transcoder *vanguard.Transcoder
}

// New creates a new Server.
func New(cfg *Config) (*Server, error) {
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{
		filterer.NewService(),
	})
	if err != nil {
		return nil, err
	}
	return &Server{
		GRPCAddr:   cfg.GRPCAddr,
		Transcoder: transcoder,
	}, nil
}

// Serve serves the grpc + rest server.
func (s *Server) Serve() error {
	fmt.Println("... Listening on", s.GRPCAddr)
	return http.ListenAndServe(s.GRPCAddr, s.Transcoder)
}
