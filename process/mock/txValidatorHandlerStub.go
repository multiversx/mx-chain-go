package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

type TxValidatorHandlerStub struct {
	SenderShardIdCalled func() uint32
	NonceCalled         func() uint64
	SenderAddressCalled func() state.AddressContainer
	FeeCalled           func() *big.Int
}

func (tvhs *TxValidatorHandlerStub) SenderShardId() uint32 {
	return tvhs.SenderShardIdCalled()
}

func (tvhs *TxValidatorHandlerStub) Nonce() uint64 {
	return tvhs.NonceCalled()
}

func (tvhs *TxValidatorHandlerStub) SenderAddress() state.AddressContainer {
	return tvhs.SenderAddressCalled()
}

func (tvhs *TxValidatorHandlerStub) Fee() *big.Int {
	return tvhs.FeeCalled()
}
