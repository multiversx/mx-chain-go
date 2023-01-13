package notifier

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewNodesConfigProvider(t *testing.T) {
	t.Parallel()

	ncp, err := NewNodesConfigProvider(nil, nil)
	require.Equal(t, process.ErrNilEpochNotifier, err)
	require.True(t, ncp.IsInterfaceNil())

	epochNotifier := forking.NewGenericEpochNotifier()
	ncp, err = NewNodesConfigProvider(epochNotifier, nil)
	require.Nil(t, err)
	require.False(t, ncp.IsInterfaceNil())
}

func TestNodesConfigProvider_GetAllNodesConfigSorted(t *testing.T) {
	t.Parallel()

	nodesConfigEpoch0 := config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            36,
		NodesToShufflePerShard: 4,
	}
	nodesConfigEpoch1 := config.MaxNodesChangeConfig{
		EpochEnable:            1,
		MaxNumNodes:            56,
		NodesToShufflePerShard: 2,
	}
	nodesConfigEpoch6 := config.MaxNodesChangeConfig{
		EpochEnable:            6,
		MaxNumNodes:            48,
		NodesToShufflePerShard: 1,
	}

	unsortedNodesConfig := []config.MaxNodesChangeConfig{
		nodesConfigEpoch6,
		nodesConfigEpoch0,
		nodesConfigEpoch1,
	}
	sortedNodesConfig := []config.MaxNodesChangeConfig{
		nodesConfigEpoch0,
		nodesConfigEpoch1,
		nodesConfigEpoch6,
	}

	epochNotifier := forking.NewGenericEpochNotifier()
	ncp, _ := NewNodesConfigProvider(epochNotifier, unsortedNodesConfig)
	require.Equal(t, sortedNodesConfig, ncp.GetAllNodesConfig())
}

func TestNodesConfigProvider_EpochConfirmedCorrectMaxNumNodesAfterNodeRestart(t *testing.T) {
	t.Parallel()

	nodesConfigEpoch0 := config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            36,
		NodesToShufflePerShard: 4,
	}
	nodesConfigEpoch1 := config.MaxNodesChangeConfig{
		EpochEnable:            1,
		MaxNumNodes:            56,
		NodesToShufflePerShard: 2,
	}
	nodesConfigEpoch6 := config.MaxNodesChangeConfig{
		EpochEnable:            6,
		MaxNumNodes:            48,
		NodesToShufflePerShard: 1,
	}

	allNodesConfig := []config.MaxNodesChangeConfig{
		nodesConfigEpoch0,
		nodesConfigEpoch1,
		nodesConfigEpoch6,
	}
	epochNotifier := forking.NewGenericEpochNotifier()
	ncp, _ := NewNodesConfigProvider(epochNotifier, allNodesConfig)

	epochNotifier.CheckEpoch(&block.Header{Epoch: 0})
	require.Equal(t, nodesConfigEpoch0, ncp.GetCurrentNodesConfig())

	epochNotifier.CheckEpoch(&block.Header{Epoch: 1})
	require.Equal(t, nodesConfigEpoch1, ncp.GetCurrentNodesConfig())

	for epoch := uint32(2); epoch <= 5; epoch++ {
		epochNotifier.CheckEpoch(&block.Header{Epoch: epoch})
		require.Equal(t, nodesConfigEpoch1, ncp.GetCurrentNodesConfig())
	}

	// simulate restart
	epochNotifier.CheckEpoch(&block.Header{Epoch: 0})
	epochNotifier.CheckEpoch(&block.Header{Epoch: 5})
	require.Equal(t, nodesConfigEpoch1, ncp.GetCurrentNodesConfig())

	epochNotifier.CheckEpoch(&block.Header{Epoch: 6})
	require.Equal(t, nodesConfigEpoch6, ncp.GetCurrentNodesConfig())

	// simulate restart
	epochNotifier.CheckEpoch(&block.Header{Epoch: 0})
	epochNotifier.CheckEpoch(&block.Header{Epoch: 6})
	require.Equal(t, nodesConfigEpoch6, ncp.GetCurrentNodesConfig())

	for epoch := uint32(7); epoch <= 20; epoch++ {
		epochNotifier.CheckEpoch(&block.Header{Epoch: epoch})
		require.Equal(t, nodesConfigEpoch6, ncp.GetCurrentNodesConfig())
	}

	// simulate restart
	epochNotifier.CheckEpoch(&block.Header{Epoch: 1})
	epochNotifier.CheckEpoch(&block.Header{Epoch: 21})
	require.Equal(t, nodesConfigEpoch6, ncp.GetCurrentNodesConfig())
}
