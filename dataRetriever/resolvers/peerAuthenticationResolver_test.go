package resolvers_test

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
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

func createMockArgPeerAuthenticationResolver() resolvers.ArgPeerAuthenticationResolver {
	return resolvers.ArgPeerAuthenticationResolver{
		ArgBaseResolver:        createMockArgBaseResolver(),
		PeerAuthenticationPool: testscommon.NewCacherStub(),
		NodesCoordinator: &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return pksMap, nil
			},
		},
		MaxNumOfPeerAuthenticationInResponse: 5,
		DataPacker:                           &mock.DataPackerStub{},
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
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.NodesCoordinator = nil
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrNilNodesCoordinator, err)
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
	t.Run("invalid max num of peer authentication  should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.MaxNumOfPeerAuthenticationInResponse = 1
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Equal(t, dataRetriever.ErrInvalidNumOfPeerAuthentication, err)
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
		assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
		assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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

	// =============== ChunkType -> resolveChunkRequest ===============

	t.Run("resolveChunkRequest: GetAllEligibleValidatorsPublicKeys returns error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return nil, expectedErr
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.ChunkType, []byte("data")), fromConnectedPeer)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("resolveChunkRequest: GetAllEligibleValidatorsPublicKeys returns empty", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
			GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
				return make(map[uint32][][]byte), nil
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.ChunkType, []byte("data")), fromConnectedPeer)
		require.Nil(t, err)
	})
	t.Run("resolveChunkRequest: chunk index is out of bounds", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		epoch := uint32(0)
		chunkIndex := uint32(10) // out of range
		err = res.ProcessReceivedMessage(createRequestMsgWithChunkIndex(dataRetriever.ChunkType, []byte(""), epoch, chunkIndex), fromConnectedPeer)
		require.Equal(t, dataRetriever.InvalidChunkIndex, err)
	})
	t.Run("resolveChunkRequest: all data not found in cache should error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return nil, false
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		wasSent := false
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				wasSent = true
				return nil
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		epoch := uint32(0)
		chunkIndex := uint32(0)
		err = res.ProcessReceivedMessage(createRequestMsgWithChunkIndex(dataRetriever.ChunkType, []byte(""), epoch, chunkIndex), fromConnectedPeer)
		assert.True(t, errors.Is(err, dataRetriever.ErrPeerAuthNotFound))
		expectedSubstrErr := fmt.Sprintf("%s %d", "from chunk", chunkIndex)
		assert.True(t, strings.Contains(fmt.Sprintf("%s", err), expectedSubstrErr))
		assert.False(t, wasSent)
	})
	t.Run("resolveChunkRequest: some data not found in cache should work", func(t *testing.T) {
		t.Parallel()

		expectedNumOfMissing := 3
		cache := testscommon.NewCacherStub()
		missingCount := 0
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			if missingCount < expectedNumOfMissing {
				missingCount++
				return nil, false
			}
			return key, true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		messagesSent := 0
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				b := &batch.Batch{}
				err := arg.Marshaller.Unmarshal(b, buff)
				assert.Nil(t, err)
				expectedDataLen := arg.MaxNumOfPeerAuthenticationInResponse - expectedNumOfMissing
				assert.Equal(t, expectedDataLen, len(b.Data))
				messagesSent++
				return nil
			},
		}
		arg.DataPacker, _ = partitioning.NewSizeDataPacker(arg.Marshaller)
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		epoch := uint32(0)
		chunkIndex := uint32(0)
		err = res.ProcessReceivedMessage(createRequestMsgWithChunkIndex(dataRetriever.ChunkType, []byte(""), epoch, chunkIndex), fromConnectedPeer)
		assert.Nil(t, err)
		assert.Equal(t, 1, messagesSent)
	})
	t.Run("resolveChunkRequest: Send returns error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return key, true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				return expectedErr
			},
		}
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.ChunkType, []byte("")), fromConnectedPeer)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("resolveChunkRequest: should work", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return key, true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		messagesSent := 0
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
				messagesSent++
				return nil
			},
		}
		arg.DataPacker, _ = partitioning.NewSizeDataPacker(arg.Marshaller)
		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		epoch := uint32(0)
		chunkIndex := uint32(1)
		err = res.ProcessReceivedMessage(createRequestMsgWithChunkIndex(dataRetriever.ChunkType, []byte(""), epoch, chunkIndex), fromConnectedPeer)
		assert.Nil(t, err)
		assert.Equal(t, 1, messagesSent)
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
			SendCalled: func(buff []byte, peer core.PeerID) error {
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
	t.Run("resolveMultipleHashesRequest: some data missing from cache should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgPeerAuthenticationResolver()

		pk1 := "pk01"
		pk2 := "pk02"
		providedKeys := make(map[string][]byte)
		providedKeys[pk1] = []byte("")
		providedKeys[pk2] = []byte("")
		pks := make([][]byte, 0)
		pks = append(pks, []byte(pk1))
		pks = append(pks, []byte(pk2))

		hashes := make([][]byte, 0)
		hashes = append(hashes, []byte("pk01")) // exists in cache
		hashes = append(hashes, []byte("pk1"))  // no entries
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
			SendCalled: func(buff []byte, peer core.PeerID) error {
				b := &batch.Batch{}
				err = arg.Marshaller.Unmarshal(b, buff)
				assert.Nil(t, err)
				assert.Equal(t, 1, len(b.Data)) // 1 entry for provided hashes
				wasSent = true
				return nil
			},
		}
		arg.DataPacker, _ = partitioning.NewSizeDataPacker(arg.Marshaller)

		res, err := resolvers.NewPeerAuthenticationResolver(arg)
		assert.Nil(t, err)
		assert.False(t, res.IsInterfaceNil())

		err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer)
		assert.Nil(t, err)
		assert.True(t, wasSent)
	})
	t.Run("resolveMultipleHashesRequest: Send returns error", func(t *testing.T) {
		t.Parallel()

		cache := testscommon.NewCacherStub()
		cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
			return key, true
		}

		arg := createMockArgPeerAuthenticationResolver()
		arg.PeerAuthenticationPool = cache
		arg.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID) error {
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
					return pk, true
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
			SendCalled: func(buff []byte, peer core.PeerID) error {
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

func TestPeerAuthenticationResolver_RequestShouldError(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerAuthenticationResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			return expectedErr
		},
	}
	res, err := resolvers.NewPeerAuthenticationResolver(arg)
	assert.Nil(t, err)
	assert.False(t, res.IsInterfaceNil())

	t.Run("RequestDataFromHash", func(t *testing.T) {
		err = res.RequestDataFromHash([]byte(""), 0)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RequestDataFromChunk", func(t *testing.T) {
		err = res.RequestDataFromChunk(0, 0)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RequestDataFromChunk - error on SendOnRequestTopic", func(t *testing.T) {
		hashes := make([][]byte, 0)
		hashes = append(hashes, []byte("pk"))
		err = res.RequestDataFromHashArray(hashes, 0)
		assert.Equal(t, expectedErr, err)
	})

}

func TestPeerAuthenticationResolver_RequestShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerAuthenticationResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			return nil
		},
	}
	res, err := resolvers.NewPeerAuthenticationResolver(arg)
	assert.Nil(t, err)
	assert.False(t, res.IsInterfaceNil())

	t.Run("RequestDataFromHash", func(t *testing.T) {
		err = res.RequestDataFromHash([]byte(""), 0)
		assert.Nil(t, err)
	})
	t.Run("RequestDataFromChunk", func(t *testing.T) {
		err = res.RequestDataFromChunk(0, 0)
		assert.Nil(t, err)
	})
}
