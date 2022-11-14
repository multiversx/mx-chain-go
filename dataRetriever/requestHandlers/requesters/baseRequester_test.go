package requesters

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgBaseRequester() ArgBaseRequester {
	return ArgBaseRequester{
		SenderResolver: &mock.TopicResolverSenderStub{},
		Marshaller:     &testscommon.MarshalizerStub{},
	}
}

func Test_createBaseRequester(t *testing.T) {
	t.Parallel()

	baseHandler := createBaseRequester(createMockArgBaseRequester())
	assert.False(t, check.IfNilReflect(baseHandler))
}

func Test_checkArgBase(t *testing.T) {
	t.Parallel()

	t.Run("nil sender resolver should error", func(t *testing.T) {
		t.Parallel()

		err := checkArgBase(ArgBaseRequester{
			SenderResolver: nil,
			Marshaller:     &testscommon.MarshalizerStub{},
		})
		assert.Equal(t, err, dataRetriever.ErrNilResolverSender)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		err := checkArgBase(ArgBaseRequester{
			SenderResolver: &mock.TopicResolverSenderStub{},
			Marshaller:     nil,
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
	senderResolver := &mock.TopicResolverSenderStub{
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
		SenderResolver: senderResolver,
		Marshaller:     &testscommon.MarshalizerStub{},
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
	senderResolver := &mock.TopicResolverSenderStub{
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
		SenderResolver: senderResolver,
		Marshaller:     &testscommon.MarshalizerStub{},
	})
	assert.False(t, check.IfNilReflect(baseHandler))

	baseHandler.SetNumPeersToQuery(providedIntra, providedCross)
	assert.True(t, wasCalled)

	intra, cross := baseHandler.NumPeersToQuery()
	assert.Equal(t, providedIntra, intra)
	assert.Equal(t, providedCross, cross)
}

func TestBaseRequester_SetResolverDebugHandler(t *testing.T) {
	t.Parallel()

	providedDebugHandler := &mock.ResolverDebugHandler{}
	senderResolver := &mock.TopicResolverSenderStub{}
	baseHandler := createBaseRequester(ArgBaseRequester{
		SenderResolver: senderResolver,
		Marshaller:     &testscommon.MarshalizerStub{},
	})
	assert.False(t, check.IfNilReflect(baseHandler))

	assert.Nil(t, baseHandler.SetResolverDebugHandler(providedDebugHandler))
	assert.Equal(t, providedDebugHandler, senderResolver.ResolverDebugHandler())
}
