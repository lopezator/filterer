package expr

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// SQL returns a database friendly format composed by a string clause and a
// slice of args
func SQL(expr *Expr) (string, []interface{}, error) {
	return walkSQL(expr.Root)
}

type sqlOperator struct {
	name        string
	argModifier func(interface{}) interface{}
}

var sqlOperatorLookup = map[string]map[FieldType]*sqlOperator{
	OperatorEquals: {
		StringFieldType:  {name: "="},
		IntegerFieldType: {name: "="},
		BoolFieldType:    {name: "="},
	},
	OperatorNotEquals: {
		StringFieldType:  {name: "<>"}, // equivalent to != but SQL-92 compliant
		IntegerFieldType: {name: "<>"}, // equivalent to != but SQL-92 compliant
	},
	OperatorGreater: {
		TimestampFieldType: {name: ">"},
	},
	OperatorGreaterEquals: {
		TimestampFieldType: {name: ">="},
	},
	OperatorLess: {
		TimestampFieldType: {name: "<"},
	},
	OperatorLessEquals: {
		TimestampFieldType: {name: "<="},
	},
	OperatorIn: {
		StringFieldType:  {name: "IN"},
		IntegerFieldType: {name: "IN"},
	},
	OperatorStartsWith: {
		StringFieldType: {
			name:        "LIKE",
			argModifier: func(v interface{}) interface{} { return escapeLikeArg(v) + "%" },
		},
	},
	OperatorEndsWith: {
		StringFieldType: {
			name:        "LIKE",
			argModifier: func(v interface{}) interface{} { return "%" + escapeLikeArg(v) },
		},
	},
	OperatorContains: {
		StringFieldType: {
			name:        "LIKE",
			argModifier: func(v interface{}) interface{} { return "%" + escapeLikeArg(v) + "%" },
		},
		StringArrayFieldType: {
			name:        "@>",
			argModifier: func(v interface{}) interface{} { return "{" + escapeLikeArg(v) + "}" },
		},
	},
}

// escapeLikeArg gets a SQL arg and returns its equivalent SQL-LIKE needed
// escaped arg. In SQL-92, backslash is used as the escape character, thus any
// arg containing a backslash needs to be doubled in order to be escaped
// correctly inside LIKE queries.
func escapeLikeArg(arg interface{}) string {
	return strings.ReplaceAll(arg.(string), `\`, `\\`)
}

func walkSQL(node Node) (string, []interface{}, error) {
	switch e := node.(type) {
	case *NotExpr:
		clause, args, err := walkSQL(e.Not)
		if err != nil {
			return "", nil, err
		}
		// Although it seems countersqlintuitive, these kind of queries do the
		// same if the NOT is used just before the IN operation or embracing
		// everything including the ident as the query planner correctly chooses
		// what to do:
		//
		// sql> EXPLAIN SELECT * FROM users WHERE NOT (name_id IN ('burt-warren'));
		// tree | field  |           description
		// +------+--------+---------------------------------+
		// scan |        |
		// 	| table  | users@primary
		// 	| spans  | ALL
		// 	| filter | name_id NOT IN ('burt-warren',)
		//
		// sql> EXPLAIN SELECT * FROM users WHERE name_id NOT IN ('burt-warren');
		// tree | field  |           description
		// +------+--------+---------------------------------+
		// scan |        |
		// 	| table  | users@primary
		// 	| spans  | ALL
		// 	| filter | name_id NOT IN ('burt-warren',)
		return fmt.Sprintf("NOT (%s)", clause), args, nil
	case *AndExpr:
		lclause, largs, err := walkSQL(e.Left)
		if err != nil {
			return "", nil, err
		}
		rclause, rargs, err := walkSQL(e.Right)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("(%s AND %s)", lclause, rclause), append(largs, rargs...), nil
	case *OrExpr:
		lclause, largs, err := walkSQL(e.Left)
		if err != nil {
			return "", nil, err
		}
		rclause, rargs, err := walkSQL(e.Right)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("(%s OR %s)", lclause, rclause), append(largs, rargs...), nil
	case *OpExpr:
		sqlOp, ok := sqlOperatorLookup[e.Op][e.Field.Ftype]
		if !ok {
			return "", nil, errors.New("expr: unsupported operation expression")
		}

		var args []interface{}
		if sqlOp.argModifier != nil {
			args = make([]interface{}, len(e.Args))
			for i := 0; i < len(e.Args); i++ {
				args[i] = sqlOp.argModifier(e.Args[i])
			}
		} else {
			args = e.Args
		}

		columnName, err := columnName(e.Field.Name, e.Field.Ftype, len(args))
		if err != nil {
			return "", nil, err
		}

		// As this SQL is supported SELECT ... WHERE name_id = ('burt-warren'),
		// we choose to always embrace with parentheses.
		switch e.Field.Ftype {
		// Enclose field and args with LOWER() in case of string query for a
		// case insensitive query.
		case StringFieldType:
			parameters := fmt.Sprintf("(%s)", strings.TrimRight(strings.Repeat("LOWER(?),", len(args)), ","))
			return fmt.Sprintf("LOWER(%s) %s %s", columnName, sqlOp.name, parameters), args, nil
		default:
			parameters := fmt.Sprintf("(%s)", strings.TrimRight(strings.Repeat("?,", len(args)), ","))
			return fmt.Sprintf("%s %s %s", columnName, sqlOp.name, parameters), args, nil
		}
	case *PresentExpr:
		columnName, err := columnName(e.Field.Name, e.Field.Ftype, 0)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("%s IS NOT NULL", columnName), []interface{}{}, nil
	default:
		return "", nil, errors.New("expr: unsupported expression")
	}
}

// This regex is used to obtain the values before and after the last operator "->"
// example: "apple_hub->apple_key->owner_name" would be:
// - $1 := "apple_hub->apple_key"
// - $2 := "owner_name"
var re = regexp.MustCompile(`(.*)->(\w+)$`)

// Using the regex above, we can transform the string to allow nested searches.
// example: "apple_hub.apple_key.owner_name" would be:
// - "apple_hub->apple_key->>'owner_name'"
func columnName(fieldName string, fieldType FieldType, numArgs int) (string, error) {
	// nested equality logic
	if strings.Contains(fieldName, ".") {
		if numArgs > 1 {
			return "", fmt.Errorf("expr: unsupported multiple args for nested type")
		}
		fieldName = strings.ReplaceAll(fieldName, ".", "->")
		fieldName = re.ReplaceAllString(fieldName, `$1->>'$2'`)
		switch fieldType {
		case BoolFieldType:
			fieldName = fmt.Sprintf("(%s)::BOOL", fieldName)
		case IntegerFieldType:
			fieldName = fmt.Sprintf("(%s)::INT", fieldName)
		case DoubleFieldType:
			fieldName = fmt.Sprintf("(%s)::FLOAT", fieldName)
		case BytesFieldType:
			fieldName = fmt.Sprintf("(%s)::BYTES", fieldName)
		case TimestampFieldType:
			fieldName = fmt.Sprintf("(%s)::TIMESTAMP", fieldName)
		}
	}
	return fieldName, nil
}
