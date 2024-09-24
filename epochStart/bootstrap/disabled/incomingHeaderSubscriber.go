package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type incomingHeaderSubscriber struct {
}

// NewIncomingHeaderSubscriber creates a disabled incoming header subscriber
func NewIncomingHeaderSubscriber() *incomingHeaderSubscriber {
	return &incomingHeaderSubscriber{}
}

// AddHeader does nothing
func (ihs *incomingHeaderSubscriber) AddHeader(_ []byte, _ sovereign.IncomingHeaderHandler) error {
	return nil
}

// CreateExtendedHeader returns an empty shard extended header
func (ihs *incomingHeaderSubscriber) CreateExtendedHeader(_ sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	return &block.ShardHeaderExtended{}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihs *incomingHeaderSubscriber) IsInterfaceNil() bool {
	return ihs == nil
}
