package disabled

// currentNetworkEpochProviderHandler -
type currentNetworkEpochProviderHandler struct {
}

// NewCurrentNetworkEpochProviderHandler will create a new currentNetworkEpochProviderHandler instance
func NewCurrentNetworkEpochProviderHandler() *currentNetworkEpochProviderHandler {
	return &currentNetworkEpochProviderHandler{}
}

// EpochConfirmed does nothing
func (ep *currentNetworkEpochProviderHandler) EpochConfirmed(_ uint32, _ uint64) {
}

// EpochIsActiveInNetwork returns true
func (cneph *currentNetworkEpochProviderHandler) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (cneph *currentNetworkEpochProviderHandler) IsInterfaceNil() bool {
	return cneph == nil
}
