package requesters

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	dataRetrieverStub "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

type requestHandlerType string

const (
	peerAuthRequester requestHandlerType = "peerAuthRequester"
	mbRequester       requestHandlerType = "miniblockRequester"
	txRequester       requestHandlerType = "transactionRequester"
	trieRequester     requestHandlerType = "trieNodeRequester"
	vInfoRequester    requestHandlerType = "validatorInfoNodeRequester"
)

var expectedErr = errors.New("expected error")

func Test_Requesters(t *testing.T) {
	t.Parallel()
	testNewRequester(t, peerAuthRequester)
	testNewRequester(t, mbRequester)
	testNewRequester(t, txRequester)
	testNewRequester(t, trieRequester)
	testNewRequester(t, vInfoRequester)

	testRequestDataFromHashArray(t, peerAuthRequester)
	testRequestDataFromHashArray(t, mbRequester)
	testRequestDataFromHashArray(t, txRequester)
	testRequestDataFromHashArray(t, trieRequester)
	testRequestDataFromHashArray(t, vInfoRequester)

	testRequestDataFromReferenceAndChunk(t, trieRequester)
}

func testNewRequester(t *testing.T, requesterType requestHandlerType) {
	t.Run("nil base arg should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockArgBaseRequester()
		argsBase.Marshaller = nil
		requester, err := getHandler(requesterType, argsBase)
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(requester))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		requester, err := getHandler(requesterType, createMockArgBaseRequester())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(requester))
	})
}

func testRequestDataFromHashArray(t *testing.T, requesterType requestHandlerType) {
	t.Run("marshaller returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgBaseRequester()
		args.Marshaller = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		requester, err := getHandler(requesterType, args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(requester))
		hashArrayHandler, ok := requester.(requestHandlers.HashSliceRequester)
		assert.True(t, ok)
		assert.Equal(t, expectedErr, hashArrayHandler.RequestDataFromHashArray(nil, 0))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedEpoch := uint32(1234)
		providedHashes := [][]byte{[]byte("hash 1"), []byte("hash 2"), []byte("hash 3")}
		args := createMockArgBaseRequester()
		wasCalled := false
		args.RequestSender = &dataRetrieverStub.TopicRequestSenderStub{
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
		requester, err := getHandler(requesterType, args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(requester))
		hashArrayHandler, ok := requester.(requestHandlers.HashSliceRequester)
		assert.True(t, ok)
		assert.Nil(t, hashArrayHandler.RequestDataFromHashArray(providedHashes, providedEpoch))
		assert.True(t, wasCalled)
	})
}

func testRequestDataFromReferenceAndChunk(t *testing.T, requesterType requestHandlerType) {
	providedChunkIndex := uint32(1234)
	providedHash := []byte("hash 1")
	providedHashes := [][]byte{providedHash}
	args := createMockArgBaseRequester()
	wasCalled := false
	args.RequestSender = &dataRetrieverStub.TopicRequestSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedHash, rd.Value)
			assert.Equal(t, providedHashes, originalHashes)
			assert.Equal(t, dataRetriever.HashType, rd.Type)
			assert.Equal(t, providedChunkIndex, rd.ChunkIndex)
			return nil
		},
	}
	requester, err := getHandler(requesterType, args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(requester))
	chunkHandler, ok := requester.(requestHandlers.ChunkRequester)
	assert.True(t, ok)
	assert.Nil(t, chunkHandler.RequestDataFromReferenceAndChunk(providedHash, providedChunkIndex))
	assert.True(t, wasCalled)
}

func getHandler(requesterType requestHandlerType, argsBase ArgBaseRequester) (check.NilInterfaceChecker, error) {
	switch requesterType {
	case peerAuthRequester:
		return NewPeerAuthenticationRequester(ArgPeerAuthenticationRequester{argsBase})
	case mbRequester:
		return NewMiniblockRequester(ArgMiniblockRequester{argsBase})
	case txRequester:
		return NewTransactionRequester(ArgTransactionRequester{argsBase})
	case trieRequester:
		return NewTrieNodeRequester(ArgTrieNodeRequester{argsBase})
	case vInfoRequester:
		return NewValidatorInfoRequester(ArgValidatorInfoRequester{argsBase})
	}
	return nil, errors.New("invalid requester type")
}
