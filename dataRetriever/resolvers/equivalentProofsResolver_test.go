package resolvers_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/p2p"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

var (
	providedHashKey  = []byte("hash-0")
	providedNonceKey = []byte("1-1")
)

func createMockArgEquivalentProofsResolver() resolvers.ArgEquivalentProofsResolver {
	return resolvers.ArgEquivalentProofsResolver{
		ArgBaseResolver:                  createMockArgBaseResolver(),
		DataPacker:                       &mock.DataPackerStub{},
		EquivalentProofsStorage:          &storageStubs.StorerStub{},
		EquivalentProofsNonceHashStorage: &storageStubs.StorerStub{},
		EquivalentProofsPool:             &dataRetrieverMocks.ProofsPoolMock{},
		IsFullHistoryNode:                false,
	}
}

func createMockRequestedProofsBuff() ([]byte, error) {
	marshaller := &marshal.GogoProtoMarshalizer{}

	return marshaller.Marshal(&batch.Batch{Data: [][]byte{[]byte("proof")}})
}

func TestNewEquivalentProofsResolver(t *testing.T) {
	t.Parallel()

	t.Run("nil SenderResolver should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.SenderResolver = nil
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Equal(t, dataRetriever.ErrNilResolverSender, err)
		require.Nil(t, res)
	})
	t.Run("nil DataPacker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.DataPacker = nil
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Equal(t, dataRetriever.ErrNilDataPacker, err)
		require.Nil(t, res)
	})
	t.Run("nil EquivalentProofsStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsStorage = nil
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.True(t, errors.Is(err, dataRetriever.ErrNilProofsStorage))
		require.Nil(t, res)
	})
	t.Run("nil EquivalentProofsNonceHashStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsNonceHashStorage = nil
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.True(t, errors.Is(err, dataRetriever.ErrNilProofsStorage))
		require.Nil(t, res)
	})
	t.Run("nil EquivalentProofsPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = nil
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Equal(t, dataRetriever.ErrNilProofsPool, err)
		require.Nil(t, res)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		res, err := resolvers.NewEquivalentProofsResolver(createMockArgEquivalentProofsResolver())
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}

func TestEquivalentProofsResolver_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createMockArgEquivalentProofsResolver()
	args.EquivalentProofsPool = nil
	res, _ := resolvers.NewEquivalentProofsResolver(args)
	require.True(t, res.IsInterfaceNil())

	res, _ = resolvers.NewEquivalentProofsResolver(createMockArgEquivalentProofsResolver())
	require.False(t, res.IsInterfaceNil())
}

func TestEquivalentProofsResolver_ProcessReceivedMessage(t *testing.T) {
	t.Parallel()

	t.Run("nil message should error", func(t *testing.T) {
		t.Parallel()

		res, _ := resolvers.NewEquivalentProofsResolver(createMockArgEquivalentProofsResolver())

		_, err := res.ProcessReceivedMessage(nil, "", nil)
		require.Equal(t, dataRetriever.ErrNilMessage, err)
	})
	t.Run("parseReceivedMessage returns error due to marshaller error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.Marshaller = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, msgID)
	})
	t.Run("invalid request type should error", func(t *testing.T) {
		t.Parallel()

		requestedBuff, err := createMockRequestedProofsBuff()
		require.Nil(t, err)

		args := createMockArgEquivalentProofsResolver()
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.ChunkType, requestedBuff), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
		require.Nil(t, msgID)
	})
	t.Run("resolveHashRequest: marshal failure before send should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return &block.HeaderProof{}, nil
			},
		}
		cnt := 0
		mockMarshaller := &marshallerMock.MarshalizerMock{}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				cnt++
				if cnt > 1 {
					return nil, expectedErr
				}

				return mockMarshaller.Marshal(obj)
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return mockMarshaller.Unmarshal(obj, buff)
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHashKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveHashRequest: invalid key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		// invalid format
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("invalidKey")), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)

		// invalid shard
		msgID, err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, []byte("hash_notAShard")), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveHashRequest: hash not found anywhere should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		wasGetProofByHashCalled := false
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				wasGetProofByHashCalled = true
				require.Equal(t, []byte("hash"), headerHash)

				return nil, expectedErr
			},
		}
		wasSearchFirstCalled := false
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				wasSearchFirstCalled = true
				require.Equal(t, []byte("hash"), key)

				return nil, expectedErr
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHashKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, msgID)
		require.True(t, wasGetProofByHashCalled)
		require.True(t, wasSearchFirstCalled)
	})
	t.Run("resolveHashRequest: should work and return from pool", func(t *testing.T) {
		t.Parallel()

		providedProof := &block.HeaderProof{}
		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return providedProof, nil
			},
		}
		wasSendCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				wasSendCalled = true

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHashKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveHashRequest: should work and return from storage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				require.Equal(t, []byte("hash"), key)

				return []byte("proof"), nil
			},
		}
		wasSendCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				wasSendCalled = true

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHashKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveMultipleHashesRequest: hashes unmarshall error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, []byte("invalid data")), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveMultipleHashesRequest: invalid key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{[]byte("invalidKey")}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveMultipleHashesRequest: hash not found anywhere should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		wasGetProofByHashCalled := false
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				wasGetProofByHashCalled = true
				require.Equal(t, []byte("hash"), headerHash)

				return nil, expectedErr
			},
		}
		wasSearchFirstCalled := false
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				wasSearchFirstCalled = true
				require.Equal(t, []byte("hash"), key)

				return nil, expectedErr
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHashKey}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, dataRetriever.ErrEquivalentProofsNotFound))
		require.Nil(t, msgID)
		require.True(t, wasGetProofByHashCalled)
		require.True(t, wasSearchFirstCalled)
	})
	t.Run("resolveMultipleHashesRequest: PackDataInChunks error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return &block.HeaderProof{}, nil
			},
		}
		args.DataPacker = &mock.DataPackerStub{
			PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
				return nil, expectedErr
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHashKey}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveMultipleHashesRequest: Send error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return &block.HeaderProof{}, nil
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				return expectedErr
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHashKey}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveMultipleHashesRequest: one hash should work and return from pool", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return &block.HeaderProof{}, nil
			},
		}
		wasSendCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				wasSendCalled = true

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHashKey}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveMultipleHashesRequest: one hash should work and return from storage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, []byte("hash"), headerHash)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				require.Equal(t, []byte("hash"), key)

				return []byte("proof"), nil
			},
		}
		wasSendCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				wasSendCalled = true

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHashKey}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveMultipleHashesRequest: one hash in pool, one in storage should work", func(t *testing.T) {
		t.Parallel()

		providedHashKey2 := []byte("hash2-2")
		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofCalled: func(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error) {
				if string(headerHash) == "hash" {
					return &block.HeaderProof{}, nil
				}
				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				if string(key) == "hash2" {
					return []byte("proof"), nil
				}
				return nil, expectedErr
			},
		}
		cntSendCalled := 0
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				cntSendCalled++

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		providedHashKeyes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHashKey, providedHashKey2}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashKeyes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.Equal(t, 2, cntSendCalled)
	})
	t.Run("resolveNonceRequest: marshal failure of proof should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
				require.Equal(t, uint64(1), headerNonce)
				require.Equal(t, uint32(1), shardID)

				return &block.HeaderProof{}, nil
			},
		}
		mockMarshaller := &marshallerMock.MarshalizerMock{}
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return mockMarshaller.Unmarshal(obj, buff)
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, providedNonceKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, msgID)
	})
	t.Run("resolveNonceRequest: invalid key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		// invalid format
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("invalidkey")), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)

		// invalid nonce
		msgID, err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("notANonce_0")), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)

		// invalid shard
		msgID, err = res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("0_notAShard")), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Error(t, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveNonceRequest: error on equivalentProofsNonceHashStorage should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		wasGetProofByNonceCalled := false
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
				wasGetProofByNonceCalled = true
				require.Equal(t, uint64(1), headerNonce)
				require.Equal(t, uint32(1), shardID)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsNonceHashStorage = &storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, providedNonceKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, msgID)
		require.True(t, wasGetProofByNonceCalled)
	})
	t.Run("resolveNonceRequest: nonce not found anywhere should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		wasGetProofByNonceCalled := false
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
				wasGetProofByNonceCalled = true
				require.Equal(t, uint64(1), headerNonce)
				require.Equal(t, uint32(1), shardID)

				return nil, expectedErr
			},
		}
		wasSearchFirstCalled := false
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				wasSearchFirstCalled = true
				require.Equal(t, []byte("hash"), key)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsNonceHashStorage = &storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return []byte("hash"), nil
			},
		}
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				require.Fail(t, "should have not been called")

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, providedNonceKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, expectedErr))
		require.Nil(t, msgID)
		require.True(t, wasGetProofByNonceCalled)
		require.True(t, wasSearchFirstCalled)
	})
	t.Run("resolveNonceRequest: should work and return from pool", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
				require.Equal(t, uint64(1), headerNonce)
				require.Equal(t, uint32(1), shardID)

				return &block.HeaderProof{}, nil
			},
		}
		wasSendCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				wasSendCalled = true

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, providedNonceKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveNonceRequest: should work and return from storage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
				require.Equal(t, uint64(1), headerNonce)
				require.Equal(t, uint32(1), shardID)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				require.Equal(t, []byte("hash"), key)

				return []byte("proof"), nil
			},
		}
		args.EquivalentProofsNonceHashStorage = &storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return []byte("hash"), nil
			},
		}
		wasSendCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
				wasSendCalled = true

				return nil
			},
		}
		res, err := resolvers.NewEquivalentProofsResolver(args)
		require.Nil(t, err)

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, providedNonceKey), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
}
