package expr

import (
	"reflect"
	"strings"
	"testing"
	"time"

	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func mustParseTimestamp(t *testing.T, value string) time.Time {
	tm, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("timestamp parse failed: %v", err)
	}
	return tm
}

func TestParse(t *testing.T) {
	t.Parallel()

	firstName := &Field{Name: "first_name", Ftype: StringFieldType}
	lastName := &Field{Name: "last_name", Ftype: StringFieldType}
	companyName := &Field{Name: "company.name", Ftype: StringFieldType}
	companyEmployeeNumber := &Field{Name: "company.employee_number", Ftype: IntegerFieldType}
	companyLocationName := &Field{Name: "company.location.name", Ftype: StringFieldType}
	companyLocationZone := &Field{Name: "company.location.zone", Ftype: IntegerFieldType}
	companyFortune500 := &Field{Name: "company.fortune500", Ftype: BoolFieldType}
	age := &Field{Name: "age", Ftype: IntegerFieldType}
	birthDate := &Field{Name: "birth_date", Ftype: TimestampFieldType}
	tags := &Field{Name: "tags", Ftype: StringArrayFieldType}

	tests := []struct {
		name    string
		input   string
		want    *Expr
		wantErr bool
	}{
		{
			name:  "equality",
			input: "first_name == 'A'",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}}},
		},
		{
			name:  "equality with nested value",
			input: "company.name == 'A'",
			want:  &Expr{Root: &OpExpr{Field: companyName, Op: "==", Args: []interface{}{"A"}}},
		},
		{
			name:  "equality with nested int value",
			input: "company.employee_number == 1",
			want:  &Expr{Root: &OpExpr{Field: companyEmployeeNumber, Op: "==", Args: []interface{}{int64(1)}}},
		},
		{
			name:  "equality with nested value of depth 2",
			input: "company.location.name == 'A'",
			want:  &Expr{Root: &OpExpr{Field: companyLocationName, Op: "==", Args: []interface{}{"A"}}},
		},
		{
			name:  "equality with nested int value of depth 2",
			input: "company.location.zone == 1",
			want:  &Expr{Root: &OpExpr{Field: companyLocationZone, Op: "==", Args: []interface{}{int64(1)}}},
		},
		{
			name:  "equality with int value",
			input: "age == 35",
			want:  &Expr{Root: &OpExpr{Field: age, Op: "==", Args: []interface{}{int64(35)}}},
		},
		{
			name:  "equality with bool value",
			input: "company.fortune500 == true",
			want:  &Expr{Root: &OpExpr{Field: companyFortune500, Op: "==", Args: []interface{}{true}}},
		},
		{
			name:  "not equals",
			input: "first_name != 'A'",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "!=", Args: []interface{}{"A"}}},
		},
		{
			name:  "not equals with int value",
			input: "age != 35",
			want:  &Expr{Root: &OpExpr{Field: age, Op: "!=", Args: []interface{}{int64(35)}}},
		},
		{
			name:  "not",
			input: "!(first_name == 'A')",
			want: &Expr{Root: &NotExpr{
				Not: &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
			}},
		},
		{
			name:  "and",
			input: "first_name == 'A' && last_name == 'B'",
			want: &Expr{Root: &AndExpr{
				Left:  &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
				Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
			}},
		},
		{
			name:  "or",
			input: "first_name == 'A' || last_name == 'B'",
			want: &Expr{Root: &OrExpr{
				Left:  &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
				Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
			}},
		},
		{
			name:  "precedence between and/or",
			input: "first_name == 'A' && last_name == 'B' || last_name == 'C'",
			want: &Expr{Root: &OrExpr{
				Left: &AndExpr{
					Left:  &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
					Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
				},
				Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"C"}},
			}},
		},
		{
			name:  "precedence between and/or overruled by using parentheses",
			input: "first_name == 'A' && (last_name == 'B' || last_name == 'C')",
			want: &Expr{Root: &AndExpr{
				Left: &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
				Right: &OrExpr{
					Left:  &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
					Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"C"}},
				},
			}},
		},
		{
			name:  "startsWith",
			input: "first_name.startsWith('A')",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "startsWith", Args: []interface{}{"A"}}},
		},
		{
			name:  "endsWith",
			input: "first_name.endsWith('A')",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "endsWith", Args: []interface{}{"A"}}},
		},
		{
			name:  "contains",
			input: "first_name.contains('A')",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "contains", Args: []interface{}{"A"}}},
		},
		{
			name:  "contains with with string array field and string value",
			input: "tags.contains('A')",
			want:  &Expr{Root: &OpExpr{Field: tags, Op: "contains", Args: []interface{}{"A"}}},
		},
		{
			name:  "in",
			input: "first_name in ['A']",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "in", Args: []interface{}{"A"}}},
		},
		{
			name:  "in with multiple values",
			input: "first_name in ['A', 'B']",
			want:  &Expr{Root: &OpExpr{Field: firstName, Op: "in", Args: []interface{}{"A", "B"}}},
		},
		{
			name:  "in with int value",
			input: "age in [35]",
			want:  &Expr{Root: &OpExpr{Field: age, Op: "in", Args: []interface{}{int64(35)}}},
		},
		{
			name:  "in with multiple int values",
			input: "age in [2, 15, 35]",
			want:  &Expr{Root: &OpExpr{Field: age, Op: "in", Args: []interface{}{int64(2), int64(15), int64(35)}}},
		},
		{
			name:    "exceed max expression depth",
			input:   strings.TrimSuffix(strings.Repeat("first_name == 'A' ||", 17), " ||"),
			wantErr: true,
		},
		{
			name:    "disallow ident in value",
			input:   "first_name == first_name",
			wantErr: true,
		},
		{
			name:    "disallow ident in list values",
			input:   "first_name in [first_name]",
			wantErr: true,
		},
		{
			name:    "disallow type mismatch in value",
			input:   "first_name == 1",
			wantErr: true,
		},
		{
			name:    "disallow type mismatch in list values",
			input:   "first_name == [1]",
			wantErr: true,
		},
		{
			name:    "disallow mixed types mismatch in list values",
			input:   `first_name in [1, 1.3, "foor"]`,
			wantErr: true,
		},
		{
			name:    "greater than with timestamp value",
			input:   `birth_date > timestamp("1983-12-10T11:03:27Z")`,
			want:    &Expr{Root: &OpExpr{Field: birthDate, Op: ">", Args: []interface{}{mustParseTimestamp(t, "1983-12-10T11:03:27Z")}}},
			wantErr: false,
		},
		{
			name:    "greaterEquals than with timestamp value",
			input:   `birth_date >= timestamp("1983-12-10T11:03:28Z")`,
			want:    &Expr{Root: &OpExpr{Field: birthDate, Op: ">=", Args: []interface{}{mustParseTimestamp(t, "1983-12-10T11:03:28Z")}}},
			wantErr: false,
		},
		{
			name:    "less than with timestamp value",
			input:   `birth_date < timestamp("1983-12-10T11:03:29Z")`,
			want:    &Expr{Root: &OpExpr{Field: birthDate, Op: "<", Args: []interface{}{mustParseTimestamp(t, "1983-12-10T11:03:29Z")}}},
			wantErr: false,
		},
		{
			name:    "lessEquals than with timestamp value",
			input:   `birth_date <= timestamp("1983-12-10T11:03:30Z")`,
			want:    &Expr{Root: &OpExpr{Field: birthDate, Op: "<=", Args: []interface{}{mustParseTimestamp(t, "1983-12-10T11:03:30Z")}}},
			wantErr: false,
		},
		{
			name:    "disallow comparing against incorrect timestamp string format",
			input:   `birth_date > timestamp("foo")`,
			wantErr: true,
		},
		{
			name:  "present string",
			input: "present(first_name)",
			want:  &Expr{Root: &PresentExpr{Field: firstName}},
		},
		{
			name:  "present int",
			input: "present(age)",
			want:  &Expr{Root: &PresentExpr{Field: age}},
		},
		{
			name:  "present nested",
			input: "present(company.name)",
			want:  &Expr{Root: &PresentExpr{Field: companyName}},
		},
	}

	parser, err := NewParser(map[string]*exprpb.Type{
		"first_name":              {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_STRING}},
		"last_name":               {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_STRING}},
		"company.name":            {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_STRING}},
		"company.employee_number": {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_INT64}},
		"company.location.name":   {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_STRING}},
		"company.location.zone":   {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_INT64}},
		"company.fortune500":      {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_BOOL}},
		"age":                     {TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_INT64}},
		"birth_date":              {TypeKind: &exprpb.Type_WellKnown{WellKnown: exprpb.Type_TIMESTAMP}},
		"tags": {TypeKind: &exprpb.Type_ListType_{ListType: &exprpb.Type_ListType{
			ElemType: &exprpb.Type{TypeKind: &exprpb.Type_Primitive{Primitive: exprpb.Type_STRING}},
		}}},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotExpr, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error: %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gotExpr, tt.want) {
				t.Errorf("Parse() got: %v, want %v", gotExpr, tt.want)
			}
		})
	}
}
