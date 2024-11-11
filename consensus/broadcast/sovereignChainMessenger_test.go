package broadcast

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

type delayedBlockBroadcasterMock struct {
	SetBroadcastHandlersCalled func(
		mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error,
		txBroadcast func(txData map[string][][]byte, pkBytes []byte) error,
		headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error,
	) error
}

// SetLeaderData -
func (mock *delayedBlockBroadcasterMock) SetLeaderData(_ *delayedBroadcastData) error {
	return nil
}

// SetHeaderForValidator -
func (mock *delayedBlockBroadcasterMock) SetHeaderForValidator(_ *validatorHeaderBroadcastData) error {
	return nil
}

// SetValidatorData -
func (mock *delayedBlockBroadcasterMock) SetValidatorData(_ *delayedBroadcastData) error {
	return nil
}

// SetBroadcastHandlers -
func (mock *delayedBlockBroadcasterMock) SetBroadcastHandlers(
	mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error,
	txBroadcast func(txData map[string][][]byte, pkBytes []byte) error,
	headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error,
) error {
	if mock.SetBroadcastHandlersCalled != nil {
		return mock.SetBroadcastHandlersCalled(mbBroadcast, txBroadcast, headerBroadcast)
	}

	return nil
}

// Close -
func (mock *delayedBlockBroadcasterMock) Close() {
}

func createSovShardMsgArgs() ArgsSovereignShardChainMessenger {
	return ArgsSovereignShardChainMessenger{
		Marshaller:       &consensusMock.MarshalizerMock{},
		Hasher:           &hashingMocks.HasherMock{},
		Messenger:        &p2pmocks.MessengerStub{},
		ShardCoordinator: &consensusMock.ShardCoordinatorMock{},
		PeerSignatureHandler: &consensusMock.PeerSignatureHandler{
			Signer: &consensusMock.SingleSignerMock{},
		},
		KeysHandler:        &testscommon.KeysHandlerStub{},
		DelayedBroadcaster: &delayedBlockBroadcasterMock{},
	}

}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func TestNewSovereignShardChainMessenger_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.Marshaller = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, spos.ErrNilMarshalizer, err)
		require.Nil(t, sovMsg)
	})
	t.Run("nil hasher", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.Hasher = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, spos.ErrNilHasher, err)
		require.Nil(t, sovMsg)
	})
	t.Run("nil messenger", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.Messenger = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, spos.ErrNilMessenger, err)
		require.Nil(t, sovMsg)
	})
	t.Run("nil shard coordinator", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.ShardCoordinator = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, spos.ErrNilShardCoordinator, err)
		require.Nil(t, sovMsg)
	})
	t.Run("nil peer signature handler", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.PeerSignatureHandler = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, spos.ErrNilPeerSignatureHandler, err)
		require.Nil(t, sovMsg)
	})
	t.Run("nil keys handler", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.KeysHandler = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, ErrNilKeysHandler, err)
		require.Nil(t, sovMsg)
	})
	t.Run("nil delayed broadcaster", func(t *testing.T) {
		args := createSovShardMsgArgs()
		args.DelayedBroadcaster = nil
		sovMsg, err := NewSovereignShardChainMessenger(args)
		require.Equal(t, errNilDelayedShardBroadCaster, err)
		require.Nil(t, sovMsg)
	})
}

func TestNewSovereignShardChainMessenger(t *testing.T) {
	t.Parallel()

	wasHandlerSet := false
	mockBroadcaster := &delayedBlockBroadcasterMock{
		SetBroadcastHandlersCalled: func(mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error, txBroadcast func(txData map[string][][]byte, pkBytes []byte) error, headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error) error {
			require.Contains(t, getFunctionName(mbBroadcast), "(*commonMessenger).BroadcastMiniBlocks")
			require.Contains(t, getFunctionName(txBroadcast), "(*commonMessenger).BroadcastTransactions")
			require.Contains(t, getFunctionName(headerBroadcast), "(*sovereignChainMessenger).BroadcastHeader")

			wasHandlerSet = true
			return nil
		},
	}

	args := createSovShardMsgArgs()
	args.DelayedBroadcaster = mockBroadcaster
	sovMsg, err := NewSovereignShardChainMessenger(args)
	require.Nil(t, err)
	require.False(t, sovMsg.IsInterfaceNil())
	require.True(t, wasHandlerSet)
}

func TestSovereignChainMessenger_BroadcastBlock(t *testing.T) {
	t.Parallel()

	args := createSovShardMsgArgs()
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{[]byte("txHash")},
				Type:     3,
			},
		},
	}
	bodyBytes, err := args.Marshaller.Marshal(body)
	require.Nil(t, err)

	hdr := &block.SovereignChainHeader{
		DevFeesInEpoch: big.NewInt(100),
	}
	hdrBytes, err := args.Marshaller.Marshal(hdr)
	require.Nil(t, err)

	broadCastCt := 0
	messenger := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			switch broadCastCt {
			case 0:
				require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), topic)
				require.Equal(t, hdrBytes, buff)
			case 1:
				require.Equal(t, fmt.Sprintf("%s_%d", factory.MiniBlocksTopic, core.SovereignChainShardId), topic)
				require.Equal(t, bodyBytes, buff)
			default:
				require.Fail(t, "should not have call this func again")
			}

			broadCastCt++
		},
	}

	args.Messenger = messenger
	sovMsg, _ := NewSovereignShardChainMessenger(args)
	err = sovMsg.BroadcastBlock(body, hdr)
	require.Nil(t, err)
	require.Equal(t, 2, broadCastCt)
}

func TestSovereignChainMessenger_BroadcastHeader(t *testing.T) {
	t.Parallel()

	args := createSovShardMsgArgs()
	hdr := &block.SovereignChainHeader{
		DevFeesInEpoch: big.NewInt(100),
	}
	hdrBytes, err := args.Marshaller.Marshal(hdr)
	require.Nil(t, err)

	broadCastCt := 0
	messenger := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), topic)
			require.Equal(t, hdrBytes, buff)
			broadCastCt++
		},
	}

	args.Messenger = messenger
	sovMsg, _ := NewSovereignShardChainMessenger(args)
	err = sovMsg.BroadcastHeader(hdr, []byte("key"))
	require.Nil(t, err)
	require.Equal(t, 1, broadCastCt)
}

func TestSovereignChainMessenger_shouldSkipShard(t *testing.T) {
	t.Parallel()

	args := createSovShardMsgArgs()
	sovMsg, _ := NewSovereignShardChainMessenger(args)
	require.False(t, sovMsg.shouldSkipShard(core.SovereignChainShardId))
	require.True(t, sovMsg.shouldSkipShard(core.SovereignChainShardId+1))
	require.True(t, sovMsg.shouldSkipShard(core.MainChainShardId))
}

func TestSovereignChainMessenger_shouldSkipTopic(t *testing.T) {
	t.Parallel()

	args := createSovShardMsgArgs()
	sovMsg, _ := NewSovereignShardChainMessenger(args)
	require.False(t, sovMsg.shouldSkipTopic("topic"))
	require.False(t, sovMsg.shouldSkipTopic(fmt.Sprintf("%s_%d", "topic", core.SovereignChainShardId)))
	require.True(t, sovMsg.shouldSkipTopic(fmt.Sprintf("%s_%d", "topic", core.MainChainShardId)))
	require.True(t, sovMsg.shouldSkipTopic(fmt.Sprintf("%s_%d_%d", "topic", core.SovereignChainShardId, core.MainChainShardId)))
}
