package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// InterceptedTxHandlerStub -
type InterceptedTxHandlerStub struct {
	SenderShardIdCalled   func() uint32
	ReceiverShardIdCalled func() uint32
	NonceCalled           func() uint64
	SenderAddressCalled   func() []byte
	FeeCalled             func() *big.Int
	TransactionCalled     func() data.TransactionHandler
}

// SenderShardId -
func (iths *InterceptedTxHandlerStub) SenderShardId() uint32 {
	return iths.SenderShardIdCalled()
}

// ReceiverShardId -
func (iths *InterceptedTxHandlerStub) ReceiverShardId() uint32 {
	return iths.ReceiverShardIdCalled()
}

// Nonce -
func (iths *InterceptedTxHandlerStub) Nonce() uint64 {
	return iths.NonceCalled()
}

// SenderAddress -
func (iths *InterceptedTxHandlerStub) SenderAddress() []byte {
	return iths.SenderAddressCalled()
}

// Fee -
func (iths *InterceptedTxHandlerStub) Fee() *big.Int {
	return iths.FeeCalled()
}

// Transaction -
func (iths *InterceptedTxHandlerStub) Transaction() data.TransactionHandler {
	return iths.TransactionCalled()
}
