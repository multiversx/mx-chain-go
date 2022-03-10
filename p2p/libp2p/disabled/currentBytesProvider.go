package disabled

// CurrentBytesProvider is the disabled implementation for the CurrentBytesProvider interface
type CurrentBytesProvider struct {
}

// BytesToSendToNewPeers will return an empty bytes slice and false
func (provider *CurrentBytesProvider) BytesToSendToNewPeers() ([]byte, bool) {
	return make([]byte, 0), false
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *CurrentBytesProvider) IsInterfaceNil() bool {
	return provider == nil
}
