package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

var errNotImplemented = errors.New("not implemented")

type AccountsHandlerMock struct {
	GetCalled         func(key string) (state.AccountsAdapter, error)
	AddCalled         func(key string, val state.AccountsAdapter) error
	AddMultipleCalled func(keys []string, interceptors []state.AccountsAdapter) error
	ReplaceCalled     func(key string, val state.AccountsAdapter) error
	RemoveCalled      func(key string)
	LenCalled         func() int
}

func (ahm *AccountsHandlerMock) Get(key string) (state.AccountsAdapter, error) {
	if ahm.GetCalled != nil {
		return ahm.GetCalled(key)
	}

	return nil, nil
}
func (ahm *AccountsHandlerMock) Add(key string, val state.AccountsAdapter) error {
	if ahm.AddCalled != nil {
		return ahm.AddCalled(key, val)
	}

	return nil
}
func (ahm *AccountsHandlerMock) AddMultiple(keys []string, interceptors []state.AccountsAdapter) error {
	if ahm.AddMultipleCalled != nil {
		return ahm.AddMultipleCalled(keys, interceptors)
	}

	return nil
}
func (ahm *AccountsHandlerMock) Replace(key string, val state.AccountsAdapter) error {
	if ahm.ReplaceCalled != nil {
		return ahm.ReplaceCalled(key, val)
	}

	return nil
}
func (ahm *AccountsHandlerMock) Remove(key string) {
	if ahm.RemoveCalled != nil {
		ahm.RemoveCalled(key)
	}

}
func (ahm *AccountsHandlerMock) Len() int {
	if ahm.LenCalled != nil {
		return ahm.LenCalled()
	}

	return 0
}
func (ahm *AccountsHandlerMock) IsInterfaceNil() bool {
	return ahm == nil
}
