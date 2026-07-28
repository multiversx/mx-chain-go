package components

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	consensusMessage "github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstantBroadcastMessenger(t *testing.T) {
	t.Parallel()

	t.Run("nil broadcastMessenger should error", func(t *testing.T) {
		t.Parallel()

		mes, err := NewInstantBroadcastMessenger(nil, nil)
		require.Equal(t, errorsMx.ErrNilBroadcastMessenger, err)
		require.Nil(t, mes)
	})
	t.Run("nil shardCoordinator should error", func(t *testing.T) {
		t.Parallel()

		mes, err := NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{}, nil)
		require.Equal(t, errorsMx.ErrNilShardCoordinator, err)
		require.Nil(t, mes)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mes, err := NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{}, &mock.ShardCoordinatorMock{})
		require.NoError(t, err)
		require.NotNil(t, mes)
	})
}

func TestInstantBroadcastMessenger_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var mes *instantBroadcastMessenger
	require.True(t, mes.IsInterfaceNil())

	mes, _ = NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{}, &mock.ShardCoordinatorMock{})
	require.False(t, mes.IsInterfaceNil())
}

func TestInstantBroadcastMessenger_BroadcastBlockDataLeader(t *testing.T) {
	t.Parallel()

	t.Run("meta should work", func(t *testing.T) {
		t.Parallel()

		providedMBs := map[uint32][]byte{
			0:                       []byte("mb shard 0"),
			1:                       []byte("mb shard 1"),
			common.MetachainShardId: []byte("mb shard meta"),
		}
		providedTxs := map[string][][]byte{
			"topic_0": {[]byte("txs topic 0")},
			"topic_1": {[]byte("txs topic 1")},
		}
		mes, err := NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{
			BroadcastMiniBlocksCalled: func(mbs map[uint32][]byte, bytes []byte) error {
				require.Equal(t, providedMBs, mbs)
				return expectedErr // for coverage only
			},
			BroadcastTransactionsCalled: func(txs map[string][][]byte, bytes []byte) error {
				require.Equal(t, providedTxs, txs)
				return expectedErr // for coverage only
			},
		}, &mock.ShardCoordinatorMock{
			ShardID: common.MetachainShardId,
		})
		require.NoError(t, err)

		err = mes.BroadcastBlockDataLeader(nil, providedMBs, providedTxs, []byte("pk"))
		require.NoError(t, err)
	})
	t.Run("shard should work", func(t *testing.T) {
		t.Parallel()

		providedMBs := map[uint32][]byte{
			0:                       []byte("mb shard 0"), // for coverage only
			common.MetachainShardId: []byte("mb shard meta"),
		}
		expectedMBs := map[uint32][]byte{
			common.MetachainShardId: []byte("mb shard meta"),
		}
		providedTxs := map[string][][]byte{
			"topic_0":      {[]byte("txs topic 1")}, // for coverage only
			"topic_0_META": {[]byte("txs topic meta")},
		}
		expectedTxs := map[string][][]byte{
			"topic_0_META": {[]byte("txs topic meta")},
		}
		mes, err := NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{
			BroadcastMiniBlocksCalled: func(mbs map[uint32][]byte, bytes []byte) error {
				require.Equal(t, expectedMBs, mbs)
				return nil
			},
			BroadcastTransactionsCalled: func(txs map[string][][]byte, bytes []byte) error {
				require.Equal(t, expectedTxs, txs)
				return nil
			},
		}, &mock.ShardCoordinatorMock{
			ShardID: 0,
		})
		require.NoError(t, err)

		err = mes.BroadcastBlockDataLeader(nil, providedMBs, providedTxs, []byte("pk"))
		require.NoError(t, err)
	})
	t.Run("shard, empty miniblocks should early exit", func(t *testing.T) {
		t.Parallel()

		mes, err := NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{
			BroadcastMiniBlocksCalled: func(mbs map[uint32][]byte, bytes []byte) error {
				require.Fail(t, "should have not been called")
				return nil
			},
			BroadcastTransactionsCalled: func(txs map[string][][]byte, bytes []byte) error {
				require.Fail(t, "should have not been called")
				return nil
			},
		}, &mock.ShardCoordinatorMock{
			ShardID: 0,
		})
		require.NoError(t, err)

		err = mes.BroadcastBlockDataLeader(nil, nil, nil, []byte("pk"))
		require.NoError(t, err)
	})
}

func TestInstantBroadcastMessenger_PrepareBroadcastBlockDataWithEquivalentProofsShouldBroadcastAllDataImmediately(t *testing.T) {
	t.Parallel()

	providedMBs := map[uint32][]byte{
		0:                       []byte("mb shard 0"),
		common.MetachainShardId: []byte("mb shard meta"),
	}
	providedTxs := map[string][][]byte{
		"topic_0":      {[]byte("txs topic 0")},
		"topic_0_META": {[]byte("txs topic meta")},
	}
	providedPk := []byte("pk")

	miniBlocksBroadcast := false
	transactionsBroadcast := false
	mes, err := NewInstantBroadcastMessenger(&consensus.BroadcastMessengerMock{
		BroadcastMiniBlocksCalled: func(mbs map[uint32][]byte, pkBytes []byte) error {
			require.Equal(t, providedMBs, mbs)
			require.Equal(t, providedPk, pkBytes)
			miniBlocksBroadcast = true
			return nil
		},
		BroadcastTransactionsCalled: func(txs map[string][][]byte, pkBytes []byte) error {
			require.Equal(t, providedTxs, txs)
			require.Equal(t, providedPk, pkBytes)
			transactionsBroadcast = true
			return nil
		},
	}, &mock.ShardCoordinatorMock{ShardID: 0})
	require.NoError(t, err)

	mes.PrepareBroadcastBlockDataWithEquivalentProofs(nil, providedMBs, providedTxs, providedPk)

	require.True(t, miniBlocksBroadcast)
	require.True(t, transactionsBroadcast)
}

func TestInstantBroadcastMessenger_V2BodyIsDeliveredImmediatelyBeforeHeader(t *testing.T) {
	t.Parallel()

	callOrder := make([]string, 0)
	broadcast := &consensus.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensusMessage.Message) error {
			assert.Equal(t, []byte("body"), message.Body)
			callOrder = append(callOrder, "body")
			return nil
		},
		BroadcastHeaderCalled: func(_ data.HeaderHandler, _ []byte) error {
			callOrder = append(callOrder, "header")
			return nil
		},
	}
	messenger, err := NewInstantBroadcastMessenger(broadcast, &mock.ShardCoordinatorMock{})
	require.NoError(t, err)

	body := &consensusMessage.Message{
		Body:       []byte("body"),
		MsgType:    int64(bls.MtBlockBody),
		RoundIndex: 7,
	}
	require.NoError(t, messenger.BroadcastConsensusMessage(body))
	assert.Empty(t, callOrder, "the v2 body must be held until its header is ready")

	require.NoError(t, messenger.BroadcastHeader(&block.MetaBlock{Round: 7}, []byte("pk")))
	assert.Equal(t, []string{"body", "header"}, callOrder)
}

func TestInstantBroadcastMessenger_EpochSwitchPrecedesV2Proposal(t *testing.T) {
	t.Parallel()

	callOrder := make([]string, 0)
	broadcast := &consensus.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(_ *consensusMessage.Message) error {
			callOrder = append(callOrder, "body")
			return nil
		},
		BroadcastHeaderCalled: func(_ data.HeaderHandler, _ []byte) error {
			callOrder = append(callOrder, "header")
			return nil
		},
	}
	messenger, err := NewInstantBroadcastMessenger(broadcast, &mock.ShardCoordinatorMock{})
	require.NoError(t, err)
	messenger.setBeforeBroadcastHeader(func(_ data.HeaderHandler) {
		callOrder = append(callOrder, "switch")
	})
	messenger.setProposalDataHandler(func(_ data.HeaderHandler, bodyBytes []byte, _ []byte) error {
		assert.Equal(t, []byte("body"), bodyBytes)
		callOrder = append(callOrder, "dependencies")
		return nil
	})

	body := &consensusMessage.Message{
		Body:       []byte("body"),
		MsgType:    int64(bls.MtBlockBody),
		RoundIndex: 11,
	}
	require.NoError(t, messenger.BroadcastConsensusMessage(body))

	header := &block.MetaBlock{
		Round: 11,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}},
		},
	}
	require.NoError(t, messenger.BroadcastHeader(header, []byte("pk")))
	assert.Equal(t, []string{"switch", "dependencies", "body", "header"}, callOrder)
}
