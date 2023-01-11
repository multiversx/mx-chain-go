package economics

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

type apiEconomicsData struct {
	*economicsData
}

// NewAPIEconomicsData will create a wrapped object over the provided economics data
func NewAPIEconomicsData(data *economicsData) (*apiEconomicsData, error) {
	if data == nil {
		return nil, process.ErrNilEconomicsData
	}

	return &apiEconomicsData{
		economicsData: data,
	}, nil
}

// CheckValidityTxValues checks if the provided transaction is economically correct. It overloads the original method
// as the check in this instance for the gas limit should be done on the MaxGasLimitPerTx value
func (ed *apiEconomicsData) CheckValidityTxValues(tx data.TransactionWithFeeHandler) error {
	if tx.GetGasLimit() > ed.MaxGasLimitPerTx() {
		return process.ErrMoreGasThanGasLimitPerMiniBlockForSafeCrossShard
	}

	return ed.economicsData.CheckValidityTxValues(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *apiEconomicsData) IsInterfaceNil() bool {
	return ed == nil
}
