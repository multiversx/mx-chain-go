package integrationtests

import (
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
