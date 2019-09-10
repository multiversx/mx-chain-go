package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

type TxValidatorHandlerStub struct {
	GetSenderShardIdCalled func() uint32
	GetNonceCalled         func() uint64
	GetSenderAddressCalled func() state.AddressContainer
	GetTotalValueCalled    func() *big.Int
}

func (tvhs *TxValidatorHandlerStub) GetSenderShardId() uint32 {
	return tvhs.GetSenderShardIdCalled()
}

func (tvhs *TxValidatorHandlerStub) GetNonce() uint64 {
	return tvhs.GetNonceCalled()
}

func (tvhs *TxValidatorHandlerStub) GetSenderAddress() state.AddressContainer {
	return tvhs.GetSenderAddressCalled()
}

func (tvhs *TxValidatorHandlerStub) GetTotalValue() *big.Int {
	return tvhs.GetTotalValueCalled()
}
