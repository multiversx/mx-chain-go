package bootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveMiniBlocksFromComponents(t *testing.T) {
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}

	mb1 := &block.MiniBlock{
		Type:          block.TxBlock,
		SenderShardID: 0,
	}
	mb2 := &block.MiniBlock{
		Type:          block.SmartContractResultBlock,
		SenderShardID: core.MetachainShardId,
	}
	mb3 := &block.MiniBlock{
		Type:            block.PeerBlock,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
	}
	mb4 := &block.MiniBlock{
		Type:            block.PeerBlock,
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 1,
	}

	mb3Hash, err := core.CalculateHash(marshaller, hasher, mb3)
	assert.Nil(t, err)

	mb4Hash, err := core.CalculateHash(marshaller, hasher, mb4)
	assert.Nil(t, err)

	components := &ComponentsNeededForBootstrap{
		PendingMiniBlocks: map[string]*block.MiniBlock{
			"mb1": mb1,
			"mb2": mb2,
		},
		PeerMiniBlocks: []*block.MiniBlock{mb3, mb4},
	}

	receivedMiniblocks := make(map[string]*block.MiniBlock)
	storageService := &storage.ChainStorerStub{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			if unitType != dataRetriever.MiniBlockUnit && unitType != dataRetriever.EpochStartMetaBlockUnit {
				assert.Fail(t, "invalid storage unit type")
			}

			mb := &block.MiniBlock{}
			err := marshaller.Unmarshal(mb, value)
			assert.Nil(t, err)

			receivedMiniblocks[string(key)] = mb

			return nil
		},
	}

	handler := &baseStorageHandler{
		marshalizer:    marshaller,
		hasher:         hasher,
		storageService: storageService,
	}

	handler.saveMiniblocksFromComponents(components)

	expectedMiniBlocks := map[string]*block.MiniBlock{
		"mb1":           mb1,
		"mb2":           mb2,
		string(mb3Hash): mb3,
		string(mb4Hash): mb4,
	}

	testMaps(t, expectedMiniBlocks, receivedMiniblocks)
}

func testMaps(tb testing.TB, expected, actual map[string]*block.MiniBlock) {
	require.Equal(tb, len(expected), len(actual))
	for key, mbExpected := range expected {
		mbActual := actual[key]
		require.Equal(tb, mbActual, mbExpected)
	}
}
