package broadcastFactory

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func createDelayData(prefix string) ([]byte, *block.Header, map[uint32][]byte, map[string][][]byte) {
	miniblocks := make(map[uint32][]byte)
	receiverShardID := uint32(1)
	miniblocks[receiverShardID] = []byte(prefix + "miniblock data")

	transactions := make(map[string][][]byte)
	topic := "txBlockBodies_0_1"
	transactions[topic] = [][]byte{
		[]byte(prefix + "tx0"),
		[]byte(prefix + "tx1"),
	}
	headerHash := []byte(prefix + "header hash")
	header := &block.Header{
		Round:        0,
		PrevRandSeed: []byte(prefix),
	}

	return headerHash, header, miniblocks, transactions
}

func createInterceptorContainer() process.InterceptorsContainer {
	return &testscommon.InterceptorsContainerStub{
		GetCalled: func(topic string) (process.Interceptor, error) {
			return &testscommon.InterceptorStub{
				ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
					return nil
				},
			}, nil
		},
	}
}

func createDefaultShardChainArgs() broadcast.ShardChainMessengerArgs {
	marshalizerMock := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	messengerMock := &p2pmocks.MessengerStub{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	headersSubscriber := &testscommon.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()
	peerSigHandler := &mock.PeerSignatureHandler{
		Signer: singleSignerMock,
	}
	alarmScheduler := &mock.AlarmSchedulerStub{}

	return broadcast.ShardChainMessengerArgs{
		CommonMessengerArgs: broadcast.CommonMessengerArgs{
			Marshalizer:                marshalizerMock,
			Hasher:                     hasher,
			Messenger:                  messengerMock,
			ShardCoordinator:           shardCoordinatorMock,
			PeerSignatureHandler:       peerSigHandler,
			HeadersSubscriber:          headersSubscriber,
			InterceptorsContainer:      interceptorsContainer,
			MaxDelayCacheSize:          1,
			MaxValidatorDelayCacheSize: 1,
			AlarmScheduler:             alarmScheduler,
			KeysHandler:                &testscommon.KeysHandlerStub{},
		},
	}
}

func TestSovereignChainMessengerFactory_CreateShardChainMessenger(t *testing.T) {
	t.Parallel()

	f := NewSovereignShardChainMessengerFactory()
	require.False(t, f.IsInterfaceNil())

	args := createDefaultShardChainArgs()
	msg, err := f.CreateShardChainMessenger(args)
	require.Nil(t, err)
	require.NotNil(t, msg)
}
