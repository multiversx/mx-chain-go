package testscommon

// AccountNonceProviderStub -
type AccountNonceProviderStub struct {
	GetAccountNonceCalled func(address []byte) (uint64, error)
}

// GetAccountNonce -
func (stub *AccountNonceProviderStub) GetAccountNonce(address []byte) (uint64, error) {
	if stub.GetAccountNonceCalled != nil {
		stub.GetAccountNonceCalled(address)
	}

	return 0, nil
}

// IsInterfaceNil -
func (stub *AccountNonceProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
