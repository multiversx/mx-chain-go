package mock

import (
	"math/big"
)

// TxValidatorHandlerStub -
type TxValidatorHandlerStub struct {
	SenderShardIdCalled   func() uint32
	ReceiverShardIdCalled func() uint32
	NonceCalled           func() uint64
	SenderAddressCalled   func() []byte
	FeeCalled             func() *big.Int
}

// SenderShardId -
func (tvhs *TxValidatorHandlerStub) SenderShardId() uint32 {
	return tvhs.SenderShardIdCalled()
}

// ReceiverShardId -
func (tvhs *TxValidatorHandlerStub) ReceiverShardId() uint32 {
	return tvhs.ReceiverShardIdCalled()
}

// Nonce -
func (tvhs *TxValidatorHandlerStub) Nonce() uint64 {
	return tvhs.NonceCalled()
}

// SenderAddress -
func (tvhs *TxValidatorHandlerStub) SenderAddress() []byte {
	return tvhs.SenderAddressCalled()
}

// Fee -
func (tvhs *TxValidatorHandlerStub) Fee() *big.Int {
	return tvhs.FeeCalled()
}
