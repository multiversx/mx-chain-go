package sovereign

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// IncomingHeaderSubscriberStub -
type IncomingHeaderSubscriberStub struct {
	AddHeaderCalled            func(headerHash []byte, header sovereign.IncomingHeaderHandler) error
	CreateExtendedHeaderCalled func(header sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error)
}

// AddHeader -
func (ihs *IncomingHeaderSubscriberStub) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	if ihs.AddHeaderCalled != nil {
		return ihs.AddHeaderCalled(headerHash, header)
	}

	return nil
}

// CreateExtendedHeader -
func (ihs *IncomingHeaderSubscriberStub) CreateExtendedHeader(header sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	if ihs.CreateExtendedHeaderCalled != nil {
		return ihs.CreateExtendedHeaderCalled(header)
	}

	return nil, nil
}

// IsInterfaceNil -
func (ihs *IncomingHeaderSubscriberStub) IsInterfaceNil() bool {
	return ihs == nil
}
