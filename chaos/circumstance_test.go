package chaos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvalExpression(t *testing.T) {
	tests := []struct {
		expression string
		variables  map[string]int
		expected   bool
		expectErr  bool
	}{
		{
			expression: "a > b",
			variables:  map[string]int{"a": 5, "b": 3},
			expected:   true,
			expectErr:  false,
		},
		{
			expression: "a < b",
			variables:  map[string]int{"a": 2, "b": 3},
			expected:   true,
			expectErr:  false,
		},
		{
			expression: "a == b",
			variables:  map[string]int{"a": 3, "b": 3},
			expected:   true,
			expectErr:  false,
		},
		{
			expression: "a != b",
			variables:  map[string]int{"a": 3, "b": 4},
			expected:   true,
			expectErr:  false,
		},
		{
			expression: "a > b",
			variables:  map[string]int{"a": 1, "b": 3},
			expected:   false,
			expectErr:  false,
		},
		{
			expression: "invalid expression",
			variables:  map[string]int{"a": 1, "b": 3},
			expected:   false,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			result, err := evalExpression(tt.expression, tt.variables)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
