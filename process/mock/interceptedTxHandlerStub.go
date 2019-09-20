package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type InterceptedTxHandlerStub struct {
	SenderShardIdCalled   func() uint32
	ReceiverShardIdCalled func() uint32
	NonceCalled           func() uint64
	SenderAddressCalled   func() state.AddressContainer
	TotalValueCalled      func() *big.Int
	HashCalled            func() []byte
	TransactionCalled     func() data.TransactionHandler
}

func (iths *InterceptedTxHandlerStub) SenderShardId() uint32 {
	return iths.SenderShardIdCalled()
}

func (iths *InterceptedTxHandlerStub) ReceiverShardId() uint32 {
	return iths.ReceiverShardIdCalled()
}

func (iths *InterceptedTxHandlerStub) Nonce() uint64 {
	return iths.NonceCalled()
}

func (iths *InterceptedTxHandlerStub) SenderAddress() state.AddressContainer {
	return iths.SenderAddressCalled()
}

func (iths *InterceptedTxHandlerStub) TotalValue() *big.Int {
	return iths.TotalValueCalled()
}

func (iths *InterceptedTxHandlerStub) Hash() []byte {
	return iths.HashCalled()
}

func (iths *InterceptedTxHandlerStub) Transaction() data.TransactionHandler {
	return iths.TransactionCalled()
}
