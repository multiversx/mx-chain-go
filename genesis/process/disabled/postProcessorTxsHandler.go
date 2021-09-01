package disabled

// PostProcessorTxsHandler implements PostProcessorTxsHandler interface but does nothing as it is a disabled component
type PostProcessorTxsHandler struct {
}

// Init does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) Init() {
}

// AddPostProcessorTx does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) AddPostProcessorTx(_ []byte) bool {
	return true
}

// IsPostProcessorTxAdded does nothing as it is a disabled component
func (ppth *PostProcessorTxsHandler) IsPostProcessorTxAdded(_ []byte) bool {
	return false
}

// IsInterfaceNil returns true if underlying object is nil
func (ppth *PostProcessorTxsHandler) IsInterfaceNil() bool {
	return ppth == nil
}
