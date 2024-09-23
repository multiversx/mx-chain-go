package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// IncomingHeaderSubscriber -
type IncomingHeaderSubscriber struct {
}

// AddHeader does nothing
func (ihs *IncomingHeaderSubscriber) AddHeader(_ []byte, _ sovereign.IncomingHeaderHandler) error {
	return nil
}

// CreateExtendedHeader does nothing
func (ihs *IncomingHeaderSubscriber) CreateExtendedHeader(_ sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	return nil, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihs *IncomingHeaderSubscriber) IsInterfaceNil() bool {
	return ihs == nil
}
