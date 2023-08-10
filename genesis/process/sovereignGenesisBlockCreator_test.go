package process

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/stretchr/testify/require"
)

func createGenesisBlockCreator(t *testing.T) *genesisBlockCreator {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	gbc, _ := NewGenesisBlockCreator(arg)
	return gbc
}

func TestNewSovereignGenesisBlockCreator(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		gbc := createGenesisBlockCreator(t)
		sgbc, err := NewSovereignGenesisBlockCreator(gbc)
		require.Nil(t, err)
		require.NotNil(t, sgbc)
	})

	t.Run("nil genesis block creator, should return error", func(t *testing.T) {
		sgbc, err := NewSovereignGenesisBlockCreator(nil)
		require.Equal(t, errNilGenesisBlockCreator, err)
		require.Nil(t, sgbc)
	})
}

func TestSovereignGenesisBlockCreator_CreateGenesisBlocksEmptyBlocks(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.StartEpochNum = 1
	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Equal(t, map[uint32]data.HeaderHandler{
		core.SovereignChainShardId: &block.SovereignChainHeader{
			Header: &block.Header{
				ShardID: core.SovereignChainShardId,
			},
		},
	}, blocks)
}

func TestSovereignGenesisBlockCreator_CreateGenesisBaseProcess(t *testing.T) {
	t.Parallel()

	gbc := createGenesisBlockCreator(t)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Len(t, blocks, 1)
	require.Contains(t, blocks, core.SovereignChainShardId)
}
