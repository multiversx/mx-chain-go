package interceptedTxMocks

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
)

// InterceptedUnsignedTxHandlerStub -
type InterceptedUnsignedTxHandlerStub struct {
	SenderShardIdCalled   func() uint32
	ReceiverShardIdCalled func() uint32
	NonceCalled           func() uint64
	SenderAddressCalled   func() []byte
	FeeCalled             func() *big.Int
	TransactionCalled     func() data.TransactionHandler
	UserTransactionCalled func() data.TransactionHandler
}

// SenderShardId -
func (iths *InterceptedUnsignedTxHandlerStub) SenderShardId() uint32 {
	if iths.SenderShardIdCalled != nil {
		return iths.SenderShardIdCalled()
	}
	return 0
}

// ReceiverShardId -
func (iths *InterceptedUnsignedTxHandlerStub) ReceiverShardId() uint32 {
	if iths.ReceiverShardIdCalled != nil {
		return iths.ReceiverShardIdCalled()
	}
	return 0
}

// Nonce -
func (iths *InterceptedUnsignedTxHandlerStub) Nonce() uint64 {
	if iths.NonceCalled != nil {
		return iths.NonceCalled()
	}
	return 0
}

// SenderAddress -
func (iths *InterceptedUnsignedTxHandlerStub) SenderAddress() []byte {
	if iths.SenderAddressCalled != nil {
		return iths.SenderAddressCalled()
	}
	return nil
}

// Fee -
func (iths *InterceptedUnsignedTxHandlerStub) Fee() *big.Int {
	if iths.FeeCalled != nil {
		return iths.FeeCalled()
	}
	return nil
}

// Transaction -
func (iths *InterceptedUnsignedTxHandlerStub) Transaction() data.TransactionHandler {
	if iths.TransactionCalled != nil {
		return iths.TransactionCalled()
	}
	return nil
}

// UserTransaction -
func (iths *InterceptedUnsignedTxHandlerStub) UserTransaction() data.TransactionHandler {
	if iths.UserTransactionCalled != nil {
		return iths.UserTransactionCalled()
	}
	return nil
}
