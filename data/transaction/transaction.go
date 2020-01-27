package transaction

import (
	"encoding/json"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// Transaction holds all the data needed for a value transfer or SC call
type Transaction struct {
	Nonce     uint64   `json:"nonce"`
	Value     *big.Int `json:"value"`
	RcvAddr   []byte   `json:"receiver"`
	SndAddr   []byte   `json:"sender"`
	GasPrice  uint64   `json:"gasPrice,omitempty"`
	GasLimit  uint64   `json:"gasLimit,omitempty"`
	Data      []byte   `json:"data,omitempty"`
	Signature []byte   `json:"signature,omitempty"`
}

// IsInterfaceNil verifies if underlying object is nil
func (tx *Transaction) IsInterfaceNil() bool {
	return tx == nil
}

// GetValue returns the value of the transaction
func (tx *Transaction) GetValue() *big.Int {
	return tx.Value
}

// GetNonce returns the transaction nonce
func (tx *Transaction) GetNonce() uint64 {
	return tx.Nonce
}

// GetData returns the data of the transaction
func (tx *Transaction) GetData() []byte {
	return tx.Data
}

// GetRecvAddress returns the receiver address from the transaction
func (tx *Transaction) GetRecvAddress() []byte {
	return tx.RcvAddr
}

// GetSndAddress returns the sender address from the transaction
func (tx *Transaction) GetSndAddress() []byte {
	return tx.SndAddr
}

// GetGasLimit returns the gas limit of the transaction
func (tx *Transaction) GetGasLimit() uint64 {
	return tx.GasLimit
}

// GetGasPrice returns the gas price of the transaction
func (tx *Transaction) GetGasPrice() uint64 {
	return tx.GasPrice
}

// SetValue sets the value of the transaction
func (tx *Transaction) SetValue(value *big.Int) {
	tx.Value = value
}

// SetData sets the data of the transaction
func (tx *Transaction) SetData(data []byte) {
	tx.Data = data
}

// SetRecvAddress sets the receiver address of the transaction
func (tx *Transaction) SetRecvAddress(addr []byte) {
	tx.RcvAddr = addr
}

// SetSndAddress sets the sender address of the transaction
func (tx *Transaction) SetSndAddress(addr []byte) {
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
		return data.ErrInvalidValue
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
