package consensus_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common/consensus"
	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/stretchr/testify/assert"
)

func Test_ShouldConsiderSelfKeyInConsensus(t *testing.T) {
	t.Parallel()

	// nil handler
	assert.False(t, consensus.ShouldConsiderSelfKeyInConsensus(nil))

	// main machine
	assert.True(t, consensus.ShouldConsiderSelfKeyInConsensus(
		&consensusMock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool { return false },
		},
	))

	// backup with active main
	assert.False(t, consensus.ShouldConsiderSelfKeyInConsensus(
		&consensusMock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled:    func() bool { return true },
			IsMainMachineActiveCalled: func() bool { return true },
		},
	))

	// backup with inactive main
	assert.True(t, consensus.ShouldConsiderSelfKeyInConsensus(
		&consensusMock.NodeRedundancyHandlerStub{
			IsRedundancyNodeCalled:    func() bool { return true },
			IsMainMachineActiveCalled: func() bool { return false },
		},
	))
}
