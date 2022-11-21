package state

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/marshallerMock"
	"github.com/ElrondNetwork/elrond-vm-common"
)

// AccountsFactoryStub -
type AccountsFactoryStub struct {
	CreateAccountCalled func(address []byte, hasher hashing.Hasher, marshaller marshal.Marshalizer) (vmcommon.AccountHandler, error)
}

// CreateAccount -
func (afs *AccountsFactoryStub) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	return afs.CreateAccountCalled(address, &hashingMocks.HasherMock{}, &marshallerMock.MarshalizerMock{})
}

// IsInterfaceNil returns true if there is no value under the interface
func (afs *AccountsFactoryStub) IsInterfaceNil() bool {
	return afs == nil
}
