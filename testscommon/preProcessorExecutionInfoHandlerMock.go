package testscommon

// PreProcessorExecutionInfoHandlerMock -
type PreProcessorExecutionInfoHandlerMock struct {
	GetNumOfCrossInterMbsAndTxsCalled func() (int, int)
	InitProcessedTxsResultsCalled     func(key []byte, parentKey []byte)
	RevertProcessedTxsResultsCalled   func(txHashes [][]byte, key []byte)
}

// GetNumOfCrossInterMbsAndTxs -
func (ppeihm *PreProcessorExecutionInfoHandlerMock) GetNumOfCrossInterMbsAndTxs() (int, int) {
	if ppeihm.GetNumOfCrossInterMbsAndTxsCalled != nil {
		return ppeihm.GetNumOfCrossInterMbsAndTxsCalled()
	}
	return 0, 0
}

// InitProcessedTxsResults -
func (ppeihm *PreProcessorExecutionInfoHandlerMock) InitProcessedTxsResults(key []byte, parentKey []byte) {
	if ppeihm.InitProcessedTxsResultsCalled != nil {
		ppeihm.InitProcessedTxsResultsCalled(key, parentKey)
	}
}

// RevertProcessedTxsResults -
func (ppeihm *PreProcessorExecutionInfoHandlerMock) RevertProcessedTxsResults(txHashes [][]byte, key []byte) {
	if ppeihm.RevertProcessedTxsResultsCalled != nil {
		ppeihm.RevertProcessedTxsResultsCalled(txHashes, key)
	}
}
