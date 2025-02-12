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

		nodeIndex:     7,
		nodePublicKey: "abba",
		amILeader:     true,
		blockNonce:    1000,
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
	require.True(t, circumstance.anyExpression([]string{"nodePublicKey == \"abba\""}))
	require.True(t, circumstance.anyExpression([]string{"nodeIndex == 7"}))
	require.True(t, circumstance.anyExpression([]string{"amILeader == true"}))
	require.False(t, circumstance.anyExpression([]string{"a == 42"}))
	require.True(t, circumstance.anyExpression([]string{"a == 42", "true"}))
	require.False(t, circumstance.anyExpression([]string{"now == \"hello\""}))
}
