package resolver

type disabledInterceptorResolver struct {
}

// NewDisabledInterceptorResolver returns a disabled instance of the debug handler
func NewDisabledInterceptorResolver() *disabledInterceptorResolver {
	return &disabledInterceptorResolver{}
}

// LogRequestedData dos nothing
func (dir *disabledInterceptorResolver) LogRequestedData(_ string, _ [][]byte, _ int, _ int) {
}

// LogReceivedHashes does nothing
func (dir *disabledInterceptorResolver) LogReceivedHashes(_ string, _ [][]byte) {
}

// LogProcessedHashes does nothing
func (dir *disabledInterceptorResolver) LogProcessedHashes(_ string, _ [][]byte, _ error) {
}

// Query returns an empty slice
func (dir *disabledInterceptorResolver) Query(_ string) []string {
	return make([]string, 0)
}

// LogFailedToResolveData does nothing
func (dir *disabledInterceptorResolver) LogFailedToResolveData(_ string, _ []byte, _ error) {
}

// LogSucceededToResolveData does nothing
func (dir *disabledInterceptorResolver) LogSucceededToResolveData(_ string, _ []byte) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (dir *disabledInterceptorResolver) IsInterfaceNil() bool {
	return dir == nil
}
