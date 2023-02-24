package requesters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/mock"
	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func createMockArgHeaderRequester(argBase ArgBaseRequester) ArgHeaderRequester {
	return ArgHeaderRequester{
		ArgBaseRequester: argBase,
		NonceConverter:   &mock.Uint64ByteSliceConverterMock{},
	}
}

func TestNewHeaderRequester(t *testing.T) {
	t.Parallel()

	t.Run("nil base arg should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockArgBaseRequester()
		argsBase.Marshaller = nil
		requester, err := NewHeaderRequester(createMockArgHeaderRequester(argsBase))
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(requester))
	})
	t.Run("nil nonce converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgHeaderRequester(createMockArgBaseRequester())
		args.NonceConverter = nil
		requester, err := NewHeaderRequester(args)
		assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
		assert.True(t, check.IfNil(requester))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		requester, err := NewHeaderRequester(createMockArgHeaderRequester(createMockArgBaseRequester()))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(requester))
	})
}

func TestHeaderRequester_RequestDataFromNonce(t *testing.T) {
	t.Parallel()

	providedNonce := uint64(1234)
	providedEpoch := uint32(1000)
	providedNonceConverter := mock.NewNonceHashConverterMock()
	argBase := createMockArgBaseRequester()
	wasCalled := false
	argBase.RequestSender = &dataRetrieverTests.TopicRequestSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedNonceConverter.ToByteSlice(providedNonce), rd.Value)
			assert.Equal(t, [][]byte{providedNonceConverter.ToByteSlice(providedNonce)}, originalHashes)
			assert.Equal(t, dataRetriever.NonceType, rd.Type)
			assert.Equal(t, providedEpoch, rd.Epoch)
			return nil
		},
	}
	args := ArgHeaderRequester{
		ArgBaseRequester: argBase,
		NonceConverter:   providedNonceConverter,
	}
	requester, err := NewHeaderRequester(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(requester))

	assert.Nil(t, requester.RequestDataFromNonce(providedNonce, providedEpoch))
	assert.True(t, wasCalled)
}

func TestHeaderRequester_RequestDataFromEpoch(t *testing.T) {
	t.Parallel()

	providedIdentifier := []byte("provided identifier")
	argBase := createMockArgBaseRequester()
	wasCalled := false
	argBase.RequestSender = &dataRetrieverTests.TopicRequestSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedIdentifier, rd.Value)
			assert.Equal(t, [][]byte{providedIdentifier}, originalHashes)
			assert.Equal(t, dataRetriever.EpochType, rd.Type)
			return nil
		},
	}
	requester, err := NewHeaderRequester(createMockArgHeaderRequester(argBase))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(requester))

	assert.Nil(t, requester.RequestDataFromEpoch(providedIdentifier))
	assert.True(t, wasCalled)
}

func TestHeaderRequester_SetEpochHandler(t *testing.T) {
	t.Parallel()

	argBase := createMockArgBaseRequester()
	requester, _ := NewHeaderRequester(createMockArgHeaderRequester(argBase))

	assert.Nil(t, requester.SetEpochHandler(nil))
}
