package integrationtests

import (
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// TestAccountFactory -
type TestAccountFactory struct {
}

// CreateAccount -
func (factory *TestAccountFactory) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	return state.NewUserAccount(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (factory *TestAccountFactory) IsInterfaceNil() bool {
	return factory == nil
}
