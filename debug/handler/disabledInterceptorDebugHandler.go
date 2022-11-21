package handler

type disabledInterceptorDebugHandler struct {
}

// NewDisabledInterceptorDebugHandler returns a disabled instance of the debug handler
func NewDisabledInterceptorDebugHandler() *disabledInterceptorDebugHandler {
	return &disabledInterceptorDebugHandler{}
}

// LogRequestedData does nothing
func (dir *disabledInterceptorDebugHandler) LogRequestedData(_ string, _ [][]byte, _ int, _ int) {
}

// LogReceivedHashes does nothing
func (dir *disabledInterceptorDebugHandler) LogReceivedHashes(_ string, _ [][]byte) {
}

// LogProcessedHashes does nothing
func (dir *disabledInterceptorDebugHandler) LogProcessedHashes(_ string, _ [][]byte, _ error) {
}

// Query returns an empty slice
func (dir *disabledInterceptorDebugHandler) Query(_ string) []string {
	return make([]string, 0)
}

// LogFailedToResolveData does nothing
func (dir *disabledInterceptorDebugHandler) LogFailedToResolveData(_ string, _ []byte, _ error) {
}

// LogSucceededToResolveData does nothing
func (dir *disabledInterceptorDebugHandler) LogSucceededToResolveData(_ string, _ []byte) {
}

// Close returns nil
func (dir *disabledInterceptorDebugHandler) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dir *disabledInterceptorDebugHandler) IsInterfaceNil() bool {
	return dir == nil
}
