package filterer

import (
	"fmt"
	"testing"

	"vitess.io/vitess/go/vt/sqlparser"
)

func TestParsing(t *testing.T) {
	tests := []struct {
		baseQuery   string
		whereClause string
		expected    string
	}{
		{
			baseQuery:   "SELECT * FROM users",
			whereClause: "user_id = 2",
			expected:    "select * from users where user_id = 2",
		},
		{
			baseQuery:   "SELECT * FROM users WHERE display_name = 'paco'",
			whereClause: "user_id = 2",
			expected:    "select * from users where display_name = 'paco' and user_id = 2",
		},
		{
			baseQuery:   "SELECT * FROM users JOIN dogs on users.id = dogs.user_id WHERE display_name = 'paco' GROUP BY users.id ORDER BY users.display_name DESC",
			whereClause: "user_id = 2",
			expected:    "select * from users join dogs on users.id = dogs.user_id where display_name = 'paco' and user_id = 2 group by users.id order by users.display_name desc",
		},
	}

	parser, err := sqlparser.New(sqlparser.Options{})
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("baseQuery: %s, whereClause: %s", tt.baseQuery, tt.whereClause), func(t *testing.T) {
			// Parse the base query
			stmt, err := parser.Parse(tt.baseQuery)
			if err != nil {
				t.Fatalf("failed to parse base query: %v", err)
			}

			// Ensure the base query is a SELECT statement
			selectStmt, ok := stmt.(*sqlparser.Select)
			if !ok {
				t.Fatalf("base query must be a SELECT statement")
			}

			// Parse the additional WHERE clause
			whereExpr, err := parser.ParseExpr(tt.whereClause)
			if err != nil {
				t.Fatalf("failed to parse additional WHERE clause: %v", err)
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

			// Generate the final query
			finalQuery := sqlparser.String(selectStmt)
			if finalQuery != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, finalQuery)
			}
		})
	}
}
