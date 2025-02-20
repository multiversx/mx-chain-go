package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/statusHandler/persister"
)

// StatusCoreComponentsStub -
type StatusCoreComponentsStub struct {
	ResourceMonitorField         statistics.ResourceMonitorHandler
	NetworkStatisticsField       statistics.NetworkStatisticsProvider
	TrieSyncStatisticsField      heartbeat.TrieSyncStatisticsProvider
	AppStatusHandlerField        core.AppStatusHandler
	AppStatusHandlerCalled       func() core.AppStatusHandler
	StatusMetricsField           external.StatusMetricsHandler
	PersistentStatusHandlerField persister.PersistentStatusHandler
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
func (stub *StatusCoreComponentsStub) ResourceMonitor() statistics.ResourceMonitorHandler {
	return stub.ResourceMonitorField
}

// NetworkStatistics -
func (stub *StatusCoreComponentsStub) NetworkStatistics() statistics.NetworkStatisticsProvider {
	return stub.NetworkStatisticsField
}

// TrieSyncStatistics -
func (stub *StatusCoreComponentsStub) TrieSyncStatistics() heartbeat.TrieSyncStatisticsProvider {
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
func (stub *StatusCoreComponentsStub) PersistentStatusHandler() persister.PersistentStatusHandler {
	return stub.PersistentStatusHandlerField
}

// IsInterfaceNil -
func (stub *StatusCoreComponentsStub) IsInterfaceNil() bool {
	return stub == nil
}
