package filterer

import (
	"context"

	"github.com/lopezator/filterer/api/lopezator/filterer/v1"
	"google.golang.org/grpc"
)

// Server is a grPC server.
type Server struct{}

// NewServer returns a Server instance.
func NewServer() (*Server, error) {
	return &Server{}, nil
}

// Register implements http.Registerer.Register.
func (s *Server) Register(srv *grpc.Server) error {
	filterer.RegisterFiltererServiceServer(srv, s)
	return nil
}

// Filter implements filterer.FiltererServiceServer.Filter.
func (s *Server) Filter(ctx context.Context, req *filterer.FilterRequest) (*filterer.FilterResponse, error) {
	return &filterer.FilterResponse{Sql: "select * from filterer"}, nil
}
