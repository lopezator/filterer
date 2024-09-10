package filterer

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"buf.build/gen/go/lopezator/filterer/connectrpc/go/lopezator/filterer/v1/filtererv1connect"
	filtererpb "buf.build/gen/go/lopezator/filterer/protocolbuffers/go/lopezator/filterer/v1"
	"connectrpc.com/connect"
	"github.com/lopezator/filterer/internal/expr"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// Service is the filterer service implementation.
type Service struct {
	filtererv1connect.UnimplementedFiltererServiceHandler
	parser *expr.Parser
}

// FieldSet is a set of filterable fields.
type FieldSet struct {
	ID     string
	Fields []*Field
}

// Field is the representation of a filterable field.
type Field struct {
	Name string
	Type string
}

// NewService returns a service instance.
func NewService(fieldSets []*FieldSet) (string, http.Handler) {
	// Convert fieldSets to a map of string to exprpb.Type
	var err error
	fieldMap := make(map[string]*exprpb.Type)
	for _, fieldSet := range fieldSets {
		for _, field := range fieldSet.Fields {
			fieldMap[field.Name], err = stringToType(field.Type)
			if err != nil {
				panic(err)
			}
		}
	}

	// Create a new parser
	parser, err := expr.NewParser(fieldMap)
	if err != nil {
		panic(err)
	}
	return filtererv1connect.NewFiltererServiceHandler(&Service{
		parser: parser,
	})
}

// stringToType converts a string representation of a type to its corresponding exprpb.Type.
func stringToType(s string) (*exprpb.Type, error) {
	switch s {
	case "bool":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_Primitive{
				Primitive: exprpb.Type_BOOL,
			},
		}, nil
	case "integer":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_Primitive{
				Primitive: exprpb.Type_INT64,
			},
		}, nil
	case "double":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_Primitive{
				Primitive: exprpb.Type_DOUBLE,
			},
		}, nil
	case "string":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_Primitive{
				Primitive: exprpb.Type_STRING,
			},
		}, nil
	case "bytes":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_Primitive{
				Primitive: exprpb.Type_BYTES,
			},
		}, nil
	case "timestamp":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_WellKnown{
				WellKnown: exprpb.Type_TIMESTAMP,
			},
		}, nil
	case "string_array":
		return &exprpb.Type{
			TypeKind: &exprpb.Type_ListType_{
				ListType: &exprpb.Type_ListType{
					ElemType: &exprpb.Type{
						TypeKind: &exprpb.Type_Primitive{
							Primitive: exprpb.Type_STRING,
						},
					},
				},
			},
		}, nil
	default:
		return nil, errors.New("filterer: unknown type")
	}
}

// Filter implements filterer.FiltererServiceServer.Filter.
func (s *Service) Filter(ctx context.Context, req *connect.Request[filtererpb.FilterRequest]) (*connect.Response[filtererpb.FilterResponse], error) {
	// Parse the expression.
	filter, err := s.parser.Parse(req.Msg.Expr)
	if err != nil {
		return nil, fmt.Errorf("filterer: %w", err)
	}

	// TODO(d.lopez): The current idea, pending to add baseQuery and baseArgs to the request.
	// Having a base query with placeholders ?, example:
	// clause: SELECT * FROM table WHERE column = ? ORDER BY id
	// args: paco
	// Maybe just a string query? I don't know yet.
	// sql: select * from table where column = 'paco' order by id
	// I should be able to append a where to that original query, example:
	// clause: "display_name=? AND age=?"
	// args: david, 30
	// And get the final query, like this:
	// SELECT * FROM table WHERE column = 'paco' AND display_name='david' AND age=30 ORDER BY id
	// Maybe modify the response to return just the final string?

	// Generate WHERE clause along with the args from the filter string expression.
	clause, args, err := expr.SQL(filter)
	if err != nil {
		return nil, fmt.Errorf("filterer: %w", err)
	}

	// Convert args to strings.
	var sargs []string
	for _, arg := range args {
		strArg, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert arg to string: %v", arg)
		}
		sargs = append(sargs, strArg)
	}

	// Return response.
	return connect.NewResponse(&filtererpb.FilterResponse{
		Where: clause,
		Args:  sargs,
	}), nil
}
