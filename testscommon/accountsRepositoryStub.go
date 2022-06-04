package testscommon

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsRepositoryStub struct {
	testAccounts             map[string]vmcommon.AccountHandler
	GetExistingAccountCalled func(address []byte, rootHash []byte) (vmcommon.AccountHandler, error)
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

// AddTestAccount -
func (stub *accountsRepositoryStub) AddTestAccount(pubKey []byte, balance *big.Int, nonce uint64) {
	account, _ := state.NewUserAccount(pubKey)
	account.Balance = balance
	account.Nonce = nonce
	stub.testAccounts[string(pubKey)] = account
}

// IsInterfaceNil -
func (stub *accountsRepositoryStub) IsInterfaceNil() bool {
	return stub == nil
}
