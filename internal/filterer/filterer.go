package filterer

import (
	"context"
	"net/http"

	"buf.build/gen/go/lopezator/filterer/connectrpc/go/lopezator/filterer/v1/filtererv1connect"
	filtererpb "buf.build/gen/go/lopezator/filterer/protocolbuffers/go/lopezator/filterer/v1"
	"connectrpc.com/connect"
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
