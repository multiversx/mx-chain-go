package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTopicsChecker_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	tc := NewDisabledTopicsChecker()
	require.False(t, check.IfNil(tc))

	require.NotPanics(t, func() {
		err := tc.CheckValidity([][]byte{[]byte("topic")})
		require.NoError(t, err)
	})
}
