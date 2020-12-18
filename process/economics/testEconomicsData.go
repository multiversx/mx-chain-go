package economics

import "math/big"

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
func (ted *TestEconomicsData) SetMaxGasLimitPerBlock(maxGasLimitPerBlock uint64) {
	ted.maxGasLimitPerBlock = maxGasLimitPerBlock
	ted.maxGasLimitPerMetaBlock = maxGasLimitPerBlock
}

// SetMinGasPrice sets the minimum gas price for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasPrice(minGasPrice uint64) {
	ted.minGasPrice = minGasPrice
}

// SetMinGasLimit sets the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasLimit(minGasLimit uint64) {
	ted.minGasLimit = minGasLimit
}

// GetMinGasLimit returns the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) GetMinGasLimit() uint64 {
	return ted.minGasLimit
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
