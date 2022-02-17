package statusHandler

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// StatusHandlersUtilsMock -
type StatusHandlersUtilsMock struct {
	StatusMetrics    external.StatusMetricsHandler
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
func (shum *StatusHandlersUtilsMock) Metrics() external.StatusMetricsHandler {
	return shum.StatusMetrics
}

// IsInterfaceNil -
func (shum *StatusHandlersUtilsMock) IsInterfaceNil() bool {
	return shum == nil
}
