package statusCore

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/external"
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
		return nil, errors.ErrNilStatusCoreComponentsFactory
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
	if check.IfNil(mscc.appStatusHandler) {
		return errors.ErrNilAppStatusHandler
	}
	if check.IfNil(mscc.statusMetrics) {
		return errors.ErrNilStatusMetrics
	}
	if check.IfNil(mscc.persistentHandler) {
		return errors.ErrNilPersistentHandler
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

// AppStatusHandler returns the app status handler instance
func (mscc *managedStatusCoreComponents) AppStatusHandler() core.AppStatusHandler {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	return mscc.statusCoreComponents.appStatusHandler
}

// StatusMetrics returns the status metrics instance
func (mscc *managedStatusCoreComponents) StatusMetrics() external.StatusMetricsHandler {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	return mscc.statusCoreComponents.statusMetrics
}

// PersistentStatusHandler returns the persistent handler instance
func (mscc *managedStatusCoreComponents) PersistentStatusHandler() factory.PersistentStatusHandler {
	mscc.mutCoreComponents.RLock()
	defer mscc.mutCoreComponents.RUnlock()

	if mscc.statusCoreComponents == nil {
		return nil
	}

	return mscc.statusCoreComponents.persistentHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (mscc *managedStatusCoreComponents) IsInterfaceNil() bool {
	return mscc == nil
}

// String returns the name of the component
func (mscc *managedStatusCoreComponents) String() string {
	return factory.StatusCoreComponentsName
}
