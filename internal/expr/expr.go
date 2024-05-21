package expr

const (
	OperatorEquals        = "=="
	OperatorNotEquals     = "!="
	OperatorGreater       = ">"
	OperatorGreaterEquals = ">="
	OperatorLess          = "<"
	OperatorLessEquals    = "<="
	OperatorIn            = "in"
	OperatorStartsWith    = "startsWith"
	OperatorEndsWith      = "endsWith"
	OperatorContains      = "contains"
)

// Node defines a node in a abstract syntax tree.
type Node any

// Expr is the abstract representation of a previously filtered, parsed and
// checked expression.
type Expr struct {
	Root Node
}

// IsZero returns true if the expression was neither created or initialized.
func (e *Expr) IsZero() bool {
	return e == nil || e.Root == nil
}

// NotExpr represents a NOT expression node.
type NotExpr struct {
	Not Node
}

// AndExpr represents an AND expression node.
type AndExpr struct {
	Left  Node
	Right Node
}

// OrExpr represents an OR expression node.
type OrExpr struct {
	Left  Node
	Right Node
}

// OpExpr represents an operation expression node.
type OpExpr struct {
	Left Node
	Op   string
	Args []any
}

// PresentExpr represents a presence expression node.
type PresentExpr struct {
	Field *Field
}

// SizeExpr represents a size expression node.
type SizeExpr struct {
	Field *Field
}

// FieldType defines a field type.
type FieldType byte

// Field type values
const (
	BoolFieldType FieldType = iota
	IntegerFieldType
	DoubleFieldType
	StringFieldType
	BytesFieldType
	TimestampFieldType
	StringArrayFieldType
)

// Field represents a field with its name and type.
type Field struct {
	Name  string
	Ftype FieldType
}
