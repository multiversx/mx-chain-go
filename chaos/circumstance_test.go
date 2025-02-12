package chaos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCircumstance_EvalExpression(t *testing.T) {
	circumstance := &failureCircumstance{
		randomNumber:    42,
		now:             1234567890,
		nodeDisplayName: "dummy",
		shard:           1,
		epoch:           7,
		round:           1001,

		blockNonce:      1000,
		nodePublicKey:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		transactionHash: []byte{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}

	require.True(t, circumstance.anyExpression([]string{"true"}))
	require.False(t, circumstance.anyExpression([]string{"false"}))

	require.True(t, circumstance.anyExpression([]string{"nodeDisplayName == \"dummy\""}))
	require.False(t, circumstance.anyExpression([]string{"nodeDisplayName != \"dummy\""}))
	require.True(t, circumstance.anyExpression([]string{"nodeDisplayName != \"foo\""}))
	require.True(t, circumstance.anyExpression([]string{"randomNumber == 42"}))
	require.True(t, circumstance.anyExpression([]string{"randomNumber == 42", "randomNumber == 43"}))
	require.True(t, circumstance.anyExpression([]string{"randomNumber > 41 && randomNumber < 43"}))
	require.False(t, circumstance.anyExpression([]string{"randomNumber % 3 != 0"}))
	require.True(t, circumstance.anyExpression([]string{"nodePublicKeyLastByte == 10"}))
	require.True(t, circumstance.anyExpression([]string{"transactionHashLastByte == 1"}))
	require.False(t, circumstance.anyExpression([]string{"a == 42"}))
	require.True(t, circumstance.anyExpression([]string{"a == 42", "true"}))
	require.False(t, circumstance.anyExpression([]string{"now == \"hello\""}))
}
