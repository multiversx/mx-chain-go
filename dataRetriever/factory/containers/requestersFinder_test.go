package containers

import (
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func createMockCoordinator(identifierPrefix string, currentShardID uint32) *mock.CoordinatorStub {
	return &mock.CoordinatorStub{
		CommunicationIdentifierCalled: func(destShardID uint32) string {
			return identifierPrefix + strconv.Itoa(int(destShardID))
		},
		SelfIdCalled: func() uint32 {
			return currentShardID
		},
	}
}

func createMockContainer(expectedKey string) *dataRetrieverMocks.RequestersContainerStub {
	return &dataRetrieverMocks.RequestersContainerStub{
		GetCalled: func(key string) (requester dataRetriever.Requester, e error) {
			if key == expectedKey {
				return &dataRetrieverMocks.RequesterStub{}, nil
			}

			return nil, nil
		},
	}
}

func TestNewRequestersFinder(t *testing.T) {
	t.Parallel()

	t.Run("nil container should error", func(t *testing.T) {
		t.Parallel()

		rf, err := NewRequestersFinder(nil, &mock.CoordinatorStub{})

		assert.Nil(t, rf)
		assert.Equal(t, dataRetriever.ErrNilRequestersContainer, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		rf, err := NewRequestersFinder(&dataRetrieverMocks.RequestersContainerStub{}, nil)

		assert.Nil(t, rf)
		assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		rf, err := NewRequestersFinder(&dataRetrieverMocks.RequestersContainerStub{}, &mock.CoordinatorStub{})

		assert.NotNil(t, rf)
		assert.Nil(t, err)
		assert.False(t, rf.IsInterfaceNil())
	})
}

func TestRequestersFinder_IntraShardRequester(t *testing.T) {
	t.Parallel()

	currentShardID := uint32(4)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(currentShardID))

	rf, _ := NewRequestersFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, currentShardID),
	)

	requester, _ := rf.IntraShardRequester(baseTopic)
	assert.NotNil(t, requester)
}

func TestRequestersFinder_MetaChainRequester(t *testing.T) {
	t.Parallel()

	currentShardID := uint32(4)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	rf, _ := NewRequestersFinder(
		createMockContainer(baseTopic),
		createMockCoordinator(identifierPrefix, currentShardID),
	)

	requester, _ := rf.MetaChainRequester(baseTopic)
	assert.NotNil(t, requester)
}

func TestRequestersFinder_CrossShardRequester(t *testing.T) {
	t.Parallel()

	crossShardID := uint32(5)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(crossShardID))

	rf, _ := NewRequestersFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, crossShardID),
	)

	requester, _ := rf.CrossShardRequester(baseTopic, crossShardID)
	assert.NotNil(t, requester)
}

func TestRequestersFinder_MetaCrossShardRequester(t *testing.T) {
	t.Parallel()

	crossShardID := uint32(5)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(crossShardID)) + identifierPrefix + common.MetachainTopicIdentifier

	rf, _ := NewRequestersFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, crossShardID),
	)

	requester, _ := rf.MetaCrossShardRequester(baseTopic, crossShardID)
	assert.NotNil(t, requester)
}
