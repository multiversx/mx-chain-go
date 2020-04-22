package resolver

type disabledInterceptorResolver struct {
}

// NewDisabledInterceptorResolver returns a disabled instance of the debug handler
func NewDisabledInterceptorResolver() *disabledInterceptorResolver {
	return &disabledInterceptorResolver{}
}

// LogRequestedData dos nothing
func (dir *disabledInterceptorResolver) LogRequestedData(_ string, _ []byte, _ int, _ int) {
}

// LogReceivedHash does nothing
func (dir *disabledInterceptorResolver) LogReceivedHash(_ string, _ []byte) {
}

// LogProcessedHash does nothing
func (dir *disabledInterceptorResolver) LogProcessedHash(_ string, _ []byte, _ error) {
}

// Enabled returns if this implementation should process data or not. Always false to optimize the caller
func (dir *disabledInterceptorResolver) Enabled() bool {
	return false
}

// Query returns an empty slice
func (dir *disabledInterceptorResolver) Query(_ string) []string {
	return make([]string, 0)
}

// LogFailedToResolveData does nothing
func (dir *disabledInterceptorResolver) LogFailedToResolveData(_ string, _ []byte, _ error) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (dir *disabledInterceptorResolver) IsInterfaceNil() bool {
	return dir == nil
}
