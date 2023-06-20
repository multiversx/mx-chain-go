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

const (
	errorOnMainNetworkString        = "on main network"
	errorOnFullArchiveNetworkString = "on full archive network"
)

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
	if check.IfNil(mnc.mainNetworkHolder.netMessenger) {
		return fmt.Errorf("%w %s", errors.ErrNilMessenger, errorOnMainNetworkString)
	}
	if check.IfNil(mnc.mainNetworkHolder.peersRatingHandler) {
		return fmt.Errorf("%w %s", errors.ErrNilPeersRatingHandler, errorOnMainNetworkString)
	}
	if check.IfNil(mnc.mainNetworkHolder.peersRatingMonitor) {
		return fmt.Errorf("%w %s", errors.ErrNilPeersRatingMonitor, errorOnMainNetworkString)
	}

	if check.IfNil(mnc.fullArchiveNetworkHolder.netMessenger) {
		return fmt.Errorf("%w %s", errors.ErrNilMessenger, errorOnFullArchiveNetworkString)
	}
	if check.IfNil(mnc.fullArchiveNetworkHolder.peersRatingHandler) {
		return fmt.Errorf("%w %s", errors.ErrNilPeersRatingHandler, errorOnFullArchiveNetworkString)
	}
	if check.IfNil(mnc.fullArchiveNetworkHolder.peersRatingMonitor) {
		return fmt.Errorf("%w %s", errors.ErrNilPeersRatingMonitor, errorOnFullArchiveNetworkString)
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

	return nil
}

// NetworkMessenger returns the p2p messenger of the main network
func (mnc *managedNetworkComponents) NetworkMessenger() p2p.Messenger {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.mainNetworkHolder.netMessenger
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

// PreferredPeersHolderHandler returns the preferred peers holder of the main network
func (mnc *managedNetworkComponents) PreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.mainNetworkHolder.preferredPeersHolder
}

// PeersRatingHandler returns the peers rating handler of the main network
func (mnc *managedNetworkComponents) PeersRatingHandler() p2p.PeersRatingHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.mainNetworkHolder.peersRatingHandler
}

// PeersRatingMonitor returns the peers rating monitor of the main network
func (mnc *managedNetworkComponents) PeersRatingMonitor() p2p.PeersRatingMonitor {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.mainNetworkHolder.peersRatingMonitor
}

// FullArchiveNetworkMessenger returns the p2p messenger of the full archive network
func (mnc *managedNetworkComponents) FullArchiveNetworkMessenger() p2p.Messenger {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.fullArchiveNetworkHolder.netMessenger
}

// FullArchivePeersRatingHandler returns the peers rating handler of the full archive network
func (mnc *managedNetworkComponents) FullArchivePeersRatingHandler() p2p.PeersRatingHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.fullArchiveNetworkHolder.peersRatingHandler
}

// FullArchivePeersRatingMonitor returns the peers rating monitor of the full archive network
func (mnc *managedNetworkComponents) FullArchivePeersRatingMonitor() p2p.PeersRatingMonitor {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.fullArchiveNetworkHolder.peersRatingMonitor
}

// FullArchivePreferredPeersHolderHandler returns the preferred peers holder of the full archive network
func (mnc *managedNetworkComponents) FullArchivePreferredPeersHolderHandler() factory.PreferredPeersHolderHandler {
	mnc.mutNetworkComponents.RLock()
	defer mnc.mutNetworkComponents.RUnlock()

	if mnc.networkComponents == nil {
		return nil
	}

	return mnc.fullArchiveNetworkHolder.preferredPeersHolder
}

// IsInterfaceNil returns true if the value under the interface is nil
func (mnc *managedNetworkComponents) IsInterfaceNil() bool {
	return mnc == nil
}

// String returns the name of the component
func (mnc *managedNetworkComponents) String() string {
	return factory.NetworkComponentsName
}
