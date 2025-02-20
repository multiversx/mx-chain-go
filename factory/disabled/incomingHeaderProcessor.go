package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// IncomingHeaderProcessor is a disabled incoming header processor
type IncomingHeaderProcessor struct {
}

// AddHeader does nothing
func (ihp *IncomingHeaderProcessor) AddHeader(_ []byte, _ sovereign.IncomingHeaderHandler) error {
	return nil
}

// CreateExtendedHeader returns an empty extended shard header
func (ihp *IncomingHeaderProcessor) CreateExtendedHeader(_ sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	return &block.ShardHeaderExtended{Header: &block.HeaderV2{}}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihp *IncomingHeaderProcessor) IsInterfaceNil() bool {
	return ihp == nil
}
