package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
)

type sovereignDataRetrieverContainersSetter struct {
}

// NewSovereignDataRetrieverContainerSetter creates a data retriever container setter for sovereign run type chain
func NewSovereignDataRetrieverContainerSetter() *sovereignDataRetrieverContainersSetter {
	return &sovereignDataRetrieverContainersSetter{}
}

// SetEpochHandlerToMetaBlockContainers does nothing for sovereign chain, since we do not need to sync any meta blocks
func (drc *sovereignDataRetrieverContainersSetter) SetEpochHandlerToMetaBlockContainers(
	_ epochStart.TriggerHandler,
	_ dataRetriever.ResolversContainer,
	_ dataRetriever.RequestersContainer,
) error {
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (drc *sovereignDataRetrieverContainersSetter) IsInterfaceNil() bool {
	return drc == nil
}
