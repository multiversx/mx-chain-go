package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/external"
)

// StatusCoreComponentsStub -
type StatusCoreComponentsStub struct {
	ResourceMonitorField         factory.ResourceMonitor
	NetworkStatisticsField       factory.NetworkStatisticsProvider
	TrieSyncStatisticsField      factory.TrieSyncStatisticsProvider
	AppStatusHandlerField        core.AppStatusHandler
	AppStatusHandlerCalled       func() core.AppStatusHandler
	StatusMetricsField           external.StatusMetricsHandler
	PersistentStatusHandlerField factory.PersistentStatusHandler
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
	return factory.StatusCoreComponentsName
}

// ResourceMonitor -
func (stub *StatusCoreComponentsStub) ResourceMonitor() factory.ResourceMonitor {
	return stub.ResourceMonitorField
}

// NetworkStatistics -
func (stub *StatusCoreComponentsStub) NetworkStatistics() factory.NetworkStatisticsProvider {
	return stub.NetworkStatisticsField
}

// TrieSyncStatistics -
func (stub *StatusCoreComponentsStub) TrieSyncStatistics() factory.TrieSyncStatisticsProvider {
	return stub.TrieSyncStatisticsField
}

// AppStatusHandler -
func (stub *StatusCoreComponentsStub) AppStatusHandler() core.AppStatusHandler {
	if stub.AppStatusHandlerCalled != nil {
		return stub.AppStatusHandlerCalled()
	}
	return stub.AppStatusHandlerField
}

// StatusMetrics -
func (stub *StatusCoreComponentsStub) StatusMetrics() external.StatusMetricsHandler {
	return stub.StatusMetricsField
}

// PersistentStatusHandler -
func (stub *StatusCoreComponentsStub) PersistentStatusHandler() factory.PersistentStatusHandler {
	return stub.PersistentStatusHandlerField
}

// IsInterfaceNil -
func (stub *StatusCoreComponentsStub) IsInterfaceNil() bool {
	return stub == nil
}
