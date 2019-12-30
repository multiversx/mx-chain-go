package economics

import "math/big"

// TestEconomicsData extends EconomicsData and is used in integration tests as it exposes some functions
// that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestEconomicsData struct {
	*EconomicsData
}

// SetMaxGasLimitPerBlock sets the maximum gas limit allowed per one block
func (ted *TestEconomicsData) SetMaxGasLimitPerBlock(maxGasLimitPerBlock uint64) {
	ted.maxGasLimitPerBlock = maxGasLimitPerBlock
}

// SetMinGasPrice sets the minimum gas price for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasPrice(minGasPrice uint64) {
	ted.minGasPrice = minGasPrice
}

// SetMinGasLimit sets the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasLimit(minGasLimit uint64) {
	ted.minGasLimit = minGasLimit
}

// SetMinGasLimit sets the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) GetMinGasLimit() uint64 {
	return ted.minGasLimit
}

// SetRewards sets the new reward value
func (ted *TestEconomicsData) SetRewards(value *big.Int) {
	ted.rewardsValue = value
}

// GetMinGasPrice returns the current min gas price
func (ted *TestEconomicsData) GetMinGasPrice() uint64 {
	return ted.minGasPrice
}

// SetGasPerDataByte sets gas per data byte for a transaction to be accepted
func (ted *TestEconomicsData) SetGasPerDataByte(gasPerDataByte uint64) {
	ted.gasPerDataByte = gasPerDataByte
}

// SetDataLimitForBaseCalc sets base calc limit for gasLimit calculation
func (ted *TestEconomicsData) SetDataLimitForBaseCalc(dataLimitForBaseCalc uint64) {
	ted.dataLimitForBaseCalc = dataLimitForBaseCalc
}
