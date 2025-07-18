package economics

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
)

// TestEconomicsData extends EconomicsData and is used in integration tests as it exposes some functions
// that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestEconomicsData struct {
	*economicsData
}

// NewTestEconomicsData -
func NewTestEconomicsData(internalData *economicsData) *TestEconomicsData {
	return &TestEconomicsData{economicsData: internalData}
}

// SetMaxGasLimitPerBlock sets the maximum gas limit allowed per one block
func (ted *TestEconomicsData) SetMaxGasLimitPerBlock(maxGasLimitPerBlock uint64, epoch uint32) {
	gc := ted.getGasConfigForEpoch(epoch)
	gc.maxGasLimitPerBlock = maxGasLimitPerBlock
	gc.maxGasLimitPerMiniBlock = maxGasLimitPerBlock
	gc.maxGasLimitPerMetaBlock = maxGasLimitPerBlock
	gc.maxGasLimitPerMetaMiniBlock = maxGasLimitPerBlock
	gc.maxGasLimitPerTx = maxGasLimitPerBlock
}

// SetMinGasPrice sets the minimum gas price for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasPrice(minGasPrice uint64) {
	ted.minGasPrice = minGasPrice
}

// SetMinGasLimit sets the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasLimit(minGasLimit uint64, epoch uint32) {
	gc := ted.getGasConfigForEpoch(epoch)
	gc.minGasLimit = minGasLimit
}

// GetMinGasLimit returns the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) GetMinGasLimit(epoch uint32) uint64 {
	gc := ted.getGasConfigForEpoch(epoch)
	return gc.minGasLimit
}

// GetMinGasPrice returns the current min gas price
func (ted *TestEconomicsData) GetMinGasPrice() uint64 {
	return ted.minGasPrice
}

// SetGasPerDataByte sets gas per data byte for a transaction to be accepted
func (ted *TestEconomicsData) SetGasPerDataByte(gasPerDataByte uint64) {
	ted.gasPerDataByte = gasPerDataByte
}

// SetTotalSupply sets the total supply when booting the network
func (ted *TestEconomicsData) SetTotalSupply(totalSupply *big.Int) {
	ted.genesisTotalSupply = totalSupply
}

// SetMaxInflationRate sets the maximum inflation rate for a transaction to be accepted
func (ted *TestEconomicsData) SetMaxInflationRate(maximumInflation float64) {
	ted.mutYearSettings.Lock()
	defer ted.mutYearSettings.Unlock()

	ted.yearSettings[0].MaximumInflation = maximumInflation
}

// MaxGasHigherFactorAccepted returns 10
func (ted *TestEconomicsData) MaxGasHigherFactorAccepted() uint64 {
	return 10
}

// SetStatusHandler returns nil
func (ted *TestEconomicsData) SetStatusHandler(_ core.AppStatusHandler) error {
	return nil
}
