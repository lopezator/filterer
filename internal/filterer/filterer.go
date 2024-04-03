package filterer

import (
	"context"
	"fmt"
	"net/http"

	"buf.build/gen/go/lopezator/filterer/connectrpc/go/lopezator/filterer/v1/filtererv1connect"
	filtererpb "buf.build/gen/go/lopezator/filterer/protocolbuffers/go/lopezator/filterer/v1"
	"connectrpc.com/connect"
	"github.com/cockscomb/cel2sql"
	"github.com/cockscomb/cel2sql/sqltypes"
	"github.com/google/cel-go/cel"
)

// Service is the filterer service implementation.
type Service struct {
	filtererv1connect.UnimplementedFiltererServiceHandler
}

// NewService returns a service instance.
func NewService() (string, http.Handler) {
	return filtererv1connect.NewFiltererServiceHandler(&Service{})
}

// Filter implements filterer.FiltererServiceServer.Filter.
func (s *Service) Filter(ctx context.Context, req *connect.Request[filtererpb.FilterRequest]) (*connect.Response[filtererpb.FilterResponse], error) {
	// TODO(lopezator): all this cel logic should be hidden somewhere else.
	var variables []cel.EnvOption
	for _, column := range req.Msg.Columns {
		var columnType *cel.Type
		switch column.Type {
		case filtererpb.FilterRequest_Column_TYPE_STRING:
			columnType = cel.StringType
		}
		variables = append(variables, cel.Variable(column.Name, columnType))
	}
	env, _ := cel.NewEnv(append([]cel.EnvOption{sqltypes.SQLTypeDeclarations}, variables...)...)
	ast, _ := env.Compile(req.Msg.Expr)

	// TODO(lopezator): consider using my own implementation instead.
	where, _ := cel2sql.Convert(ast)

	// Return response.
	return connect.NewResponse(&filtererpb.FilterResponse{
		Where: fmt.Sprintf("%q", where),
	}), nil
}
