package network

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

var _ factory.ComponentHandler = (*managedNetworkComponents)(nil)
var _ factory.NetworkComponentsHolder = (*managedNetworkComponents)(nil)
var _ factory.NetworkComponentsHandler = (*managedNetworkComponents)(nil)

// managedNetworkComponents creates the data components handler that can create, close and access the data components
type managedNetworkComponents struct {
	*networkComponents
	networkComponentsFactory *networkComponentsFactory
	mutNetworkComponents     sync.RWMutex
}

// NewManagedNetworkComponents creates a new data components handler
func NewManagedNetworkComponents(ncf *networkComponentsFactory) (*managedNetworkComponents, error) {
	if ncf == nil {
		return nil, errors.ErrNilNetworkComponentsFactory
	}

	return &managedNetworkComponents{
		networkComponents:        nil,
		networkComponentsFactory: ncf,
	}, nil
}

// Create creates the network components
func (mnc *managedNetworkComponents) Create() error {
	nc, err := mnc.networkComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrNetworkComponentsFactoryCreate, err)
	}

	mnc.mutNetworkComponents.Lock()
	mnc.networkComponents = nc
	mnc.mutNetworkComponents.Unlock()

	return nil
}

// Close closes the network components
func (mnc *managedNetworkComponents) Close() error {
	mnc.mutNetworkComponents.Lock()
	defer mnc.mutNetworkComponents.Unlock()

	if mnc.networkComponents == nil {
		return nil
	}

	err := mnc.networkComponents.Close()
	if err != nil {
		return err
	}
	mnc.networkComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mnc *managedNetworkComponents) CheckSubcomponents() error {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return errors.ErrNilNetworkComponents
	}
	if check.IfNil(mnc.netMessenger) {
		return errors.ErrNilMessenger
	}
	if check.IfNil(mnc.inputAntifloodHandler) {
		return errors.ErrNilInputAntiFloodHandler
	}
	if check.IfNil(mnc.outputAntifloodHandler) {
		return errors.ErrNilOutputAntiFloodHandler
	}
	if check.IfNil(mnc.peerBlackListHandler) {
		return errors.ErrNilPeerBlackListHandler
	}
	if check.IfNil(mnc.peerHonestyHandler) {
		return errors.ErrNilPeerHonestyHandler
	}
	if check.IfNil(mnc.peersRatingHandler) {
		return errors.ErrNilPeersRatingHandler
	}
	if check.IfNil(mnc.peersRatingMonitor) {
		return errors.ErrNilPeersRatingMonitor
	}

	return nil
}

// NetworkMessenger returns the p2p messenger
func (mnc *managedNetworkComponents) NetworkMessenger() p2p.Messenger {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.netMessenger
}

// InputAntiFloodHandler returns the input p2p anti-flood handler
func (mnc *managedNetworkComponents) InputAntiFloodHandler() factory.P2PAntifloodHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.inputAntifloodHandler
}

// OutputAntiFloodHandler returns the output p2p anti-flood handler
func (mnc *managedNetworkComponents) OutputAntiFloodHandler() factory.P2PAntifloodHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.outputAntifloodHandler
}

// PubKeyCacher returns the public keys time cacher
func (mnc *managedNetworkComponents) PubKeyCacher() process.TimeCacher {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.pubKeyTimeCacher
}

// PeerBlackListHandler returns the blacklist handler
func (mnc *managedNetworkComponents) PeerBlackListHandler() process.PeerBlackListCacher {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.networkComponents.peerBlackListHandler
}

// PeerHonestyHandler returns the blacklist handler
func (mnc *managedNetworkComponents) PeerHonestyHandler() factory.PeerHonestyHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.networkComponents.peerHonestyHandler
}

// PreferredPeersHolderHandler returns the preferred peers holder
func (mnc *managedNetworkComponents) PreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.networkComponents.peersHolder
}

// PeersRatingHandler returns the peers rating handler
func (mnc *managedNetworkComponents) PeersRatingHandler() p2p.PeersRatingHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.networkComponents.peersRatingHandler
}

// PeersRatingMonitor returns the peers rating monitor
func (mnc *managedNetworkComponents) PeersRatingMonitor() p2p.PeersRatingMonitor {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.networkComponents.peersRatingMonitor
}

// IsInterfaceNil returns true if the value under the interface is nil
func (mnc *managedNetworkComponents) IsInterfaceNil() bool {
	return mnc == nil
}

// String returns the name of the component
func (mnc *managedNetworkComponents) String() string {
	return factory.NetworkComponentsName
}
