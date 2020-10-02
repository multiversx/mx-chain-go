package mock

// NilCurrentNetworkEpochProviderHandler -
type NilCurrentNetworkEpochProviderHandler struct {
}

// SetNetworkEpochAtBootstrap will update the component's current epoch
func (ncneph *NilCurrentNetworkEpochProviderHandler) SetNetworkEpochAtBootstrap(_ uint32) {
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (ncneph *NilCurrentNetworkEpochProviderHandler) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// CurrentEpoch returns the current network epoch
func (ncneph *NilCurrentNetworkEpochProviderHandler) CurrentEpoch() uint32 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (ncneph *NilCurrentNetworkEpochProviderHandler) IsInterfaceNil() bool {
	return ncneph == nil
}
