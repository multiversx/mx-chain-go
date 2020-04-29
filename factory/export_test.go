package factory

// SetSkPkProviderHandler updates the handler for testing reasons
func (cspf *cryptoSigningParamsLoader) SetSkPkProviderHandler(handler func() ([]byte, []byte, error)) {
	cspf.skPkProviderHandler = handler
}

// GetSkPk will call the inner function
func (cspf *cryptoSigningParamsLoader) GetSkPk() ([]byte, []byte, error) {
	return cspf.getSkPk()
}

// SetListenAddress will update the listen address for testing reasons
func (ncf *networkComponentsFactory) SetListenAddress(address string) {
	ncf.listenAddress = address
}
