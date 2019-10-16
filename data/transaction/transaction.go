package transaction

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/transaction/capnp"
	capn "github.com/glycerine/go-capnproto"
)

// Transaction holds all the data needed for a value transfer
type Transaction struct {
	Nonce     uint64   `capid:"0" json:"nonce"`
	Value     *big.Int `capid:"1" json:"value"`
	RcvAddr   []byte   `capid:"2" json:"receiver"`
	SndAddr   []byte   `capid:"3" json:"sender"`
	GasPrice  uint64   `capid:"4" json:"gasPrice,omitempty"`
	GasLimit  uint64   `capid:"5" json:"gasLimit,omitempty"`
	Data      string   `capid:"6" json:"data,omitempty"`
	Signature []byte   `capid:"7" json:"signature,omitempty"`
	Challenge []byte   `capid:"8" json:"challenge,omitempty"`
}

// Save saves the serialized data of a Transaction into a stream through Capnp protocol
func (tx *Transaction) Save(w io.Writer) error {
	seg := capn.NewBuffer(nil)
	TransactionGoToCapn(seg, tx)
	_, err := seg.WriteTo(w)
	return err
}

// Load loads the data from the stream into a Transaction object through Capnp protocol
func (tx *Transaction) Load(r io.Reader) error {
	capMsg, err := capn.ReadFromStream(r, nil)
	if err != nil {
		return err
	}
	z := capnp.ReadRootTransactionCapn(capMsg)
	TransactionCapnToGo(z, tx)
	return nil
}

// TransactionCapnToGo is a helper function to copy fields from a TransactionCapn object to a Transaction object
func TransactionCapnToGo(src capnp.TransactionCapn, dest *Transaction) *Transaction {
	if dest == nil {
		dest = &Transaction{}
	}

	if dest.Value == nil {
		dest.Value = big.NewInt(0)
	}

	// Nonce
	dest.Nonce = src.Nonce()
	// Value
	err := dest.Value.GobDecode(src.Value())

	if err != nil {
		return nil
	}

	// RcvAddr
	dest.RcvAddr = src.RcvAddr()
	// SndAddr
	dest.SndAddr = src.SndAddr()
	// GasPrice
	dest.GasPrice = src.GasPrice()
	// GasLimit
	dest.GasLimit = src.GasLimit()
	// Data
	dest.Data = src.Data()
	// Signature
	dest.Signature = src.Signature()
	// Challenge
	dest.Challenge = src.Challenge()

	return dest
}

// TransactionGoToCapn is a helper function to copy fields from a Transaction object to a TransactionCapn object
func TransactionGoToCapn(seg *capn.Segment, src *Transaction) capnp.TransactionCapn {
	dest := capnp.AutoNewTransactionCapn(seg)

	value, _ := src.Value.GobEncode()
	dest.SetNonce(src.Nonce)
	dest.SetValue(value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetSndAddr(src.SndAddr)
	dest.SetGasPrice(src.GasPrice)
	dest.SetGasLimit(src.GasLimit)
	dest.SetData(src.Data)
	dest.SetSignature(src.Signature)
	dest.SetChallenge(src.Challenge)

	return dest
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
func (tx *Transaction) GetData() string {
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
func (tx *Transaction) SetData(data string) {
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
