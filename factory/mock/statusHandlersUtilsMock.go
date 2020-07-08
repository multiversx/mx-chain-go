package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type StatusHandlersUtilsMock struct {
}

func (shum *StatusHandlersUtilsMock) UpdateStorerAndMetricsForPersistentHandler(_ storage.Storer) error {
	return nil
}

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

func (shum *StatusHandlersUtilsMock) IsInterfaceNil() bool {
	return shum == nil
}
