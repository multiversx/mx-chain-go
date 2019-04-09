package containers

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
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

func createMockContainer(expectedKey string) *mock.ResolversContainerStub {
	return &mock.ResolversContainerStub{
		GetCalled: func(key string) (resolver dataRetriever.Resolver, e error) {
			if key == expectedKey {
				return &mock.ResolverStub{}, nil
			}

			return nil, nil
		},
	}
}

//------- NewResolversFinder

func TestNewResolversFinder_NilContainerShouldErr(t *testing.T) {
	t.Parallel()

	rf, err := NewResolversFinder(nil, &mock.CoordinatorStub{})

	assert.Nil(t, rf)
	assert.Equal(t, dataRetriever.ErrNilResolverContainer, err)
}

func TestNewResolversFinder_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rf, err := NewResolversFinder(&mock.ResolversContainerStub{}, nil)

	assert.Nil(t, rf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewResolversFinder_ShouldWork(t *testing.T) {
	t.Parallel()

	rf, err := NewResolversFinder(&mock.ResolversContainerStub{}, &mock.CoordinatorStub{})

	assert.NotNil(t, rf)
	assert.Nil(t, err)
}

func TestResolversFinder_IntraShardResolver(t *testing.T) {
	currentShardID := uint32(4)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(currentShardID))

	rf, _ := NewResolversFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, currentShardID),
	)

	resolver, _ := rf.IntraShardResolver(baseTopic)
	assert.NotNil(t, resolver)
}

func TestResolversFinder_CrossShardResolver(t *testing.T) {
	crossShardID := uint32(5)
	identifierPrefix := "_"
	baseTopic := "baseTopic"

	expectedTopic := baseTopic + identifierPrefix + strconv.Itoa(int(crossShardID))

	rf, _ := NewResolversFinder(
		createMockContainer(expectedTopic),
		createMockCoordinator(identifierPrefix, crossShardID),
	)

	resolver, _ := rf.CrossShardResolver(baseTopic, crossShardID)
	assert.NotNil(t, resolver)
}
