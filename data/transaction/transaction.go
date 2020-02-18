//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf  --gogoslick_out=. transaction.proto
package transaction

import (
	"encoding/json"
	io "io"
	"io/ioutil"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

var _ = data.TransactionHandler(&Transaction{})

// IsInterfaceNil verifies if underlying object is nil
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

// MarshalJSON converts the Transaction data type into its corresponding equivalent in byte slice.
// Note that Value data type is converted in a string
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	valAsString := "nil"
	if tx.Value != nil {
		valAsString = tx.Value.String()
	}
	return json.Marshal(&struct {
		Nonce     uint64 `json:"nonce"`
		Value     string `json:"value"`
		RcvAddr   []byte `json:"receiver"`
		SndAddr   []byte `json:"sender"`
		GasPrice  uint64 `json:"gasPrice,omitempty"`
		GasLimit  uint64 `json:"gasLimit,omitempty"`
		Data      []byte `json:"data,omitempty"`
		Signature []byte `json:"signature,omitempty"`
	}{
		Nonce:     tx.Nonce,
		Value:     valAsString,
		RcvAddr:   tx.RcvAddr,
		SndAddr:   tx.SndAddr,
		GasPrice:  tx.GasPrice,
		GasLimit:  tx.GasLimit,
		Data:      tx.Data,
		Signature: tx.Signature,
	})
}

// UnmarshalJSON converts the provided bytes into a Transaction data type.
func (tx *Transaction) UnmarshalJSON(dataBuff []byte) error {
	aux := &struct {
		Nonce     uint64 `json:"nonce"`
		Value     string `json:"value"`
		RcvAddr   []byte `json:"receiver"`
		SndAddr   []byte `json:"sender"`
		GasPrice  uint64 `json:"gasPrice,omitempty"`
		GasLimit  uint64 `json:"gasLimit,omitempty"`
		Data      []byte `json:"data,omitempty"`
		Signature []byte `json:"signature,omitempty"`
	}{}
	if err := json.Unmarshal(dataBuff, &aux); err != nil {
		return err
	}
	tx.Nonce = aux.Nonce
	tx.RcvAddr = aux.RcvAddr
	tx.SndAddr = aux.SndAddr
	tx.GasPrice = aux.GasPrice
	tx.GasLimit = aux.GasLimit
	tx.Data = aux.Data
	tx.Signature = aux.Signature

	var ok bool
	tx.Value, ok = big.NewInt(0).SetString(aux.Value, 10)
	if !ok {
		tx.Value = nil
	}

	return nil
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
