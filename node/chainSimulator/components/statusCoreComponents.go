package components

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/cmd/termui/presenter"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/common/statistics/machine"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/statusHandler/persister"
	statisticsTrie "github.com/multiversx/mx-chain-go/trie/statistics"
)

type statusCoreComponentsHolder struct {
	resourceMonitor            factory.ResourceMonitor
	networkStatisticsProvider  factory.NetworkStatisticsProvider
	trieSyncStatisticsProvider factory.TrieSyncStatisticsProvider
	statusHandler              core.AppStatusHandler
	statusMetrics              external.StatusMetricsHandler
	persistentStatusHandler    factory.PersistentStatusHandler
}

// CreateStatusCoreComponentsHolder will create a new instance of factory.StatusCoreComponentsHolder
func CreateStatusCoreComponentsHolder(cfg config.Config, coreComponents factory.CoreComponentsHolder) (factory.StatusCoreComponentsHolder, error) {
	var err error
	instance := &statusCoreComponentsHolder{
		networkStatisticsProvider:  machine.NewNetStatistics(),
		trieSyncStatisticsProvider: statisticsTrie.NewTrieSyncStatistics(),
		statusHandler:              presenter.NewPresenterStatusHandler(),
		statusMetrics:              statusHandler.NewStatusMetrics(),
	}

	instance.resourceMonitor, err = statistics.NewResourceMonitor(cfg, instance.networkStatisticsProvider)
	if err != nil {
		return nil, err
	}
	instance.persistentStatusHandler, err = persister.NewPersistentStatusHandler(coreComponents.InternalMarshalizer(), coreComponents.Uint64ByteSliceConverter())
	if err != nil {
		return nil, err
	}

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

// IsInterfaceNil returns true if there is no value under the interface
func (s *statusCoreComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}
