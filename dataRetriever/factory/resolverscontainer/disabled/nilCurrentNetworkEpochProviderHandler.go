package disabled

// NilCurrentNetworkEpochProviderHandler implements CurrentNetworkEpochSetter and does nothing
type NilCurrentNetworkEpochProviderHandler struct {
}

// SetNetworkEpochAtBootstrap does nothing
func (ncneph *NilCurrentNetworkEpochProviderHandler) SetNetworkEpochAtBootstrap(_ uint32) {
}

// EpochIsActiveInNetwork returns true
func (ncneph *NilCurrentNetworkEpochProviderHandler) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// CurrentEpoch returns 0
func (ncneph *NilCurrentNetworkEpochProviderHandler) CurrentEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncneph *NilCurrentNetworkEpochProviderHandler) IsInterfaceNil() bool {
	return ncneph == nil
}
