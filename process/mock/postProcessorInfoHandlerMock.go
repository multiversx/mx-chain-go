package mock

// PostProcessorInfoHandlerMock -
type PostProcessorInfoHandlerMock struct {
	GetNumOfCrossInterMbsAndTxsCalled func() (int, int)
	InitProcessedTxsResultsCalled     func(key []byte)
	RevertProcessedTxsResultsCalled   func(txHashes [][]byte, key []byte)
}

// GetNumOfCrossInterMbsAndTxs -
func (ppihm *PostProcessorInfoHandlerMock) GetNumOfCrossInterMbsAndTxs() (int, int) {
	if ppihm.GetNumOfCrossInterMbsAndTxsCalled != nil {
		return ppihm.GetNumOfCrossInterMbsAndTxsCalled()
	}
	return 0, 0
}

// InitProcessedTxsResults -
func (ppihm *PostProcessorInfoHandlerMock) InitProcessedTxsResults(key []byte) {
	if ppihm.InitProcessedTxsResultsCalled != nil {
		ppihm.InitProcessedTxsResultsCalled(key)
	}
}

// RevertProcessedTxsResults -
func (ppihm *PostProcessorInfoHandlerMock) RevertProcessedTxsResults(txHashes [][]byte, key []byte) {
	if ppihm.RevertProcessedTxsResultsCalled != nil {
		ppihm.RevertProcessedTxsResultsCalled(txHashes, key)
	}
}
