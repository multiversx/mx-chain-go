package testscommon

// PostProcessorTxsStub -
type PostProcessorTxsStub struct {
	InitCalled                   func()
	AddPostProcessorTxCalled     func([]byte) bool
	IsPostProcessorTxAddedCalled func([]byte) bool
}

// Init -
func (ppts *PostProcessorTxsStub) Init() {
	if ppts.InitCalled != nil {
		ppts.InitCalled()
	}
}

// AddPostProcessorTx -
func (ppts *PostProcessorTxsStub) AddPostProcessorTx(txHash []byte) bool {
	if ppts.AddPostProcessorTxCalled != nil {
		return ppts.AddPostProcessorTxCalled(txHash)
	}
	return true
}

// IsPostProcessorTxAdded -
func (ppts *PostProcessorTxsStub) IsPostProcessorTxAdded(txHash []byte) bool {
	if ppts.IsPostProcessorTxAddedCalled != nil {
		return ppts.IsPostProcessorTxAddedCalled(txHash)
	}
	return false
}

// IsInterfaceNil -
func (ppts *PostProcessorTxsStub) IsInterfaceNil() bool {
	return ppts == nil
}
