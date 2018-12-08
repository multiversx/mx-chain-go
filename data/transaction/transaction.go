package transaction

import (
	"io"

	"ElrondNetwork/elrond-go-sandbox/data/transaction/capnproto1"
	"github.com/glycerine/go-capnproto"
)

// Transaction holds all the data needed for a value transfer
type Transaction struct {
	Nonce     uint64 `capid:"0"`
	Value     []byte `capid:"1"`
	RcvAddr   []byte `capid:"2"`
	SndAddr   []byte `capid:"3"`
	GasPrice  uint64 `capid:"4"`
	GasLimit  uint64 `capid:"5"`
	Data      []byte `capid:"6"`
	Signature []byte `capid:"7"`
	Challenge []byte `capid:"8"`
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
	z := capnproto1.ReadRootTransactionCapn(capMsg)
	TransactionCapnToGo(z, tx)
	return nil
}

// TransactionCapnToGo is a helper function to copy fields from a TransactionCapn object to a Transaction object
func TransactionCapnToGo(src capnproto1.TransactionCapn, dest *Transaction) *Transaction {
	if dest == nil {
		dest = &Transaction{}
	}

	// Nonce
	dest.Nonce = src.Nonce()
	// Value
	dest.Value = src.Value()
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
func TransactionGoToCapn(seg *capn.Segment, src *Transaction) capnproto1.TransactionCapn {
	dest := capnproto1.AutoNewTransactionCapn(seg)

	dest.SetNonce(src.Nonce)
	dest.SetValue(src.Value)
	dest.SetRcvAddr(src.RcvAddr)
	dest.SetSndAddr(src.SndAddr)
	dest.SetGasPrice(src.GasPrice)
	dest.SetGasLimit(src.GasLimit)
	dest.SetData(src.Data)
	dest.SetSignature(src.Signature)
	dest.SetChallenge(src.Challenge)

	return dest
}
