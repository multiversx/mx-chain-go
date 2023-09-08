package interceptorscontainer

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

// IncomingHeaderSubscriber defines a subscriber to incoming headers
type IncomingHeaderSubscriber interface {
	AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error
	IsInterfaceNil() bool
}
