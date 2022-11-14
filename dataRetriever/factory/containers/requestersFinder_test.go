package containers

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	dataRetrieverTests "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
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

func createMockContainer(expectedKey string) *dataRetrieverTests.RequestersContainerStub {
	return &dataRetrieverTests.RequestersContainerStub{
		GetCalled: func(key string) (resolver dataRetriever.Requester, e error) {
			if key == expectedKey {
				return &dataRetrieverTests.RequesterStub{}, nil
			}

			return nil, nil
		},
	}
}

//------- NewRequestersFinder

func TestNewRequestersFinder_NilContainerShouldErr(t *testing.T) {
	t.Parallel()

	rf, err := NewRequestersFinder(nil, &mock.CoordinatorStub{})

	assert.Nil(t, rf)
	assert.Equal(t, dataRetriever.ErrNilRequestersContainer, err)
}

func TestNewRequestersFinder_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rf, err := NewRequestersFinder(&dataRetrieverTests.RequestersContainerStub{}, nil)

	assert.Nil(t, rf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewRequestersFinder_ShouldWork(t *testing.T) {
	t.Parallel()

	rf, err := NewRequestersFinder(&dataRetrieverTests.RequestersContainerStub{}, &mock.CoordinatorStub{})

	assert.NotNil(t, rf)
	assert.Nil(t, err)
	assert.False(t, rf.IsInterfaceNil())
}

func TestRequestersFinder_IntraShardRequester(t *testing.T) {
	currentShardID := uint32(4)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(currentShardID))

	rf, _ := NewRequestersFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, currentShardID),
	)

	resolver, _ := rf.IntraShardRequester(baseTopic)
	assert.NotNil(t, resolver)
}

func TestRequestersFinder_MetaChainRequester(t *testing.T) {
	currentShardID := uint32(4)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	rf, _ := NewRequestersFinder(
		createMockContainer(baseTopic),
		createMockCoordinator(identifierPrefix, currentShardID),
	)

	resolver, _ := rf.MetaChainRequester(baseTopic)
	assert.NotNil(t, resolver)
}

func TestRequestersFinder_CrossShardRequester(t *testing.T) {
	crossShardID := uint32(5)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(crossShardID))

	rf, _ := NewRequestersFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, crossShardID),
	)

	resolver, _ := rf.CrossShardRequester(baseTopic, crossShardID)
	assert.NotNil(t, resolver)
}
