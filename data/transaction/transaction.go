//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. transaction.proto
package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&Transaction{})

// IsInterfaceNil vesorrifies if underlying object is nil
func (tx *Transaction) IsInterfaceNil() bool {
	return tx == nil
}

// SetValue sets the value of the transaction
func (tx *Transaction) SetValue(value *big.Int) {
	tx.Value = value
}

// SetData sets the data of the transaction
func (tx *Transaction) SetData(data []byte) {
	tx.Data = data
}

// SetRcvAddr sets the receiver address of the transaction
func (tx *Transaction) SetRcvAddr(addr []byte) {
	tx.RcvAddr = addr
}

// SetSndAddr sets the sender address of the transaction
func (tx *Transaction) SetSndAddr(addr []byte) {
	tx.SndAddr = addr
}

// TrimSlicePtr creates a copy of the provided slice without the excess capacity
func TrimSlicePtr(in []*Transaction) []*Transaction {
	if len(in) == 0 {
		return []*Transaction{}
	}
	ret := make([]*Transaction, len(in))
	copy(ret, in)
	return ret
}

// TrimSliceHandler creates a copy of the provided slice without the excess capacity
func TrimSliceHandler(in []data.TransactionHandler) []data.TransactionHandler {
	if len(in) == 0 {
		return []data.TransactionHandler{}
	}
	ret := make([]data.TransactionHandler, len(in))
	copy(ret, in)
	return ret
}

// frontendTransaction represents the DTO used in transaction signing/validation.
type frontendTransaction struct {
	Nonce            uint64 `json:"nonce"`
	Value            string `json:"value"`
	Receiver         string `json:"receiver"`
	Sender           string `json:"sender"`
	SenderUsername   []byte `json:"senderusername,omitempty"`
	ReceiverUsername []byte `json:"receiverusername,omitempty"`
	GasPrice         uint64 `json:"gasPrice"`
	GasLimit         uint64 `json:"gasLimit"`
	Data             []byte `json:"data,omitempty"`
	Signature        string `json:"signature,omitempty"`
}

// GetFataForSigning returns the serialized transaction having the
func (tx *Transaction) GetFataForSigning(encoder Encoder, marshalizer Marshalizer) ([]byte, error) {
	if check.IfNil(encoder) {
		return nil, ErrNilEncoder
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

	ftx := &frontendTransaction{
		Nonce:            tx.Nonce,
		Value:            tx.Value.String(),
		Receiver:         encoder.Encode(tx.RcvAddr),
		Sender:           encoder.Encode(tx.SndAddr),
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		SenderUsername:   tx.SndUserName,
		ReceiverUsername: tx.RcvUserName,
		Data:             tx.Data,
	}

	return marshalizer.Marshal(ftx)
}
