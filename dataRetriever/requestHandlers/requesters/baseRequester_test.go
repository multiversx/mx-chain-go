package requesters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
)

func createMockArgBaseRequester() ArgBaseRequester {
	return ArgBaseRequester{
		RequestSender: &dataRetrieverMocks.TopicRequestSenderStub{},
		Marshaller:    &marshallerMock.MarshalizerStub{},
	}
}

func Test_createBaseRequester(t *testing.T) {
	t.Parallel()

	baseHandler := createBaseRequester(createMockArgBaseRequester())
	assert.False(t, check.IfNilReflect(baseHandler))
}

func Test_checkArgBase(t *testing.T) {
	t.Parallel()

	t.Run("nil request sender should error", func(t *testing.T) {
		t.Parallel()

		err := checkArgBase(ArgBaseRequester{
			RequestSender: nil,
			Marshaller:    &marshallerMock.MarshalizerStub{},
		})
		assert.Equal(t, err, dataRetriever.ErrNilRequestSender)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		err := checkArgBase(ArgBaseRequester{
			RequestSender: &dataRetrieverMocks.TopicRequestSenderStub{},
			Marshaller:    nil,
		})
		assert.Equal(t, err, dataRetriever.ErrNilMarshalizer)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		err := checkArgBase(createMockArgBaseRequester())
		assert.Nil(t, err)
	})
}

func TestBaseRequester_RequestDataFromHash(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(1234)
	providedHash := []byte("provided hash")
	providedHashes := [][]byte{providedHash}
	wasCalled := false
	requestSender := &dataRetrieverMocks.TopicRequestSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
			wasCalled = true
			assert.Equal(t, providedHash, rd.Value)
			assert.Equal(t, providedHashes, originalHashes)
			assert.Equal(t, dataRetriever.HashType, rd.Type)
			assert.Equal(t, providedEpoch, rd.Epoch)
			return nil
		},
	}
	baseHandler := createBaseRequester(ArgBaseRequester{
		RequestSender: requestSender,
		Marshaller:    &marshallerMock.MarshalizerStub{},
	})
	assert.False(t, check.IfNilReflect(baseHandler))

	assert.Nil(t, baseHandler.RequestDataFromHash(providedHash, providedEpoch))
	assert.True(t, wasCalled)
}

func TestBaseRequester_NumPeersToQuery(t *testing.T) {
	t.Parallel()

	providedIntra := 123
	providedCross := 100
	wasCalled := false
	requestSender := &dataRetrieverMocks.TopicRequestSenderStub{
		SetNumPeersToQueryCalled: func(intra int, cross int) {
			wasCalled = true
			assert.Equal(t, providedIntra, intra)
			assert.Equal(t, providedCross, cross)
		},
		GetNumPeersToQueryCalled: func() (int, int) {
			return providedIntra, providedCross
		},
	}
	baseHandler := createBaseRequester(ArgBaseRequester{
		RequestSender: requestSender,
		Marshaller:    &marshallerMock.MarshalizerStub{},
	})
	assert.False(t, check.IfNilReflect(baseHandler))

	baseHandler.SetNumPeersToQuery(providedIntra, providedCross)
	assert.True(t, wasCalled)

	intra, cross := baseHandler.NumPeersToQuery()
	assert.Equal(t, providedIntra, intra)
	assert.Equal(t, providedCross, cross)
}

func TestBaseRequester_SetDebugHandler(t *testing.T) {
	t.Parallel()

	providedDebugHandler := &mock.DebugHandler{}
	requestSender := &dataRetrieverMocks.TopicRequestSenderStub{
		SetDebugHandlerCalled: func(handler dataRetriever.DebugHandler) error {
			assert.Equal(t, providedDebugHandler, handler)
			return nil
		},
		DebugHandlerCalled: func() dataRetriever.DebugHandler {
			return providedDebugHandler
		},
	}
	baseHandler := createBaseRequester(ArgBaseRequester{
		RequestSender: requestSender,
		Marshaller:    &marshallerMock.MarshalizerStub{},
	})
	assert.False(t, check.IfNilReflect(baseHandler))

	assert.Nil(t, baseHandler.SetDebugHandler(providedDebugHandler))
	assert.Equal(t, providedDebugHandler, requestSender.DebugHandler())
}
