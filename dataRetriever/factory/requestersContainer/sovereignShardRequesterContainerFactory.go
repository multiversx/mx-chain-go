package requesterscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

// TODO: Implement this in MX-14516

type sovereignShardRequestersContainerFactory struct {
	*baseRequestersContainerFactory
}

// NewSovereignShardRequestersContainerFactory creates a new container filled with topic requesters for sovereign shards
func NewSovereignShardRequestersContainerFactory(args FactoryArgs) (*sovereignShardRequestersContainerFactory, error) {
	return &sovereignShardRequestersContainerFactory{}, nil
}

// Create returns a requesters container that will hold all sovereign requesters
func (srcf *sovereignShardRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *sovereignShardRequestersContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
