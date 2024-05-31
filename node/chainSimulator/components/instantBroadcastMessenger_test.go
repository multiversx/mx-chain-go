package components

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
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

		mes, err := NewInstantBroadcastMessenger(&mock.BroadcastMessengerMock{}, nil)
		require.Equal(t, errorsMx.ErrNilShardCoordinator, err)
		require.Nil(t, mes)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mes, err := NewInstantBroadcastMessenger(&mock.BroadcastMessengerMock{}, &mock.ShardCoordinatorMock{})
		require.NoError(t, err)
		require.NotNil(t, mes)
	})
}

func TestInstantBroadcastMessenger_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var mes *instantBroadcastMessenger
	require.True(t, mes.IsInterfaceNil())

	mes, _ = NewInstantBroadcastMessenger(&mock.BroadcastMessengerMock{}, &mock.ShardCoordinatorMock{})
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
		mes, err := NewInstantBroadcastMessenger(&mock.BroadcastMessengerMock{
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
		mes, err := NewInstantBroadcastMessenger(&mock.BroadcastMessengerMock{
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

		mes, err := NewInstantBroadcastMessenger(&mock.BroadcastMessengerMock{
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
