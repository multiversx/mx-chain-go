package broadcast

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type delayedBlockBroadcasterMock struct {
	mock.Mock
}

// SetLeaderData sets the data for consensus leader delayed broadcast
func (mock *delayedBlockBroadcasterMock) SetLeaderData(broadcastData *delayedBroadcastData) error {
	return nil
}

// SetHeaderForValidator sets the header to be broadcast by validator if leader fails to broadcast it
func (mock *delayedBlockBroadcasterMock) SetHeaderForValidator(vData *validatorHeaderBroadcastData) error {
	return nil
}

// SetValidatorData sets the data for consensus validator delayed broadcast
func (mock *delayedBlockBroadcasterMock) SetValidatorData(broadcastData *delayedBroadcastData) error {
	return nil
}

// SetBroadcastHandlers sets the broadcast handlers for miniBlocks and transactions
func (mock *delayedBlockBroadcasterMock) SetBroadcastHandlers(
	mbBroadcast func(mbData map[uint32][]byte, pkBytes []byte) error,
	txBroadcast func(txData map[string][][]byte, pkBytes []byte) error,
	headerBroadcast func(header data.HeaderHandler, pkBytes []byte) error,
) error {
	args := mock.Called(mbBroadcast, txBroadcast, headerBroadcast)
	return args.Error(0)
}

// Close closes all the started infinite looping goroutines and subcomponents
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

func TestNewSovereignShardChainMessenger(t *testing.T) {
	t.Parallel()

	mockBroadcaster := new(delayedBlockBroadcasterMock)
	mockBroadcaster.On("SetBroadcastHandlers", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mbBroadcast := args.Get(0)
		txBroadcast := args.Get(1)
		headerBroadcast := args.Get(2)

		require.Contains(t, getFunctionName(mbBroadcast), "(*commonMessenger).BroadcastMiniBlocks")
		require.Contains(t, getFunctionName(txBroadcast), "(*commonMessenger).BroadcastTransactions")
		require.Contains(t, getFunctionName(headerBroadcast), "(*sovereignChainMessenger).BroadcastHeader")
	}).Return(nil)

	args := createSovShardMsgArgs()
	args.DelayedBroadcaster = mockBroadcaster
	sovMsg, err := NewSovereignShardChainMessenger(args)
	require.Nil(t, err)
	require.False(t, sovMsg.IsInterfaceNil())

	mockBroadcaster.AssertCalled(t, "SetBroadcastHandlers", mock.Anything, mock.Anything, mock.Anything)
}
