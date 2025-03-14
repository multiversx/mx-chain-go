package spos

import (
	"crypto/rand"
	"fmt"
	"sync"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-communication-go/p2p/data"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus"
	consensusMock "github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func createMockArgs() ArgInvalidSignersCache {
	return ArgInvalidSignersCache{
		Hasher:         &testscommon.HasherStub{},
		SigningHandler: &consensusMock.MessageSigningHandlerStub{},
		Marshaller:     &marshallerMock.MarshalizerStub{},
	}
}

func TestInvalidSignersCache_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var cache *invalidSignersCache
	require.True(t, cache.IsInterfaceNil())

	cache, _ = NewInvalidSignersCache(createMockArgs())
	require.False(t, cache.IsInterfaceNil())
}

func TestNewInvalidSignersCache(t *testing.T) {
	t.Parallel()

	t.Run("nil Hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Hasher = nil

		cache, err := NewInvalidSignersCache(args)
		require.Equal(t, ErrNilHasher, err)
		require.Nil(t, cache)
	})
	t.Run("nil SigningHandler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.SigningHandler = nil

		cache, err := NewInvalidSignersCache(args)
		require.Equal(t, ErrNilSigningHandler, err)
		require.Nil(t, cache)
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.Marshaller = nil

		cache, err := NewInvalidSignersCache(args)
		require.Equal(t, ErrNilMarshalizer, err)
		require.Nil(t, cache)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cache, err := NewInvalidSignersCache(createMockArgs())
		require.NoError(t, err)
		require.NotNil(t, cache)
	})
}

func TestInvalidSignersCache(t *testing.T) {
	t.Parallel()

	t.Run("all ops should work", func(t *testing.T) {
		t.Parallel()

		headerHash1 := []byte("headerHash11")
		invalidSigners1 := []byte("invalidSigners1")
		pubKeys1 := []string{"pk0", "pk1"}
		invalidSigners2 := []byte("invalidSigners2")

		args := createMockArgs()
		args.Hasher = &testscommon.HasherStub{
			ComputeCalled: func(s string) []byte {
				return []byte(s)
			},
		}
		args.SigningHandler = &consensusMock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				if string(messagesBytes) == string(invalidSigners1) {
					m1, _ := libp2p.NewMessage(createDummyP2PMessage(), &testscommon.ProtoMarshalizerMock{}, "")
					m2, _ := libp2p.NewMessage(createDummyP2PMessage(), &testscommon.ProtoMarshalizerMock{}, "")
					return []p2p.MessageP2P{m1, m2}, nil
				}

				m1, _ := libp2p.NewMessage(createDummyP2PMessage(), &testscommon.ProtoMarshalizerMock{}, "")
				return []p2p.MessageP2P{m1}, nil
			},
		}
		cnt := 0
		args.Marshaller = &marshallerMock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				message := obj.(*consensus.Message)
				message.PubKey = []byte(fmt.Sprintf("pk%d", cnt))
				cnt++

				return nil
			},
		}
		cache, _ := NewInvalidSignersCache(args)
		require.NotNil(t, cache)

		cache.AddInvalidSigners(nil, nil, nil) // early return, for coverage only

		require.False(t, cache.CheckKnownInvalidSigners(headerHash1, invalidSigners1))

		cache.AddInvalidSigners(headerHash1, invalidSigners1, pubKeys1)
		require.True(t, cache.CheckKnownInvalidSigners(headerHash1, invalidSigners1)) // should find in signers by hashes map

		require.True(t, cache.CheckKnownInvalidSigners(headerHash1, invalidSigners2)) // should have different hash but the known signers

		cache.Reset()
		require.False(t, cache.CheckKnownInvalidSigners(headerHash1, invalidSigners1))
	})
	t.Run("concurrent ops should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		args := createMockArgs()
		cache, _ := NewInvalidSignersCache(args)
		require.NotNil(t, cache)

		numCalls := 1000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 3 {
				case 0:
					cache.AddInvalidSigners([]byte("hash"), []byte("invalidSigners"), []string{"pk0", "pk1"})
				case 1:
					cache.CheckKnownInvalidSigners([]byte("hash"), []byte("invalidSigners"))
				case 2:
					cache.Reset()
				default:
					require.Fail(t, "should not happen")
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}

func createDummyP2PMessage() *pubsub.Message {
	marshaller := &testscommon.ProtoMarshalizerMock{}
	topicMessage := &data.TopicMessage{
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
		Version:   1,
	}
	buff, _ := marshaller.Marshal(topicMessage)
	topic := "topic"
	mes := &pb.Message{
		From:  getRandomID().Bytes(),
		Data:  buff,
		Topic: &topic,
	}

	return &pubsub.Message{Message: mes}
}

func getRandomID() core.PeerID {
	prvKey, _, _ := crypto.GenerateSecp256k1Key(rand.Reader)
	id, _ := peer.IDFromPublicKey(prvKey.GetPublic())

	return core.PeerID(id)
}
