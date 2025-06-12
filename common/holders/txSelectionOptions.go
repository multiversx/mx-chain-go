package holders

type txSelectionOptions struct {
	gasRequested              int
	maxNumTxs                 int
	loopMaximumDuration       int
	loopDurationCheckInterval int
}

// NewTxSelectionOptions returns a new instance of a selectionOptions struct
func NewTxSelectionOptions(gasRequested int, maxNumTxs int, loopMaximumDuration int, loopDurationCheckInterval int) *txSelectionOptions {
	return &txSelectionOptions{
		gasRequested:              gasRequested,
		maxNumTxs:                 maxNumTxs,
		loopMaximumDuration:       loopMaximumDuration,
		loopDurationCheckInterval: loopDurationCheckInterval,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (options *txSelectionOptions) IsInterfaceNil() bool {
	return options == nil
}
