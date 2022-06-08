package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsRepositoryStub struct {
	testAccounts              map[string]vmcommon.AccountHandler
	GetAccountOnFinalCalled   func(address []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error)
	GetAccountOnCurrentCalled func(address []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error)
	GetCodeOnFinalCalled      func(codeHash []byte) ([]byte, state.AccountBlockInfo)
	GetCodeOnCurrentCalled    func(codeHash []byte) ([]byte, state.AccountBlockInfo)
	CloseCalled               func() error
}

// NewAccountsRepositoryStub -
func NewAccountsRepositoryStub() *accountsRepositoryStub {
	return &accountsRepositoryStub{
		testAccounts: make(map[string]vmcommon.AccountHandler),
	}
}

// GetAccountOnFinal -
func (stub *accountsRepositoryStub) GetAccountOnFinal(address []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error) {
	if stub.GetAccountOnFinalCalled != nil {
		return stub.GetAccountOnFinalCalled(address)
	}

	value, ok := stub.testAccounts[string(address)]
	if ok {
		return value, &AccountBlockInfoStub{}, nil
	}

	return nil, nil, nil
}

// GetAccountOnCurrent -
func (stub *accountsRepositoryStub) GetAccountOnCurrent(address []byte) (vmcommon.AccountHandler, state.AccountBlockInfo, error) {
	if stub.GetAccountOnCurrentCalled != nil {
		return stub.GetAccountOnCurrentCalled(address)
	}

	value, ok := stub.testAccounts[string(address)]
	if ok {
		return value, &AccountBlockInfoStub{}, nil
	}

	return nil, nil, nil
}

// GetCodeOnFinal -
func (stub *accountsRepositoryStub) GetCodeOnFinal(codeHash []byte) ([]byte, state.AccountBlockInfo) {
	if stub.GetCodeOnFinalCalled != nil {
		return stub.GetCodeOnFinalCalled(codeHash)
	}

	return nil, nil
}

// GetCodeOnCurrent -
func (stub *accountsRepositoryStub) GetCodeOnCurrent(codeHash []byte) ([]byte, state.AccountBlockInfo) {
	if stub.GetCodeOnCurrentCalled != nil {
		return stub.GetCodeOnCurrentCalled(codeHash)
	}

	return nil, nil
}

// Close -
func (stub *accountsRepositoryStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// AddTestAccount -
func (stub *accountsRepositoryStub) AddTestAccount(account vmcommon.AccountHandler) {
	stub.testAccounts[string(account.AddressBytes())] = account
}

// IsInterfaceNil -
func (stub *accountsRepositoryStub) IsInterfaceNil() bool {
	return stub == nil
}
