package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type TransactionInterceptorMock struct {
	IsChecked                   bool
	IsVerified                  bool
	RcvShardVal                 uint32
	SndShardVal                 uint32
	IsAddressedToOtherShardsVal bool
	AddrConverter               state.AddressConverter
	Tx                          *transaction.Transaction
}

func (tim *TransactionInterceptorMock) Check() bool {
	return tim.IsChecked
}

func (tim *TransactionInterceptorMock) VerifySig() bool {
	return tim.IsVerified
}

func (tim *TransactionInterceptorMock) New() p2p.Newer {
	return &TransactionInterceptorMock{}
}

func (tim *TransactionInterceptorMock) ID() string {
	panic("implement me")
}

func (tim *TransactionInterceptorMock) RcvShard() uint32 {
	return tim.RcvShardVal
}

func (tim *TransactionInterceptorMock) SndShard() uint32 {
	return tim.SndShardVal
}

func (tim *TransactionInterceptorMock) IsAddressedToOtherShards() bool {
	return tim.IsAddressedToOtherShardsVal
}

func (tim *TransactionInterceptorMock) SetAddressConverter(converter state.AddressConverter) {
	tim.AddrConverter = converter
}

func (tim *TransactionInterceptorMock) AddressConverter() state.AddressConverter {
	return tim.AddrConverter
}

func (tim *TransactionInterceptorMock) GetTransaction() *transaction.Transaction {
	return tim.Tx
}
