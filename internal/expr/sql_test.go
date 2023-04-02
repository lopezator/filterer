package expr

import (
	"reflect"
	"testing"
)

func TestSQL(t *testing.T) {
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
		name     string
		input    *Expr
		want     string
		wantArgs []interface{}
	}{
		{
			name:     "equality",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}}},
			want:     "LOWER(first_name) = (LOWER(?))",
			wantArgs: []interface{}{"A"},
		},
		{
			name:     "equality with nested string Field",
			input:    &Expr{Root: &OpExpr{Field: companyName, Op: "==", Args: []interface{}{"A"}}},
			want:     "LOWER(company->>'name') = (LOWER(?))",
			wantArgs: []interface{}{"A"},
		},
		{
			name:     "equality with nested int64 value",
			input:    &Expr{Root: &OpExpr{Field: companyEmployeeNumber, Op: "==", Args: []interface{}{1}}},
			want:     "(company->>'employee_number')::INT = (?)",
			wantArgs: []interface{}{1},
		},
		{
			name:     "equality with nested bool value",
			input:    &Expr{Root: &OpExpr{Field: companyFortune500, Op: "==", Args: []interface{}{true}}},
			want:     "(company->>'fortune500')::BOOL = (?)",
			wantArgs: []interface{}{true},
		},
		{
			name:     "equality with nested string value of depth 2",
			input:    &Expr{Root: &OpExpr{Field: companyLocationName, Op: "==", Args: []interface{}{"A"}}},
			want:     "LOWER(company->location->>'name') = (LOWER(?))",
			wantArgs: []interface{}{"A"},
		},
		{
			name:     "equality with nested int64 value of depth 2",
			input:    &Expr{Root: &OpExpr{Field: companyLocationZone, Op: "==", Args: []interface{}{1}}},
			want:     "(company->location->>'zone')::INT = (?)",
			wantArgs: []interface{}{1},
		},
		{
			name:     "equality with int64 value",
			input:    &Expr{Root: &OpExpr{Field: age, Op: "==", Args: []interface{}{"1"}}},
			want:     "age = (?)",
			wantArgs: []interface{}{"1"},
		},
		{
			name:     "not equals",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "!=", Args: []interface{}{"A"}}},
			want:     "LOWER(first_name) <> (LOWER(?))",
			wantArgs: []interface{}{"A"},
		},
		{
			name:     "not equals with int64 value",
			input:    &Expr{Root: &OpExpr{Field: age, Op: "!=", Args: []interface{}{"1"}}},
			want:     "age <> (?)",
			wantArgs: []interface{}{"1"},
		},
		{
			name: "not",
			input: &Expr{Root: &NotExpr{
				Not: &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
			}},
			want:     "NOT (LOWER(first_name) = (LOWER(?)))",
			wantArgs: []interface{}{"A"},
		},
		{
			name: "and",
			input: &Expr{Root: &AndExpr{
				Left:  &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
				Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
			}},
			want:     "(LOWER(first_name) = (LOWER(?)) AND LOWER(last_name) = (LOWER(?)))",
			wantArgs: []interface{}{"A", "B"},
		},
		{
			name: "or",
			input: &Expr{Root: &OrExpr{
				Left:  &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
				Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
			}},
			want:     "(LOWER(first_name) = (LOWER(?)) OR LOWER(last_name) = (LOWER(?)))",
			wantArgs: []interface{}{"A", "B"},
		},
		{
			name: "precedence between and/or",
			input: &Expr{Root: &OrExpr{
				Left: &AndExpr{
					Left:  &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
					Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
				},
				Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"C"}},
			}},
			want:     "((LOWER(first_name) = (LOWER(?)) AND LOWER(last_name) = (LOWER(?))) OR LOWER(last_name) = (LOWER(?)))",
			wantArgs: []interface{}{"A", "B", "C"},
		},
		{
			name: "precedence between and/or overruled by using parentheses",
			input: &Expr{Root: &AndExpr{
				Left: &OpExpr{Field: firstName, Op: "==", Args: []interface{}{"A"}},
				Right: &OrExpr{
					Left:  &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"B"}},
					Right: &OpExpr{Field: lastName, Op: "==", Args: []interface{}{"C"}},
				},
			}},
			want:     "(LOWER(first_name) = (LOWER(?)) AND (LOWER(last_name) = (LOWER(?)) OR LOWER(last_name) = (LOWER(?))))",
			wantArgs: []interface{}{"A", "B", "C"},
		},
		{
			name:     "startsWith",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "startsWith", Args: []interface{}{"A"}}},
			want:     "LOWER(first_name) LIKE (LOWER(?))",
			wantArgs: []interface{}{"A%"},
		},
		{
			name:     "startsWith containing backslash",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "startsWith", Args: []interface{}{`A\B`}}},
			want:     "LOWER(first_name) LIKE (LOWER(?))",
			wantArgs: []interface{}{`A\\B%`},
		},
		{
			name:     "endsWith",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "endsWith", Args: []interface{}{"A"}}},
			want:     "LOWER(first_name) LIKE (LOWER(?))",
			wantArgs: []interface{}{"%A"},
		},
		{
			name:     "endsWith containing backslash",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "endsWith", Args: []interface{}{`A\B`}}},
			want:     "LOWER(first_name) LIKE (LOWER(?))",
			wantArgs: []interface{}{`%A\\B`},
		},
		{
			name:     "contains",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "contains", Args: []interface{}{"A"}}},
			want:     "LOWER(first_name) LIKE (LOWER(?))",
			wantArgs: []interface{}{"%A%"},
		},
		{
			name:     "contains containing backslash",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "contains", Args: []interface{}{`A\B`}}},
			want:     "LOWER(first_name) LIKE (LOWER(?))",
			wantArgs: []interface{}{`%A\\B%`},
		},
		{
			name:     "contains with with string array field type",
			input:    &Expr{Root: &OpExpr{Field: tags, Op: "contains", Args: []interface{}{"A"}}},
			want:     "tags @> (?)",
			wantArgs: []interface{}{"{A}"},
		},
		{
			name:     "in",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "in", Args: []interface{}{"A"}}},
			want:     "LOWER(first_name) IN (LOWER(?))",
			wantArgs: []interface{}{"A"},
		},
		{
			name:     "in with multiple values",
			input:    &Expr{Root: &OpExpr{Field: firstName, Op: "in", Args: []interface{}{"A", "B"}}},
			want:     "LOWER(first_name) IN (LOWER(?),LOWER(?))",
			wantArgs: []interface{}{"A", "B"},
		},
		{
			name:     "in with int value",
			input:    &Expr{Root: &OpExpr{Field: age, Op: "in", Args: []interface{}{"35"}}},
			want:     "age IN (?)",
			wantArgs: []interface{}{"35"},
		},
		{
			name:     "in with multiple int values",
			input:    &Expr{Root: &OpExpr{Field: age, Op: "in", Args: []interface{}{"2", "15", "35"}}},
			want:     "age IN (?,?,?)",
			wantArgs: []interface{}{"2", "15", "35"},
		},
		{
			name:     "greater than with timestamp value",
			input:    &Expr{Root: &OpExpr{Field: birthDate, Op: ">", Args: []interface{}{"1983-12-10T11:03:27Z"}}},
			want:     "birth_date > (?)",
			wantArgs: []interface{}{"1983-12-10T11:03:27Z"},
		},
		{
			name:     "greaterEquals than with timestamp value",
			input:    &Expr{Root: &OpExpr{Field: birthDate, Op: ">=", Args: []interface{}{"1983-12-10T11:03:27Z"}}},
			want:     "birth_date >= (?)",
			wantArgs: []interface{}{"1983-12-10T11:03:27Z"},
		},
		{
			name:     "less than with timestamp value",
			input:    &Expr{Root: &OpExpr{Field: birthDate, Op: "<", Args: []interface{}{"1983-12-10T11:03:27Z"}}},
			want:     "birth_date < (?)",
			wantArgs: []interface{}{"1983-12-10T11:03:27Z"},
		},
		{
			name:     "lessEquals than with timestamp value",
			input:    &Expr{Root: &OpExpr{Field: birthDate, Op: "<=", Args: []interface{}{"1983-12-10T11:03:27Z"}}},
			want:     "birth_date <= (?)",
			wantArgs: []interface{}{"1983-12-10T11:03:27Z"},
		},
		{
			name:     "present",
			input:    &Expr{Root: &PresentExpr{Field: firstName}},
			want:     "first_name IS NOT NULL",
			wantArgs: []interface{}{},
		},
		{
			name:     "present with nested",
			input:    &Expr{Root: &PresentExpr{Field: companyName}},
			want:     "company->>'name' IS NOT NULL",
			wantArgs: []interface{}{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotClause, gotArgs, err := SQL(tt.input)
			if err != nil {
				t.Errorf("SQL() error: %v", err)
			}
			if gotClause != tt.want {
				t.Errorf("SQL() got: %v, want %v", gotClause, tt.want)
			}
			if !reflect.DeepEqual(tt.wantArgs, gotArgs) {
				t.Errorf("SQL() got: %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
