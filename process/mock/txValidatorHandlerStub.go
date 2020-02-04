package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// TxValidatorHandlerStub -
type TxValidatorHandlerStub struct {
	SenderShardIdCalled func() uint32
	NonceCalled         func() uint64
	SenderAddressCalled func() state.AddressContainer
	FeeCalled           func() *big.Int
}

// SenderShardId -
func (tvhs *TxValidatorHandlerStub) SenderShardId() uint32 {
	return tvhs.SenderShardIdCalled()
}

// Nonce -
func (tvhs *TxValidatorHandlerStub) Nonce() uint64 {
	return tvhs.NonceCalled()
}

// SenderAddress -
func (tvhs *TxValidatorHandlerStub) SenderAddress() state.AddressContainer {
	return tvhs.SenderAddressCalled()
}

// Fee -
func (tvhs *TxValidatorHandlerStub) Fee() *big.Int {
	return tvhs.FeeCalled()
}
