package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"math/big"
)

type BlockChainHookHandlerMock struct {
	AddTempAccountCalled    func(address []byte, balance *big.Int, nonce uint64)
	CleanTempAccountsCalled func()
	TempAccountCalled       func(address []byte) state.AccountHandler
	SetCurrentHeaderCalled  func(hdr data.HeaderHandler)
}

func (e *BlockChainHookHandlerMock) AddTempAccount(address []byte, balance *big.Int, nonce uint64) {
	if e.AddTempAccountCalled != nil {
		e.AddTempAccountCalled(address, balance, nonce)
	}
}

func (e *BlockChainHookHandlerMock) CleanTempAccounts() {
	if e.CleanTempAccountsCalled != nil {
		e.CleanTempAccountsCalled()
	}
}

func (e *BlockChainHookHandlerMock) TempAccount(address []byte) state.AccountHandler {
	if e.TempAccountCalled != nil {
		return e.TempAccountCalled(address)
	}
	return nil
}

func (e *BlockChainHookHandlerMock) IsInterfaceNil() bool {
	if e == nil {
		return true
	}
	return false
}

func (e *BlockChainHookHandlerMock) SetCurrentHeader(hdr data.HeaderHandler) {
	if e.SetCurrentHeaderCalled != nil {
		e.SetCurrentHeaderCalled(hdr)
	}
}
