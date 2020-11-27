package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// StatusHandlersUtilsMock -
type StatusHandlersUtilsMock struct {
	AppStatusHandler core.AppStatusHandler
}

// UpdateStorerAndMetricsForPersistentHandler -
func (shum *StatusHandlersUtilsMock) UpdateStorerAndMetricsForPersistentHandler(_ storage.Storer) error {
	return nil
}

// LoadTpsBenchmarkFromStorage -
func (shum *StatusHandlersUtilsMock) LoadTpsBenchmarkFromStorage(_ storage.Storer, _ marshal.Marshalizer) *statistics.TpsPersistentData {
	return &statistics.TpsPersistentData{
		BlockNumber:           1,
		RoundNumber:           1,
		PeakTPS:               0,
		AverageBlockTxCount:   big.NewInt(0),
		TotalProcessedTxCount: big.NewInt(0),
		LastBlockTxCount:      0,
	}
}

// StatusHandler -
func (shum *StatusHandlersUtilsMock) StatusHandler() core.AppStatusHandler {
	return shum.AppStatusHandler
}

// Metrics -
func (shum *StatusHandlersUtilsMock) Metrics() external.StatusMetricsHandler {
	return nil
}

// SignalStartViews -
func (shum *StatusHandlersUtilsMock) SignalStartViews() {

}

// SignalLogRewrite -
func (shum *StatusHandlersUtilsMock) SignalLogRewrite() {

}

// IsInterfaceNil -
func (shum *StatusHandlersUtilsMock) IsInterfaceNil() bool {
	return shum == nil
}
