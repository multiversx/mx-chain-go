package holders

type txSelectionOptions struct {
	gasRequested              uint64
	maxNumTxs                 int
	loopDurationCheckInterval int
	haveTimeForSelection      func() bool
}

// NewTxSelectionOptions returns a new instance of a selectionOptions struct
func NewTxSelectionOptions(gasRequested uint64, maxNumTxs int, loopDurationCheckInterval int, haveTimeForSelection func() bool) (*txSelectionOptions, error) {
	if haveTimeForSelection == nil {
		return nil, errNilHaveTimeForSelectionFunc
	}
	return &txSelectionOptions{
		gasRequested:              gasRequested,
		maxNumTxs:                 maxNumTxs,
		haveTimeForSelection:      haveTimeForSelection,
		loopDurationCheckInterval: loopDurationCheckInterval,
	}, nil
}

// GetGasRequested returns a selection constraint parameter (for gas)
func (options *txSelectionOptions) GetGasRequested() uint64 {
	return options.gasRequested
}

// GetMaxNumTxs returns a selection constraint parameter (for number of transactions)
func (options *txSelectionOptions) GetMaxNumTxs() int {
	return options.maxNumTxs
}

// HaveTimeForSelection returns true if there is any time left for the selection (related to selection duration)
func (options *txSelectionOptions) HaveTimeForSelection() bool {
	return options.haveTimeForSelection()
}

// GetLoopDurationCheckInterval returns a selection constraint parameter (related to selection duration)
func (options *txSelectionOptions) GetLoopDurationCheckInterval() int {
	return options.loopDurationCheckInterval
}

// IsInterfaceNil returns true if there is no value under the interface
func (options *txSelectionOptions) IsInterfaceNil() bool {
	return options == nil
}
