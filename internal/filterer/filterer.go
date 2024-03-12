package filterer

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	filtererpb "github.com/lopezator/filterer/api/v1"
	"github.com/lopezator/filterer/api/v1/filtererv1connect"
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
func (s *Service) Filter(context.Context, *connect.Request[filtererpb.FilterRequest]) (*connect.Response[filtererpb.FilterResponse], error) {
	return connect.NewResponse(&filtererpb.FilterResponse{
		Sql: "select * from filterer",
	}), nil
}
