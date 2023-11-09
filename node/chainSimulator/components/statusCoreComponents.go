package components

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/node/external"
)

type statusCoreComponentsHolder struct {
	closeHandler               *closeHandler
	resourceMonitor            factory.ResourceMonitor
	networkStatisticsProvider  factory.NetworkStatisticsProvider
	trieSyncStatisticsProvider factory.TrieSyncStatisticsProvider
	statusHandler              core.AppStatusHandler
	statusMetrics              external.StatusMetricsHandler
	persistentStatusHandler    factory.PersistentStatusHandler
}

// CreateStatusCoreComponentsHolder will create a new instance of factory.StatusCoreComponentsHolder
func CreateStatusCoreComponentsHolder(configs config.Configs, coreComponents factory.CoreComponentsHolder) (factory.StatusCoreComponentsHandler, error) {
	var err error

	statusCoreComponentsFactory, err := statusCore.NewStatusCoreComponentsFactory(statusCore.StatusCoreComponentsFactoryArgs{
		Config:          *configs.GeneralConfig,
		EpochConfig:     *configs.EpochConfig,
		RoundConfig:     *configs.RoundConfig,
		RatingsConfig:   *configs.RatingsConfig,
		EconomicsConfig: *configs.EconomicsConfig,
		CoreComp:        coreComponents,
	})
	if err != nil {
		return nil, err
	}

	managedStatusCoreComponents, err := statusCore.NewManagedStatusCoreComponents(statusCoreComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedStatusCoreComponents.Create()
	if err != nil {
		return nil, err
	}

	instance := &statusCoreComponentsHolder{
		closeHandler:               NewCloseHandler(),
		resourceMonitor:            managedStatusCoreComponents.ResourceMonitor(),
		networkStatisticsProvider:  managedStatusCoreComponents.NetworkStatistics(),
		trieSyncStatisticsProvider: managedStatusCoreComponents.TrieSyncStatistics(),
		statusHandler:              managedStatusCoreComponents.AppStatusHandler(),
		statusMetrics:              managedStatusCoreComponents.StatusMetrics(),
		persistentStatusHandler:    managedStatusCoreComponents.PersistentStatusHandler(),
	}

	instance.collectClosableComponents()

	return instance, nil
}

// ResourceMonitor will return the resource monitor
func (s *statusCoreComponentsHolder) ResourceMonitor() factory.ResourceMonitor {
	return s.resourceMonitor
}

// NetworkStatistics will return the network statistics provider
func (s *statusCoreComponentsHolder) NetworkStatistics() factory.NetworkStatisticsProvider {
	return s.networkStatisticsProvider
}

// TrieSyncStatistics will return trie sync statistics provider
func (s *statusCoreComponentsHolder) TrieSyncStatistics() factory.TrieSyncStatisticsProvider {
	return s.trieSyncStatisticsProvider
}

// AppStatusHandler will return the status handler
func (s *statusCoreComponentsHolder) AppStatusHandler() core.AppStatusHandler {
	return s.statusHandler
}

// StatusMetrics will return the status metrics handler
func (s *statusCoreComponentsHolder) StatusMetrics() external.StatusMetricsHandler {
	return s.statusMetrics
}

// PersistentStatusHandler will return the persistent status handler
func (s *statusCoreComponentsHolder) PersistentStatusHandler() factory.PersistentStatusHandler {
	return s.persistentStatusHandler
}

func (s *statusCoreComponentsHolder) collectClosableComponents() {
	s.closeHandler.AddComponent(s.resourceMonitor)
	s.closeHandler.AddComponent(s.networkStatisticsProvider)
	s.closeHandler.AddComponent(s.statusHandler)
	s.closeHandler.AddComponent(s.persistentStatusHandler)
}

// Close will call the Close methods on all inner components
func (s *statusCoreComponentsHolder) Close() error {
	return s.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *statusCoreComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}

// Create will do nothing
func (s *statusCoreComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (s *statusCoreComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (s *statusCoreComponentsHolder) String() string {
	return ""
}
