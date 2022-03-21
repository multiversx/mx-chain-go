package disabled

// CurrentPayloadProvider is the disabled implementation for the CurrentPayloadProvider interface
type CurrentPayloadProvider struct {
}

// BytesToSendToNewPeers will return an empty bytes slice and false
func (provider *CurrentPayloadProvider) BytesToSendToNewPeers() ([]byte, bool) {
	return make([]byte, 0), false
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *CurrentPayloadProvider) IsInterfaceNil() bool {
	return provider == nil
}
