package testscommon

import "errors"

// AccountNonceProviderStub -
type AccountNonceProviderStub struct {
	GetAccountNonceCalled func(address []byte) (uint64, error)
}

func NewAccountNonceProviderStub() *AccountNonceProviderStub {
	return &AccountNonceProviderStub{}
}

// GetAccountNonce -
func (stub *AccountNonceProviderStub) GetAccountNonce(address []byte) (uint64, error) {
	if stub.GetAccountNonceCalled != nil {
		return stub.GetAccountNonceCalled(address)
	}

	return 0, errors.New("AccountNonceProviderStub.GetAccountNonceCalled is not set")
}

// IsInterfaceNil -
func (stub *AccountNonceProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
