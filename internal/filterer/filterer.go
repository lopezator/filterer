package filterer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"buf.build/gen/go/lopezator/filterer/connectrpc/go/lopezator/filterer/v1/filtererv1connect"
	filtererpb "buf.build/gen/go/lopezator/filterer/protocolbuffers/go/lopezator/filterer/v1"
	"connectrpc.com/connect"
	"github.com/lopezator/filterer/internal/expr"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"vitess.io/vitess/go/vt/sqlparser"
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
	baseQuery := "SELECT * FROM users WHERE id = 3 ORDER BY id LIMIT 10"

	// Create a new parser instance
	parser, err := sqlparser.New(sqlparser.Options{})
	if err != nil {
		log.Fatalf("failed to create parser: %v", err)
	}

	// Parse the base query.
	stmt, err := parser.Parse(baseQuery)
	if err != nil {
		log.Fatalf("filterer: invalid base query: %v", err)
	}

	// Only allow select stmts.
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		log.Fatalf("filterer: base query must be a valid SELECT stmt")
	}

	// Parse the expression.
	filter, err := s.parser.Parse(req.Msg.Expr)
	if err != nil {
		return nil, fmt.Errorf("filterer: %w", err)
	}

	// Generate WHERE clause along with the args from the filter string expression.
	clause, args, err := expr.SQL(filter)
	if err != nil {
		return nil, fmt.Errorf("filterer: %w", err)
	}

	// Replace placeholders with actual arguments
	var formattedClause string
	for _, arg := range args {
		formattedClause = strings.Replace(clause, "?", fmt.Sprintf("'%v'", arg), 1)
	}

	// Parse the additional WHERE clause
	whereExpr, err := parser.ParseExpr(formattedClause)
	if err != nil {
		return nil, fmt.Errorf("filterer: failed to parse additional WHERE clause: %w", err)
	}

	// Append the additional WHERE clause to the base query
	if selectStmt.Where == nil {
		selectStmt.Where = &sqlparser.Where{
			Type: sqlparser.WhereClause,
			Expr: whereExpr,
		}
	} else {
		selectStmt.Where.Expr = &sqlparser.AndExpr{
			Left:  selectStmt.Where.Expr,
			Right: whereExpr,
		}
	}

	// Return response.
	return connect.NewResponse(&filtererpb.FilterResponse{
		Query: sqlparser.String(selectStmt),
	}), nil
}
