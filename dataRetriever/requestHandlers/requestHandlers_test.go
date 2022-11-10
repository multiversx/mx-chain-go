package requestHandlers

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

type requestHandlerType string

const (
	peerAuthRequestHandler requestHandlerType = "peerAuthRH"
	mbRequestHandler       requestHandlerType = "miniblockRH"
	txRequestHandler       requestHandlerType = "transactionRH"
	trieRequestHandler     requestHandlerType = "trieNodeRH"
	vInfoRequestHandler    requestHandlerType = "validatorInfoNodeRH"
)

func Test_RequestHandlers(t *testing.T) {
	t.Parallel()
	testNewRequestHandler(t, peerAuthRequestHandler)
	testNewRequestHandler(t, mbRequestHandler)
	testNewRequestHandler(t, txRequestHandler)
	testNewRequestHandler(t, trieRequestHandler)
	testNewRequestHandler(t, vInfoRequestHandler)

	testRequestDataFromHashArray(t, peerAuthRequestHandler)
	testRequestDataFromHashArray(t, mbRequestHandler)
	testRequestDataFromHashArray(t, txRequestHandler)
	testRequestDataFromHashArray(t, trieRequestHandler)
	testRequestDataFromHashArray(t, vInfoRequestHandler)

	testRequestDataFromReferenceAndChunk(t, trieRequestHandler)
}

func testNewRequestHandler(t *testing.T, handlerType requestHandlerType) {
	t.Run("nil base arg should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockArgBaseRequestHandler()
		argsBase.Marshaller = nil
		handler, err := getHandler(handlerType, argsBase)
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := getHandler(handlerType, createMockArgBaseRequestHandler())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func testRequestDataFromHashArray(t *testing.T, handlerType requestHandlerType) {
	t.Run("marshaller returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgBaseRequestHandler()
		args.Marshaller = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, errExpected
			},
		}
		handler, err := getHandler(handlerType, args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		hashArrayHandler, ok := handler.(HashSliceResolver)
		assert.True(t, ok)
		assert.Equal(t, errExpected, hashArrayHandler.RequestDataFromHashArray(nil, 0))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedEpoch := uint32(1234)
		providedHashes := [][]byte{[]byte("hash 1"), []byte("hash 2"), []byte("hash 3")}
		args := createMockArgBaseRequestHandler()
		wasCalled := false
		args.SenderResolver = &mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
				wasCalled = true
				b := &batch.Batch{
					Data: providedHashes,
				}
				batchBytes, err := args.Marshaller.Marshal(b)
				assert.Nil(t, err)
				assert.Equal(t, batchBytes, rd.Value)
				assert.Equal(t, providedHashes, originalHashes)
				assert.Equal(t, dataRetriever.HashArrayType, rd.Type)
				assert.Equal(t, providedEpoch, rd.Epoch)
				return nil
			},
		}
		handler, err := getHandler(handlerType, args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
		hashArrayHandler, ok := handler.(HashSliceResolver)
		assert.True(t, ok)
		assert.Nil(t, hashArrayHandler.RequestDataFromHashArray(providedHashes, providedEpoch))
		assert.True(t, wasCalled)
	})
}

func testRequestDataFromReferenceAndChunk(t *testing.T, handlerType requestHandlerType) {
	providedChunkIndex := uint32(1234)
	providedHash := []byte("hash 1")
	providedHashes := [][]byte{providedHash}
	args := createMockArgBaseRequestHandler()
	wasCalled := false
	args.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedHash, rd.Value)
			assert.Equal(t, providedHashes, originalHashes)
			assert.Equal(t, dataRetriever.HashType, rd.Type)
			assert.Equal(t, providedChunkIndex, rd.ChunkIndex)
			return nil
		},
	}
	handler, err := getHandler(handlerType, args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(handler))
	chunkHandler, ok := handler.(ChunkResolver)
	assert.True(t, ok)
	assert.Nil(t, chunkHandler.RequestDataFromReferenceAndChunk(providedHash, providedChunkIndex))
	assert.True(t, wasCalled)
}

func getHandler(handlerType requestHandlerType, argsBase ArgBaseRequestHandler) (check.NilInterfaceChecker, error) {
	switch handlerType {
	case peerAuthRequestHandler:
		return NewPeerAuthenticationRequestHandler(ArgPeerAuthenticationRequestHandler{argsBase})
	case mbRequestHandler:
		return NewMiniblockRequestHandler(ArgMiniblockRequestHandler{argsBase})
	case txRequestHandler:
		return NewTransactionRequestHandler(ArgTransactionRequestHandler{argsBase})
	case trieRequestHandler:
		return NewTrieNodeRequestHandler(ArgTrieNodeRequestHandler{argsBase})
	case vInfoRequestHandler:
		return NewValidatorInfoRequestHandler(ArgValidatorInfoRequestHandler{argsBase})
	}
	return nil, errors.New("invalid handler type")
}
