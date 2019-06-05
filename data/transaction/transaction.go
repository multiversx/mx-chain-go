package transaction

import (
	"io"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction/capnp"
	"github.com/glycerine/go-capnproto"
)

// Transaction holds all the data needed for a value transfer
type Transaction struct {
	Nonce     uint64   `capid:"0" json:"nonce"`
	Value     *big.Int `capid:"1" json:"value,omitempty"`
	RcvAddr   []byte   `capid:"2" json:"receiver,omitempty"`
	SndAddr   []byte   `capid:"3" json:"sender,omitempty"`
	GasPrice  uint64   `capid:"4" json:"gasPrice,omitempty"`
	GasLimit  uint64   `capid:"5" json:"gasLimit,omitempty"`
	Data      []byte   `capid:"6" json:"data,omitempty"`
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
