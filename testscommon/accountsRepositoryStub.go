package testscommon

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsRepositoryStub struct {
	testAccounts             map[string]vmcommon.AccountHandler
	GetExistingAccountCalled func(address []byte, rootHash []byte) (vmcommon.AccountHandler, error)
	GetCodeCalled            func(codeHash []byte, rootHash []byte) []byte
	CloseCalled              func() error
}

// NewAccountsRepositoryStub -
func NewAccountsRepositoryStub() *accountsRepositoryStub {
	return &accountsRepositoryStub{
		testAccounts: make(map[string]vmcommon.AccountHandler),
	}
}

// GetExistingAccount -
func (stub *accountsRepositoryStub) GetExistingAccount(address []byte, rootHash []byte) (vmcommon.AccountHandler, error) {
	if stub.GetExistingAccountCalled != nil {
		return stub.GetExistingAccountCalled(address, rootHash)
	}

	value, ok := stub.testAccounts[string(address)]
	if ok {
		return value, nil
	}

	return nil, nil
}

// GetCode -
func (stub *accountsRepositoryStub) GetCode(codeHash []byte, rootHash []byte) []byte {
	if stub.GetCodeCalled != nil {
		return stub.GetCodeCalled(codeHash, rootHash)
	}

	return nil
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
