package mock

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/data"
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

// GetUserTxSenderInRelayedTx -
func (tvhs *TxValidatorHandlerStub) GetUserTxSenderInRelayedTx() ([]byte, error) {
	return nil, errors.New("error")
}

// Transaction -
func (tvhs *TxValidatorHandlerStub) Transaction() data.TransactionHandler {
	panic("implement me")
}

func (tvhs *TxValidatorHandlerStub) CheckValidity() error {
	panic("implement me")
}

func (tvhs *TxValidatorHandlerStub) IsForCurrentShard() bool {
	panic("implement me")
}

func (tvhs *TxValidatorHandlerStub) IsInterfaceNil() bool {
	return tvhs == nil
}

func (tvhs *TxValidatorHandlerStub) Hash() []byte {
	panic("implement me")
}

func (tvhs *TxValidatorHandlerStub) Type() string {
	panic("implement me")
}

func (tvhs *TxValidatorHandlerStub) Identifiers() [][]byte {
	panic("implement me")
}

func (tvhs *TxValidatorHandlerStub) String() string {
	panic("implement me")
}
