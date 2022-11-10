package requestHandlers

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	dataRetrieverMocks "github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgHeaderRequestHandler(argBase ArgBaseRequestHandler) ArgHeaderRequestHandler {
	return ArgHeaderRequestHandler{
		ArgBaseRequestHandler: argBase,
		NonceConverter:        &mock.Uint64ByteSliceConverterMock{},
	}
}

func TestNewHeaderRequestHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil base arg should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockArgBaseRequestHandler()
		argsBase.Marshaller = nil
		handler, err := NewHeaderRequestHandler(createMockArgHeaderRequestHandler(argsBase))
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("nil nonce converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgHeaderRequestHandler(createMockArgBaseRequestHandler())
		args.NonceConverter = nil
		handler, err := NewHeaderRequestHandler(args)
		assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
		assert.True(t, check.IfNil(handler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := NewHeaderRequestHandler(createMockArgHeaderRequestHandler(createMockArgBaseRequestHandler()))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func TestHeaderRequestHandler_RequestDataFromNonce(t *testing.T) {
	t.Parallel()

	providedNonce := uint64(1234)
	providedEpoch := uint32(1000)
	providedNonceConverter := mock.NewNonceHashConverterMock()
	argBase := createMockArgBaseRequestHandler()
	wasCalled := false
	argBase.SenderResolver = &dataRetrieverMocks.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedNonceConverter.ToByteSlice(providedNonce), rd.Value)
			assert.Equal(t, [][]byte{providedNonceConverter.ToByteSlice(providedNonce)}, originalHashes)
			assert.Equal(t, dataRetriever.NonceType, rd.Type)
			assert.Equal(t, providedEpoch, rd.Epoch)
			return nil
		},
	}
	args := ArgHeaderRequestHandler{
		ArgBaseRequestHandler: argBase,
		NonceConverter:        providedNonceConverter,
	}
	handler, err := NewHeaderRequestHandler(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(handler))

	assert.Nil(t, handler.RequestDataFromNonce(providedNonce, providedEpoch))
	assert.True(t, wasCalled)
}

func TestHeaderRequestHandler_RequestDataFromEpoch(t *testing.T) {
	t.Parallel()

	providedIdentifier := []byte("provided identifier")
	argBase := createMockArgBaseRequestHandler()
	wasCalled := false
	argBase.SenderResolver = &dataRetrieverMocks.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedIdentifier, rd.Value)
			assert.Equal(t, [][]byte{providedIdentifier}, originalHashes)
			assert.Equal(t, dataRetriever.EpochType, rd.Type)
			return nil
		},
	}
	handler, err := NewHeaderRequestHandler(createMockArgHeaderRequestHandler(argBase))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(handler))

	assert.Nil(t, handler.RequestDataFromEpoch(providedIdentifier))
	assert.True(t, wasCalled)
}
