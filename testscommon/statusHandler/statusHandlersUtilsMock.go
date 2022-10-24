package statusHandler

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// StatusHandlersUtilsMock -
type StatusHandlersUtilsMock struct {
	StatusMetrics    statusHandler.StatusMetricsHandler
	AppStatusHandler core.AppStatusHandler
}

// UpdateStorerAndMetricsForPersistentHandler -
func (shum *StatusHandlersUtilsMock) UpdateStorerAndMetricsForPersistentHandler(_ storage.Storer) error {
	return nil
}

// StatusHandler -
func (shum *StatusHandlersUtilsMock) StatusHandler() core.AppStatusHandler {
	return shum.AppStatusHandler
}

// Metrics -
func (shum *StatusHandlersUtilsMock) Metrics() statusHandler.StatusMetricsHandler {
	return shum.StatusMetrics
}

// IsInterfaceNil -
func (shum *StatusHandlersUtilsMock) IsInterfaceNil() bool {
	return shum == nil
}
