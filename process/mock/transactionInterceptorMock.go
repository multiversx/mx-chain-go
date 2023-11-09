package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// TransactionInterceptorMock -
type TransactionInterceptorMock struct {
	IsChecked                   bool
	IsVerified                  bool
	RcvShardVal                 uint32
	SndShardVal                 uint32
	IsAddressedToOtherShardsVal bool
	pubkeyConverter             core.PubkeyConverter
	Tx                          *transaction.Transaction
	hash                        []byte
}

// Check -
func (tim *TransactionInterceptorMock) Check() bool {
	return tim.IsChecked
}

// VerifySig -
func (tim *TransactionInterceptorMock) VerifySig() bool {
	return tim.IsVerified
}

// ID -
func (tim *TransactionInterceptorMock) ID() string {
	panic("implement me")
}

// RcvShard -
func (tim *TransactionInterceptorMock) RcvShard() uint32 {
	return tim.RcvShardVal
}

// SndShard -
func (tim *TransactionInterceptorMock) SndShard() uint32 {
	return tim.SndShardVal
}

// IsAddressedToOtherShards -
func (tim *TransactionInterceptorMock) IsAddressedToOtherShards() bool {
	return tim.IsAddressedToOtherShardsVal
}

// SetPubkeyConverter -
func (tim *TransactionInterceptorMock) SetPubkeyConverter(converter core.PubkeyConverter) {
	tim.pubkeyConverter = converter
}

// PubkeyConverter -
func (tim *TransactionInterceptorMock) PubkeyConverter() core.PubkeyConverter {
	return tim.pubkeyConverter
}

// GetTransaction -
func (tim *TransactionInterceptorMock) GetTransaction() *transaction.Transaction {
	return tim.Tx
}

// SetHash -
func (tim *TransactionInterceptorMock) SetHash(hash []byte) {
	tim.hash = hash
}

// Hash -
func (tim *TransactionInterceptorMock) Hash() []byte {
	return tim.hash
}

// IsInterfaceNil returns true if there is no value under the interface
func (tim *TransactionInterceptorMock) IsInterfaceNil() bool {
	return tim == nil
}
