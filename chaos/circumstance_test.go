package chaos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCircumstance_EvalExpression(t *testing.T) {
	circumstance := &failureCircumstance{
		nodePublicKey:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		transactionHash: []byte{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}

	require.True(t, doEvalExpression(t, circumstance, "nodePublicKeyLastByte == 10"))
	require.True(t, doEvalExpression(t, circumstance, "transactionHashLastByte == 1"))

	doEvalExpressionExpectError(t, circumstance, "a == 42", "undefined: a")
	doEvalExpressionExpectError(t, circumstance, "now == \"hello\"", "mismatched types uint64 and untyped string")
}

func doEvalExpression(t *testing.T, circumstance *failureCircumstance, expression string) bool {
	result, err := circumstance.evalExpression(expression)
	require.NoError(t, err)
	return result
}

func doEvalExpressionExpectError(t *testing.T, circumstance *failureCircumstance, expression string, errorMessage string) {
	result, err := circumstance.evalExpression(expression)
	require.ErrorContains(t, err, errorMessage)
	require.False(t, result)
}
