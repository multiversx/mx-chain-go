package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/node/external"
)

// StatusCoreComponentsStub -
type StatusCoreComponentsStub struct {
	ResourceMonitorField         factory.ResourceMonitor
	NetworkStatisticsField       factory.NetworkStatisticsProvider
	AppStatusHandlerField        core.AppStatusHandler
	StatusMetricsField           external.StatusMetricsHandler
	PersistentStatusHandlerField factory.PersistenStatusHandler
}

// Create -
func (stub *StatusCoreComponentsStub) Create() error {
	return nil
}

// Close -
func (stub *StatusCoreComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (stub *StatusCoreComponentsStub) CheckSubcomponents() error {
	return nil
}

// String -
func (stub *StatusCoreComponentsStub) String() string {
	return ""
}

// ResourceMonitor -
func (stub *StatusCoreComponentsStub) ResourceMonitor() factory.ResourceMonitor {
	return stub.ResourceMonitorField
}

// NetworkStatistics -
func (stub *StatusCoreComponentsStub) NetworkStatistics() factory.NetworkStatisticsProvider {
	return stub.NetworkStatisticsField
}

// AppStatusHandler -
func (stub *StatusCoreComponentsStub) AppStatusHandler() core.AppStatusHandler {
	return stub.AppStatusHandlerField
}

// StatusMetrics -
func (stub *StatusCoreComponentsStub) StatusMetrics() external.StatusMetricsHandler {
	return stub.StatusMetricsField
}

// PersistentStatusHandler -
func (stub *StatusCoreComponentsStub) PersistentStatusHandler() factory.PersistenStatusHandler {
	return stub.PersistentStatusHandlerField
}

// IsInterfaceNil -
func (stub *StatusCoreComponentsStub) IsInterfaceNil() bool {
	return stub == nil
}
