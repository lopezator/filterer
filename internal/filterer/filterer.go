package filterer

import (
	"context"
	"fmt"

	filtererpb "github.com/lopezator/filterer/api/lopezator/filterer/v1"
	"github.com/lopezator/filterer/internal/expr"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server is the filterer server.
type Server struct {
	parser *expr.Parser
}

// NewServer returns a Server instance.
func NewServer() (*Server, error) {
	filterables := map[string]*exprpb.Type{
		"display_name": {
			TypeKind: &exprpb.Type_Primitive{
				Primitive: exprpb.Type_STRING,
			},
		},
	}
	parser, err := expr.NewParser(filterables)
	if err != nil {
		return nil, err
	}
	return &Server{parser: parser}, nil
}

// Register implements http.Registerer.Register.
func (s *Server) Register(srv *grpc.Server) error {
	filtererpb.RegisterFiltererServiceServer(srv, s)
	return nil
}

// Filter implements filterer.FiltererServiceServer.Filter.
func (s *Server) Filter(ctx context.Context, req *filtererpb.FilterRequest) (*filtererpb.FilterResponse, error) {
	if req.Table == "" {
		return nil, status.Error(codes.InvalidArgument, "filterer: you must provide a table name")
	}
	if req.Filter == "" {
		return nil, status.Error(codes.InvalidArgument, "filterer: you must provide a filter")
	}
	e, err := s.parser.Parse(req.Filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	clause, args, err := expr.SQL(e)
	fmt.Println(clause, args)
	return &filtererpb.FilterResponse{Sql: "select * from filterer"}, nil
}
