package economics

// TestEconomicsData extends EconomicsData and is used in integration tests as it exposes some functions
// that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestEconomicsData struct {
	*EconomicsData
}

// SetMaxGasLimitPerMiniBlock sets the maximum gas limit allowed per one mini block
func (ted *TestEconomicsData) SetMaxGasLimitPerMiniBlock(maxGasLimitPerMiniBlock uint64) {
	ted.maxGasLimitPerMiniBlock = maxGasLimitPerMiniBlock
}

// SetMinGasPrice sets the minimum gas price for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasPrice(minGasPrice uint64) {
	ted.minGasPrice = minGasPrice
}

// SetMinGasLimit sets the minimum gas limit for a transaction to be accepted
func (ted *TestEconomicsData) SetMinGasLimit(minGasLimit uint64) {
	ted.minGasLimit = minGasLimit
}
