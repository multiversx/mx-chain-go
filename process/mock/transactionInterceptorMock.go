package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

type TransactionInterceptorMock struct {
	IsChecked                   bool
	IsVerified                  bool
	RcvShardVal                 uint32
	SndShardVal                 uint32
	IsAddressedToOtherShardsVal bool
	AddrConverter               state.AddressConverter
	Tx                          *transaction.Transaction
	hash                        []byte
}

func (tim *TransactionInterceptorMock) Check() bool {
	return tim.IsChecked
}

func (tim *TransactionInterceptorMock) VerifySig() bool {
	return tim.IsVerified
}

//func (tim *TransactionInterceptorMock) Create() p2p.Creator {
//	return &TransactionInterceptorMock{}
//}

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

func (tim *TransactionInterceptorMock) SetHash(hash []byte) {
	tim.hash = hash
}

func (tim *TransactionInterceptorMock) Hash() []byte {
	return tim.hash
}
