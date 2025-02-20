package dataRetriever

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignDataRetrieverContainersSetter struct {
}

// NewSovereignDataRetrieverContainerSetter creates a data retriever container setter for sovereign run type chain
func NewSovereignDataRetrieverContainerSetter() *sovereignDataRetrieverContainersSetter {
	return &sovereignDataRetrieverContainersSetter{}
}

// SetEpochHandlerToMetaBlockContainers sets epoch handler to header resolvers/requesters for ShardBlocksTopic
func (drc *sovereignDataRetrieverContainersSetter) SetEpochHandlerToMetaBlockContainers(
	epochStartTrigger epochStart.TriggerHandler,
	resolversContainer dataRetriever.ResolversContainer,
	requestersContainer dataRetriever.RequestersContainer,
) error {
	err := setEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return err
	}

	return setEpochHandlerToHdrRequester(requestersContainer, epochStartTrigger)
}

func setEpochHandlerToHdrResolver(
	resolversContainer dataRetriever.ResolversContainer,
	epochHandler dataRetriever.EpochHandler,
) error {
	topic := getShardTopic()
	resolver, err := resolversContainer.Get(topic)
	if err != nil {
		return err
	}

	hdrResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		return dataRetriever.ErrWrongTypeInContainer
	}

	return hdrResolver.SetEpochHandler(epochHandler)
}

func getShardTopic() string {
	return factory.ShardBlocksTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
}

func setEpochHandlerToHdrRequester(
	requestersContainer dataRetriever.RequestersContainer,
	epochHandler dataRetriever.EpochHandler,
) error {
	topic := getShardTopic()
	requester, err := requestersContainer.Get(topic)
	if err != nil {
		return err
	}

	hdrRequester, ok := requester.(dataRetriever.HeaderRequester)
	if !ok {
		return dataRetriever.ErrWrongTypeInContainer
	}

	return hdrRequester.SetEpochHandler(epochHandler)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (drc *sovereignDataRetrieverContainersSetter) IsInterfaceNil() bool {
	return drc == nil
}
