package disabled

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"
)

func TestIncomingHeaderProcessor_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	ihp := NewDisabledIncomingHeaderProcessor()
	require.False(t, check.IfNil(ihp))

	require.NotPanics(t, func() {
		err := ihp.AddHeader([]byte("hash"), &sovereign.IncomingHeader{})
		require.NoError(t, err)

		extendedHeader, err := ihp.CreateExtendedHeader(&block.ShardHeaderExtended{})
		require.Equal(t, &block.ShardHeaderExtended{}, extendedHeader)
		require.NoError(t, err)
	})
}
