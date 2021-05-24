package disabled

// epochProvider will be used for regular validator/observer nodes
type epochProvider struct {
}

// NewEpochProvider returns a new disabled epoch provider instance
func NewEpochProvider() *epochProvider {
	return &epochProvider{}
}

// EpochIsActiveInNetwork returns true
func (ep *epochProvider) EpochIsActiveInNetwork(_ uint32) bool {
	return true
}

// EpochConfirmed does nothing
func (ep *epochProvider) EpochConfirmed(_ uint32, _ uint64) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ep *epochProvider) IsInterfaceNil() bool {
	return ep == nil
}
