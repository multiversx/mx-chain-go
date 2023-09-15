package sovereign

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

type IncomingHeaderSubscriberStub struct {
	AddHeaderCalled func(headerHash []byte, header sovereign.IncomingHeaderHandler) error
}

func (ihs *IncomingHeaderSubscriberStub) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	if ihs.AddHeaderCalled != nil {
		return ihs.AddHeaderCalled(headerHash, header)
	}

	return nil
}

func (ihs *IncomingHeaderSubscriberStub) IsInterfaceNil() bool {
	return ihs == nil
}
