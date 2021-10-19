package mock

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// UnitOpenerStub -
type UnitOpenerStub struct {
}

// OpenDB -
func (u *UnitOpenerStub) OpenDB(_ config.DBConfig, _ uint32, _ uint32) (storage.Storer, error) {
	return &StorerMock{}, nil
}

// GetMostRecentStorageUnit -
func (u *UnitOpenerStub) GetMostRecentStorageUnit(_ config.DBConfig) (storage.Storer, error) {
	return NewStorerMock(), nil
}

// IsInterfaceNil -
func (u *UnitOpenerStub) IsInterfaceNil() bool {
	return u == nil
}
