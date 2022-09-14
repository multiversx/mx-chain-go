package broadcast_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
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
		privateKeyMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHolderStub{},
	)

	msg := &consensus.Message{}
	err2 := cm.BroadcastConsensusMessage(msg)
	assert.Equal(t, err, err2)
}

func TestCommonMessenger_BroadcastConsensusMessageShouldWork(t *testing.T) {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
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
		privateKeyMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHolderStub{},
	)

	msg := &consensus.Message{}
	err := cm.BroadcastConsensusMessage(msg)
	assert.Nil(t, err)
}

func TestCommonMessenger_SignMessageShouldErrWhenSignFail(t *testing.T) {
	err := errors.New("sign message error")
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
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
		privateKeyMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHolderStub{},
	)

	msg := &consensus.Message{}
	_, err2 := cm.SignMessage(msg)
	assert.Equal(t, err, err2)
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
	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	privateKeyMock := &mock.PrivateKeyMock{}
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
		privateKeyMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHolderStub{},
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

	messengerMock := &mock.MessengerStub{
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
	privateKeyMock := &mock.PrivateKeyMock{}
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
		privateKeyMock,
		shardCoordinatorMock,
		peerSigHandler,
		&testscommon.KeysHolderStub{},
	)

	miniBlocks := map[uint32][]byte{0: []byte("mbs data1"), 1: []byte("mbs data2")}
	transactions := map[string][][]byte{"topic1": {[]byte("txdata1"), []byte("txdata2")}, "topic2": {[]byte("txdata3")}}
	delay := time.Millisecond * 10

	t.Run("original public key of the node", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		pkBytes, _ := privateKeyMock.GeneratePublic().ToByteArray()
		cm.BroadcastBlockData(miniBlocks, transactions, pkBytes, delay)
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

	messengerMock := &mock.MessengerStub{
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
	privateKeyMock := &mock.PrivateKeyMock{}
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
			privateKeyMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHolderStub{},
		)

		pkBytes, _ := privateKeyMock.GeneratePublic().ToByteArray()
		cm.Broadcast(testTopic, []byte("data"), pkBytes)

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
			privateKeyMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHolderStub{},
		)

		cm.Broadcast(testTopic, []byte("data"), []byte("managed key"))

		mutCounters.Lock()
		assert.Equal(t, 1, countersBroadcast[broadcastUsingPrivateKeyCalledMethodPrefix+testTopic])
		assert.Equal(t, 0, countersBroadcast[broadcastMethodPrefix+testTopic])
		mutCounters.Unlock()
	})
	t.Run("keys holder fails", func(t *testing.T) {
		mutCounters.Lock()
		countersBroadcast = make(map[string]int)
		mutCounters.Unlock()

		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			privateKeyMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHolderStub{
				GetP2PIdentityCalled: func(pkBytes []byte) ([]byte, core.PeerID, error) {
					return nil, "", expectedErr
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

func TestCommonMessenger_getPrivateKey(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	marshallerMock := &mock.MarshalizerMock{}

	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte(""), nil
		},
	}
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}
	managedKeyBytes := []byte("managed key")
	managedPrivateKey := &mock.PrivateKeyMock{}

	t.Run("using the original public key bytes of the node", func(t *testing.T) {
		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			privateKeyMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHolderStub{},
		)

		pkBytes, _ := privateKeyMock.GeneratePublic().ToByteArray()
		retrievedPrivateKey := cm.GetPrivateKey(&consensus.Message{
			PubKey: pkBytes,
		})

		assert.True(t, privateKeyMock == retrievedPrivateKey) // pointer testing
	})
	t.Run("using a managed key", func(t *testing.T) {
		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			privateKeyMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHolderStub{
				GetPrivateKeyCalled: func(pkBytes []byte) (crypto.PrivateKey, error) {
					if bytes.Equal(pkBytes, managedKeyBytes) {
						return managedPrivateKey, nil
					}

					return privateKeyMock, nil
				},
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return bytes.Equal(pkBytes, managedKeyBytes)
				},
			},
		)

		retrievedPrivateKey := cm.GetPrivateKey(&consensus.Message{
			PubKey: managedKeyBytes,
		})

		assert.True(t, managedPrivateKey == retrievedPrivateKey) // pointer testing
	})
	t.Run("keys holder fails", func(t *testing.T) {
		cm, _ := broadcast.NewCommonMessenger(
			marshallerMock,
			messengerMock,
			privateKeyMock,
			shardCoordinatorMock,
			peerSigHandler,
			&testscommon.KeysHolderStub{
				GetPrivateKeyCalled: func(pkBytes []byte) (crypto.PrivateKey, error) {
					if bytes.Equal(pkBytes, managedKeyBytes) {
						return nil, expectedErr
					}

					return privateKeyMock, nil
				},
			},
		)

		retrievedPrivateKey := cm.GetPrivateKey(&consensus.Message{
			PubKey: managedKeyBytes,
		})

		assert.True(t, privateKeyMock == retrievedPrivateKey) // pointer testing with the original private key node
	})
}
