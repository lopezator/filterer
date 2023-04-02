package expr

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/saltosystems/x/errors"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const (
	// Maximum value for the expression depth in order to validate and
	// prematurely exit in case a too depth expression was found.
	maxDepth = 5
)

var primitiveTypeLookup = map[exprpb.Type_PrimitiveType]FieldType{
	exprpb.Type_BOOL:   BoolFieldType,
	exprpb.Type_INT64:  IntegerFieldType,
	exprpb.Type_UINT64: IntegerFieldType,
	exprpb.Type_DOUBLE: DoubleFieldType,
	exprpb.Type_STRING: StringFieldType,
	exprpb.Type_BYTES:  BytesFieldType,
}

var wellKnownTypeLookup = map[exprpb.Type_WellKnownType]FieldType{
	exprpb.Type_TIMESTAMP: TimestampFieldType,
}

var listTypeLookup = map[exprpb.Type_PrimitiveType]FieldType{
	exprpb.Type_STRING: StringArrayFieldType,
}

// Parser is our expr parser
type Parser struct {
	env          *cel.Env
	fields       map[string]*Field
	declarations []*exprpb.Decl
}

// ParserOpt sets options such as validators.
type ParserOpt func(parser *Parser)

// WithDeclarations overrides default declarations
func WithDeclarations(declarations []*exprpb.Decl) ParserOpt {
	return func(m *Parser) {
		m.declarations = declarations
	}
}

// NewParser creates a new parser
func NewParser(allowedFields map[string]*exprpb.Type, opts ...ParserOpt) (*Parser, error) {
	parser := &Parser{
		fields:       make(map[string]*Field),
		declarations: StandardDeclarations(),
	}
	for _, opt := range opts {
		opt(parser)
	}

	// extract dynamic declarations from proto.Message + allowedFields
	for allowedField, exprType := range allowedFields {
		parser.declarations = append(parser.declarations, decls.NewVar(allowedField, exprType))

		// detect field type
		var ftype FieldType
		var ok bool
		switch kind := exprType.TypeKind.(type) {
		case *exprpb.Type_Primitive:
			ftype, ok = primitiveTypeLookup[kind.Primitive]
			if !ok {
				return nil, fmt.Errorf("expr: unsupported primitive field type for %s", allowedField)
			}
		case *exprpb.Type_WellKnown:
			ftype, ok = wellKnownTypeLookup[kind.WellKnown]
			if !ok {
				return nil, fmt.Errorf("expr: unsupported well known field type for %s", allowedField)
			}
		case *exprpb.Type_ListType_:
			if kind, ok := kind.ListType.ElemType.TypeKind.(*exprpb.Type_Primitive); ok {
				ftype, ok = listTypeLookup[kind.Primitive]
				if !ok {
					return nil, fmt.Errorf("expr: unsupported list field type element type for %s", allowedField)
				}
			} else {
				return nil, fmt.Errorf("expr: unsupported list field type for %s", allowedField)
			}
		default:
			return nil, fmt.Errorf("expr: unsupported field type %T for %s", kind, allowedField)
		}
		parser.fields[allowedField] = &Field{Name: allowedField, Ftype: ftype}
	}

	// build custom environment with provided declarations
	env, err := cel.NewCustomEnv(
		cel.HomogeneousAggregateLiterals(),
		cel.Declarations(parser.declarations...),
	)
	if err != nil {
		return nil, err
	}
	parser.env = env

	// return the parser
	return parser, nil
}

// StandardDeclarations returns a set of standard declarations to use within out parser
func StandardDeclarations() []*exprpb.Decl {
	return []*exprpb.Decl{
		decls.NewFunction(operators.LogicalNot,
			decls.NewOverload(overloads.LogicalNot, []*exprpb.Type{decls.Bool}, decls.Bool),
		),
		decls.NewFunction(operators.LogicalAnd,
			decls.NewOverload(overloads.LogicalAnd, []*exprpb.Type{decls.Bool, decls.Bool}, decls.Bool),
		),
		decls.NewFunction(operators.LogicalOr,
			decls.NewOverload(overloads.LogicalOr, []*exprpb.Type{decls.Bool, decls.Bool}, decls.Bool),
		),
		decls.NewFunction(operators.Equals,
			decls.NewOverload(overloads.Equals, []*exprpb.Type{decls.String, decls.String}, decls.Bool),
			decls.NewOverload(overloads.Equals, []*exprpb.Type{decls.Bool, decls.Bool}, decls.Bool),
			decls.NewOverload(overloads.Equals, []*exprpb.Type{decls.Int, decls.Int}, decls.Bool),
		),
		decls.NewFunction(operators.NotEquals,
			decls.NewOverload(overloads.NotEquals, []*exprpb.Type{decls.String, decls.String}, decls.Bool),
			decls.NewOverload(overloads.NotEquals, []*exprpb.Type{decls.Int, decls.Int}, decls.Bool),
		),
		decls.NewFunction(operators.In,
			decls.NewOverload(overloads.InList, []*exprpb.Type{decls.String, decls.NewListType(decls.String)}, decls.Bool),
			decls.NewOverload(overloads.InList, []*exprpb.Type{decls.Int, decls.NewListType(decls.Int)}, decls.Bool),
		),
		decls.NewFunction(overloads.Contains,
			decls.NewInstanceOverload(overloads.ContainsString, []*exprpb.Type{decls.String, decls.String}, decls.Bool),
			decls.NewInstanceOverload(overloads.ContainsString, []*exprpb.Type{decls.NewListType(decls.String), decls.String}, decls.Bool),
		),
		decls.NewFunction(overloads.EndsWith,
			decls.NewInstanceOverload(overloads.EndsWithString, []*exprpb.Type{decls.String, decls.String}, decls.Bool),
		),
		decls.NewFunction(overloads.StartsWith,
			decls.NewInstanceOverload(overloads.StartsWithString, []*exprpb.Type{decls.String, decls.String}, decls.Bool),
		),
		decls.NewFunction(operators.Less,
			decls.NewOverload(overloads.LessTimestamp, []*exprpb.Type{decls.Timestamp, decls.Timestamp}, decls.Bool),
		),
		decls.NewFunction(operators.LessEquals,
			decls.NewOverload(overloads.LessEqualsTimestamp, []*exprpb.Type{decls.Timestamp, decls.Timestamp}, decls.Bool),
		),
		decls.NewFunction(operators.Greater,
			decls.NewOverload(overloads.GreaterTimestamp, []*exprpb.Type{decls.Timestamp, decls.Timestamp}, decls.Bool),
		),
		decls.NewFunction(operators.GreaterEquals,
			decls.NewOverload(overloads.GreaterEqualsTimestamp, []*exprpb.Type{decls.Timestamp, decls.Timestamp}, decls.Bool),
		),
		decls.NewFunction(overloads.TypeConvertTimestamp,
			decls.NewOverload(overloads.StringToTimestamp, []*exprpb.Type{decls.String}, decls.Timestamp),
		),
		decls.NewFunction("present",
			decls.NewOverload("present_string", []*exprpb.Type{decls.String}, decls.Bool),
			decls.NewOverload("present_int", []*exprpb.Type{decls.Int}, decls.Bool),
			decls.NewOverload("present_bool", []*exprpb.Type{decls.Bool}, decls.Bool),
			decls.NewOverload("present_uint", []*exprpb.Type{decls.Uint}, decls.Bool),
			decls.NewOverload("present_bytes", []*exprpb.Type{decls.Bytes}, decls.Bool),
			decls.NewOverload("present_double", []*exprpb.Type{decls.Double}, decls.Bool),
			decls.NewOverload("present_timestamp", []*exprpb.Type{decls.Timestamp}, decls.Bool),
		),
	}
}

// Parse produces a database friendly expr from a cel string expr
func (p *Parser) Parse(filter string) (*Expr, error) {
	if filter == "" {
		return &Expr{}, nil
	}

	// compile filter to ast
	ast, iss := p.env.Compile(filter)
	if iss.Err() != nil {
		return nil, iss.Err()
	}

	// custom check ast
	return p.check(ast)
}

func (p *Parser) check(ast *cel.Ast) (*Expr, error) {
	switch exprKind := ast.Expr().ExprKind.(type) {
	case *exprpb.Expr_CallExpr:
		n, err := p.walk(exprKind.CallExpr, 0)
		if err != nil {
			return nil, err
		}
		return &Expr{Root: n}, nil
	default:
		return nil, fmt.Errorf("expr: unsupported expression of kind %T", exprKind)
	}
}

func (p *Parser) walk(callExpr *exprpb.Expr_Call, depth int) (Node, error) {
	depth++
	if depth > maxDepth {
		return nil, fmt.Errorf("expr: limit of %d depth level exceed", maxDepth)
	}

	switch callExpr.Function {
	case operators.Equals, operators.NotEquals, operators.In, operators.Greater, operators.GreaterEquals, operators.Less, operators.LessEquals:
		if len(callExpr.Args) != 2 {
			return nil, errors.New("expr: invalid number of arguments")
		}
		return p.opExpr(callExpr.Function, callExpr.Args[0], callExpr.Args[1])
	case overloads.StartsWith, overloads.EndsWith, overloads.Contains:
		if len(callExpr.Args) != 1 {
			return nil, errors.New("expr: invalid number of arguments")
		}
		return p.opExpr(callExpr.Function, callExpr.Target, callExpr.Args[0])
	case operators.LogicalNot:
		expr, err := p.walk(callExpr.Args[0].ExprKind.(*exprpb.Expr_CallExpr).CallExpr, depth)
		if err != nil {
			return nil, err
		}
		return &NotExpr{Not: expr}, nil
	case operators.LogicalAnd, operators.LogicalOr:
		left, err := p.walk(callExpr.Args[0].ExprKind.(*exprpb.Expr_CallExpr).CallExpr, depth)
		if err != nil {
			return nil, err
		}
		right, err := p.walk(callExpr.Args[1].ExprKind.(*exprpb.Expr_CallExpr).CallExpr, depth)
		if err != nil {
			return nil, err
		}
		if callExpr.Function == operators.LogicalAnd {
			return &AndExpr{Left: left, Right: right}, nil
		}
		return &OrExpr{Left: left, Right: right}, nil
	case "present":
		identExpr, ok := callExpr.Args[0].ExprKind.(*exprpb.Expr_IdentExpr)
		if !ok {
			return nil, errors.New("expr: failed to cast to ident expression")
		}
		return &PresentExpr{Field: p.fields[identExpr.IdentExpr.Name]}, nil
	default:
		return nil, errors.New("expr: unsupported call expression function")
	}
}

func (p *Parser) opExpr(op string, leftExpr, rightExpr *exprpb.Expr) (*OpExpr, error) {
	identExpr, ok := leftExpr.ExprKind.(*exprpb.Expr_IdentExpr)
	if !ok {
		return nil, errors.New("expr: failed to cast to ident expression")
	}

	var args []interface{}
	if listExpr, ok := rightExpr.ExprKind.(*exprpb.Expr_ListExpr); ok {
		for _, elemExpr := range listExpr.ListExpr.GetElements() {
			arg, err := value(elemExpr.ExprKind)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	} else {
		arg, err := value(rightExpr.ExprKind)
		if err != nil {
			return nil, err
		}
		args = []interface{}{arg}
	}

	return &OpExpr{
		Field: p.fields[identExpr.IdentExpr.Name],
		Op:    strings.Trim(strings.Trim(op, "_"), "@"),
		Args:  args,
	}, nil
}

func value(expr interface{}) (interface{}, error) {
	var constant *exprpb.Constant
	var isTimestamp bool
	switch valueExpr := expr.(type) {
	case *exprpb.Expr_ConstExpr:
		constant = valueExpr.ConstExpr
	case *exprpb.Expr_CallExpr:
		if len(valueExpr.CallExpr.Args) != 1 {
			return nil, errors.New("expr: invalid number of arguments")
		}
		if valueExpr.CallExpr.Function != overloads.TypeConvertTimestamp {
			return nil, errors.New("expr: unsupported type for call expression")
		}
		constExpr, ok := valueExpr.CallExpr.Args[0].ExprKind.(*exprpb.Expr_ConstExpr)
		if !ok {
			return nil, errors.New("expr: invalid argument type")
		}
		constant = constExpr.ConstExpr
		isTimestamp = true
	default:
		return nil, errors.New("expr: unsupported type for value")
	}

	switch constKind := constant.ConstantKind.(type) {
	case *exprpb.Constant_BoolValue:
		return constKind.BoolValue, nil
	case *exprpb.Constant_BytesValue:
		return constKind.BytesValue, nil
	case *exprpb.Constant_DoubleValue:
		return constKind.DoubleValue, nil
	case *exprpb.Constant_Int64Value:
		return constKind.Int64Value, nil
	case *exprpb.Constant_NullValue:
		return nil, nil
	case *exprpb.Constant_StringValue:
		if isTimestamp {
			t, err := time.Parse(time.RFC3339, constKind.StringValue)
			if err != nil {
				return nil, fmt.Errorf("expr: failed to parse time: %v", err)
			}
			return t.UTC(), nil
		}
		return constKind.StringValue, nil
	case *exprpb.Constant_Uint64Value:
		return constKind.Uint64Value, nil
	default:
		return nil, fmt.Errorf("expr: constant expression of kind %T not supported", constKind)
	}
}
