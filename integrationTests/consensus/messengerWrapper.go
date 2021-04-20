package consensus

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

//TODO refactor consensus node so this wrapper will not be required
type messengerWrapper struct {
	p2p.Messenger
}

// ConnectTo will try to initiate a connection to the provided parameter
func (mw *messengerWrapper) ConnectTo(connectable integrationTests.Connectable) error {
	if check.IfNil(connectable) {
		return fmt.Errorf("trying to connect to a nil Connectable parameter")
	}

	return mw.ConnectToPeer(connectable.GetConnectableAddress())
}

// GetConnectableAddress returns a non circuit, non windows default connectable p2p address
func (mw *messengerWrapper) GetConnectableAddress() string {
	if mw == nil {
		return "nil"
	}

	return integrationTests.GetConnectableAddress(mw)
}
