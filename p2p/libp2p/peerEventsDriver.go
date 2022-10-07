package libp2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type peerEventsDriver struct {
	mutHandlers sync.RWMutex
	handlers    []p2p.PeerEventsHandler
}

// NewPeerEventsDriver will create a new peer events driver
func NewPeerEventsDriver() *peerEventsDriver {
	return &peerEventsDriver{
		handlers: make([]p2p.PeerEventsHandler, 0),
	}
}

// AddHandler will add a new handler to the internal list
func (driver *peerEventsDriver) AddHandler(handler p2p.PeerEventsHandler) error {
	if check.IfNil(handler) {
		return p2p.ErrNilPeerEventsHandler
	}

	driver.mutHandlers.Lock()
	driver.handlers = append(driver.handlers, handler)
	driver.mutHandlers.Unlock()

	return nil
}

// Connected calls the Connected methods on all registered handlers
func (driver *peerEventsDriver) Connected(pid core.PeerID, connection string) {
	driver.mutHandlers.RLock()
	defer driver.mutHandlers.RUnlock()

	for _, h := range driver.handlers {
		h.Connected(pid, connection)
	}
}

// Disconnected calls the Disconnected methods on all registered handlers
func (driver *peerEventsDriver) Disconnected(pid core.PeerID) {
	driver.mutHandlers.RLock()
	defer driver.mutHandlers.RUnlock()

	for _, h := range driver.handlers {
		h.Disconnected(pid)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (driver *peerEventsDriver) IsInterfaceNil() bool {
	return driver == nil
}
