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

var providedHash = []byte("headerHash")

func createMockArgEquivalentProofsResolver() resolvers.ArgEquivalentProofsResolver {
	return resolvers.ArgEquivalentProofsResolver{
		ArgBaseResolver:         createMockArgBaseResolver(),
		DataPacker:              &mock.DataPackerStub{},
		EquivalentProofsStorage: &storageStubs.StorerStub{},
		EquivalentProofsPool:    &dataRetrieverMocks.ProofsPoolMock{},
		IsFullHistoryNode:       false,
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
		require.Equal(t, dataRetriever.ErrNilProofsStorage, err)
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
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

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

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHash), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveHashRequest: hash not found anywhere should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		wasGetProofByHashCalled := false
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				wasGetProofByHashCalled = true
				require.Equal(t, providedHash, headerHash)

				return nil, expectedErr
			},
		}
		wasSearchFirstCalled := false
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				wasSearchFirstCalled = true
				require.Equal(t, providedHash, key)

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

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHash), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
		require.True(t, wasGetProofByHashCalled)
		require.True(t, wasSearchFirstCalled)
	})
	t.Run("resolveHashRequest: should work and return from pool", func(t *testing.T) {
		t.Parallel()

		providedProof := &block.HeaderProof{}
		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

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

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHash), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveHashRequest: should work and return from storage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				require.Equal(t, providedHash, key)

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

		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, providedHash), fromConnectedPeer, &p2pmocks.MessengerStub{})
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
	t.Run("resolveMultipleHashesRequest: hash not found anywhere should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		wasGetProofByHashCalled := false
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				wasGetProofByHashCalled = true
				require.Equal(t, providedHash, headerHash)

				return nil, expectedErr
			},
		}
		wasSearchFirstCalled := false
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				wasSearchFirstCalled = true
				require.Equal(t, providedHash, key)

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

		providedHashes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHash}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.True(t, errors.Is(err, dataRetriever.ErrEquivalentProofsNotFound))
		require.Nil(t, msgID)
		require.True(t, wasGetProofByHashCalled)
		require.True(t, wasSearchFirstCalled)
	})
	t.Run("resolveMultipleHashesRequest: PackDataInChunks error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

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

		providedHashes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHash}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveMultipleHashesRequest: Send error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

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

		providedHashes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHash}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.Equal(t, expectedErr, err)
		require.Nil(t, msgID)
	})
	t.Run("resolveMultipleHashesRequest: one hash should work and return from pool", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

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

		providedHashes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHash}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveMultipleHashesRequest: one hash should work and return from storage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				require.Equal(t, providedHash, headerHash)

				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				require.Equal(t, providedHash, key)

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

		providedHashes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHash}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.True(t, wasSendCalled)
	})
	t.Run("resolveMultipleHashesRequest: one hash in pool, one in storage should work", func(t *testing.T) {
		t.Parallel()

		providedHash2 := []byte("headerHash2")
		args := createMockArgEquivalentProofsResolver()
		args.EquivalentProofsPool = &dataRetrieverMocks.ProofsPoolMock{
			GetProofByHashCalled: func(headerHash []byte) (data.HeaderProofHandler, error) {
				if string(headerHash) == string(providedHash) {
					return &block.HeaderProof{}, nil
				}
				return nil, expectedErr
			},
		}
		args.EquivalentProofsStorage = &storageStubs.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				if string(key) == string(providedHash2) {
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

		providedHashes, err := args.Marshaller.Marshal(batch.Batch{Data: [][]byte{providedHash, providedHash2}})
		require.Nil(t, err)
		msgID, err := res.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashArrayType, providedHashes), fromConnectedPeer, &p2pmocks.MessengerStub{})
		require.NoError(t, err)
		require.Nil(t, msgID)
		require.Equal(t, 2, cntSendCalled)
	})
}
