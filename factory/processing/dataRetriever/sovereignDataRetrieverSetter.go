package dataRetriever

import (
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

// SetEpochHandlerToMetaBlockContainers does nothing for sovereign chain, since we do not need to sync any meta blocks
func (drc *sovereignDataRetrieverContainersSetter) SetEpochHandlerToMetaBlockContainers(
	epochStartTrigger epochStart.TriggerHandler,
	resolversContainer dataRetriever.ResolversContainer,
	requestersContainer dataRetriever.RequestersContainer,
) error {
	err := SetEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return err
	}

	return SetEpochHandlerToHdrRequester(requestersContainer, epochStartTrigger)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (drc *sovereignDataRetrieverContainersSetter) IsInterfaceNil() bool {
	return drc == nil
}

// SetEpochHandlerToHdrResolver sets the epoch handler to the metablock hdr resolver
func SetEpochHandlerToHdrResolver(
	resolversContainer dataRetriever.ResolversContainer,
	epochHandler dataRetriever.EpochHandler,
) error {
	resolver, err := resolversContainer.Get(factory.ShardBlocksTopic + "_0")
	if err != nil {
		return err
	}

	hdrResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		return dataRetriever.ErrWrongTypeInContainer
	}

	err = hdrResolver.SetEpochHandler(epochHandler)
	if err != nil {
		return err
	}

	return nil
}

// SetEpochHandlerToHdrRequester sets the epoch handler to the metablock hdr requester
func SetEpochHandlerToHdrRequester(
	requestersContainer dataRetriever.RequestersContainer,
	epochHandler dataRetriever.EpochHandler,
) error {
	requester, err := requestersContainer.Get(factory.ShardBlocksTopic + "_0")
	if err != nil {
		return err
	}

	hdrRequester, ok := requester.(dataRetriever.HeaderRequester)
	if !ok {
		return dataRetriever.ErrWrongTypeInContainer
	}

	err = hdrRequester.SetEpochHandler(epochHandler)
	if err != nil {
		return err
	}

	return nil
}
