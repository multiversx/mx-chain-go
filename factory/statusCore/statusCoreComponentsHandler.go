package statusCore

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
)

var _ factory.ComponentHandler = (*managedStatusCoreComponents)(nil)
var _ factory.StatusCoreComponentsHolder = (*managedStatusCoreComponents)(nil)
var _ factory.StatusCoreComponentsHandler = (*managedStatusCoreComponents)(nil)

// managedStatusCoreComponents is an implementation of status core components handler that can create, close and access the status core components
type managedStatusCoreComponents struct {
	statusCoreComponentsFactory *statusCoreComponentsFactory
	*statusCoreComponents
	mutCoreComponents sync.RWMutex
}

// NewManagedStatusCoreComponents creates a new status core components handler implementation
func NewManagedStatusCoreComponents(sccf *statusCoreComponentsFactory) (*managedStatusCoreComponents, error) {
	if sccf == nil {
		return nil, errors.ErrNilCoreComponentsFactory
	}

	mcc := &managedStatusCoreComponents{
		statusCoreComponents:        nil,
		statusCoreComponentsFactory: sccf,
	}
	return mcc, nil
}

// Create creates the core components
func (mscc *managedStatusCoreComponents) Create() error {
	scc, err := mscc.statusCoreComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrStatusCoreComponentsFactoryCreate, err)
	}

	mscc.mutCoreComponents.Lock()
	mscc.statusCoreComponents = scc
	mscc.mutCoreComponents.Unlock()

	return nil
}

// Close closes the managed core components
func (mscc *managedStatusCoreComponents) Close() error {
	mscc.mutCoreComponents.Lock()
	defer mscc.mutCoreComponents.Unlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	err := mscc.statusCoreComponents.Close()
	if err != nil {
		return err
	}
	mscc.statusCoreComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mscc *managedStatusCoreComponents) CheckSubcomponents() error {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return errors.ErrNilStatusCoreComponents
	}
	if check.IfNil(mscc.networkStatistics) {
		return errors.ErrNilNetworkStatistics
	}
	if check.IfNil(mscc.resourceMonitor) {
		return errors.ErrNilResourceMonitor
	}
	if check.IfNil(mscc.trieSyncStatistics) {
		return errors.ErrNilTrieSyncStatistics
	}

	return nil
}

// NetworkStatistics returns the network statistics instance
func (mscc *managedStatusCoreComponents) NetworkStatistics() factory.NetworkStatisticsProvider {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	return mscc.statusCoreComponents.networkStatistics
}

// ResourceMonitor returns the resource monitor instance
func (mscc *managedStatusCoreComponents) ResourceMonitor() factory.ResourceMonitor {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	return mscc.statusCoreComponents.resourceMonitor
}

// TrieSyncStatistics returns the trie sync statistics instance
func (mscc *managedStatusCoreComponents) TrieSyncStatistics() factory.TrieSyncStatisticsProvider {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	return mscc.statusCoreComponents.trieSyncStatistics
}

// IsInterfaceNil returns true if there is no value under the interface
func (mscc *managedStatusCoreComponents) IsInterfaceNil() bool {
	return mscc == nil
}

// String returns the name of the component
func (mscc *managedStatusCoreComponents) String() string {
	return factory.StatusCoreComponentsName
}
