package disabled

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers/disabled"
)

type resolversContainerFactory struct {
}

// NewDisabledResolversContainerFactory returns a new instance of disabled resolvers container factory
func NewDisabledResolversContainerFactory() *resolversContainerFactory {
	return &resolversContainerFactory{}
}

// Create returns a disabled resolvers container and nil
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	return disabled.NewDisabledResolversContainer(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}
