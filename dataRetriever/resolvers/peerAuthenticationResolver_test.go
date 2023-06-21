package resolvers_test

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")
var pksMap = map[uint32][][]byte{
	0: {[]byte("pk00"), []byte("pk01"), []byte("pk02")},
	1: {[]byte("pk10"), []byte("pk11")},
	2: {[]byte("pk21"), []byte("pk21"), []byte("pk32"), []byte("pk33")},
}

func getKeysSlice() [][]byte {
	pks := make([][]byte, 0)
	for _, pk := range pksMap {
		pks = append(pks, pk...)
	}
	sort.Slice(pks, func(i, j int) bool {
		return bytes.Compare(pks[i], pks[j]) < 0
	})
	return pks
}

func createMockPeerAuthenticationObject() interface{} {
	arg := createMockArgPeerAuthenticationResolver()

	payload := &heartbeat.Payload{
		Timestamp: time.Now().Unix(),
	}
	payloadBuff, _ := arg.Marshaller.Marshal(payload)
	return &heartbeat.PeerAuthentication{
		Payload: payloadBuff,
	}
}

func createMockArgPeerAuthenticationResolver() resolvers.ArgPeerAuthenticationResolver {
	return resolvers.ArgPeerAuthenticationResolver{
		ArgBaseResolver:        createMockArgBaseResolver(),
		PeerAuthenticationPool: testscommon.NewCacherStub(),
		DataPacker:             &mock.DataPackerStub{},
		PayloadValidator:       &testscommon.PeerAuthenticationPayloadValidatorStub{},
	}
}

func createPublicKeys(prefix string, numOfPks int) [][]byte {
	var pkList [][]byte
	for i := 0; i < numOfPks; i++ {
		pk := []byte(fmt.Sprintf("%s%d", prefix, i))
		pkList = append(pkList, pk)
	}
	return pkList
}

func createMockRequestedBuff(numOfPks int) ([]byte, error) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	return marshaller.Marshal(&batch.Batch{Data: createPublicKeys("pk", numOfPks)})
}

func TestNewPeerAuthenticationResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil SenderResolver should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.SenderResolver = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
		assert.Nil(t, res)
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.Marshaller = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.Nil(t, res)
	})
	t.Run("nil AntifloodHandler should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.AntifloodHandler = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
		assert.Nil(t, res)
	})
	t.Run("nil Throttler should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.Throttler = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilThrottler, err)
		assert.Nil(t, res)
	})
	t.Run("nil PeerAuthenticationPool should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilPeerAuthenticationPool, err)
		assert.Nil(t, res)
	})
	t.Run("nil DataPacker should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.DataPacker = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
		assert.Nil(t, res)
	})
	t.Run("nil payload validator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.PayloadValidator = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilPayloadValidator, err)
		assert.Nil(t, res)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())
	})
}

func TestPeerAuthenticationResolver_ProcessReceivedMessage(t *testing.T) {
	t.Parallel()

	t.Run("nil message should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(nil, fromConnectedPeer)
		assert.Equal(t, dataRetriever.ErrNilMessage, err)
	})
	t.Run("canProcessMessage due to antiflood handler error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return expectedErr
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.ChunkType, nil), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
		assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
		assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
	})
	t.Run("parseReceivedMessage returns error due to marshaller error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.Marshaller = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.ChunkType, nil), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("invalid request type should error", func(t *testing.T) {
		t.Parallel()

		numOfPks := 3
		requestedBuff, err := createMockRequestedBuff(numOfPks)
		require.Nil(t, err)

		arg := createMockArgPeerAuthenticationResolver()
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedBuff), fromConnectedPeer)
		assert.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
	})

	// =============== HashArrayType -> resolveMultipleHashesRequest ===============

	t.Run("resolveMultipleHashesRequest: Unmarshal returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, []byte("invalid data")), fromConnectedPeer)
		assert.NotNil(t, err)
	})
	t.Run("resolveMultipleHashesRequest: all hashes missing from cache should error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		wasSent := false
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
				wasSent = true
				return nil
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		hashes := getKeysSlice()
		providedHashes, err := arg.Marshaller.Marshal(batch.Batch{Data: hashes})
		assert.Nil(t, err)
		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		expectedSubstrErr := fmt.Sprintf("%s %x", "from buff", providedHashes)
		assert.True(t, strings.Contains(fmt.Sprintf("%s", err), expectedSubstrErr))
		assert.False(t, wasSent)
	})
	t.Run("resolveMultipleHashesRequest: all hashes will return wrong objects should error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return "wrong object", true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		wasSent := false
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
				wasSent = true
				return nil
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		hashes := getKeysSlice()
		providedHashes, err := arg.Marshaller.Marshal(batch.Batch{Data: hashes})
		assert.Nil(t, err)
		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		expectedSubstrErr := fmt.Sprintf("%s %x", "from buff", providedHashes)
		assert.True(t, strings.Contains(fmt.Sprintf("%s", err), expectedSubstrErr))
		assert.False(t, wasSent)
	})
	t.Run("resolveMultipleHashesRequest: all hashes will return objects with invalid payload should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return createMockPeerAuthenticationObject(), true
		}

		numValidationCalls := 0
		arg.PayloadValidator = &testscommon.PeerAuthenticationPayloadValidatorStub{
			ValidateTimestampCalled: func(payloadTimestamp int64) error {
				numValidationCalls++
				return expectedErr
			},
		}
		arg.PeerAuthenticationPool = cache
		wasSent := false
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
				wasSent = true
				return nil
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		hashes := getKeysSlice()
		providedHashes, err := arg.Marshaller.Marshal(batch.Batch{Data: hashes})
		assert.Nil(t, err)
		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		expectedSubstrErr := fmt.Sprintf("%s %x", "from buff", providedHashes)
		assert.True(t, strings.Contains(fmt.Sprintf("%s", err), expectedSubstrErr))
		assert.False(t, wasSent)
		assert.Equal(t, len(hashes), numValidationCalls)
	})
	t.Run("resolveMultipleHashesRequest: some data missing from cache should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()

		pk1 := "pk01"
		pk2 := "pk02"
		pk3 := "pk03"
		providedKeys := make(map[string]interface{})
		providedKeys[pk1] = createMockPeerAuthenticationObject()
		providedKeys[pk2] = createMockPeerAuthenticationObject()
		providedKeys[pk3] = createMockPeerAuthenticationObject()
		pks := make([][]byte, 0)
		pks = append(pks, []byte(pk1))
		pks = append(pks, []byte(pk2))
		pks = append(pks, []byte(pk3))

		hashes := make([][]byte, 0)
		hashes = append(hashes, []byte("pk01")) // exists in cache
		hashes = append(hashes, []byte("pk1"))  // no entries
		hashes = append(hashes, []byte("pk03")) // unmarshal fails
		providedHashes, err := arg.Marshaller.Marshal(batch.Batch{Data: hashes})
		assert.Nil(t, err)

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			val, ok := providedKeys[string(key)]
			return val, ok
		}
		cache.KeysCalled = func() [][]byte {
			return pks
		}

		arg.PeerAuthenticationPool = cache
		wasSent := false
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
				b := &batch.Batch{}
				err = arg.Marshaller.Unmarshal(b, buff)
				assert.Nil(t, err)
				assert.Equal(t, 1, len(b.Data)) // 1 entry for provided hashes
				wasSent = true
				return nil
			},
		}
		arg.DataPacker, _ = partitioning.NewSizeDataPacker(arg.Marshaller)
		initialMarshaller := arg.Marshaller
		cnt := 0
		arg.Marshaller = &mock.MarshalizerStub{
			MarshalCalled: initialMarshaller.Marshal,
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				cnt++
				if cnt == 4 { // pk03
					return expectedErr
				}
				return initialMarshaller.Unmarshal(obj, buff)
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		assert.Nil(t, err)
		assert.True(t, wasSent)
	})
	t.Run("resolveMultipleHashesRequest: PackDataInChunks returns error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return createMockPeerAuthenticationObject(), true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		arg.DataPacker = &mock.DataPackerStub{
			PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
				return nil, expectedErr
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		hashes := getKeysSlice()
		providedHashes, err := arg.Marshaller.Marshal(batch.Batch{Data: hashes})
		assert.Nil(t, err)
		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("resolveMultipleHashesRequest: Send returns error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return createMockPeerAuthenticationObject(), true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
				return expectedErr
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		hashes := getKeysSlice()
		providedHashes, err := arg.Marshaller.Marshal(batch.Batch{Data: hashes})
		assert.Nil(t, err)
		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("resolveMultipleHashesRequest: send large data buff", func(t *testing.T) {
		t.Parallel()

		providedKeys := getKeysSlice()
		expectedLen := len(providedKeys)
		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			for _, pk := range providedKeys {
				if bytes.Equal(pk, key) {
					return createMockPeerAuthenticationObject(), true
				}
			}
			return nil, false
		}
		cache.KeysCalled = func() [][]byte {
			return getKeysSlice()
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		messagesSent := 0
		hashesReceived := 0
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
				b := &batch.Batch{}
				err := arg.Marshaller.Unmarshal(b, buff)
				assert.Nil(t, err)

				hashesReceived += len(b.Data)
				messagesSent++
				return nil
			},
		}
		// split data into 2 packs
		arg.DataPacker = &mock.DataPackerStub{
			PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
				middle := len(data) / 2
				b := &batch.Batch{
					Data: data[middle:],
				}
				buff1, err := arg.Marshaller.Marshal(b)
				assert.Nil(t, err)

				b = &batch.Batch{
					Data: data[:middle],
				}
				buff2, err := arg.Marshaller.Marshal(b)
				assert.Nil(t, err)
				return [][]byte{buff1, buff2}, nil
			},
		}

		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		epoch := uint32(0)
		chunkIndex := uint32(0)
		providedHashes, err := arg.Marshaller.Marshal(&batch.Batch{Data: providedKeys})
		assert.Nil(t, err)
		err = res.ProcessReceivedMessage(createRequestMsgWithChunkIndex(dataRetriever.HashArrayType, providedHashes, epoch, chunkIndex), fromConnectedPeer)
		assert.Nil(t, err)
		assert.Equal(t, 2, messagesSent)
		assert.Equal(t, expectedLen, hashesReceived)
	})
}
