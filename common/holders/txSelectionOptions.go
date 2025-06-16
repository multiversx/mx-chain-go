package holders

type txSelectionOptions struct {
	gasRequested              uint64
	maxNumTxs                 int
	loopMaximumDuration       int
	loopDurationCheckInterval int
}

// NewTxSelectionOptions returns a new instance of a selectionOptions struct
func NewTxSelectionOptions(gasRequested uint64, maxNumTxs int, loopMaximumDuration int, loopDurationCheckInterval int) *txSelectionOptions {
	return &txSelectionOptions{
		gasRequested:              gasRequested,
		maxNumTxs:                 maxNumTxs,
		loopMaximumDuration:       loopMaximumDuration,
		loopDurationCheckInterval: loopDurationCheckInterval,
	}
}

// GetGasRequested returns a selection constraint parameter (for gas)
func (options *txSelectionOptions) GetGasRequested() uint64 {
	return options.gasRequested
}

// GetMaxNumTxs returns a selection constraint parameter (for number of transactions)
func (options *txSelectionOptions) GetMaxNumTxs() int {
	return options.maxNumTxs
}

// GetLoopMaximumDurationMs returns a selection constraint parameter (related to selection duration)
func (options *txSelectionOptions) GetLoopMaximumDurationMs() int {
	return options.loopMaximumDuration
}

// GetLoopDurationCheckInterval returns a selection constraint parameter (related to selection duration)
func (options *txSelectionOptions) GetLoopDurationCheckInterval() int {
	return options.loopDurationCheckInterval
}

// IsInterfaceNil returns true if there is no value under the interface
func (options *txSelectionOptions) IsInterfaceNil() bool {
	return options == nil
}
