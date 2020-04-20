package resolver

type disabledInterceptorResolver struct {
}

// NewDisabledInterceptorResolver returns a disabled instance of the debug handler
func NewDisabledInterceptorResolver() *disabledInterceptorResolver {
	return &disabledInterceptorResolver{}
}

// RequestedData dos nothing
func (dir *disabledInterceptorResolver) RequestedData(_ string, _ []byte, _ int, _ int) {
}

// ReceivedHash does nothing
func (dir *disabledInterceptorResolver) ReceivedHash(_ string, _ []byte) {
}

// ProcessedHash does nothing
func (dir *disabledInterceptorResolver) ProcessedHash(_ string, _ []byte, _ error) {
}

// Enabled returns if this implementation should process data or not. Always false to optimize the caller
func (dir *disabledInterceptorResolver) Enabled() bool {
	return false
}

// Query returns an empty slice
func (dir *disabledInterceptorResolver) Query(_ string) []string {
	return make([]string, 0)
}

// FailedToResolveData does nothing
func (dir *disabledInterceptorResolver) FailedToResolveData(_ string, _ []byte, _ error) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (dir *disabledInterceptorResolver) IsInterfaceNil() bool {
	return dir == nil
}
