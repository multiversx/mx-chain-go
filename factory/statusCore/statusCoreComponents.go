package statusCore

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common/statistics"
	"github.com/ElrondNetwork/elrond-go/common/statistics/machine"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
)

// StatusCoreComponentsFactoryArgs holds the arguments needed for creating a status core components factory
type StatusCoreComponentsFactoryArgs struct {
	Config config.Config
}

// statusCoreComponentsFactory is responsible for creating the status core components
type statusCoreComponentsFactory struct {
	config config.Config
}

// statusCoreComponents is the DTO used for core components
type statusCoreComponents struct {
	resourceMonitor   factory.ResourceMonitor
	networkStatistics factory.NetworkStatisticsProvider
}

// NewStatusCoreComponentsFactory initializes the factory which is responsible to creating status core components
func NewStatusCoreComponentsFactory(args StatusCoreComponentsFactoryArgs) *statusCoreComponentsFactory {
	return &statusCoreComponentsFactory{
		config: args.Config,
	}
}

// Create creates the status core components
func (sccf *statusCoreComponentsFactory) Create() (*statusCoreComponents, error) {
	netStats := machine.NewNetStatistics()

	resourceMonitor, err := statistics.NewResourceMonitor(sccf.config, netStats)
	if err != nil {
		return nil, err
	}

	if sccf.config.ResourceStats.Enabled {
		resourceMonitor.StartMonitoring()
	}

	return &statusCoreComponents{
		resourceMonitor:   resourceMonitor,
		networkStatistics: netStats,
	}, nil
}

// Close closes all underlying components
func (scc *statusCoreComponents) Close() error {
	var errNetStats error
	var errResourceMonitor error
	if !check.IfNil(scc.networkStatistics) {
		errNetStats = scc.networkStatistics.Close()
	}
	if !check.IfNil(scc.resourceMonitor) {
		errResourceMonitor = scc.resourceMonitor.Close()
	}

	if errNetStats != nil {
		return errNetStats
	}
	return errResourceMonitor
}
