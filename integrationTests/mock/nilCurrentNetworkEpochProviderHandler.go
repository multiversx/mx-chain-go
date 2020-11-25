package mock

// NilCurrentNetworkEpochProviderHandler -
type NilCurrentNetworkEpochProviderHandler struct {
}

// SetNetworkEpochAtBootstrap -
func (ncneph *NilCurrentNetworkEpochProviderHandler) SetNetworkEpochAtBootstrap(_ uint32) {
}

// EpochIsActiveInNetwork -
func (ncneph *NilCurrentNetworkEpochProviderHandler) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// CurrentEpoch -
func (ncneph *NilCurrentNetworkEpochProviderHandler) CurrentEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncneph *NilCurrentNetworkEpochProviderHandler) IsInterfaceNil() bool {
	return ncneph == nil
}
