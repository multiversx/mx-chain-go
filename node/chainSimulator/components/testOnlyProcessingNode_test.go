package components

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsTestOnlyProcessingNode(t *testing.T) ArgsTestOnlyProcessingNode {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:           3,
		OriginalConfigsPath:   "../../../cmd/node/config/",
		GenesisTimeStamp:      0,
		RoundDurationInMillis: 6000,
		TempDir:               t.TempDir(),
	})
	require.Nil(t, err)

	return ArgsTestOnlyProcessingNode{
		Configs:             *outputConfigs.Configs,
		GasScheduleFilename: outputConfigs.GasScheduleFilename,
		NumShards:           3,

		SyncedBroadcastNetwork: NewSyncedBroadcastNetwork(),
		ChanStopNodeProcess:    make(chan endProcess.ArgEndProcess),
		APIInterface:           api.NewNoApiInterface(),
	}
}

func TestNewTestOnlyProcessingNode(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		if testing.Short() {
			t.Skip("cannot run with -race -short; requires Wasm VM fix")
		}

		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})

	t.Run("try commit a block", func(t *testing.T) {
		if testing.Short() {
			t.Skip("cannot run with -race -short; requires Wasm VM fix")
		}

		t.Parallel()

		args := createMockArgsTestOnlyProcessingNode(t)
		node, err := NewTestOnlyProcessingNode(args)
		assert.Nil(t, err)
		assert.NotNil(t, node)

		newHeader, err := node.ProcessComponentsHolder.BlockProcessor().CreateNewHeader(1, 1)
		assert.Nil(t, err)

		err = newHeader.SetPrevHash(node.ChainHandler.GetGenesisHeaderHash())
		assert.Nil(t, err)

		header, block, err := node.ProcessComponentsHolder.BlockProcessor().CreateBlock(newHeader, func() bool {
			return true
		})
		assert.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, block)

		err = node.ProcessComponentsHolder.BlockProcessor().ProcessBlock(header, block, func() time.Duration {
			return 1000
		})
		assert.Nil(t, err)

		err = node.ProcessComponentsHolder.BlockProcessor().CommitBlock(header, block)
		assert.Nil(t, err)
	})
}
