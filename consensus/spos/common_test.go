package spos

import (
	"testing"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/stretchr/testify/require"
)

func TestShouldConsiderSelfKeyInConsensus(t *testing.T) {
	t.Parallel()

	require.False(t, ShouldConsiderSelfKeyInConsensus(nil))
	require.True(t, ShouldConsiderSelfKeyInConsensus(&mock.NodeRedundancyHandlerStub{
		IsRedundancyNodeCalled: func() bool {
			return false
		},
	}))
	require.True(t, ShouldConsiderSelfKeyInConsensus(&mock.NodeRedundancyHandlerStub{
		IsMainMachineActiveCalled: func() bool {
			return true
		},
	}))
}
