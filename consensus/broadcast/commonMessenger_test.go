package broadcast_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	broadcastMethodPrefix                      = "broadcast"
	broadcastUsingPrivateKeyCalledMethodPrefix = "broadcastUsingPrivateKeyCalled"
)

func newTestBlockBody() *block.Body {
	return &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{[]byte("txh1"), []byte("txh2")},
				ReceiverShardID: 0,
				SenderShardID:   0,
				Type:            0,
			},
		},
	}
}

func TestCommonMessenger_BroadcastConsensusMessageShouldErrWhenSignMessageFail(t *testing.T) {
	err := errors.New("sign message error")
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &p2pmocks.MessengerStub{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return nil, err
		},
	}
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHandlerStub{},
	)

	msg := &consensus.Message{}
	err2 := cm.BroadcastConsensusMessage(msg)
	assert.Equal(t, err, err2)
}

func TestCommonMessenger_BroadcastConsensusMessageShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHandlerStub{},
	)

	msg := &consensus.Message{}
	err := cm.BroadcastConsensusMessage(msg)
	assert.Nil(t, err)
}

func TestSubroundEndRound_ExtractMiniBlocksAndTransactionsShouldWork(t *testing.T) {
	t.Parallel()

	miniBlocks := make(map[uint32][]byte)
	transactions := make(map[string][][]byte)

	miniBlocks[1] = []byte("mbs_shard_1")
	miniBlocks[core.MetachainShardId] = []byte("mbs_shard_meta")
	miniBlocks[2] = []byte("mbs_shard_2")

	transactions["transactions_0_1"] = [][]byte{
		[]byte("tx1_shard_1"),
		[]byte("tx2_shard_1"),
		[]byte("tx3_shard_1"),
	}

	transactions["transactions_0_META"] = [][]byte{
		[]byte("tx1_shard_meta"),
		[]byte("tx2_shard_meta"),
		[]byte("tx3_shard_meta"),
	}

	transactions["transactions_0_2"] = [][]byte{
		[]byte("tx1_shard_2"),
		[]byte("tx2_shard_2"),
		[]byte("tx3_shard_2"),
	}

	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHandlerStub{},
	)

	metaMiniBlocks, metaTransactions := cm.ExtractMetaMiniBlocksAndTransactions(miniBlocks, transactions)

	require.Equal(t, 2, len(miniBlocks))
	require.Equal(t, 2, len(transactions))
	require.Equal(t, 1, len(metaMiniBlocks))
	require.Equal(t, 1, len(metaTransactions))

	assert.Nil(t, miniBlocks[core.MetachainShardId])
	assert.Nil(t, transactions["transactions_0_META"])
	assert.NotNil(t, metaMiniBlocks[core.MetachainShardId])
	assert.NotNil(t, metaTransactions["transactions_0_META"])
}

func TestCommonMessenger_BroadcastBlockData(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	countersBroadcast := make(map[string]int)
	mutCounters := &sync.Mutex{}

	messengerMock := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutCounters.Lock()
			countersBroadcast[broadcastMethodPrefix+topic]++
			mutCounters.Unlock()
		},
		BroadcastUsingPrivateKeyCalled: func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			mutCounters.Lock()
			countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+topic]++
			mutCounters.Unlock()
		},
	}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}

	cm, _ := broadcast.NewCommonMessenger(
		marshalizerMock,
		messengerMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHandlerStub{
			IsOriginalPublicKeyOfTheNodeCalled: func(pkBytes []byte) bool {
				return bytes.Equal(pkBytes, nodePkBytes)
			},
		},
	)

	miniBlocks := map[uint32][]byte{0: []byte("mbs data1"), 1: []byte("mbs data2")}
	transactions := map[string][][]byte{"topic1": {[]byte("txdata1"), []byte("txdata2")}, "topic2": {[]byte("txdata3")}}
	delay := time.Millisecond * 10

	t.Run("original public key of the node", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		cm.BroadcastBlockData(miniBlocks, transactions, nodePkBytes, delay)
		time.Sleep(delay * 2)

		mutCounters.Lock()
		defer mutCounters.Unlock()

		numBroadcast := countersBroadcast[broadcastMethodPrefix+"txBlockBodies_0"]
		numBroadcast += countersBroadcast[broadcastMethodPrefix+"txBlockBodies_0_1"]
		assert.Equal(t, len(miniBlocks), numBroadcast)

		numBroadcast = countersBroadcast[broadcastMethodPrefix+"topic1"]
		numBroadcast += countersBroadcast[broadcastMethodPrefix+"topic2"]
		assert.Equal(t, len(transactions), numBroadcast)
	})
	t.Run("managed key", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		cm.BroadcastBlockData(miniBlocks, transactions, []byte("managed key"), delay)
		time.Sleep(delay * 2)

		mutCounters.Lock()
		defer mutCounters.Unlock()

		numBroadcast := countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"txBlockBodies_0"]
		numBroadcast += countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"txBlockBodies_0_1"]
		assert.Equal(t, len(miniBlocks), numBroadcast)

		numBroadcast = countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"topic1"]
		numBroadcast += countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+"topic2"]
		assert.Equal(t, len(transactions), numBroadcast)
	})
}

func TestCommonMessenger_broadcast(t *testing.T) {
	t.Parallel()

	testTopic := "test"
	expectedErr := errors.New("expected error")
	marshallerMock := &mock.MarshalizerMock{}
	countersBroadcast := make(map[string]int)
	mutCounters := &sync.Mutex{}

	messengerMock := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutCounters.Lock()
			countersBroadcast[broadcastMethodPrefix+topic]++
			mutCounters.Unlock()
		},
		BroadcastUsingPrivateKeyCalled: func(topic string, buff []byte, pid core.PeerID, skBytes []byte) {
			mutCounters.Lock()
			countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+topic]++
			mutCounters.Unlock()
		},
	}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}

	t.Run("using the original public key bytes of the node", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHandlerStub{
				IsOriginalPublicKeyOfTheNodeCalled: func(pkBytes []byte) bool {
					return bytes.Equal(nodePkBytes, pkBytes)
				},
			},
		)

		cm.Broadcast(testTopic, []byte("data"), nodePkBytes)

		mutCounters.Lock()
		assert.Equal(t, 1, countersBroadcast[broadcastMethodPrefix+testTopic])
		assert.Equal(t, 0, countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+testTopic])
		mutCounters.Unlock()
	})
	t.Run("using a managed key", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHandlerStub{
				IsOriginalPublicKeyOfTheNodeCalled: func(pkBytes []byte) bool {
					return false
				},
			},
		)

		cm.Broadcast(testTopic, []byte("data"), []byte("managed key"))

		mutCounters.Lock()
		assert.Equal(t, 1, countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+testTopic])
		assert.Equal(t, 0, countersBroadcast[broadcastMethodPrefix+testTopic])
		mutCounters.Unlock()
	})
	t.Run("managed key and the keys handler fails", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHandlerStub{
				GetP2PIdentityCalled: func(pkBytes []byte) ([]byte, core.PeerID, error) {
					return nil, "", expectedErr
				},
				IsOriginalPublicKeyOfTheNodeCalled: func(pkBytes []byte) bool {
					return false
				},
			},
		)

		cm.Broadcast(testTopic, []byte("data"), []byte("managed key"))

		mutCounters.Lock()
		assert.Equal(t, 0, countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+testTopic])
		assert.Equal(t, 0, countersBroadcast[broadcastMethodPrefix+testTopic])
		mutCounters.Unlock()
	})
}
