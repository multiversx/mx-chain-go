package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
)

type dataRetrieverContainersSetter struct {
}

// NewDataRetrieverContainerSetter creates a data retriever container setter for regular run type chain
func NewDataRetrieverContainerSetter() *dataRetrieverContainersSetter {
	return &dataRetrieverContainersSetter{}
}

// SetEpochHandlerToMetaBlockContainers sets the epoch start trigger to meta block resolvers and requesters container
func (drc *dataRetrieverContainersSetter) SetEpochHandlerToMetaBlockContainers(
	epochStartTrigger epochStart.TriggerHandler,
	resolversContainer dataRetriever.ResolversContainer,
	requestersContainer dataRetriever.RequestersContainer,
) error {
	err := dataRetriever.SetEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return err
	}

	return dataRetriever.SetEpochHandlerToHdrRequester(requestersContainer, epochStartTrigger)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (drc *dataRetrieverContainersSetter) IsInterfaceNil() bool {
	return drc == nil
}
