package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

type sovereignTotalStakedValueProcessorFactory struct {
}

// NewSovereignTotalStakedValueProcessorFactory create a new sovereign total staked value handler
func NewSovereignTotalStakedValueProcessorFactory() *sovereignTotalStakedValueProcessorFactory {
	return &sovereignTotalStakedValueProcessorFactory{}
}

// CreateTotalStakedValueProcessorHandler will create a new instance of total staked value processor for sovereign chain
func (d *sovereignTotalStakedValueProcessorFactory) CreateTotalStakedValueProcessorHandler(args trieIterators.ArgTrieIteratorProcessor) (external.TotalStakedValueHandler, error) {
	return trieIterators.NewTotalStakedValueProcessor(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (d *sovereignTotalStakedValueProcessorFactory) IsInterfaceNil() bool {
	return d == nil
}
