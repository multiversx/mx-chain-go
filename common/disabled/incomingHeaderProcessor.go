package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type incomingHeaderProcessor struct {
}

func NewDisabledIncomingHeaderProcessor() *incomingHeaderProcessor {
	return &incomingHeaderProcessor{}
}

// AddHeader return nothing
func (ihp *incomingHeaderProcessor) AddHeader(_ []byte, _ sovereign.IncomingHeaderHandler) error {
	return nil
}

// CreateExtendedHeader return nothing
func (ihp *incomingHeaderProcessor) CreateExtendedHeader(_ sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	return &block.ShardHeaderExtended{}, nil
}

// IsInterfaceNil - returns true if there is no value under the interface
func (ihp *incomingHeaderProcessor) IsInterfaceNil() bool {
	return ihp == nil
}
