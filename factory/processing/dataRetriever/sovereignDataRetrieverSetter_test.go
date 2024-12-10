package dataRetriever

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	mockRetriever "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestDataRetrieverContainersSetter_SetEpochHandlerToMetaBlockContainers(t *testing.T) {
	t.Parallel()

	sovSetter := NewSovereignDataRetrieverContainerSetter()
	require.False(t, sovSetter.IsInterfaceNil())

	expectedTopic := fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId)
	triggerStub := &testscommon.EpochStartTriggerStub{}

	wasEpochHandlerSetInResolver := false
	resolver := &mockRetriever.HeaderResolverStub{
		SetEpochHandlerCalled: func(epochHandler retriever.EpochHandler) error {
			require.Equal(t, triggerStub, epochHandler)
			wasEpochHandlerSetInResolver = true
			return nil
		},
	}
	resolversContainer := &dataRetriever.ResolversContainerStub{
		GetCalled: func(key string) (retriever.Resolver, error) {
			require.Equal(t, expectedTopic, key)
			return resolver, nil
		},
	}

	wasEpochHandlerSetInRequester := false
	requester := &mockRetriever.HeaderRequesterStub{
		SetEpochHandlerCalled: func(epochHandler retriever.EpochHandler) error {
			require.Equal(t, triggerStub, epochHandler)
			wasEpochHandlerSetInRequester = true
			return nil
		},
	}
	requestersContainer := &dataRetriever.RequestersContainerStub{
		GetCalled: func(key string) (retriever.Requester, error) {
			require.Equal(t, expectedTopic, key)
			return requester, nil
		},
	}
	err := sovSetter.SetEpochHandlerToMetaBlockContainers(triggerStub, resolversContainer, requestersContainer)
	require.Nil(t, err)
	require.True(t, wasEpochHandlerSetInResolver)
	require.True(t, wasEpochHandlerSetInRequester)
}
